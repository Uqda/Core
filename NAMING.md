# Uqda Naming Specification

Status: **specification only.** No file, package, binary, or identifier in this
repository has been renamed yet. This document is the single source of truth
that future rebranding work must follow. See [RESTRUCTURING.md](RESTRUCTURING.md)
for how/when it gets applied.

## Origin of the name

**Uqda** — Arabic: **عُقدة** — meaning a knot, a point of connection, a node.
In networking, a node is a participating point in a network: it connects to
other nodes, forwards traffic, establishes routes, and contributes to the
network topology. The name is not an acronym; it is the Latin-script project
name derived from the Arabic word.

> **EN:** Uqda Core is a hardened, high-performance implementation of the
> Yggdrasil networking protocol, designed to remain fully interoperable with
> the existing Yggdrasil network.
>
> **AR:** Uqda Core هو تنفيذ مطوّر ومحصّن عالي الأداء لبروتوكول Yggdrasil،
> مصمم للعمل بتوافق كامل مع شبكة Yggdrasil الحالية.

## The non-negotiable rule

**Uqda renames the implementation, the product, and the tooling. It does not
rename or alter the Yggdrasil wire protocol, network, or node identity.**

A Uqda node must interoperate transparently with an unmodified Yggdrasil node,
in both directions, with no configuration workaround required. Concretely,
do **not** change:

- Wire message formats
- Peer negotiation / peer URI behavior
- Node identity format (ed25519-based)
- IPv6 address derivation from node identity
- Routing message format (owned upstream by Ironwood)
- Session protocol
- Anything that would make `Uqda ↔ Yggdrasil` fail

Everything else — code structure, tooling, packaging, operations, security
hardening, diagnostics, documentation, branding — is fair game and is the
whole point of the project.

There is **no "Uqda Protocol"** and **no "Uqda Network."** The correct framing
is always: *Uqda Core, a Yggdrasil-compatible implementation*, running on
*the Yggdrasil network*.

## Product / concept naming

| Name | Role |
|---|---|
| **Uqda** | Project and product name |
| **Uqda Core** | Primary networking engine |
| **Uqda Node** | A running network node (a device running Uqda) |
| **Uqda Gateway** | A public, professionally operated Yggdrasil-compatible peer — not a separate network, not a protocol translator |
| **Uqda Guard** | Security / abuse-resistance boundary in front of the peer engine |
| **Uqda Route** | Routing integration layer wrapping Ironwood (Ironwood itself keeps its own name, license, and attribution) |
| **Uqda Link** | Transport / direct-peer connectivity layer (TCP/TLS/QUIC/WS/SOCKS/Unix) |
| **Uqda Session** | End-to-end encrypted session handling |
| **Uqda Identity** | Node identity and key handling |
| **Uqda Address** | IPv6 identity-derived addressing |
| **Uqda Discover** | Local peer discovery (multicast) |
| **Uqda Control** | Administration subsystem (local API) |
| **Uqda Doctor** | Health and security diagnostics |
| **Uqda Update** | Verified update subsystem (does not exist upstream today — new work) |
| **Uqda Lab** | Fuzzing / chaos / interoperability / performance test environment |
| **Uqda Build** | Build, packaging, and release infrastructure |

## Technical / filesystem naming

| Item | Name |
|---|---|
| Repository | `Uqda/Core` |
| Daemon binary | `uqda` |
| Admin CLI binary | `uqdactl` |
| Config file | `uqda.conf` |
| Default interface name | `uqda0` |
| Linux service | `uqda.service` |
| Config directory | `/etc/uqda/` |
| State directory | `/var/lib/uqda/` |
| Runtime directory | `/run/uqda/` |
| Admin socket | `/run/uqda/admin.sock` |

Per-platform paths:

| Platform | Program | Config | Logs |
|---|---|---|---|
| Windows | `UQDA` (display name), `uqda.exe`, `uqdactl.exe`, service `UQDA` | `C:\ProgramData\UQDA\uqda.conf` | `C:\ProgramData\UQDA\logs\` |
| macOS | `uqda`, `uqdactl`, launchd service `com.uqda.core` | `/Library/Application Support/UQDA/` | `/Library/Logs/UQDA/` |
| Docker | image `ghcr.io/uqda/core`, tags `latest`, `edge`, `vX.Y.Z` | — | — |

Gateway identifiers follow `<region>-<location>-<sequence>`, e.g.
`eu-at-01`, `eu-de-01`, `us-east-01`, `asia-sg-01`; once a domain is owned,
`eu-at-01.gateway.<domain>`. Do not commit to a domain before it is actually
registered.

## Capitalization rule

- Brand name in prose: **`Uqda`** (capital U, rest lowercase). Never `UQDA`,
  `UqDa`, `uqDa`, or `U.Q.D.A`.
- Commands, binaries, paths, config keys: all lowercase — `uqda`, `uqdactl`,
  `uqda.conf`, `uqda0`.
- Arabic form: **عُقدة**.

## Versioning

Uqda Core's own version is tracked separately from the Yggdrasil wire version
it targets, so the product can iterate quickly without implying a protocol
break, e.g.:

```
Uqda Core 2.0.0
Yggdrasil compatibility: 0.5
```

## Attribution

Uqda is derived from the open-source `yggdrasil-go` implementation (LGPLv3)
and, transitively, from `Arceliar/ironwood`. Do not remove upstream copyright
notices, license files, or attribution. State plainly in documentation that
Uqda is independently maintained but derived from prior open-source work, and
distinguish Uqda-specific improvements from inherited Yggdrasil/Ironwood
functionality. Do not rename Ironwood itself or present its functionality as
an original Uqda invention — see [ARCHITECTURE.md](ARCHITECTURE.md#legal--attribution-state)
for the current (currently minimal) state of legal files in this repo.

## Resolved: hardened handshake — dropped

Earlier drafts of the Uqda concept mentioned a "hardened handshake" that two
Uqda peers could negotiate, gated by a `secure=required` option. Any such
negotiation would add a wire message or field a stock Yggdrasil peer would
not recognize — a **wire-level protocol extension**, which conflicts with
the non-negotiable rule above regardless of how gracefully it degrades.

**Decision: this feature is dropped.** Uqda Guard (see [RESTRUCTURING.md](RESTRUCTURING.md),
Phase 4) is implemented entirely through implementation-side hardening —
safer parsing, stronger validation, resource/queue limits, connection
admission control, rate limiting, and timeouts — none of which add, remove,
or alter anything on the wire. No Uqda-only handshake, mandatory extension,
or negotiated capability will be introduced. If any prototype code for the
dropped handshake is found later in the source tree, it must be removed or
fully isolated from the supported build unless independently proven to be
byte-for-byte wire-compatible with stock Yggdrasil peers.
