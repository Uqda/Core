#!/usr/bin/env bash

set -euo pipefail
source "$(dirname "$0")/common.sh"
prepare_e2e

LEGACY_BIN=${UQDA_LEGACY_PEER_BIN:-"$ROOT/tests/e2e/legacy-peer"}
[[ -x "$LEGACY_BIN" ]] || fail "legacy peer fixture not built: $LEGACY_BIN"

A="uqs-a-$$"
B="uqs-b-$$"
S="uqs-s-$$"
L="uqs-l-$$"
STRICT_PORT=12105
LEGACY_PORT=12106

# First prove that strict mode succeeds when both peers support transcript-bound
# handshake confirmation.
create_namespace "$A"
create_namespace "$B"
link_namespaces "$A" a0 10.243.1.1/30 "$B" b0 10.243.1.2/30

CFG_B=$(write_node_config b "tcp://0.0.0.0:$STRICT_PORT?secure=required")
CFG_A=$(write_node_config a - "tcp://10.243.1.2:$STRICT_PORT?secure=required")
start_node "$B" b "$CFG_B"
start_node "$A" a "$CFG_A"
wait_peer_count a 1 25
wait_peer_count b 1 25

ADDR_B=$(node_address b)
[[ -n "$ADDR_B" ]] || fail "failed to read strict peer overlay address"
wait_for_ping "$A" "$ADDR_B" 25

echo "[E2E] secure=required accepted a hardened peer"

# Then connect a strict current node to a valid legacy 0.5 handshake fixture.
# The legacy hello is correctly signed but advertises no confirmation capability,
# so the connection must be rejected specifically because strict mode is enabled.
create_namespace "$S"
create_namespace "$L"
link_namespaces "$S" s0 10.243.2.1/30 "$L" l0 10.243.2.2/30

LEGACY_LOG="$E2E_TMP/legacy.log"
ip netns exec "$L" "$LEGACY_BIN" -listen "10.243.2.2:$LEGACY_PORT" >"$LEGACY_LOG" 2>&1 &
E2E_PIDS[legacy]=$!
sleep 0.2
kill -0 "${E2E_PIDS[legacy]}" 2>/dev/null || { cat "$LEGACY_LOG" >&2; fail "legacy peer fixture failed to start"; }

CFG_S=$(write_node_config s - "tcp://10.243.2.2:$LEGACY_PORT?secure=required")
start_node "$S" s "$CFG_S"
wait_peer_error s "peer does not support transcript-bound handshakes" 20

if (( $(peer_up_count s) != 0 )); then
  fail "secure=required unexpectedly accepted a legacy 0.5 peer"
fi

echo "[E2E] PASS: secure=required accepts hardened peers and rejects legacy 0.5 peers"
