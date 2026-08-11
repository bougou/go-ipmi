#!/usr/bin/env bash
# sol_test.sh — E2E test for SOL (Serial over LAN, spec v2.0 §15) against
# goipmi-server.
#
# Starts goipmi-server with a PTY-backed console (GOIPMI_SERVER_CONSOLE=pty)
# and drives real SOL clients. The PTY slave plays the role of the system
# serial port: bytes written there are console output, bytes read there are
# remote-console keystrokes.
#
# Clients exercised (each available one gets the full interactive suite):
#   - local ipmitool           (ipmitool sol activate)
#   - Docker ipmitool          (fallback when no local binary; entrypoint of
#                               the image is ipmitool, stdin via docker -i)
#   - local goipmi binary      (goipmi sol activate — our own client, which
#                               polls with SOL packets unlike ipmitool)
#
# Requires Linux (PTY allocation). With neither local ipmitool nor Docker
# the ipmitool suites are skipped; the goipmi suite always runs.
#
# Environment variables:
#   GOIPMI_SERVER_PORT – port for the server to listen on (default: random)
#   IPMITOOL_BIN       – path to ipmitool   (auto-detected if unset)
#   IPMITOOL_IMAGE     – Docker image fallback (default: ghcr.io/halfcrazy/ipmitool:faea53b)
#
# Requires: make build  (or: make test-e2e-sol)
#
# Usage:
#   ./test/e2e/sol_test.sh

set -euo pipefail

# shellcheck source=test/e2e/common.sh
source "$(dirname "$0")/common.sh"
e2e_init

GOIPMI_SERVER_PORT="${GOIPMI_SERVER_PORT:-$((10700 + RANDOM % 1000))}"
GOIPMI_USER="${GOIPMI_USER:-ADMIN}"
GOIPMI_PASS="${GOIPMI_PASS:-ADMIN}"
IPMITOOL_IMAGE="${IPMITOOL_IMAGE:-ghcr.io/halfcrazy/ipmitool:faea53b}"

IPMITOOL_BIN="${IPMITOOL_BIN:-}"
if [ -z "${IPMITOOL_BIN}" ] && command -v ipmitool &>/dev/null; then
	IPMITOOL_BIN="ipmitool"
fi

FORCE_DOCKER=0
if [ "${IPMITOOL_BIN}" = "docker" ]; then
	# Escape hatch to exercise the container path on hosts that also have a
	# local ipmitool.
	FORCE_DOCKER=1
	IPMITOOL_BIN=""
fi

HAVE_DOCKER=0
if [ -z "${IPMITOOL_BIN}" ] && command -v docker &>/dev/null && docker info &>/dev/null; then
	HAVE_DOCKER=1
fi

if [ -n "${IPMITOOL_BIN}" ]; then
	echo "==> Using ipmitool: ${IPMITOOL_BIN}"
elif [ "${HAVE_DOCKER}" -eq 1 ]; then
	echo "==> No local ipmitool; using Docker image: ${IPMITOOL_IMAGE}"
else
	echo "==> No ipmitool (local or Docker); ipmitool suites will be skipped"
fi

# ---------------------------------------------------------------------------
# Start the server with a PTY console
# ---------------------------------------------------------------------------
WORK="$(mktemp -d)"
SERVER_LOG="${WORK}/server.log"

CLIENT_PID=""
SOL_CAT_PID=""
DOCKER_NAME="goipmi-sol-e2e-$$"

cleanup() {
	if [ -n "${CLIENT_PID:-}" ] && kill -0 "${CLIENT_PID}" 2>/dev/null; then
		kill -9 "${CLIENT_PID}" 2>/dev/null || true
	fi
	if [ -n "${SOL_CAT_PID:-}" ] && kill -0 "${SOL_CAT_PID}" 2>/dev/null; then
		kill "${SOL_CAT_PID}" 2>/dev/null || true
	fi
	if [ "${HAVE_DOCKER}" -eq 1 ] && [ -z "${IPMITOOL_BIN}" ]; then
		docker rm -f "${DOCKER_NAME}" >/dev/null 2>&1 || true
	fi
	if [ -n "${SERVER_PID:-}" ] && kill -0 "${SERVER_PID}" 2>/dev/null; then
		kill "${SERVER_PID}" 2>/dev/null || true
		wait "${SERVER_PID}" 2>/dev/null || true
	fi
	exec 9>&- 2>/dev/null || true
	# Keep the work dir (server trace, client log, slave capture) when a
	# case failed; the diagnostics dump below then stays inspectable.
	if [ -z "${KEEP_WORK:-}" ]; then
		rm -rf "${WORK}"
	fi
}
trap cleanup EXIT

