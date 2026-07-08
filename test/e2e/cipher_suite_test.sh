#!/usr/bin/env bash
# cipher_suite_test.sh — E2E for RMCP+ cipher suite coverage.
#
# Starts three goipmi-server instances:
#   1) default config ({3, 17})        — verifies suites 3/17 succeed and that
#      suite 0 (AuthAlgNone) is rejected, guarding against the auth bypass.
#   2) extended config ({1,2,15,16,3,17}) — verifies the remaining spec-defined
#      suites (1, 2, 15, 16) succeed end-to-end with an external, spec-faithful
#      ipmitool client. This exercises RAKP4 ICV derivation by the *auth*
#      algorithm (spec §13.28.1/§13.28.1b) for suites that pair a non-None auth
#      algorithm with Integrity=None (suites 1 and 15).
#   3) cross-suite config ({2, 17})    — verifies that suite 3 (SHA1+SHA1-96+AES)
#      is rejected even though each algorithm appears individually (SHA1/SHA1-96
#      from suite 2, AES from suite 17). This guards against cross-suite
#      algorithm recombination in Open Session negotiation.
#
# Covers the acceptance criteria from docs/upstream-plan-2026-06.md §A.8:
#   - ipmitool -C 17 (RAKP-HMAC-SHA256, HMAC-SHA256-128, AES-CBC-128) succeeds
#   - ipmitool -C 3  (RAKP-HMAC-SHA1,  HMAC-SHA1-96,   AES-CBC-128) still works
#
# Environment variables:
#   GOIPMI_SERVER_PORT – default server listen port (default: random high port)
#   GOIPMI_SERVER_PORT_EXTENDED – extended-config server port (default: random)
#   GOIPMI_SERVER_PORT_CROSS – cross-suite-test server port (default: random)
#   IPMITOOL_BIN       – path to ipmitool   (auto-detected if unset)
#
# Requires: make build  (or: make test-e2e-cipher)
#
# Usage:
#   ./test/e2e/cipher_suite_test.sh
#   GOIPMI_SERVER_PORT=9623 ./test/e2e/cipher_suite_test.sh

set -euo pipefail

# shellcheck source=test/e2e/common.sh
source "$(dirname "$0")/common.sh"
e2e_init

PORT="${GOIPMI_SERVER_PORT:-$((9900 + RANDOM % 1000))}"
PORT_EXTENDED="${GOIPMI_SERVER_PORT_EXTENDED:-$((11100 + RANDOM % 1000))}"
PORT_CROSS="${GOIPMI_SERVER_PORT_CROSS:-$((11300 + RANDOM % 1000))}"
USER="${GOIPMI_USER:-ADMIN}"
PASS="${GOIPMI_PASS:-ADMIN}"
IPMITOOL_IMAGE="${IPMITOOL_IMAGE:-ghcr.io/halfcrazy/ipmitool:eecd64f}"

# ---------------------------------------------------------------------------
# Find or choose an ipmitool (local install, else Docker image).
# ---------------------------------------------------------------------------
IPMITOOL_BIN="${IPMITOOL_BIN:-}"
if [ -z "${IPMITOOL_BIN}" ]; then
	if command -v ipmitool &>/dev/null; then
		IPMITOOL_BIN="ipmitool"
	fi
fi

if [ -n "${IPMITOOL_BIN}" ]; then
	echo "==> Using ipmitool: ${IPMITOOL_BIN}"
	IPMITOOL_RUN="${IPMITOOL_BIN}"
else
	echo "==> ipmitool not found locally, will use Docker: ${IPMITOOL_IMAGE}"
	IPMITOOL_RUN="docker run --rm --network host ${IPMITOOL_IMAGE}"
fi

