# Security model

Uqda provides encrypted sessions on the Yggdrasil network. It does not provide
anonymity or traffic-analysis resistance, and does not secure applications exposed
on the resulting IPv6 interface. Use host firewalls and application authentication.
No independent external security audit is claimed.

| Boundary | Controls | Remaining exposure |
|---|---|---|
| Direct-peer handshake (`src/core/version.go`, `link.go`) | Bounded metadata fields, length checks, full reads and handshake deadline | No global cap on simultaneous handshakes |
| Established peers (Ironwood) | Peer liveness deadlines and keepalives | Liveness does not prevent low-throughput connection hoarding |
| Local discovery (`src/multicast/advertisement.go`) | Header and hash bounds checks | A valid advertisement is not evidence of a trusted operator |
| Decrypted IPv6 (`src/ipv6rwc`) | Address/header processing before delivery to TUN | Host network stack and exposed services remain attack surfaces |
| Admin API (`src/admin`) | Local defaults, Unix socket permissions, warning on non-loopback TCP | No authentication; local TCP is accessible to other local users |
| Identity (`src/config`) | Unix permission warnings; Ansible vault files use 0600 | Warnings do not repair permissions; Windows ACLs require operator review |
| Dependencies | Pinned module versions and vulnerability scanning | Scan results age and do not establish absence of vulnerabilities |
| Logs and diagnostics | Admin response types expose public identifiers and connection data | Logs/configuration exports need careful handling; no exhaustive logging audit is claimed |

`FuzzVersionMetadataDecode` and `FuzzMulticastAdvertisementUnmarshalBinary`
exercise the two local binary parsers. See [testing](testing.md) for commands.
The IPv6 payload surface does not have equivalent fuzz coverage in this tree.

Ironwood manages read deadlines on established links. Adding a second deadline
manager on the same `net.Conn` can overwrite its deadlines. Configure and test
liveness through the dependency's own options instead.

A relay sees connection metadata and traffic patterns but cannot decrypt other
nodes' end-to-end session payloads. `AllowedPublicKeys` does not restrict multicast
peerings; disable discovery on untrusted interfaces. `GroupPassword` limits
session communication to nodes with the same secret, not transport peering.

Protect configuration backups and external PEM keys as carefully as active keys.
A stolen private key permits node impersonation. Keep the admin API local and
limit its socket group. There is no built-in updater; obtain packages and source
from trusted channels. Report vulnerabilities according to [SECURITY](../SECURITY.md).