echo "==> Starting goipmi-server on :${GOIPMI_SERVER_PORT} with PTY console ..."
# GOIPMI_SERVER_TRACE: per-packet SOL trace (sol< / sol> lines with
# timestamps) goes to SERVER_LOG; failures below dump it.
GOIPMI_SERVER_PORT="${GOIPMI_SERVER_PORT}" \
GOIPMI_SERVER_USER="${GOIPMI_USER}" \
GOIPMI_SERVER_PASS="${GOIPMI_PASS}" \
GOIPMI_SERVER_CONSOLE=pty \
GOIPMI_SERVER_TRACE=1 \
	"${SERVER_BIN}" > "${SERVER_LOG}" 2>&1 &
SERVER_PID=$!
sleep 2

if ! kill -0 "${SERVER_PID}" 2>/dev/null; then
	echo -e "${RED}ERROR: goipmi-server exited early${NC}" >&2
	cat "${SERVER_LOG}" >&2
	exit 1
fi

SLAVE="$(grep "console pty slave" "${SERVER_LOG}" | awk '{print $NF}')"
if [ -z "${SLAVE}" ]; then
	echo -e "${RED}ERROR: server did not report a PTY slave path${NC}" >&2
	cat "${SERVER_LOG}" >&2
	exit 1
fi
echo "==> Console PTY slave: ${SLAVE}"
# Raw bytes both ways: no echo, no CR/LF translation on the system side.
stty -F "${SLAVE}" raw -echo

wait_for() { # wait_for <file> <pattern> <tenths-of-second>
	local file="$1" pattern="$2" ticks="${3:-50}" i
	for ((i = 0; i < ticks; i++)); do
		grep -qa "${pattern}" "${file}" 2>/dev/null && return 0
		sleep 0.1
	done
	return 1
}

failures=0

# ---------------------------------------------------------------------------
# Client abstractions
# ---------------------------------------------------------------------------
# MODE is one of: ipmitool | ipmitool-docker | goipmi
# SOL_OUT/SOL_FIFO/SOL_IN_GOT are per-suite files under WORK.

ipmi_oneshot() { # one-shot ipmitool command in the current mode
	case "${MODE}" in
	ipmitool-docker)
		docker run --rm --network host "${IPMITOOL_IMAGE}" \
			-I lanplus -H 127.0.0.1 -p "${GOIPMI_SERVER_PORT}" \
			-U "${GOIPMI_USER}" -P "${GOIPMI_PASS}" "$@"
		;;
	*)
		"${IPMITOOL_BIN}" -I lanplus -H 127.0.0.1 -p "${GOIPMI_SERVER_PORT}" \
			-U "${GOIPMI_USER}" -P "${GOIPMI_PASS}" "$@"
		;;
	esac
}

sol_activate_bg() { # start the interactive client of the current mode
	case "${MODE}" in
	ipmitool)
		# stdbuf -o0: ipmitool block-buffers stdout when redirected to a
		# file, which would hide the activation banner.
		stdbuf -o0 "${IPMITOOL_BIN}" -I lanplus -H 127.0.0.1 -p "${GOIPMI_SERVER_PORT}" \
			-U "${GOIPMI_USER}" -P "${GOIPMI_PASS}" sol activate \
			< "${SOL_FIFO}" > "${SOL_OUT}" 2>&1 &
		;;
	ipmitool-docker)
		# No stdbuf inside the image: activation is detected by polling
		# Get Payload Activation Status instead of the banner (see
		# wait_activated). Console data itself is fflush'd per packet by
		# ipmitool, so it needs no unbuffering help.
		docker run --rm -i --network host --name "${DOCKER_NAME}" "${IPMITOOL_IMAGE}" \
			-I lanplus -H 127.0.0.1 -p "${GOIPMI_SERVER_PORT}" \
			-U "${GOIPMI_USER}" -P "${GOIPMI_PASS}" sol activate \
			< "${SOL_FIFO}" > "${SOL_OUT}" 2>&1 &
		;;
	goipmi)
		# Go writes are unbuffered; the banner is reliable as-is. --debug
		# logs every client packet exchange to SOL_OUT (dumped on failure).
		"${GOIPMI}" --debug -I lanplus -H 127.0.0.1 -p "${GOIPMI_SERVER_PORT}" \
			-U "${GOIPMI_USER}" -P "${GOIPMI_PASS}" sol activate \
			< "${SOL_FIFO}" > "${SOL_OUT}" 2>&1 &
		;;
	esac
	CLIENT_PID=$!
}