# ---------------------------------------------------------------------------
# Start the server
# ---------------------------------------------------------------------------
cleanup() {
	if [ -n "${SERVER_PID:-}" ] && kill -0 "${SERVER_PID}" 2>/dev/null; then
		echo "==> Stopping goipmi-server (pid ${SERVER_PID}) ..."
		kill "${SERVER_PID}" 2>/dev/null || true
		wait "${SERVER_PID}" 2>/dev/null || true
	fi
	if [ -n "${EXTENDED_PID:-}" ] && kill -0 "${EXTENDED_PID}" 2>/dev/null; then
		echo "==> Stopping extended-config goipmi-server (pid ${EXTENDED_PID}) ..."
		kill "${EXTENDED_PID}" 2>/dev/null || true
		wait "${EXTENDED_PID}" 2>/dev/null || true
	fi
	if [ -n "${CROSS_PID:-}" ] && kill -0 "${CROSS_PID}" 2>/dev/null; then
		echo "==> Stopping cross-suite-test goipmi-server (pid ${CROSS_PID}) ..."
		kill "${CROSS_PID}" 2>/dev/null || true
		wait "${CROSS_PID}" 2>/dev/null || true
	fi
}
trap cleanup EXIT

echo "==> Starting goipmi-server on :${PORT} (default cipher suites) ..."
env \
	GOIPMI_SERVER_PORT="${PORT}" \
	GOIPMI_SERVER_USER="${USER}" \
	GOIPMI_SERVER_PASS="${PASS}" \
	"${SERVER_BIN}" &
SERVER_PID=$!
sleep 1

if ! ss -uln | grep -q ":${PORT} "; then
	echo -e "${RED}ERROR: server failed to start${NC}" >&2
	exit 1
fi

echo "==> Starting goipmi-server on :${PORT_EXTENDED} (extended cipher suites 1,2,15,16,3,17) ..."
env \
	GOIPMI_SERVER_PORT="${PORT_EXTENDED}" \
	GOIPMI_SERVER_USER="${USER}" \
	GOIPMI_SERVER_PASS="${PASS}" \
	GOIPMI_SERVER_CIPHER_SUITES="1,2,15,16,3,17" \
	"${SERVER_BIN}" &
EXTENDED_PID=$!
sleep 1

if ! ss -uln | grep -q ":${PORT_EXTENDED} "; then
	echo -e "${RED}ERROR: extended-config server failed to start${NC}" >&2
	exit 1
fi

echo "==> Starting goipmi-server on :${PORT_CROSS} (cross-suite-test cipher suites 2,17) ..."
env \
	GOIPMI_SERVER_PORT="${PORT_CROSS}" \
	GOIPMI_SERVER_USER="${USER}" \
	GOIPMI_SERVER_PASS="${PASS}" \
	GOIPMI_SERVER_CIPHER_SUITES="2,17" \
	"${SERVER_BIN}" &
CROSS_PID=$!
sleep 1

if ! ss -uln | grep -q ":${PORT_CROSS} "; then
	echo -e "${RED}ERROR: cross-suite-test server failed to start${NC}" >&2
	exit 1
fi

IPMI_ARGS=(-H 127.0.0.1 -p "${PORT}" -U "${USER}" -P "${PASS}" -I lanplus)
IPMI_ARGS_EXTENDED=(-H 127.0.0.1 -p "${PORT_EXTENDED}" -U "${USER}" -P "${PASS}" -I lanplus)
IPMI_ARGS_CROSS=(-H 127.0.0.1 -p "${PORT_CROSS}" -U "${USER}" -P "${PASS}" -I lanplus)

# run_cipher <suite> <ipmitool-args...>          — against the default server.
# run_cipher_extended <suite> <ipmitool-args...> — against the extended server.
# run_cipher_cross <suite> <ipmitool-args...>    — against the cross-suite-test server.
run_cipher() {
	local suite="$1"
	shift
	# shellcheck disable=SC2086
	${IPMITOOL_RUN} "${IPMI_ARGS[@]}" -C "${suite}" "$@"
}
run_cipher_extended() {
	local suite="$1"
	shift
	# shellcheck disable=SC2086
	${IPMITOOL_RUN} "${IPMI_ARGS_EXTENDED[@]}" -C "${suite}" "$@"
}
run_cipher_cross() {
	local suite="$1"
	shift
	# shellcheck disable=SC2086
	${IPMITOOL_RUN} "${IPMI_ARGS_CROSS[@]}" -C "${suite}" "$@"
}

