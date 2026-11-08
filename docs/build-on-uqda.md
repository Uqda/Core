# Build on Uqda

Uqda gives every node a permanent, encrypted IPv6 address that no one else can
use. Everything else (websites, chat, file sharing, games, tools, devices) is
for you to build. This page shows how.

The short version: **anything that listens on IPv6 already works on Uqda.**
There is no special SDK to learn and no API key to request.

## What you get as a builder

| Property | What it means for your app |
|---|---|
| A stable address per node | Your server keeps the same address wherever it moves: home, office, phone network |
| Addresses bound to keys | The source address of an incoming connection proves which node sent it. You can use it as a login |
| End-to-end encryption | Traffic between nodes is already encrypted. Your app can still add TLS |
| No port forwarding | Users behind NAT or carrier-grade NAT can host services |
| Local-first | Nodes on the same network reach each other even without internet |
| No central server | Nobody can switch off your service's network |

## 1. Host a website in one minute

On your node:

```sh
python3 -m http.server 8080 --bind ::
```

Allow it through your node's [firewall](firewall.md) in `uqda.conf`, then restart Uqda:

```hjson
Firewall: {
  Enabled: true
  AllowPing: true
  Rules: [
    { Sources: ["*"], Protocol: "tcp", Ports: ["8080"], Comment: "my website" }
  ]
}
```

Find your address with `uqda info` and share the link. IPv6 addresses go in
square brackets:

```
http://[200:1a2b:3c4d::7]:8080/
```

On Windows, also allow the port in Defender Firewall, which filters the Uqda
adapter separately. Limit it to Uqda addresses:
`New-NetFirewallRule -DisplayName "My website over Uqda" -Direction Inbound -Protocol TCP -LocalPort 8080 -RemoteAddress 200::/7 -Action Allow`
(administrator PowerShell).

Any Uqda user can open it in a normal browser. For a private group only, use
the address of each member in `Sources` instead of `"*"`.

## 2. Use the address as identity

Inside Uqda, a connection's source address is authenticated: it can only come
from the node holding that key. That makes simple, strong access control
possible without passwords.

```go
package main

import (
	"fmt"
	"net"
	"net/http"
	"net/netip"
)

var uqda = netip.MustParsePrefix("200::/7")

// Only these nodes may use the service; get each address from `uqda info`.
var allowed = map[netip.Addr]bool{
	netip.MustParseAddr("200:1a2b:3c4d::7"): true,
}

func handler(w http.ResponseWriter, r *http.Request) {
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	addr, err := netip.ParseAddr(host)
	if err != nil || !uqda.Contains(addr) || !allowed[addr] {
		http.Error(w, "this service is for known Uqda nodes only", http.StatusForbidden)
		return
	}
	fmt.Fprintf(w, "hello, %s\n", addr)
}

func main() {
	http.HandleFunc("/", handler)
	// Listen on the node's Uqda address, or on [::] to include other networks.
	_ = http.ListenAndServe("[::]:8080", nil)
}
```

Rules worth following:

- Trust the address only for connections that arrived over Uqda, from `200::/7`
  (a node) or `300::/7` (a device in that node's subnet).
- Listening on `[::]` also accepts your LAN and the internet. Bind to your
  Uqda address if the service is for Uqda only.
- The network hides content, not who is talking: do not build on Uqda for
  anonymity.

## 3. Ideas people can build

| Idea | Why Uqda fits |
|---|---|
| Personal or family websites, blogs, photo galleries | Host at home, no domain or hosting bill |
| Chat and voice servers (Matrix, XMPP, Mumble) | Members are identified by address; no open ports |
| File sync and backups (Syncthing, rsync, Nextcloud) | Direct encrypted paths between your own devices |
| Smart-home control (Home Assistant) | Reach your home without a vendor cloud |
| Game servers | Friends join by address, even behind carrier-grade NAT |
| Community services for a neighbourhood or campus | Keep working on the local network when internet access fails |
| Monitoring and admin tools | Read node state through the admin API (below) |
| Mobile apps | Embed a node with the Android and Apple SDKs |

## 4. Tools that talk to the node

Every node has a local **admin API**: JSON over a Unix socket (Linux, macOS)
or `tcp://localhost:9001` (Windows). One request per line:

```json
{"request": "getSelf"}
```

```json
{"status": "success", "response": {"address": "200:1a2b:3c4d::7", "key": "...", "subnet": "300:1a2b:3c4d::/64"}}
```

| Request | Returns |
|---|---|
| `list` | Every command the node supports |
| `getSelf` | Address, subnet, public key, version |
| `getPeers` | Peer links, traffic, latency, whether each proved its key |
| `getSessions`, `getPaths`, `getTree` | Encrypted sessions, routes and routing tree |
| `getFirewall` | Firewall rules and counters |
| `getAutoPeers` | Automatic peer selection state |
| `createInvite` | A one-time invite for another device |
| `addPeer`, `removePeer` | Change peers at runtime |

The quickest way to explore is `uqdactl -json COMMAND`. The admin API has no
authentication, so keep it local: anything that can reach it can control the
node. See [commands](commands.md) for every command and its arguments.

## 5. Apps with an embedded node

The `uqda-VERSION-android.aar` and `uqda-VERSION-apple-xcframework.zip`
release files are SDKs that run a node inside your app. Your app supplies the
virtual network interface (Android `VpnService`, Apple `NetworkExtension`). See
[mobile](mobile.md).

## 6. Make your service easy to reach

- Share the address as a link or QR code.
- Run your service on a node with healthy peers. `AutoPeers` keeps them
  ([public peers](public-peers.md#automatic-peers)).
- For services that should never sleep, run the node on an always-on machine
  such as a home server, an Umbrel or a small VPS.

## Responsible building

- Keep the [firewall](firewall.md) on and open only the ports you serve.
- Your service still needs its own security: updates, input validation and
  limits.
- Uqda is an experimental project. Do not build anything on it where a failure
  would endanger people.

Built something? Share it in the project's
[issues](https://github.com/Uqda/Core/issues/new?template=feature_request.yml)
so others can find it.
