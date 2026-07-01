#!/usr/bin/env bash
# chassis_codec_test.sh — E2E for the typed chassis codec + Set/Get System
# Boot Options + PowerCycle dispatch (branch feat/typed-chassis-codec).
#
# Starts goipmi-server (mock HAL) and drives it with ipmitool, so the server is
# validated by an external, spec-faithful client rather than our own client.
#
# Covers the acceptance criteria from docs/upstream-plan-2026-06.md §B.8:
#   - `chassis power cycle` no longer returns CodeParamOutOfRange
#   - Set/Get System Boot Options (Boot Flags) round-trips the typed structure
#
# Environment variables:
#   GOIPMI_SERVER_PORT – server listen port (default: random high port)
#   IPMITOOL_BIN       – path to ipmitool   (auto-detected if unset)
#
# Requires: make build  (or: make test-e2e-chassis-codec)
#
# Usage:
#   ./test/e2e/chassis_codec_test.sh
#   GOIPMI_SERVER_PORT=9623 ./test/e2e/chassis_codec_test.sh

set -euo pipefail

# shellcheck source=test/e2e/common.sh
source "$(dirname "$0")/common.sh"
e2e_init

GOIPMI_SERVER_PORT="${GOIPMI_SERVER_PORT:-$((9800 + RANDOM % 1000))}"
PORT="${GOIPMI_SERVER_PORT}"
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
run_ipmitool() {
	# shellcheck disable=SC2086
	${IPMITOOL_RUN} "${IPMI_ARGS[@]}" "$@"
}
# ---------------------------------------------------------------------------
# Run the tests
# ---------------------------------------------------------------------------
echo ""
echo "========================================"
echo " Chassis Codec E2E: ipmitool → goipmi-server (:${PORT})"
echo "========================================"

# chassis power cycle: previously fell through to CodeParamOutOfRange on the
# reference server. The mock HAL now exposes PowerCycle, so ipmitool must
# report "Cycle" and exit 0.
test_power_cycle() {
	local out
	out=$(run_ipmitool chassis power cycle 2>&1) || { echo "${out}" >&2; return 1; }
	if echo "${out}" | grep -q "Cycle"; then
		return 0
	fi
	echo "  unexpected power cycle output: ${out}" >&2
	return 1
}
# `chassis bootdev pxe` issues Set System Boot Options (param 5, Boot Flags)
# with BootDeviceSelector=ForcePXE.  Read it back with `bootparam get 5` and
# verify the typed structure round-trips to "Force PXE".
test_bootdev_pxe_round_trip() {
	local out
	out=$(run_ipmitool chassis bootdev pxe 2>&1) || { echo "${out}" >&2; return 1; }
	out=$(run_ipmitool chassis bootparam get 5 2>&1) || { echo "${out}" >&2; return 1; }
	if echo "${out}" | grep -q "Force PXE"; then
		return 0
	fi
	echo "  PXE round-trip mismatch, got: ${out}" >&2
	return 1
}
# Switch the boot device to CD/DVD and verify the selector round-trips to
# "Force Boot from CD/DVD" on read-back.
test_bootdev_cdrom_round_trip() {
	local out
	out=$(run_ipmitool chassis bootdev cdrom 2>&1) || { echo "${out}" >&2; return 1; }
	out=$(run_ipmitool chassis bootparam get 5 2>&1) || { echo "${out}" >&2; return 1; }
	if echo "${out}" | grep -q "Force Boot from CD/DVD"; then
		return 0
	fi
	echo "  CDROM round-trip mismatch, got: ${out}" >&2
	return 1
}

