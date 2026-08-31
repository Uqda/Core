#!/usr/bin/env bash

set -euo pipefail
source "$(dirname "$0")/common.sh"
prepare_e2e

A="uqf-a-$$"
B="uqf-b-$$"
C="uqf-c-$$"
D="uqf-d-$$"
PORT=12103

create_namespace "$A"
create_namespace "$B"
create_namespace "$C"
create_namespace "$D"

# Ring topology: A-B-C-D-A. A and C therefore have two independent mesh paths.
link_namespaces "$A" a1 10.241.1.1/30 "$B" b1 10.241.1.2/30
link_namespaces "$B" b2 10.241.2.1/30 "$C" c2 10.241.2.2/30
link_namespaces "$C" c3 10.241.3.1/30 "$D" d3 10.241.3.2/30
link_namespaces "$D" d4 10.241.4.1/30 "$A" a4 10.241.4.2/30

CFG_A=$(write_node_config a "tcp://0.0.0.0:$PORT" "tcp://10.241.1.2:$PORT")
CFG_B=$(write_node_config b "tcp://0.0.0.0:$PORT" "tcp://10.241.2.2:$PORT")
CFG_C=$(write_node_config c "tcp://0.0.0.0:$PORT" "tcp://10.241.3.2:$PORT")
CFG_D=$(write_node_config d "tcp://0.0.0.0:$PORT" "tcp://10.241.4.2:$PORT")

start_node "$A" a "$CFG_A"
start_node "$B" b "$CFG_B"
start_node "$C" c "$CFG_C"
start_node "$D" d "$CFG_D"

wait_peer_count a 2 30
wait_peer_count b 2 30
wait_peer_count c 2 30
wait_peer_count d 2 30

ADDR_C=$(node_address c)
[[ -n "$ADDR_C" ]] || fail "failed to read node C overlay address"
wait_for_ping "$A" "$ADDR_C" 30

echo "[E2E] stopping node B to force A-C route convergence through D"
stop_node b

# The remaining A-D-C side of the ring must keep the destination reachable.
wait_peer_count a 1 20
wait_peer_count c 1 20
wait_peer_count d 2 20
wait_for_ping "$A" "$ADDR_C" 40

echo "[E2E] PASS: route survived node B failure via alternate A-D-C path"
