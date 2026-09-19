#!/usr/bin/env sh

set -e

CONF_DIR="/etc/uqda"

# If an existing Yggdrasil container's volume (which used the filename
# config.conf) is mounted at $CONF_DIR, migrate it to uqda.conf instead of
# silently generating a brand new identity. Validate first, and never
# delete the original config.conf - it stays right where it was mounted.
if [ ! -f "$CONF_DIR/uqda.conf" ] && [ -f "$CONF_DIR/config.conf" ]; then
  echo "found an existing config.conf - validating before migrating it to uqda.conf"
  if uqda -useconffile "$CONF_DIR/config.conf" -checkconf; then
    cp "$CONF_DIR/config.conf" "$CONF_DIR/uqda.conf"
  else
    echo "config.conf did not pass validation - not migrating it automatically"
  fi
fi

if [ ! -f "$CONF_DIR/uqda.conf" ]; then
  echo "generate $CONF_DIR/uqda.conf"
  (umask 037 && uqda --genconf > "$CONF_DIR/uqda.conf")
fi

if [ -n "$ALLOW_IPV6_FORWARDING" ]; then
  echo "set sysctl -w net.ipv6.conf.all.forwarding=1"
  sysctl -w net.ipv6.conf.all.forwarding=1
fi

uqda --useconf < "$CONF_DIR/uqda.conf"
exit $?
