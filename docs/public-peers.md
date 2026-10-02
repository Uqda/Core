# Public peer suggestions (26.0.4)

This feature requires 26.0.4 or newer; it is not present in 26.0.3 binaries.
It uses [the public peer status site](https://publicpeers.neilalexander.dev/),
whose endpoints come from [Yggdrasil's public-peers repository](https://github.com/yggdrasil-network/public-peers).

```sh
uqda find
uqda find country=germany
uqda find country=austria limit=2 --json
uqda help find
```

No daemon or administrator privileges are needed. Bare `find` fetches the catalog
and lists countries with online TLS candidates; it does not contact those peers
or query a geolocation service. Choose a nearby country using its catalog name
(for example `united-kingdom`). Country names are case-insensitive.

With a country, the command selects at most eight online TLS candidates,
prioritizing the site's reported seven-day uptime and avoiding duplicate
hostnames or pinned keys. It checks TCP port reachability, then suggests up to
three candidates by measured TCP connection time. `limit=1..5` controls output,
not the probing budget. The site updates hourly; snapshots older than six hours
or more than five minutes in the future are refused. Keep your system clock set.

## What a result does and does not prove

A successful check means a TCP connection was accepted at a published TLS
endpoint from this computer, at this time. It is **not** a TLS handshake, public-key
verification, authenticated Uqda session, overlay ping, throughput measurement,
or guarantee that an operator is trustworthy. Site uptime and country labels are
third-party claims. QUIC, WebSocket, plain TCP and offline entries are not suggested.
The measured time excludes DNS resolution and must not be called ping RTT.

The command does not add a peer, restart a service, read configuration, or change
`GroupPassword` or allowlists. Public peers will not normally authenticate into
your private group. Do not disable private-group protection to use a suggestion.

For a node intentionally participating in the public network, connecting remains
an explicit action, separate from discovery:

```sh
uqdactl addPeer uri="TLS_URI_FROM_THE_RESULT"
uqda peers
uqda test REMOTE_UQDA_IPV6
```

Use `sudo` on macOS/Linux only if your admin socket requires it. Keep any `key`
and `sni` parameters in the suggested URI. `addPeer` changes runtime state only;
for persistence, add the URI to `Peers` in your existing protected configuration.
Never overwrite the configuration or regenerate the node's identity.

## Bounds, privacy and failures

The source URL is fixed HTTPS with normal certificate verification. Cross-origin
and non-HTTPS redirects are refused. HTML is capped at 1 MiB, and timestamps,
country labels, URI schemes, ports and query parameters are checked. Credentialed
URIs and password parameters are excluded from both text and JSON suggestions.
No arbitrary-source, arbitrary-target, auto-connect or configuration option exists.

Candidate DNS results must all be public addresses. Loopback, private, link-local,
metadata, multicast, documentation and selected reserved/transition ranges are
rejected. Connections use validated IP literals rather than re-resolving hostnames.
At most three candidates are checked concurrently, each with a three-second
deadline and at most two IP attempts (up to sixteen TCP attempts total). The fetch
has a ten-second timeout; the whole command has a twenty-five-second deadline.
Public operators see your source IP during these short connection attempts.

Exit codes: `0` for countries or at least one reachable suggestion, `1` for a
source/probe failure or no reachable suggestions, `2` for invalid arguments.
An empty successful catalog probe prints JSON with an empty `suggestions` array
but returns `1`. No raw transport error or URI credential is printed.