wait_activated() { # wait until the payload is active (≈5s)
	case "${MODE}" in
	ipmitool-docker)
		local i status
		# Poll Get Payload Activation Status (§24.4, App 4Ah): byte 2 of the
		# response is the instance 1-8 bitmask; bit0 = instance 1 active.
		for ((i = 0; i < 20; i++)); do
			status="$(ipmi_oneshot raw 0x06 0x4a 0x01 2>/dev/null | tr -d ' \n')"
			[ "${status:2:2}" = "01" ] && return 0
			sleep 0.25
		done
		return 1
		;;
	ipmitool)
		wait_for "${SOL_OUT}" "SOL Session operational" 50
		;;
	goipmi)
		wait_for "${SOL_OUT}" "SOL payload activated" 50
		;;
	esac
}

stop_client() { # ask the client to exit via the ~. escape, then reap
	printf '\r~.' >&9 2>/dev/null || true
	local i
	for ((i = 0; i < 25; i++)); do
		[ -n "${CLIENT_PID}" ] && kill -0 "${CLIENT_PID}" 2>/dev/null || break
		sleep 0.2
	done
	if [ -n "${CLIENT_PID}" ] && kill -0 "${CLIENT_PID}" 2>/dev/null; then
		kill -9 "${CLIENT_PID}" 2>/dev/null || true
		[ "${MODE}" = "ipmitool-docker" ] && docker rm -f "${DOCKER_NAME}" >/dev/null 2>&1 || true
	fi
	wait "${CLIENT_PID}" 2>/dev/null || true
	CLIENT_PID=""
}

# ---------------------------------------------------------------------------
# The interactive suite, run once per available client
# ---------------------------------------------------------------------------
case_outbound() { # console output reaches the remote client
	printf 'E2E-OUT-%s\r\n' "${MODE}" > "${SLAVE}"
	wait_for "${SOL_OUT}" "E2E-OUT-${MODE}" 50
}

case_streaming() { # multi-packet output arrives complete and in order
	local i
	for ((i = 1; i <= 40; i++)); do
		printf 'stream-line-%02d\r\n' "${i}" > "${SLAVE}"
	done
	printf 'E2E-STREAM-END-%s\r\n' "${MODE}" > "${SLAVE}"
	wait_for "${SOL_OUT}" "E2E-STREAM-END-${MODE}" 100 &&
		grep -qa "stream-line-01" "${SOL_OUT}" && grep -qa "stream-line-40" "${SOL_OUT}"
}

case_inbound() { # remote keystrokes land on the system serial port
	# Reap the previous inbound case's reader: its 5 s window outlives the
	# case, and a lingering reader shares the slave with this case's cat,
	# splitting — or, having blocked first, entirely consuming — the
	# keystrokes. pkill -P catches the cat itself in case timeout does not
	# forward the signal.
	if [ -n "${SOL_CAT_PID:-}" ] && kill -0 "${SOL_CAT_PID}" 2>/dev/null; then
		kill "${SOL_CAT_PID}" 2>/dev/null || true
		wait "${SOL_CAT_PID}" 2>/dev/null || true
		pkill -TERM -P "${SOL_CAT_PID}" 2>/dev/null || true
	fi
	# EIO on stderr is expected: the SOL session ends (PTY master closes)
	# before the 5 s timeout, and reads on an orphaned slave return EIO.
	timeout 5 cat "${SLAVE}" 2>/dev/null > "${SOL_IN_GOT}" &
	SOL_CAT_PID=$!
	printf 'E2E-IN-%s' "${MODE}" >&9
	wait_for "${SOL_IN_GOT}" "E2E-IN-${MODE}" 50
}

