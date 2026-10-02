# Uqda's pinned Ironwood snapshot

Base: `github.com/Arceliar/ironwood`
commit `cc7fdd2b785ff05f1ec5d7a48c4d2bffb2f50588`
module `v0.0.0-20260924233544-cc7fdd2b785f`.

The upstream sources, COPYRIGHT and MPL-2.0 LICENSE are preserved. This is a local
replacement for that exact snapshot, not a new upstream release or endorsement.

Local changes:

- `network/debug.go`: snapshot registry membership first, then read mutable
  signature timing on each peer's actor. Previously the registry actor read
  `srst`/`srrt` while the peer actor could write them. Do not nest blocking actor
  calls: registry callbacks may otherwise deadlock with connection shutdown.
- `network/debug_test.go`: concurrent timestamp writes and peer snapshot reads,
  plus snapshot reads while peers are removed and added. Run with the race detector.
- `encrypted/crypto.go`: Uqda private groups derive their signature preimage with
  Argon2id (64 MiB, three passes, four lanes, 32-byte output) once at startup.
  The fixed `uqda/group-auth/argon2id/v1` salt is a protocol domain separator,
  not a unique password-storage salt. Keep group passwords strong and random.
- `encrypted/crypto_test.go`: pinned KDF vector, unchanged empty/public mode,
  same/different/absent password handshakes, and rejection of legacy signatures.

No packet formats, routing decisions or public-mode wire versions are changed.
Private-group authentication IS changed: every group member must upgrade together
from 26.0.1 or earlier. No legacy fallback is offered. The independent pinned-upstream
public interoperability gate remains mandatory. Remove the local replacement only
when upstream contains equivalent verified fixes AND a deliberate migration plan
preserves Uqda private-group authentication. The root module retains the original
version and checksum as provenance.

Go consumers do not inherit dependency-module `replace` directives. This local
patch is included by builds using Core as the main module, including all release
packages. A downstream Go main module must opt into the replacement explicitly;
otherwise it selects the original upstream module version.
