#!/usr/bin/env bash

set -euo pipefail
source "$(dirname "$0")/common.sh"
prepare_e2e

A="uql-a-$$"
B="uql-b-$$"
PORT=12104

create_namespace "$A"
create_namespace "$B"
link_namespaces "$A" a0 10.242.1.1/30 "$B" b0 10.242.1.2/30

# Exercise the overlay over an intentionally poor underlay. TCP should absorb
# packet loss while the overlay remains reachable, at the cost of higher RTT.
ip netns exec "$A" tc qdisc add dev a0 root netem delay 80ms 10ms loss 5%
ip netns exec "$B" tc qdisc add dev b0 root netem delay 80ms 10ms loss 5%

CFG_B=$(write_node_config b "tcp://0.0.0.0:$PORT")
CFG_A=$(write_node_config a - "tcp://10.242.1.2:$PORT")

start_node "$B" b "$CFG_B"
start_node "$A" a "$CFG_A"
wait_peer_count a 1 30
wait_peer_count b 1 30

ADDR_B=$(node_address b)
[[ -n "$ADDR_B" ]] || fail "failed to read node B overlay address"

PING_OUTPUT=$(ip netns exec "$A" ping -6 -n -c 12 -W 3 "$ADDR_B" || true)
printf '%s\n' "$PING_OUTPUT"

grep -q '100% packet loss' <<<"$PING_OUTPUT" && fail "overlay became unreachable under degraded underlay"
AVG_RTT=$(awk -F/ '/^rtt min\/avg\/max/ {print $5}' <<<"$PING_OUTPUT" | tail -1)
[[ -n "$AVG_RTT" ]] || fail "could not parse overlay ping RTT"
awk -v rtt="$AVG_RTT" 'BEGIN { exit !(rtt >= 100.0) }' || fail "expected degraded overlay RTT >= 100ms, got ${AVG_RTT}ms"

ip netns exec "$A" tc qdisc show dev a0
ip netns exec "$B" tc qdisc show dev b0

echo "[E2E] PASS: overlay remained reachable with 80ms delay and 5% underlay loss per direction"
