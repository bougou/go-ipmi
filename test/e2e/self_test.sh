#!/usr/bin/env bash
# self_test.sh — go-ipmi self-test: use goipmi (client) to talk to
# goipmi-server (server), validating both ends in a single test.
#
# Exercises dual-stack serving: IPMI v2.0 (lanplus) and v1.5 (lan) on the
# same UDP port.
#
# Environment variables:
#   GOIPMI_SERVER_PORT – server listen port (default: random high port)
#
# Requires: make build  (or: make test-e2e-self)
#
# Usage:
#   ./test/e2e/self_test.sh
#   GOIPMI_SERVER_PORT=9623 ./test/e2e/self_test.sh

set -euo pipefail

# shellcheck source=test/e2e/common.sh
source "$(dirname "$0")/common.sh"
e2e_init

PORT="${GOIPMI_SERVER_PORT:-$((9623 + RANDOM % 1000))}"
USER="${GOIPMI_USER:-ADMIN}"
PASS="${GOIPMI_PASS:-ADMIN}"

# ---------------------------------------------------------------------------
# Start server
# ---------------------------------------------------------------------------
cleanup() {
	if [ -n "${SERVER_PID:-}" ] && kill -0 "${SERVER_PID}" 2>/dev/null; then
		echo "==> Stopping goipmi-server (pid ${SERVER_PID}) ..."
		kill "${SERVER_PID}" 2>/dev/null || true
		wait "${SERVER_PID}" 2>/dev/null || true
	fi
}
trap cleanup EXIT

echo "==> Starting goipmi-server on :${PORT} (dual-stack: lanplus + lan) ..."
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

# ---------------------------------------------------------------------------
# Run tests
# ---------------------------------------------------------------------------
echo ""
echo "========================================"
echo " Self E2E: goipmi → goipmi-server (:${PORT})"
echo " (lanplus + lan dual-stack)"
echo "========================================"

failures=0

e2e_run_chassis_cases_lanplus failures \
	"${GOIPMI}" -H 127.0.0.1 -p "${PORT}" -U "${USER}" -P "${PASS}" -I lanplus

e2e_run_chassis_cases_lan failures \
	"${GOIPMI}" -H 127.0.0.1 -p "${PORT}" -U "${USER}" -P "${PASS}" -I lan

e2e_report "Self E2E" "${failures}"
