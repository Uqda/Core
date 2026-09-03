#!/bin/sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
TOOL=$ROOT/contrib/gateway/uqda-gateway

sh -n "$TOOL"

OUT=$(sh "$TOOL" plan --backend networkmanager --wan eth0 --lan wlan0 --ssid Home-UQDA)
printf '%s\n' "$OUT" | grep -F 'backend=networkmanager' >/dev/null
printf '%s\n' "$OUT" | grep -F 'WAN=eth0; Wi-Fi/LAN=wlan0; SSID=Home-UQDA' >/dev/null
printf '%s\n' "$OUT" | grep -F 'No changes made' >/dev/null && exit 1 || true
if printf '%s\n' "$OUT" | grep "$(printf '\033')" >/dev/null; then
	echo 'automatic color escaped into redirected gateway output' >&2
	exit 1
fi

COLOR_OUT=$(sh "$TOOL" plan --backend networkmanager --wan eth0 --lan wlan0 --color always)
printf '%s\n' "$COLOR_OUT" | grep "$(printf '\033')" >/dev/null
NO_COLOR_OUT=$(NO_COLOR=1 sh "$TOOL" plan --backend networkmanager --wan eth0 --lan wlan0)
if printf '%s\n' "$NO_COLOR_OUT" | grep "$(printf '\033')" >/dev/null; then
	echo 'NO_COLOR gateway output contained ANSI escapes' >&2
	exit 1
fi

OUT=$(sh "$TOOL" plan --backend openwrt)
printf '%s\n' "$OUT" | grep -F 'WAN=REQUIRED; Wi-Fi/LAN=REQUIRED' >/dev/null
printf '%s\n' "$OUT" | grep -F 'No changes made' >/dev/null

OUT=$(sh "$TOOL" plan --backend networkmanager --profile cafe --wan eth0 --lan wlan0 --ssid Cafe-UQDA)
printf '%s\n' "$OUT" | grep -F 'profile=cafe' >/dev/null
printf '%s\n' "$OUT" | grep -F 'Isolate Wi-Fi clients' >/dev/null
printf '%s\n' "$OUT" | grep -F 'requires --public' >/dev/null

ERR_FILE=$ROOT/gateway-test.err
if sh "$TOOL" apply --backend networkmanager --profile cafe --wan eth0 --lan wlan0 >"$ERR_FILE" 2>&1; then
	echo 'cafe apply accepted without public acknowledgement' >&2
	exit 1
fi
grep -F 'requires --public acknowledgement' "$ERR_FILE" >/dev/null
rm -f "$ERR_FILE"

if sh "$TOOL" plan --backend networkmanager --wan 'eth0;bad' --lan wlan0 >/dev/null 2>&1; then
	echo 'unsafe interface name was accepted' >&2
	exit 1
fi

printf '%s\n' 'gateway tests passed'
