# Threat Model

This document models the concrete threats this codebase faces today, based
on an actual audit of the untrusted-input boundaries in this repository
(not a generic checklist). It exists to make security work traceable: every
mitigation elsewhere in the codebase should map back to a threat listed
here, and every threat listed here should eventually map to a mitigation,
a test, or an explicitly accepted residual risk.

Scope: this document covers the `yggdrasil-go` implementation in this
repository. Routing and session-layer wire framing is owned by the external
`github.com/Arceliar/ironwood` dependency and is out of scope for this
document — see that project's own security posture. This repository's own
untrusted-input surface is smaller than the whole networking stack, and
this document is precise about where that surface actually is.

## Assets

- **Node identity** — the ed25519 private key that determines a node's
  network address (`src/config`, loaded via `-useconffile`/`-useconf`).
- **Local admin control** — the ability to add/remove peers, read session
  and routing state, and (via the TUN module) read/write packets
  (`src/admin`).
- **Traffic confidentiality/integrity** between two peers, and between two
  nodes communicating end-to-end over the mesh (delegated to Ironwood's
  session layer, not re-implemented here).
- **Availability** of the daemon process itself (a crash affects every
  service depending on the node's connectivity).

## Trust boundaries and threats

### 1. Malicious direct peer — handshake

**Boundary:** `src/core/link.go`, `l.handler()` (line ~627), calling
`version_metadata.decode()` in `src/core/version.go`. This runs on **every**
inbound TCP/TLS/QUIC/WS/WSS connection to a listening port, before any
authentication, pinned-key check, or `AllowedPublicKeys` check has
happened. Anyone who can reach a configured listener can send arbitrary
bytes into this parser.

- **Threat:** a crafted handshake message causes a parser panic (DoS via
  crash), an out-of-bounds read, or excessive memory allocation.
- **Mitigation today:** all length fields are `uint16` (bounding
  allocation to 64KB), every field-length claim is checked against
  remaining buffer length before slicing (`src/core/version.go:119-152`),
  and `io.ReadFull` is used for all reads so partial/short writes don't
  desync parsing.
- **Verification:** `FuzzVersionMetadataDecode` (`src/core/version_test.go`)
  — 2.5M+ mutated inputs run with zero panics as of the commit that added
  it. Table-driven tests (`TestVersionDecodeRejectsMalformedFieldLengths`,
  `TestVersionDecodeRejectsTrailingBytes`) cover the specific edge cases
  identified during the audit.
- **Residual risk:** a handshake attempt holds a goroutine and a 6-second
  connection deadline (`conn.SetDeadline(... 6 * time.Second)`,
  `src/core/link.go:635`) per attempt; there is currently no global cap on
  concurrent in-progress handshakes, so many slow/incomplete handshake
  attempts could accumulate goroutines and open file descriptors. Tracked
  for Phase 4 (Uqda Guard) as a connection-admission control item, not yet
  implemented.

### 2. Malicious direct peer — post-handshake

**Boundary:** after `version_metadata.decode()` succeeds, the connection is
handed to `l.core.HandleConn()` (`src/core/link.go:708`), which delegates
to Ironwood. `conn.SetDeadline(time.Time{})` clears the handshake deadline
entirely (`src/core/link.go:657`) — there is no read/write deadline on an
established link at the yggdrasil-go layer.

- **Threat:** a peer that completes the handshake but then sends slowly,
  never sends, or never reads (holding the TCP connection open) ties up
  the connection's goroutines/buffers indefinitely.
- **Mitigation today:** none at this layer; this is delegated to Ironwood
  and to OS-level TCP keepalive behavior.
- **Residual risk:** open — tracked for Phase 4/Phase 8 (Uqda Guard /
  Reliability Engineering). Any future timeout added here must be
  verified against real Ironwood traffic patterns first, since an
  aggressive timeout could disconnect legitimate slow links.

### 3. Malicious or spoofed local-network host — multicast discovery

**Boundary:** `src/multicast/advertisement.go`,
`multicastAdvertisement.UnmarshalBinary()`. This runs on every UDP
multicast/broadcast packet received on an interface configured for local
discovery — no authentication, and (unlike the TCP handshake) no port-level
access control; anyone on the same L2 segment can send crafted packets
here directly, without even attempting a real peering.

- **Threat:** same class as #1 — parser panic or resource exhaustion from
  a malformed beacon.
- **Mitigation today:** header length is checked before every slice
  (`src/multicast/advertisement.go:29-42`); the declared hash length is
  bounds-checked against the actual remaining buffer before use.
- **Verification:** `FuzzMulticastAdvertisementUnmarshalBinary`
  (`src/multicast/advertisement_test.go`) — 3.1M+ mutated inputs, zero
  panics. `TestMulticastAdvertisementRejectsTruncatedHash` covers the
  specific truncation edge case.
- **Residual risk:** a validly-parsed but forged beacon (correct format,
  attacker's own real key) is not a vulnerability in the parser — it's
  just an invitation to peer, which then goes through the same handshake
  scrutiny as #1. Local discovery inherently trusts "reachable on this L2
  segment" as its only admission criterion; that's a design property of
  multicast discovery generally, not a defect introduced here, and it's
  why `AllowedPublicKeys` and `GroupPassword` exist as additional gates.

### 4. Malicious remote network participant — routed traffic / decrypted packets

**Boundary:** `src/ipv6rwc/` and `src/ipv6rwc/icmpv6.go`. Once a packet has
traversed the mesh and been decrypted, ipv6rwc treats its payload as an
arbitrary IPv6 packet (potentially with attacker-controlled contents) and
writes it to the OS TUN interface, or reads local packets and encrypts
them outward.

- **Threat:** a crafted "arrived" packet triggers a bug in the
  address/keystore bookkeeping in `keyStore` (`src/ipv6rwc/ipv6rwc.go`), or
  a crafted ICMPv6 payload triggers unexpected OS/network-stack behavior
  once written to TUN.
- **Mitigation today:** the packet payload itself is opaque to Uqda beyond
  address/header parsing needed for internal routing bookkeeping; the OS
  kernel's own IPv6 stack is the actual consumer of TUN-written packets, so
  most payload-content risk is inherited by the OS, not this code.
- **Residual risk:** not yet fuzzed. Tracked for a future pass once Phase 3
  restructuring settles this subsystem's boundaries (moving to
  `src/tun/` per RESTRUCTURING.md) — fuzzing a moving target is wasted
  effort.

### 5. Local unprivileged user — admin socket

**Boundary:** `src/admin/admin.go`. Explicitly marked upstream with
`// TODO: Add authentication` (line 18). Default listen addresses are
local-only (`unix:///var/run/yggdrasil.sock` on Linux/macOS/*BSD,
`tcp://localhost:9001` on Windows — see `src/config/defaults_*.go`), and
the Unix socket is created with mode `0660` (`src/admin/admin.go:118`).

- **Threat:** any local process that can reach the socket (same user, or
  same group if the socket's group permissions allow it, or any process at
  all if an operator reconfigures `AdminListen` to a non-local address) can
  fully control the node: add/remove peers, read routing/session state,
  reconfigure TUN — with **no credential check of any kind**.
- **Mitigation today:** the default listen address is local-only, and the
  Unix socket's file permissions provide OS-level access control on
  Unix-likes. This is access control by *placement*, not by
  *authentication*.
- **Residual risk:** the Windows default (`tcp://localhost:9001`) has no
  filesystem permission equivalent — any local process/user on a
  multi-user Windows box can connect. And an operator who sets
  `AdminListen` to a non-loopback address with no additional protection
  exposes full node control to the network. Tracked for Phase 14 (Uqda
  Control hardening) — the fix is authentication on the admin socket
  itself, not just documentation, but that is a real behavior change
  requiring its own design and is not implemented in this pass.

### 6. Compromised or malicious public gateway

Not applicable to this repository's code today — this is an operational
threat against whoever runs a public peer (Uqda Gateway), not a code-level
threat. A compromised gateway operator can see traffic transiting their
node exactly as any other relay node can (Ironwood's session encryption is
end-to-end between the two communicating nodes, not link-by-link secret to
each relay — relays forward encrypted session traffic they cannot decrypt).
Full treatment belongs in Phase 10 (Uqda Gateway) operational docs, not
here.

### 7. Compromised dependency

See [BASELINE.md](../BASELINE.md) for the current `govulncheck` results (0
reachable vulnerabilities, 5 unreachable ones in transitive dependencies
tracked for Phase 22). This threat is process-level (keep scanning,
minimize dependency surface) rather than something resolvable by a single
code change.

### 8. Stolen identity key

**Boundary:** wherever the private key touches disk or memory:
`src/config` (loading/generating), `cmd/yggdrasil/main.go` (holds it in
process memory for the process lifetime).

- **Threat:** if the config file (or `PrivateKeyPath` file) is readable by
  an unauthorized local user, or if it's ever logged, an attacker can fully
  impersonate the node (its network address is derived from this key).
- **Mitigation today:** none enforced by this codebase — file permissions
  on the config file are entirely the operator's responsibility today; the
  code does not check or warn about permissive file modes.
- **Residual risk:** open. Tracked for Phase 12 (Identity/key safety) —
  candidate mitigation is warning (not silently failing) when the config
  or key file is group/world-readable on platforms where that's
  meaningful.

### 9. Malicious update source

Not applicable — see [ARCHITECTURE.md](../ARCHITECTURE.md): no update
mechanism exists in this codebase today. This threat becomes relevant only
once Phase 20 (Uqda Update) is implemented, and that subsystem must be
designed with this threat in mind from the start (see the "Do Not
Implement" and "Uqda Update" sections of the original engineering
directive: HTTPS-only sources, checksum/signature verification before
install, staged verification, safe rollback).

### 10. Log/diagnostic data leakage

**Boundary:** `cmd/yggdrasil/main.go` logging setup, `src/admin`'s
`getSelf`/`getPeers`/etc. handlers.

- **Threat:** private keys, passwords, or `GroupPassword` values ending up
  in logs or admin socket responses.
- **Mitigation today:** spot-checked during this audit —
  `getSelf`/`getPeers` responses (`src/admin/getself.go`,
  `src/admin/getpeers.go`) expose only public keys, derived addresses, and
  connection metadata; no private key or password field is present in any
  admin response struct reviewed. Not exhaustively audited across every
  log call site in this pass.
- **Residual risk:** a full log-call audit across the codebase is tracked
  for Phase 15 (Logging standardization), not completed here.

## Explicitly out of scope for Uqda (per project policy)

Anonymity/traffic-analysis resistance is not a goal — Uqda (like
Yggdrasil) provides confidentiality and integrity for session traffic, not
sender/receiver anonymity or protection against a global passive observer
correlating traffic patterns. This must be stated plainly in user-facing
documentation (Phase 11), not just here.

## How to keep this document honest

Every new fuzz target, every new admin authentication mechanism, every
resource-limit added under Uqda Guard should update the relevant section
above with a concrete file reference and test name — the same way section
1 and 3 above cite the fuzz tests that back their claims. A mitigation
without a linked test is a claim, not a fact; avoid adding one without the
other.
