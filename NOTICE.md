# Notice and Attribution

This project is planned to be renamed and rebranded as **Uqda** (Arabic:
عُقدة) — see [NAMING.md](NAMING.md) and
[RESTRUCTURING.md](RESTRUCTURING.md) for the naming and restructuring
plan. As of this file, the module path, binaries, and most identifiers in
this repository are still the unmodified upstream ones; this notice
applies regardless of which name the checkout currently uses.

## This project's own license

This software is licensed under the **LGPLv3**, with a special exception
permitting static/dynamic linking without triggering the LGPL's
source-disclosure requirements for the linking application. See
[LICENSE](LICENSE) for the full text — that file is authoritative; this
document only summarizes and points to it.

## Upstream origin

This repository is derived from **Yggdrasil**
(`github.com/yggdrasil-network/yggdrasil-go`), an open-source
implementation of an end-to-end encrypted, self-arranging IPv6 mesh
network. Uqda's stated goal — see [NAMING.md](NAMING.md) — is to evolve
the implementation while remaining a fully interoperable participant on
the same Yggdrasil network and wire protocol, not to fork the network
itself. Capabilities inherited from Yggdrasil (routing, addressing,
peering, the admin API, TUN integration, and everything else present
before this repository's own Uqda-specific changes began) are Yggdrasil's
work, not a Uqda invention — see
[docs/IMPLEMENTATION_STATUS.md](docs/IMPLEMENTATION_STATUS.md) and this
repository's commit history for what has actually changed since that
baseline (`27dd82b`).

## Key direct dependencies

The routing and end-to-end encrypted session layer this project builds on
top of is **not implemented in this repository**. It comes from:

- **`github.com/Arceliar/ironwood`** — licensed under the **Mozilla Public
  License 2.0**. Provides the DHT-based routing and encrypted session
  protocol.
- **`github.com/Arceliar/phony`** — licensed under the **Mozilla Public
  License 2.0**. Provides the actor-model concurrency primitive
  (`phony.Inbox`/`phony.Block`) used throughout `src/core`.

Both were verified directly against the LICENSE file vendored in each
module (via the local Go module cache) at the time this notice was
written, rather than assumed.

## The rest of the dependency tree

This project's `go.mod`/`go.sum` pull in several dozen further direct and
transitive dependencies (transport libraries, TUN backends, logging,
config parsing, CLI formatting, and their own transitive dependencies —
run `go list -m all` for the complete, current list). Rather than
hand-transcribe license text for each one here — which would drift out of
date the moment any dependency is upgraded, added, or removed, and which
this document has not independently verified name-by-name the way it did
for Ironwood and phony above — full third-party license compliance for a
release should be generated mechanically (e.g. `go-licenses`, an SBOM
tool) as part of Phase 22 (Supply-Chain Security) / Phase 19 (Verifiable
Builds), not maintained by hand in this file. Each dependency's own
repository is the authoritative source for its license in the meantime.

## What this document does not do

It does not claim upstream Yggdrasil or Ironwood authorship for anything
they didn't write, and it does not mechanically relabel their copyright as
Uqda's own. Branding described in [NAMING.md](NAMING.md) applies to this
project's own product, tooling, documentation, and code organization — it
does not extend to erasing where the underlying implementation came from.