reconnect_case() { # console fault (SIGUSR1) → reconnect → recovery
	# The SOL session from the interactive suite is still active. Break the
	# console link: console reads fail, the instance flips broken, and the
	# first reconnect attempt (1s later, default policy) fails too.
	kill -USR1 "${SERVER_PID}"
	sleep 1.5

	# The remote console must be unaffected by the outage: process alive and
	# quiet. (ipmitool keeps the session alive with Get Device ID keepalives;
	# it only exits after 6 missed keepalives, and the RMCP+ path is fine.)
	if ! kill -0 "${CLIENT_PID}" 2>/dev/null; then
		echo "client exited during console outage" >&2
		return 1
	fi
	if grep -qaE "Error|FAIL|connection" "${SOL_OUT}"; then
		echo "client reported an error during console outage" >&2
		cat "${SOL_OUT}" >&2
		return 1
	fi

	# Restore the console link; the next reconnect attempt succeeds and the
	# fresh console's output must reach the still-open session.
	kill -USR2 "${SERVER_PID}"
	local i
	for ((i = 0; i < 40; i++)); do
		printf 'E2E-RECONNECT-%s
' "${MODE}" > "${SLAVE}"
		if grep -qa "E2E-RECONNECT-${MODE}" "${SOL_OUT}"; then
			return 0
		fi
		sleep 0.25
	done
	echo "console output did not resume after reconnect" >&2
	return 1
}

run_interactive_suite() {
	local label="$1"
	SOL_FIFO="${WORK}/${MODE}_in"
	SOL_OUT="${WORK}/${MODE}_out.txt"
	SOL_IN_GOT="${WORK}/${MODE}_slave.txt"
	rm -f "${SOL_FIFO}"
	mkfifo "${SOL_FIFO}"
	: > "${SOL_OUT}"

	# O_RDWR avoids the reader/writer open deadlock on the fifo.
	exec 9<> "${SOL_FIFO}"
	sol_activate_bg

	e2e_run_test "${label}: sol activate" wait_activated || ((FAILURES++)) || true
	e2e_run_test "${label}: outbound (BMC → console)" case_outbound || ((FAILURES++)) || true
	e2e_run_test "${label}: outbound streaming (40 lines)" case_streaming || ((FAILURES++)) || true
	e2e_run_test "${label}: inbound (console → BMC)" case_inbound || ((FAILURES++)) || true
	e2e_run_test "${label}: deactivate via ~." stop_client || ((FAILURES++)) || true
	exec 9>&-
}

# FAILURES is global; run_interactive_suite increments it (uppercase to tell
# it apart from the e2e_run_test callers in other suites).
FAILURES=0

if [ -n "${IPMITOOL_BIN}" ]; then
	MODE="ipmitool"
	e2e_run_test "ipmitool: sol info" ipmi_oneshot sol info || ((FAILURES++)) || true
	run_interactive_suite "ipmitool"

	# Recovery: a fresh `sol deactivate` (different session, admin privilege)
	# force-frees a stale payload, after which activation works again (§24.2).
	force_deactivate_case() {
		exec 9<> "${SOL_FIFO}"
		sol_activate_bg
		if ! wait_activated; then
			echo "re-activate failed" >&2
			return 1
		fi
		if ! ipmi_oneshot sol deactivate >/dev/null 2>&1; then
			echo "sol deactivate failed" >&2
			return 1
		fi
		stop_client
		exec 9>&-
		: > "${SOL_OUT}"
		exec 9<> "${SOL_FIFO}"
		sol_activate_bg
		local rc=0
		wait_activated || rc=1
		stop_client
		exec 9>&-
		return "${rc}"
	}
	e2e_run_test "ipmitool: force-deactivate frees payload" force_deactivate_case || ((FAILURES++)) || true

elif [ "${HAVE_DOCKER}" -eq 1 ]; then
	MODE="ipmitool-docker"
	e2e_run_test "ipmitool(docker): sol info" ipmi_oneshot sol info || ((FAILURES++)) || true
	run_interactive_suite "ipmitool(docker)"
fi

MODE="goipmi"
run_interactive_suite "goipmi"

if [ "${FAILURES}" -gt 0 ]; then
	KEEP_WORK=1
	echo "==> goipmi-server SOL trace (full):" >&2
	grep -E "sol[<>!]" "${SERVER_LOG}" >&2
	echo "==> goipmi client log (full):" >&2
	cat "${WORK}/goipmi_out.txt" 2>/dev/null >&2
	echo "==> slave capture:" >&2
	cat "${WORK}/goipmi_slave.txt" 2>/dev/null >&2
	echo "==> work dir kept at: ${WORK}" >&2
fi

e2e_report "SOL E2E" "${FAILURES}"
