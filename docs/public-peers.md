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

## Automatic peers

`uqda find` suggests peers once. `AutoPeers` keeps doing it inside the daemon,
so a node stays connected when public peers come and go:

```hjson
AutoPeers: {
  Enabled: true
  Country: germany   # a catalog country; run uqda find for the list
  Count: 2           # wanted healthy outbound peers, 1 to 5
  Interval: 30m      # how often to re-measure, 10m to 24h
}
```

| Situation | What Uqda does |
|---|---|
| Fewer than `Count` outbound peers are up | Adds the fastest reachable public TLS peers in `Country` |
| A peer goes down | Within 2 seconds, adds the next peer from a standby list measured in advance; no catalog fetch is needed. The lost peer is still retried and removed after five minutes down |
| A new peer does not come up within 20 seconds | Removes it, tries the next candidate and skips the failed one for ten minutes |
| More peers are up than `Count` (for example a lost peer came back) | After 30 seconds of stable links, removes the one with the slowest measured round trip |
| A free candidate is much faster (at least 1.5 times plus 30 ms) than its slowest peer | Swaps that one peer, adding the new one before removing the old, at most one per check |
| The host's network changes (Wi-Fi reconnects, cable, new address) | Redials every peer at once instead of waiting out its reconnect backoff, and re-measures candidates |
| The catalog cannot be reached | Keeps the current peers and the standby list, and warns once |
| A peer you listed in `Peers` | Never removes it; it counts towards `Count` while it is up |

### What happens when a link fails

Routing does not wait for any of this: traffic moves to the node's other
peers on its own. With `Count: 2` or more, a node always has a second path
while the lost peer is replaced.

Measured with four nodes, two paths and 100 packets a second, cutting the
path in use (`TestFailoverAroundLostRelay`, run by CI on every change in its own
job; locally with `UQDA_TEST_FAILOVER=1 go test ./src/core -run Failover -v`):

| How the link fails | Traffic stops for | Packets lost |
|---|---|---|
| The connection closes (peer shut down, cable to it unplugged) | about 10 ms in most cases | none |
| The same, when the lost link was the node's parent in the routing tree | 3 to 4 seconds while the tree re-forms | those seconds of traffic |
| The link goes silent (router or Wi-Fi died, no close) | about 3 seconds, the dead-link timeout | those seconds of traffic |

The first row comes from looking a destination up again within 250 ms of a
broken path; before 26.1.1 it took about 2.7 seconds in every case. The
other two are limits of the routing protocol shared with Yggdrasil: the
dead-link timeout is kept at three seconds so slow mobile links are not
dropped by mistake. Packets lost in that window are resent by TCP and other
reliable protocols, as on any IP network.

Inbound connections and local multicast peers do not count, because they are
not a path to the rest of the network. `uqda setup country=NAME` writes two
peers as a starting point and turns `AutoPeers` on. Check it with
`uqdactl getAutoPeers -json` or `uqda status`.

Privacy: the catalog operator sees your IP address each time it is fetched
(at start, every `Interval`, after a network change, and to refill the standby
list after a peer was replaced), and measured peers see TCP connection
attempts. Choosing a public peer does not change who can reach you; the
[firewall](firewall.md) and private groups still decide that.