# Get System Boot Options for param 0 (Set In Progress) – handler does not
# implement this selector; must return completion code 80h per §28.13 Table 28-13.
test_get_unsupported_boot_param() {
	local out
	out=$(run_ipmitool raw 0x00 0x09 0x00 0x00 0x00 2>&1) && { echo "${out}" >&2; return 1; }
	if echo "${out}" | grep -q "rsp=0x80"; then
		return 0
	fi
	echo "  expected rsp=0x80, got: ${out}" >&2
	return 1
}
# Set System Boot Options for param 0 (Set In Progress) – handler does not
# implement this selector; must return completion code 80h per §28.12 Table 28-12.
test_set_unsupported_boot_param() {
	local out
	out=$(run_ipmitool raw 0x00 0x08 0x00 0x01 2>&1) && { echo "${out}" >&2; return 1; }
	if echo "${out}" | grep -q "rsp=0x80"; then
		return 0
	fi
	echo "  expected rsp=0x80, got: ${out}" >&2
	return 1
}
# Set Boot Flags (param 5) with 0 bytes of parameter data – spec §28.12 allows
# valid-bit toggling without affecting the current parameter value.
test_set_boot_flags_zero_data() {
	# First store real flags so a subsequent read-back succeeds.
	run_ipmitool chassis bootdev pxe 2>&1 || { echo "  failed to set boot flags" >&2; return 1; }
	# Now set param 5 with no data bytes.
	run_ipmitool raw 0x00 0x08 0x05 2>&1 || { echo "  raw 0x00 0x08 0x05 failed" >&2; return 1; }
	# Read-back must still show Force PXE.
	local out
	out=$(run_ipmitool chassis bootparam get 5 2>&1) || { echo "${out}" >&2; return 1; }
	if echo "${out}" | grep -q "Force PXE"; then
		return 0
	fi
	echo "  PXE flags lost after 0-byte set: ${out}" >&2
	return 1
}

# Chassis Control with a reserved action (0x0F) — must return C9h (Parameter out
# of range) per §5.2 Table 5-2.  This validates the CodeParamOutOfRange fix
# (was incorrectly 0xC9 before the spec alignment).
test_chassis_control_unknown_action() {
	local out
	out=$(run_ipmitool raw 0x00 0x02 0x0F 2>&1) && { echo "${out}" >&2; return 1; }
	if echo "${out}" | grep -q "rsp=0xc9"; then
		return 0
	fi
	echo "  expected rsp=0xc9, got: ${out}" >&2
	return 1
}

# Get Chassis Status response must always include byte 3 (front-panel button
# disables) per §28.2 Table 28-3: "Return as 00h if the panel button disable
# function is not supported."  This validates the Pack() always-4-bytes fix.
test_chassis_status_always_4bytes() {
	local out
	out=$(run_ipmitool raw 0x00 0x01 2>&1)
	# Raw response: " 00 00 40 00" (4 data bytes after completion code).
	local bytes
	bytes=$(echo "${out}" | tr ' ' '\n' | grep -cE '^[0-9a-fA-F]{2}$')
	if [ "${bytes}" -ge 4 ]; then
		return 0
	fi
	echo "  want >=4 data bytes, got ${bytes}: ${out}" >&2
	return 1
}
failures=0
e2e_run_test "chassis power cycle (PowerCycle dispatch)" test_power_cycle || ((failures++)) || true
e2e_run_test "set/get boot flags PXE round-trip" test_bootdev_pxe_round_trip || ((failures++)) || true
e2e_run_test "set/get boot flags CDROM round-trip" test_bootdev_cdrom_round_trip || ((failures++)) || true
e2e_run_test "get unsupported boot param → 80h" test_get_unsupported_boot_param || ((failures++)) || true
e2e_run_test "set unsupported boot param → 80h" test_set_unsupported_boot_param || ((failures++)) || true
e2e_run_test "set boot flags with 0 data bytes → OK" test_set_boot_flags_zero_data || ((failures++)) || true
e2e_run_test "chassis control unknown action → C9h" test_chassis_control_unknown_action || ((failures++)) || true
e2e_run_test "chassis status always 4 bytes" test_chassis_status_always_4bytes || ((failures++)) || true

e2e_report "Chassis Codec E2E" "${failures}"
