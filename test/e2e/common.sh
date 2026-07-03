# common.sh — shared helpers for go-ipmi E2E test scripts.
# Source from test/e2e/*.sh; do not execute directly.

E2E_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GOIPMI="${E2E_DIR}/../../_output/goipmi"
SERVER_BIN="${E2E_DIR}/../../_output/goipmi-server"

e2e_init() {
	RED='\033[0;31m'
	GREEN='\033[0;32m'
	NC='\033[0m'
}

e2e_run_test() {
	local name="$1"
	shift
	echo ""
	echo "--- ${name} ---"
	if "$@"; then
		echo -e "${GREEN}✓ ${name}${NC}"
	else
		echo -e "${RED}✗ ${name}${NC}" >&2
		return 1
	fi
}

# Run the three standard chassis cases.  Increments the caller's failures counter
# (pass the variable name as the first argument).
e2e_run_chassis_cases() {
	local -n _fail="${1:?failures counter variable name required}"
	shift
	e2e_run_test "chassis status" "$@" chassis status || ((_fail++)) || true
	e2e_run_test "chassis power on" "$@" chassis power on || ((_fail++)) || true
	e2e_run_test "chassis power off" "$@" chassis power off || ((_fail++)) || true
}

# Run chassis cases over IPMI v2.0 / RMCP+ (ipmitool -I lanplus).
e2e_run_chassis_cases_lanplus() {
	local -n _fail="${1:?failures counter variable name required}"
	shift
	e2e_run_test "lanplus chassis status" "$@" chassis status || ((_fail++)) || true
	e2e_run_test "lanplus chassis power on" "$@" chassis power on || ((_fail++)) || true
	e2e_run_test "lanplus chassis power off" "$@" chassis power off || ((_fail++)) || true
}

# Run chassis cases over IPMI v1.5 (ipmitool -I lan -A MD5).
e2e_run_chassis_cases_lan() {
	local -n _fail="${1:?failures counter variable name required}"
	shift
	e2e_run_test "lan chassis status" "$@" chassis status || ((_fail++)) || true
	e2e_run_test "lan chassis power on" "$@" chassis power on || ((_fail++)) || true
	e2e_run_test "lan chassis power off" "$@" chassis power off || ((_fail++)) || true
}

e2e_report() {
	local suite="$1"
	local failures="$2"
	echo ""
	if [ "${failures}" -eq 0 ]; then
		echo -e "${GREEN}========================================"
		echo " ${suite}: PASSED"
		echo -e "========================================${NC}"
	else
		echo -e "${RED}========================================"
		echo " ${suite}: ${failures} test(s) FAILED"
		echo -e "========================================${NC}"
		exit 1
	fi
}
