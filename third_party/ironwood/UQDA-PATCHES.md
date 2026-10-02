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

No packet formats, authentication, routing decisions or wire versions are changed.
The independent pinned-upstream interoperability gate remains mandatory. Remove
the local replacement when an upstream version contains an equivalent verified
fix. The root module retains the original version and checksum as provenance.
