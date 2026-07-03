#!/usr/bin/env bash
# server_test.sh — E2E test for go-ipmi as an IPMI BMC server.
#
# Starts goipmi-server, then connects with ipmitool (local or Docker) over
# both IPMI v2.0 (lanplus) and v1.5 (lan -A MD5) to verify dual-stack serving.
#
# Environment variables:
#   GOIPMI_SERVER_PORT – port for the server to listen on (default: 9623)
#                         Use a port >1024 to avoid sudo when testing locally.
#   IPMITOOL_BIN       – path to ipmitool   (auto-detected if unset)
#   IPMITOOL_IMAGE     – Docker image to use when ipmitool is not found
#                         (default: ghcr.io/halfcrazy/ipmitool:eecd64f)
#
# Requires: make build  (or: make test-e2e-server)
#
# Usage:
#   ./test/e2e/server_test.sh                          # port 623   (needs root)
#   GOIPMI_SERVER_PORT=9623 ./test/e2e/server_test.sh  # port 9623 (no root needed)

set -euo pipefail

# shellcheck source=test/e2e/common.sh
source "$(dirname "$0")/common.sh"
e2e_init

GOIPMI_SERVER_PORT="${GOIPMI_SERVER_PORT:-9623}"
GOIPMI_USER="${GOIPMI_USER:-ADMIN}"
GOIPMI_PASS="${GOIPMI_PASS:-ADMIN}"
IPMITOOL_IMAGE="${IPMITOOL_IMAGE:-ghcr.io/halfcrazy/ipmitool:eecd64f}"

# ---------------------------------------------------------------------------
# Find or choose an ipmitool
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

USE_SUDO=""
if [ "${GOIPMI_SERVER_PORT}" -lt 1024 ]; then
	USE_SUDO="sudo"
fi

echo "==> Starting goipmi-server on :${GOIPMI_SERVER_PORT} (dual-stack: lanplus + lan) ..."
${USE_SUDO} env \
	GOIPMI_SERVER_PORT="${GOIPMI_SERVER_PORT}" \
	GOIPMI_SERVER_USER="${GOIPMI_USER}" \
	GOIPMI_SERVER_PASS="${GOIPMI_PASS}" \
	"${SERVER_BIN}" &
SERVER_PID=$!
sleep 2

if ! ss -uln | grep -q ":${GOIPMI_SERVER_PORT} "; then
	echo "ERROR: server failed to bind port ${GOIPMI_SERVER_PORT}" >&2
	exit 1
fi
echo "==> Server is listening on :${GOIPMI_SERVER_PORT}"

# ---------------------------------------------------------------------------
# Run the tests
# ---------------------------------------------------------------------------
echo ""
echo "========================================"
echo " Server E2E: ipmitool → goipmi-server"
echo " (lanplus + lan dual-stack)"
echo "========================================"

run_ipmitool() {
	# shellcheck disable=SC2086
	${IPMITOOL_RUN} "$@"
}

failures=0

e2e_run_chassis_cases_lanplus failures run_ipmitool \
	-H 127.0.0.1 -p "${GOIPMI_SERVER_PORT}" -U "${GOIPMI_USER}" -P "${GOIPMI_PASS}" -I lanplus

e2e_run_chassis_cases_lan failures run_ipmitool \
	-H 127.0.0.1 -p "${GOIPMI_SERVER_PORT}" -U "${GOIPMI_USER}" -P "${GOIPMI_PASS}" -I lan -A MD5

e2e_report "Server E2E" "${failures}"
