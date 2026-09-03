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
if printf '%s\n' "$OUT" | grep "$(printf '\033')" >/dev/null; then
	echo 'automatic color escaped into redirected output' >&2
	exit 1
fi

COLOR_OUT=$(UQDA_LATENCY_PING_OUTPUT="$GOOD" sh "$TOOL" 201::2 --count 3 --target-ms 20 --color always)
printf '%s\n' "$COLOR_OUT" | grep "$(printf '\033')" >/dev/null
NO_COLOR_OUT=$(NO_COLOR=1 UQDA_LATENCY_PING_OUTPUT="$GOOD" sh "$TOOL" 201::2 --count 3 --target-ms 20)
if printf '%s\n' "$NO_COLOR_OUT" | grep "$(printf '\033')" >/dev/null; then
	echo 'NO_COLOR output contained ANSI escapes' >&2
	exit 1
fi

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

INTERRUPT_DIR=$ROOT/latency-interrupt-test
mkdir -p "$INTERRUPT_DIR/bin" "$INTERRUPT_DIR/tmp"
cat >"$INTERRUPT_DIR/bin/ping" <<'EOF'
#!/bin/sh
case " $* " in *' -c 1 '*) exit 0 ;; esac
printf '%s\n' "$$" >"$UQDA_TEST_PING_PID"
: >"$UQDA_TEST_PING_READY"
trap 'exit 143' HUP INT TERM
while :; do sleep 1; done
EOF
chmod +x "$INTERRUPT_DIR/bin/ping"
UQDA_TEST_PING_READY=$INTERRUPT_DIR/ready
UQDA_TEST_PING_PID=$INTERRUPT_DIR/pid
export UQDA_TEST_PING_READY UQDA_TEST_PING_PID
PATH="$INTERRUPT_DIR/bin:$PATH" TMPDIR="$INTERRUPT_DIR/tmp" sh "$TOOL" 201::2 --count 100 >"$INTERRUPT_DIR/out" 2>"$INTERRUPT_DIR/err" &
TOOL_PID=$!
TRIES=0
while [ ! -e "$UQDA_TEST_PING_READY" ] && [ "$TRIES" -lt 50 ]; do
	sleep 0.1
	TRIES=$((TRIES + 1))
done
[ -e "$UQDA_TEST_PING_READY" ] || { echo 'interrupt test ping did not start' >&2; exit 1; }
kill -TERM "$(cat "$UQDA_TEST_PING_PID")" "$TOOL_PID" 2>/dev/null || true
STATUS=0
wait "$TOOL_PID" || STATUS=$?
[ "$STATUS" -eq 143 ] || { echo "interrupted measurement exited $STATUS, expected 143" >&2; exit 1; }
grep -F 'CANCELLED: measurement interrupted' "$INTERRUPT_DIR/err" >/dev/null
if grep -F 'No such file or directory' "$INTERRUPT_DIR/err" >/dev/null; then
	echo 'interrupted measurement accessed a removed temporary file' >&2
	exit 1
fi
if find "$INTERRUPT_DIR/tmp" -mindepth 1 -print -quit | grep . >/dev/null; then
	echo 'interrupted measurement left temporary files behind' >&2
	exit 1
fi
rm -rf "$INTERRUPT_DIR"

printf '%s\n' 'latency tests passed'
