#!/usr/bin/env sh

set -e

CONF_DIR="/etc/uqda-network"

if [ ! -f "$CONF_DIR/config.conf" ]; then
  echo "generate $CONF_DIR/config.conf"
  (umask 037 && uqda -genconf > "$CONF_DIR/config.conf")
fi

if [ -n "$ALLOW_IPV6_FORWARDING" ]; then
  echo "set sysctl -w net.ipv6.conf.all.forwarding=1"
  sysctl -w net.ipv6.conf.all.forwarding=1
fi

uqda -useconf < "$CONF_DIR/config.conf"
exit $?
