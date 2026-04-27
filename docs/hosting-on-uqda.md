# Hosting a website on the Uqda network

Uqda does **not** include a built-in web server or “hosting panel”. It gives you a **normal IPv6 presence** on the overlay (`200::/7`). Anything that can **bind to an IPv6 address** and speak **HTTP** (or TLS) can be your site — the same way you would on the public Internet, with a few overlay-specific details below.

---

## 1. Checklist before you expose a service

| Step | Why |
|------|-----|
| **`uqda` is running** with a stable **`uqda.conf`** | Without the daemon, the overlay address does not exist. |
| **You have peerings** so visitors can reach you | Remote users must have **their own path** through the mesh to your node (multi-hop is fine if routes exist). |
| **You know your Uqda IPv6** | You will bind the server or firewall rules to it. |
| **OS firewall rules are intentional** | **Every** node on the mesh can try to open TCP to your overlay address ([FAQ](faq.md)). Only open ports you need. |

---

## 2. Find your overlay address

On the machine that will host the site:

```sh
uqdactl getSelf
```

Or (no running daemon needed if you only have the config file):

```sh
uqda -useconffile /path/to/uqda.conf -address
```

Use the **`200:…`** address shown as the host identity. If you advertise a **routed `/64`** to a LAN (see [Configuration — advertising a prefix](configuration.md#advertising-a-prefix)), clients on that LAN get addresses under your **`300:…::/64`**; the web server might listen there instead — same principles.

---

## 3. Run a normal web server

### Option A — Bind only to your Uqda address (recommended default)

Restricts the site to the overlay (not your LAN’s global IPv6 unless you duplicate listeners).

**Caddy** example (`Caddyfile`):

```caddyfile
http://[200:1111:2222:3333:4444:5555:6666:7777] {
    root * /var/www/uqda-site
    file_server
}
```

Replace the address with yours. For **HTTPS** on that literal IPv6, you typically use **internal TLS** (see section 6) — public CAs do not issue certificates for arbitrary overlay addresses.

**nginx** example (`server` block):

```nginx
server {
    listen [200:1111:2222:3333:4444:5555:6666:7777]:80;
    root /var/www/uqda-site;
    index index.html;
}
```

### Option B — Listen on all interfaces (`[::]:80`)

```nginx
listen [::]:80;
```

This also listens on **other** IPv6 addresses on the machine (e.g. public WAN). Prefer **Option A** unless you understand the exposure.

### Option C — Application stack (Node, Python, etc.)

Bind the app to your **`200:…`** address and a port, e.g. `200:…:8080`, and optionally put nginx/Caddy in front on `:80`.

---

## 4. Firewall (do not skip)

Examples — adapt interface names and addresses.

**Linux (nftables)** — allow HTTP from overlay only:

```nft
table inet filter {
  chain input {
    type filter hook input priority filter; policy drop;
    ip6 saddr 200::/7 tcp dport { 80, 443 } accept
    # … keep rules for ssh, lo, established, etc.
  }
}
```

**Linux (ufw)** — UFW can be awkward with source ranges; many operators use **nftables** or **iptables** for fine IPv6 rules.

**Windows** — Advanced Security → Inbound Rules → IPv6, scope **remote IP** `200::/7`, local port 80/443.

Tighten further if you only want **specific** remote keys or subnets — that is **not** what `AllowedPublicKeys` does; it filters **peering**, not HTTP ([Configuration reference](configuration-reference.md)).

---

## 5. How visitors open your site

- **By raw address:** `http://[200:1111:2222:3333:4444:5555:6666:7777]/`  
  Brackets are required in URLs for literal IPv6.

- **By name:** Uqda has **no global DNS** for `200::/7`. Practical options:
  - **`/etc/hosts`** (or Windows `hosts`) on each client: map a fake name to your `200:…` address.
  - Your own **DNS server** for the whole community — see **[Unified DNS for your mesh](mesh-dns.md)** (resolver on a `200:` address, private zone, client settings).
  - A **public DNS name** pointing at your overlay address only works if resolvers and clients can **route** to `200::/7` (unusual on the clearnet; typical on mesh-only resolvers).

---

## 6. HTTPS (TLS) on the overlay

- **Let’s Encrypt / public CA:** almost never issues for **random-looking `200:`** addresses you cannot prove control of in DNS the way CAs expect.
- **Practical approaches:**
  - **HTTP only** on the mesh for static content (acceptable only if you accept cleartext on the overlay — overlay hop encryption still applies between Uqda nodes, but **HTTP payload** is visible to your server process as usual).
  - **Self-signed certificate** for `IP` SAN or a private hostname; visitors trust your CA once.
  - **Caddy / nginx with your own CA** for a mesh-only domain name.

---

## 7. Common problems

| Symptom | Things to check |
|---------|------------------|
| **Connection time out** | Visitor has **no route** (no peering path to you). Add/repair **Peers** / **Listen** / public peer lists. |
| **Refused** | Web server not running, wrong **bind** address, or **firewall** blocking. |
| **Works locally, not remotely** | Server bound to **`127.0.0.1` / `::1`** only — bind to **`200:…`** or `::` with firewall. |
| **Large uploads fail** | Check **MTU** along the path (`IfMTU` in config, intermediate tunnels). |

---

## 8. Hosting for devices behind your Uqda router

If this machine is a **gateway** that advertises your **`300:…::/64`** on a LAN, you can run the web server on a **LAN host** with a `300:…` address, or port-forward from the gateway’s Uqda address using **nftables** / **socat**. Design depends on your topology; start from [Advertising a prefix](configuration.md#advertising-a-prefix).

---

## 9. Security reminders

- Do **not** expose **`uqdactl` / AdminListen** to the mesh ([FAQ](faq.md)).  
- Keep the web server and OS **patched**; overlay reachability is wide by default.  
- Separate **high‑risk** apps from production using VMs or containers if needed.

---

## See also

- [FAQ](faq.md) — firewall, anonymity, address range.  
- [Network capabilities & full config guide](network-services-and-config.md) — all config surfaces.  
- [Private two-node](private-two-nodes.md) — if only a closed group should reach you at the **peering** layer (still use a firewall for HTTP).  
- [Documentation hub](README.md).
