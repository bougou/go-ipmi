#!/bin/sh

set -eu

socat \
	PTY,link=/dev/vtty,raw,echo=0 \
	EXEC:/bin/sh,pty,stderr,setsid,sigint,sane &

exec ipmi_sim -n -c /tmp/ipmisim/lan.conf -f /tmp/ipmisim/sim.emu