# ---------------------------------------------------------------------------
# Run the tests
# ---------------------------------------------------------------------------
echo ""
echo "========================================"
echo " Cipher Suite E2E: ipmitool → goipmi-server"
echo "   default   :${PORT}  (suites 3, 17)"
echo "   extended  :${PORT_EXTENDED}  (suites 1, 2, 15, 16, 3, 17)"
echo "   cross     :${PORT_CROSS}  (suites 2, 17)"
echo "========================================"

# Suite 17 = RAKP-HMAC-SHA256 / HMAC-SHA256-128 / AES-CBC-128.  The Get Channel
# Cipher Suites / Open Session / RAKP handshake and the first authenticated,
# encrypted command must all succeed.
test_suite17_chassis_status() {
	local out
	out=$(run_cipher 17 chassis status 2>&1) || { echo "${out}" >&2; return 1; }
	echo "${out}" | grep -q "System Power" && return 0
	echo "  suite 17 chassis status mismatch: ${out}" >&2
	return 1
}

# Suite 17 over a different NetFn (App Get Device ID) to exercise SHA256-128
# integrity on more than one payload type.
test_suite17_mc_info() {
	local out
	out=$(run_cipher 17 mc info 2>&1) || { echo "${out}" >&2; return 1; }
	echo "${out}" | grep -q "Device ID" && return 0
	echo "  suite 17 mc info mismatch: ${out}" >&2
	return 1
}

# Suite 3 = RAKP-HMAC-SHA1 / HMAC-SHA1-96 / AES-CBC-128.  Regression guard: the
# SHA256 additions must not break the existing SHA1 path.
test_suite3_chassis_status() {
	local out
	out=$(run_cipher 3 chassis status 2>&1) || { echo "${out}" >&2; return 1; }
	echo "${out}" | grep -q "System Power" && return 0
	echo "  suite 3 chassis status mismatch: ${out}" >&2
	return 1
}

# Suite 0 = RAKP-None / Integrity-None / Confidentiality-None.  The default
# cipher suite set ({3, 17}) does not advertise suite 0, so the server must
# reject the Open Session Request (error 0x04 invalid auth alg) and no
# authenticated session may be established.  This guards against the
# authentication bypass where AuthAlgNone skips RAKP password verification.
test_suite0_rejected_by_default() {
	local out rc
	out=$(run_cipher 0 chassis status 2>&1) && rc=0 || rc=$?
	if [ "${rc}" -eq 0 ]; then
		echo "  suite 0 (AuthAlgNone) unexpectedly established a session: ${out}" >&2
		return 1
	fi
	if echo "${out}" | grep -q "System Power"; then
		echo "  suite 0 (AuthAlgNone) executed a command despite default config: ${out}" >&2
		return 1
	fi
	return 0
}

# Cross-suite recombination guard: server configured with {2, 17} must reject
# suite 3 (SHA1+SHA1-96+AES).  Each algorithm appears in some configured suite
# (SHA1/SHA1-96 from suite 2, AES from suite 17), but the triple as a unit was
# never advertised.  Accepting it would violate the cipher suite model and,
# with certain configs (e.g. {3, 15}), could silently enable suite 1
# (authenticated but no integrity check).
test_suite3_rejected_by_cross_suite_server() {
	local out rc
	out=$(run_cipher_cross 3 chassis status 2>&1) && rc=0 || rc=$?
	if [ "${rc}" -eq 0 ]; then
		echo "  suite 3 (cross-suite recombination) unexpectedly established a session: ${out}" >&2
		return 1
	fi
	if echo "${out}" | grep -q "System Power"; then
		echo "  suite 3 (cross-suite recombination) executed a command despite not being configured: ${out}" >&2
		return 1
	fi
	return 0
}

# Suite 2 against the cross-suite server ({2, 17}) must succeed — it is a
# legitimately configured suite.
test_suite2_cross_suite_server() {
	local out
	out=$(run_cipher_cross 2 chassis status 2>&1) || { echo "${out}" >&2; return 1; }
	echo "${out}" | grep -q "System Power" && return 0
	echo "  suite 2 (cross-suite server) chassis status mismatch: ${out}" >&2
	return 1
}

