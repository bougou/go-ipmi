FROM alpine:3.22 AS builder

# Pin OpenIPMI v2.0.37
ARG OPENIPMI_VERSION=v2.0.37

RUN apk add --no-cache \
    alpine-sdk \
    autoconf \
    automake \
    glib-dev \
    libtool \
    linux-headers \
    ncurses-dev \
    net-snmp-dev \
    openssl-dev \
    popt-dev \
    readline-dev

WORKDIR /build
RUN git clone --branch "${OPENIPMI_VERSION}" --depth 1 \
        https://github.com/cminyard/openipmi.git && \
    cd openipmi && \
    autoreconf -ivf && \
    ./configure \
        --sysconfdir=/etc \
        --with-openssl \
        --with-glib \
        --enable-shared \
        --disable-static \
        --without-perl \
        --without-python \
        --without-tcl && \
    make -j"$(nproc)" && \
    make install

FROM alpine:3.22

RUN apk add --no-cache popt socat tini

COPY --from=builder /usr/local /usr/local

COPY --from=builder /build/openipmi/lanserv/ipmi_sim_chassiscontrol /usr/local/bin/ipmi_sim_chassiscontrol
COPY lan.conf /tmp/ipmisim/lan.conf
COPY sim.emu /tmp/ipmisim/sim.emu
COPY --chmod=0755 entrypoint.sh /usr/local/bin/ipmi-simulator

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/usr/local/bin/ipmi-simulator"]
