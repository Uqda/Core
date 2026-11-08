# Security model

Uqda provides encrypted sessions on the Yggdrasil network. It does not provide
anonymity or traffic-analysis resistance, and does not secure applications exposed
on the resulting IPv6 interface. Use host firewalls and application authentication.
No independent external security audit is claimed.

| Boundary | Controls | Remaining exposure |
|---|---|---|
| Direct-peer handshake (`src/core/version.go`, `link.go`, `link_limits.go`) | Bounded metadata fields, length checks, full reads, handshake deadline; at most 256 pending inbound handshakes per node and 8 per source (IPv4 address or IPv6 /64), excess connections closed immediately | Distributed floods from many sources can still occupy the global slots and delay new legitimate peerings; the handshake signature is not bound to a fresh challenge (see below) |
| Established peers (Ironwood) | Peer liveness deadlines and keepalives | Liveness does not prevent low-throughput connection hoarding |
| Local discovery (`src/multicast/advertisement.go`) | Header and hash bounds checks | A valid advertisement is not evidence of a trusted operator |
| Decrypted IPv6 (`src/ipv6rwc`, `src/firewall`) | Address/header processing before delivery to TUN; stateful inbound firewall (default deny on new configurations) with rules by authenticated key, address or /64, protocol and port; fuzzed parser | Allowed services remain attack surfaces; existing configurations keep the firewall off until enabled. See [firewall](firewall.md) |
| Admin API (`src/admin`) | Local defaults, Unix socket permissions, warning on non-loopback TCP; at most 64 concurrent connections, 1 MiB per request and a five-minute idle timeout | No authentication; local TCP is accessible to other local users |
| Identity (`src/config`) | Unix permission warnings; Ansible vault files use 0600 | Warnings do not repair permissions; Windows ACLs require operator review |
| Dependencies | Pinned module versions and vulnerability scanning | Scan results age and do not establish absence of vulnerabilities |
| Logs and diagnostics | Admin response types expose public identifiers and connection data | Logs/configuration exports need careful handling; no exhaustive logging audit is claimed |

`FuzzVersionMetadataDecode`, `FuzzMulticastAdvertisementUnmarshalBinary` and the
firewall's `FuzzParse` exercise the local binary parsers. See [testing](testing.md) for commands.
The firewall parser covers IPv6 and extension headers; the rest of the decrypted
IPv6 path (`src/ipv6rwc`) does not yet have its own fuzz target.

Ironwood manages read deadlines on established links. Adding a second deadline
manager on the same `net.Conn` can overwrite its deadlines. Configure and test
liveness through the dependency's own options instead.

A relay sees connection metadata and traffic patterns but cannot decrypt other
nodes' end-to-end session payloads. `AllowedPublicKeys` does not restrict multicast
peerings; disable discovery on untrusted interfaces. `GroupPassword` limits
session communication to nodes with the same secret, not transport peering.
In 26.0.2, that secret is derived with Argon2id at startup, using the RFC 9106
memory-constrained cost profile (64 MiB, three passes, four lanes). This is not a
PAKE or a guarantee against offline guessing: use a random strong password. The
shared protocol salt is deterministic, not unique to each group. Older private
groups must migrate all members together; no weak legacy fallback is accepted.

The Yggdrasil handshake signs only the node's public key (keyed by the link
password), not a fresh per-connection challenge, so a recorded handshake can be
replayed. Uqda adds a **fresh handshake**: each Uqda node sends a random 32-byte
nonce in an extra metadata field, and when both sides sent one, each signs the
other's nonce together with both keys before the link is handed to Ironwood. A
replayed handshake cannot produce that signature and is rejected, which also
protects `AllowedPublicKeys` and pinned `?key=` peerings. `uqdactl getPeers -json`
reports `fresh_handshake` for each link.

Yggdrasil nodes and older Uqda releases do not send the nonce and are still
accepted, because compatibility is a requirement. An attacker can strip the
field from a recorded handshake to look like such a node. In networks where
every node runs this Uqda release or later, set `RequireFreshHandshake: true`
to reject them and close that downgrade. Elsewhere, prefer `tls://` or `quic://`
between known peers and a per-link `password` for restricted peerings.

Protect configuration backups and external PEM keys as carefully as active keys.
A stolen private key permits node impersonation. Keep the admin API local and
limit its socket group. There is no built-in updater; obtain packages and source
from trusted channels. Report vulnerabilities according to [SECURITY](../SECURITY.md).