# Suite 17 against the cross-suite server ({2, 17}) must succeed — it is a
# legitimately configured suite.
test_suite17_cross_suite_server() {
	local out
	out=$(run_cipher_cross 17 chassis status 2>&1) || { echo "${out}" >&2; return 1; }
	echo "${out}" | grep -q "System Power" && return 0
	echo "  suite 17 (cross-suite server) chassis status mismatch: ${out}" >&2
	return 1
}

# The following suites are negotiated against the extended-config server
# (GOIPMI_SERVER_CIPHER_SUITES="1,2,15,16,3,17"). Each must complete the full
# RAKP handshake and an authenticated (optionally encrypted) command.

# Suite 1 = RAKP-HMAC-SHA1 / Integrity-None / Confidentiality-None.
# Exercises the spec-correct RAKP4 ICV: selected by the *auth* algorithm
# (HMAC-SHA1-96, 12 bytes) even though Integrity=None (spec §13.28.1, §13.31).
test_suite1_chassis_status() {
	local out
	out=$(run_cipher_extended 1 chassis status 2>&1) || { echo "${out}" >&2; return 1; }
	echo "${out}" | grep -q "System Power" && return 0
	echo "  suite 1 chassis status mismatch: ${out}" >&2
	return 1
}

# Suite 2 = RAKP-HMAC-SHA1 / HMAC-SHA1-96 / Confidentiality-None.
# Authenticated + integrity-protected, unencrypted.
test_suite2_chassis_status() {
	local out
	out=$(run_cipher_extended 2 chassis status 2>&1) || { echo "${out}" >&2; return 1; }
	echo "${out}" | grep -q "System Power" && return 0
	echo "  suite 2 chassis status mismatch: ${out}" >&2
	return 1
}

# Suite 15 = RAKP-HMAC-SHA256 / Integrity-None / Confidentiality-None.
# Like suite 1 but with SHA256 auth: RAKP4 ICV = HMAC-SHA256-128 (16 bytes)
# even though Integrity=None (spec §13.28.1b, §13.31).
test_suite15_chassis_status() {
	local out
	out=$(run_cipher_extended 15 chassis status 2>&1) || { echo "${out}" >&2; return 1; }
	echo "${out}" | grep -q "System Power" && return 0
	echo "  suite 15 chassis status mismatch: ${out}" >&2
	return 1
}

# Suite 16 = RAKP-HMAC-SHA256 / HMAC-SHA256-128 / Confidentiality-None.
# Authenticated + integrity-protected, unencrypted.
test_suite16_chassis_status() {
	local out
	out=$(run_cipher_extended 16 chassis status 2>&1) || { echo "${out}" >&2; return 1; }
	echo "${out}" | grep -q "System Power" && return 0
	echo "  suite 16 chassis status mismatch: ${out}" >&2
	return 1
}

failures=0
e2e_run_test "suite 17 (SHA256) chassis status" test_suite17_chassis_status || ((failures++)) || true
e2e_run_test "suite 17 (SHA256) mc info"        test_suite17_mc_info        || ((failures++)) || true
e2e_run_test "suite 3 (SHA1) chassis status"    test_suite3_chassis_status  || ((failures++)) || true
e2e_run_test "suite 0 (AuthAlgNone) rejected by default" test_suite0_rejected_by_default || ((failures++)) || true
e2e_run_test "suite 3 rejected by {2,17} cross-suite server" test_suite3_rejected_by_cross_suite_server || ((failures++)) || true
e2e_run_test "suite 2 (SHA1, no crypt) on cross-suite server" test_suite2_cross_suite_server || ((failures++)) || true
e2e_run_test "suite 17 (SHA256) on cross-suite server" test_suite17_cross_suite_server || ((failures++)) || true
e2e_run_test "suite 1 (SHA1, no integ/crypt)"   test_suite1_chassis_status  || ((failures++)) || true
e2e_run_test "suite 2 (SHA1, no crypt)"         test_suite2_chassis_status  || ((failures++)) || true
e2e_run_test "suite 15 (SHA256, no integ/crypt)" test_suite15_chassis_status || ((failures++)) || true
e2e_run_test "suite 16 (SHA256, no crypt)"       test_suite16_chassis_status || ((failures++)) || true

e2e_report "Cipher Suite E2E" "${failures}"
