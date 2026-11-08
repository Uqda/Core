# What Yggdrasil users ask for, and where Uqda stands

Reviewed 2026-11-08: all 169 open issues in
[yggdrasil-network/yggdrasil-go](https://github.com/yggdrasil-network/yggdrasil-go/issues),
ranked by reactions plus comments. This page records what Uqda took from them,
what it built, and why the rest waits. Uqda must stay wire-compatible with
Yggdrasil, so changes that need a new routing protocol are left to upstream.

## By topic

| Topic | Open issues | Summary |
|---|---|---|
| Routing and performance | 28 | Path choice, latency versus bandwidth, MTU, root selection |
| Packaging | 18 | winget, Chocolatey, RPM, OpenWrt, Debian package quality |
| Windows | 12 | Service crashes, DNS, adapter behaviour |
| Censorship and transports | 9 | Tor, SOCKS, Bluetooth and radio |
| DNS and naming | 8 | Names instead of addresses, DNS servers for the interface |
| Security | 7 | Firewall, key brute force, TLS certificates |
| Mobile | 3 | iOS and Android apps |
| Other | 84 | Questions, discussions, small bugs |

## Built in Uqda

| Upstream issue | Request | In Uqda |
|---|---|---|
| [#1244](https://github.com/yggdrasil-network/yggdrasil-go/issues/1244), the most requested | A built-in firewall, because services are exposed by default | Stateful inbound [firewall](../firewall.md), on for new nodes, with rules by key, address, port |
| [#1083](https://github.com/yggdrasil-network/yggdrasil-go/issues/1083) | Read peer passwords from a file | `?password_file=PATH` on peers and listeners, plus `GroupPasswordFile` |
| [#172](https://github.com/yggdrasil-network/yggdrasil-go/issues/172) | Limit peerings on public nodes | `MaxInboundPeers`. Outbound, local and trusted peers are never refused |
| [#1173](https://github.com/yggdrasil-network/yggdrasil-go/issues/1173), [#1192](https://github.com/yggdrasil-network/yggdrasil-go/issues/1192) | Multicast errors logged every second | First occurrence is a warning; identical repeats go to debug for 10 minutes |
| [#1357](https://github.com/yggdrasil-network/yggdrasil-go/issues/1357) | `ctl` tables overflow on small screens | Tables fit the terminal width (or `COLUMNS`); piped output is unchanged |
| [#1354](https://github.com/yggdrasil-network/yggdrasil-go/issues/1354) | A distinctive TUN name instead of `tun0` | `IfName: auto` gives `uqda0`, `uqda1` on Linux; Windows uses `Uqda` |
| [#885](https://github.com/yggdrasil-network/yggdrasil-go/issues/885), [#779](https://github.com/yggdrasil-network/yggdrasil-go/issues/779) | Stronger keys against address collisions | `uqda -genconf -strength=N`; `setup` and `join` create strength-8 identities |
| [#417](https://github.com/yggdrasil-network/yggdrasil-go/issues/417), [#1067](https://github.com/yggdrasil-network/yggdrasil-go/issues/1067) | Confusing permission errors, setup under a service user | Commands explain when to use `sudo`; `setup` and `join` keep the service user's file ownership |
| [#719](https://github.com/yggdrasil-network/yggdrasil-go/issues/719) | What is the threat model? | [Security model](../security.md) and [networks](../networks.md) |
| Fake TLS and handshake trust, [#1333](https://github.com/yggdrasil-network/yggdrasil-go/issues/1333) | Peers could be impersonated | The [fresh handshake](../security.md) proves the key on every connection; invites use TLS 1.3 pinned to the node key |

## Planned (compatible, not yet built)

| Upstream issue | Request | Plan |
|---|---|---|
| [#1125](https://github.com/yggdrasil-network/yggdrasil-go/issues/1125) | DNS servers for the interface, as WireGuard does | Optional `DNS` setting applied per platform when the interface comes up |
| [#713](https://github.com/yggdrasil-network/yggdrasil-go/issues/713), [#611](https://github.com/yggdrasil-network/yggdrasil-go/issues/611), [#1128](https://github.com/yggdrasil-network/yggdrasil-go/issues/1128), [#1010](https://github.com/yggdrasil-network/yggdrasil-go/issues/1010) | winget, Chocolatey, RPM, OpenWrt packages | Manifests from the existing MSI, RPM and portable builds |
| [#1106](https://github.com/yggdrasil-network/yggdrasil-go/issues/1106) | Names instead of IPv6 addresses | `.uqda` names bound to keys; see the [roadmap](../roadmap.md) |
| [#881](https://github.com/yggdrasil-network/yggdrasil-go/issues/881), [#1407](https://github.com/yggdrasil-network/yggdrasil-go/issues/1407) | Use a pre-created TUN, control the prefix route | Options for container and rootless setups on Linux |
| [#1025](https://github.com/yggdrasil-network/yggdrasil-go/issues/1025) | iOS and Android apps | Apps on top of the existing SDKs; see the [roadmap](../roadmap.md) |

## Left to upstream

These need changes to the routing protocol or Ironwood. Doing them only in
Uqda would split the network.

- Path choice and bandwidth-aware routing:
  [#1222](https://github.com/yggdrasil-network/yggdrasil-go/issues/1222),
  [#1022](https://github.com/yggdrasil-network/yggdrasil-go/issues/1022),
  [#998](https://github.com/yggdrasil-network/yggdrasil-go/issues/998).
- Link telemetry with new wire messages:
  [#1403](https://github.com/yggdrasil-network/yggdrasil-go/issues/1403), planned upstream for v0.6.
- Root selection behaviour: [#1339](https://github.com/yggdrasil-network/yggdrasil-go/issues/1339).
- Performance regressions inside Ironwood:
  [#978](https://github.com/yggdrasil-network/yggdrasil-go/issues/978),
  [#1167](https://github.com/yggdrasil-network/yggdrasil-go/issues/1167).

## Not goals

- Anonymity or censorship circumvention as a guarantee
  ([#1265](https://github.com/yggdrasil-network/yggdrasil-go/issues/1265),
  [#1291](https://github.com/yggdrasil-network/yggdrasil-go/issues/1291)). Uqda does not hide who you
  are; see [networks](../networks.md#what-neither-mode-gives-you).
- A separate address space or IANA registration
  ([#1130](https://github.com/yggdrasil-network/yggdrasil-go/issues/1130)): Uqda shares Yggdrasil's.
