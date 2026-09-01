#!/bin/sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
TOOL=$ROOT/contrib/performance/uqda-latency

sh -n "$TOOL"

GOOD='64 bytes from 201::2: icmp_seq=1 ttl=64 time=8.0 ms
64 bytes from 201::2: icmp_seq=2 ttl=64 time=10.0 ms
64 bytes from 201::2: icmp_seq=3 ttl=64 time=12.0 ms'
OUT=$(UQDA_LATENCY_PING_OUTPUT="$GOOD" sh "$TOOL" 201::2 --count 3 --target-ms 20)
printf '%s\n' "$OUT" | grep -F 'p50=10.000ms p95=12.000ms' >/dev/null
printf '%s\n' "$OUT" | grep -F 'PASS' >/dev/null

SLOW='64 bytes from 201::2: icmp_seq=1 ttl=64 time=8.0 ms
64 bytes from 201::2: icmp_seq=2 ttl=64 time=11.0 ms
64 bytes from 201::2: icmp_seq=3 ttl=64 time=48.0 ms'
if UQDA_LATENCY_PING_OUTPUT="$SLOW" sh "$TOOL" 201::2 --count 3 --target-ms 20 >"$ROOT/latency-test.out"; then
	echo 'slow p95 unexpectedly passed' >&2
	exit 1
fi
grep -F 'p95=48.000ms' "$ROOT/latency-test.out" >/dev/null
grep -F 'WARN' "$ROOT/latency-test.out" >/dev/null
rm -f "$ROOT/latency-test.out"

LOSS='64 bytes from 201::2: icmp_seq=1 ttl=64 time=8.0 ms
64 bytes from 201::2: icmp_seq=3 ttl=64 time=9.0 ms'
if UQDA_LATENCY_PING_OUTPUT="$LOSS" sh "$TOOL" 201::2 --count 3 --target-ms 20 >"$ROOT/latency-test.out"; then
	echo 'packet loss unexpectedly passed' >&2
	exit 1
fi
grep -F 'samples=2/3' "$ROOT/latency-test.out" >/dev/null
grep -F 'packet loss' "$ROOT/latency-test.out" >/dev/null
rm -f "$ROOT/latency-test.out"

printf '%s\n' 'latency tests passed'
