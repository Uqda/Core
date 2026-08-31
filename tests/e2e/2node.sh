#!/usr/bin/env bash

set -euo pipefail
source "$(dirname "$0")/common.sh"
prepare_e2e

A="uq2-a-$$"
B="uq2-b-$$"
PORT=12101

create_namespace "$A"
create_namespace "$B"
link_namespaces "$A" a0 10.240.1.1/30 "$B" b0 10.240.1.2/30

CFG_B=$(write_node_config b "tcp://0.0.0.0:$PORT")
CFG_A=$(write_node_config a - "tcp://10.240.1.2:$PORT")

start_node "$B" b "$CFG_B"
start_node "$A" a "$CFG_A"

wait_peer_count a 1
wait_peer_count b 1

ADDR_A=$(node_address a)
ADDR_B=$(node_address b)
[[ -n "$ADDR_A" && -n "$ADDR_B" ]] || fail "failed to read overlay addresses"

wait_for_ping "$A" "$ADDR_B"
wait_for_ping "$B" "$ADDR_A"
assert_routing_entries_at_least a 2
assert_routing_entries_at_least b 2

echo "[E2E] PASS: two-node encrypted overlay connectivity"
