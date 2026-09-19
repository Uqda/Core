#!/usr/bin/env sh
set -eu
CONF_DIR="/etc/uqda"
/usr/bin/uqda-install-config "$CONF_DIR/uqda.conf" "$CONF_DIR/config.conf"
if [ -n "${ALLOW_IPV6_FORWARDING:-}" ]; then
  sysctl -w net.ipv6.conf.all.forwarding=1
fi
exec uqda -useconffile "$CONF_DIR/uqda.conf" "$@"
