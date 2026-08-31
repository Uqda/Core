#!/usr/bin/env bash

set -euo pipefail
source "$(dirname "$0")/common.sh"
prepare_e2e

A="uq3-a-$$"
B="uq3-b-$$"
C="uq3-c-$$"
PORT=12102

create_namespace "$A"
create_namespace "$B"
create_namespace "$C"

link_namespaces "$A" a0 10.240.11.1/30 "$B" b1 10.240.11.2/30
link_namespaces "$B" b2 10.240.12.1/30 "$C" c0 10.240.12.2/30

CFG_B=$(write_node_config b "tcp://0.0.0.0:$PORT")
CFG_A=$(write_node_config a - "tcp://10.240.11.2:$PORT")
CFG_C=$(write_node_config c - "tcp://10.240.12.1:$PORT")

start_node "$B" b "$CFG_B"
start_node "$A" a "$CFG_A"
start_node "$C" c "$CFG_C"

wait_peer_count a 1
wait_peer_count b 2
wait_peer_count c 1

# A and C have no direct IPv4 underlay route. Any successful overlay ping
# therefore has to traverse UQDA through B.
if ip netns exec "$A" ip route get 10.240.12.2 >/dev/null 2>&1; then
  fail "node A unexpectedly has a direct underlay route to node C"
fi
if ip netns exec "$C" ip route get 10.240.11.1 >/dev/null 2>&1; then
  fail "node C unexpectedly has a direct underlay route to node A"
fi

ADDR_A=$(node_address a)
ADDR_C=$(node_address c)
[[ -n "$ADDR_A" && -n "$ADDR_C" ]] || fail "failed to read overlay addresses"

wait_for_ping "$A" "$ADDR_C" 30
wait_for_ping "$C" "$ADDR_A" 30
assert_routing_entries_at_least a 3
assert_routing_entries_at_least b 3
assert_routing_entries_at_least c 3

echo "[E2E] PASS: A <-> C multi-hop routing through B"
