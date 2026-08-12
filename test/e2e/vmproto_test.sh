#!/usr/bin/env bash
# vmproto_test.sh — E2E test for go-ipmi's OpenIPMI VM protocol server.
#
# Starts goipmi-server with both its network port and its VM protocol mode (on a
# unix socket), sharing one BMC, then proves the two frontends share state: a
# chassis power action taken in band (goipmi-vmprobe over the VM socket, the
# stand-in for QEMU's ipmi-bmc-extern / the in-guest OpenIPMI driver) is observed
# out of band over LAN (goipmi chassis power status). This is the VM-protocol
# counterpart of self_test.sh.
#
# Requires: make build  (or: make test-e2e-vmproto)
#
# Usage:
#   ./test/e2e/vmproto_test.sh

set -euo pipefail

# shellcheck source=test/e2e/common.sh
source "$(dirname "$0")/common.sh"
e2e_init

# The other e2e suites' random port ranges span roughly 9700-12299 (and the CI
# ipmi-simulator uses 9623); pick a base above all of them so a future parallel
# `make -j test-e2e` cannot collide with a sibling suite. CI runs the suites
# sequentially, so this is belt-and-suspenders.
PORT="${GOIPMI_SERVER_PORT:-$((12300 + RANDOM % 1000))}"
USER="${GOIPMI_USER:-ADMIN}"
PASS="${GOIPMI_PASS:-ADMIN}"
# Short, unique socket path: unix socket paths are length-limited (~104 bytes),
# and this avoids the GNU/BSD differences in `mktemp -t`.
SOCK="${TMPDIR:-/tmp}/goipmi-vm-$$.sock"
SOCK="${SOCK//\/\//\/}"

cleanup() {
	if [ -n "${SERVER_PID:-}" ] && kill -0 "${SERVER_PID}" 2>/dev/null; then
		echo "==> Stopping goipmi-server (pid ${SERVER_PID}) ..."
		kill "${SERVER_PID}" 2>/dev/null || true
		wait "${SERVER_PID}" 2>/dev/null || true
	fi
	rm -f "${SOCK}"
}
trap cleanup EXIT

echo "==> Starting goipmi-server on :${PORT} with VM protocol on ${SOCK} ..."
env \
	GOIPMI_SERVER_PORT="${PORT}" \
	GOIPMI_SERVER_USER="${USER}" \
	GOIPMI_SERVER_PASS="${PASS}" \
	GOIPMI_SERVER_VM_SOCKET="${SOCK}" \
	"${SERVER_BIN}" &
SERVER_PID=$!

# The VM socket is created after the UDP bind, so its appearance means both
# frontends are ready.
for _ in $(seq 1 50); do
	[ -S "${SOCK}" ] && break
	if ! kill -0 "${SERVER_PID}" 2>/dev/null; then
		echo -e "${RED}ERROR: goipmi-server exited before creating the VM socket${NC}" >&2
		exit 1
	fi
	sleep 0.1
done
if [ ! -S "${SOCK}" ]; then
	echo -e "${RED}ERROR: VM socket ${SOCK} was not created${NC}" >&2
	exit 1
fi

# In-band Chassis Control (NetFn 0x00, cmd 0x02) over the system interface;
# no session, no credentials, by design. data 1 = power up, 0 = power down.
vm_chassis_power() {
	"${VMPROBE_BIN}" -socket "${SOCK}" -netfn 0 -cmd 2 -data "$1" >/dev/null
}

# Out-of-band power state over LAN. `chassis power status` prints exactly one
# line ("Chassis Power is on"/"...off"), so an exact-string check is stable.
lan_power_is() {
	local out
	out="$("${GOIPMI}" -H 127.0.0.1 -p "${PORT}" -U "${USER}" -P "${PASS}" -I lanplus chassis power status 2>/dev/null)"
	[ "${out}" = "Chassis Power is $1" ]
}

echo ""
echo "========================================"
echo " VM protocol E2E: goipmi-vmprobe → goipmi-server (${SOCK}), shared with LAN (:${PORT})"
echo "========================================"

failures=0

# Basic wiring: an in-band Get Device ID round-trips over the VM socket.
e2e_run_test "vm get device id" "${VMPROBE_BIN}" -socket "${SOCK}" -netfn 6 -cmd 1 \
	|| ((failures++)) || true

# Shared state: power on in band, then observe it over LAN, then off, then LAN.
# Power on first — the mock BMC boots powered off, so on-then-verify proves a
# real transition rather than the initial state.
e2e_run_test "vm power on (in band)" vm_chassis_power 1 || ((failures++)) || true
e2e_run_test "lan sees power on" lan_power_is on || ((failures++)) || true
e2e_run_test "vm power off (in band)" vm_chassis_power 0 || ((failures++)) || true
e2e_run_test "lan sees power off" lan_power_is off || ((failures++)) || true

e2e_report "VM protocol E2E" "${failures}"
