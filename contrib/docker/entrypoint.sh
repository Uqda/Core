#!/usr/bin/env sh

set -e

CONF_DIR="/etc/uqda-network"

if [ ! -f "$CONF_DIR/config.conf" ]; then
  echo "generate $CONF_DIR/config.conf"
  uqda -genconf > "$CONF_DIR/config.conf"
fi

uqda -useconf < "$CONF_DIR/config.conf"
exit $?
