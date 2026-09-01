#!/bin/sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
TOOL=$ROOT/contrib/gateway/uqda-gateway

sh -n "$TOOL"

OUT=$(sh "$TOOL" plan --backend networkmanager --wan eth0 --lan wlan0 --ssid Home-UQDA)
printf '%s\n' "$OUT" | grep -F 'backend=networkmanager' >/dev/null
printf '%s\n' "$OUT" | grep -F 'WAN=eth0; Wi-Fi/LAN=wlan0; SSID=Home-UQDA' >/dev/null
printf '%s\n' "$OUT" | grep -F 'No changes made' >/dev/null && exit 1 || true

OUT=$(sh "$TOOL" plan --backend openwrt)
printf '%s\n' "$OUT" | grep -F 'WAN=REQUIRED; Wi-Fi/LAN=REQUIRED' >/dev/null
printf '%s\n' "$OUT" | grep -F 'No changes made' >/dev/null

if sh "$TOOL" plan --backend networkmanager --wan 'eth0;bad' --lan wlan0 >/dev/null 2>&1; then
	echo 'unsafe interface name was accepted' >&2
	exit 1
fi

printf '%s\n' 'gateway tests passed'
