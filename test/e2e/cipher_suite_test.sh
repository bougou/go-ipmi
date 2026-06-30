#!/usr/bin/env bash
# cipher_suite_test.sh — E2E for RMCP+ cipher suite 17 (SHA256) support
# (branch feat/server-cipher-suite-17).
#
# Starts goipmi-server (which advertises {3, 17} by default) and connects with
# ipmitool using explicit cipher-suite selectors, so the server's RAKP/HMAC and
# integrity handling for SHA256 is validated by an external, spec-faithful
# client rather than our own client.
#
# Covers the acceptance criteria from docs/upstream-plan-2026-06.md §A.8:
#   - ipmitool -C 17 (RAKP-HMAC-SHA256, HMAC-SHA256-128, AES-CBC-128) succeeds
#   - ipmitool -C 3  (RAKP-HMAC-SHA1,  HMAC-SHA1-96,   AES-CBC-128) still works
#
# Environment variables:
#   GOIPMI_SERVER_PORT – server listen port (default: random high port)
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
}
trap cleanup EXIT

echo "==> Starting goipmi-server on :${PORT} ..."
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

IPMI_ARGS=(-H 127.0.0.1 -p "${PORT}" -U "${USER}" -P "${PASS}" -I lanplus)

# run_cipher <suite> <ipmitool-args...>
run_cipher() {
	local suite="$1"
	shift
	# shellcheck disable=SC2086
	${IPMITOOL_RUN} "${IPMI_ARGS[@]}" -C "${suite}" "$@"
}

# ---------------------------------------------------------------------------
# Run the tests
# ---------------------------------------------------------------------------
echo ""
echo "========================================"
echo " Cipher Suite E2E: ipmitool → goipmi-server (:${PORT})"
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

failures=0
e2e_run_test "suite 17 (SHA256) chassis status" test_suite17_chassis_status || ((failures++)) || true
e2e_run_test "suite 17 (SHA256) mc info"        test_suite17_mc_info        || ((failures++)) || true
e2e_run_test "suite 3 (SHA1) chassis status"    test_suite3_chassis_status  || ((failures++)) || true

e2e_report "Cipher Suite E2E" "${failures}"
