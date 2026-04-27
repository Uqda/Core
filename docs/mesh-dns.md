# Unified DNS for your Uqda network

**Uqda Core does not implement DNS.** DNS is a separate service you run (or rent) like on any IP network. This page describes how to build a **single, community-wide DNS view** so members can use **names** instead of raw `200:…` addresses for sites and services on the overlay.

Nothing here requires changing Uqda source code — only **your** DNS software, firewalls, and client settings.

---

## 1. What you are building

| Piece | Role |
|--------|------|
| **One or two DNS servers** | Processes listening on **`200:…:53`** (UDP and TCP) reachable from every mesh member who should use them. |
| **A private zone** | Names you control, e.g. **`svc.yourmesh.internal.`** → AAAA records pointing at overlay IPv6 addresses. |
| **Recursive resolution (optional but usual)** | For names **outside** your zone (e.g. `github.com`), the same server **forwards** to upstream resolvers (`1.1.1.1`, `8.8.8.8`, …) or your org policy. |
| **Clients** | Every phone/laptop/server sets **this resolver** as its DNS (manually, via DHCP/RDNSS on a LAN gateway, or MDM). |

Result: **one unified namespace** for your mesh, independent of the public Internet’s view of those names.

---

## 2. Choose a zone name (namespace)

- Use a suffix **you own and agree on** inside the community, e.g. **`yourmesh.internal.`**  
  The **`.internal`** DNS zone is intended for **private application** use (not the public DNS root). Pick a **unique** second-level name for your group so you do not collide with another private network. Alternatively use a subdomain of a domain you already own, e.g. **`mesh.example.com`**, only served inside the mesh.
- **Avoid `*.local`** for unicast DNS: **`*.local`** is heavily associated with **multicast DNS (mDNS)**; mixing unicast DNS on `.local` causes confusing failures on many operating systems.
- If you already own **`example.com`**, you can instead use **`mesh.example.com`** on your **authoritative** servers and only answer that zone from resolvers **inside** the mesh (split horizon). Public visitors then do not see mesh-only addresses unless you want them to.

---

## 3. Where to run the server

1. Pick a node with a **stable** Uqda identity (long-lived key / address) and good connectivity — often the same machine as a “bootstrap” peer or a small VPS-style node **on the mesh** (still a `200:` address).
2. Install **CoreDNS**, **BIND**, **Unbound**, **dnsmasq**, or similar on that host.
3. Configure the daemon to **listen** on the host’s **Uqda IPv6** (see `uqdactl getSelf`), e.g. bind to that address only:

   ```text
   listen-on-v6 { 200:1111:2222:3333:4444:5555:6666:7777; };
   ```

   (Syntax varies by software — equivalent in CoreDNS is `bind 200:1111:…` in the server block.)

4. Open the OS firewall for **`200::/7` → UDP/TCP 53** to that host **only** if you want the whole mesh to query it; tighten the source if your community is smaller.

---

## 4. Recommended pattern: recursive resolver + stub zone

Most communities want **one IP** that answers everything:

- **Known mesh names** → your **AAAA** records (and optional **A** if you dual-stack elsewhere).
- **Everything else** → **forward** to trusted upstream resolvers (or block, if this is a locked-down lab).

**Small mesh — `dnsmasq`** (single config file, easy to start):

```text
# Listen on Uqda IPv6 only (example)
bind-interfaces
listen-address=200:1111:2222:3333:4444:5555:6666:7777

# Static names → overlay addresses
address=/www.yourmesh.internal/200:aaaa:bbbb:cccc:dddd:eeee:ffff:1111
address=/git.yourmesh.internal/200:aaaa:bbbb:cccc:dddd:eeee:ffff:2222

# Everything else → upstream resolvers
server=1.1.1.1
server=8.8.8.8
```

**Larger mesh — CoreDNS / BIND / Unbound:** use one **server block or view** for **`yourmesh.internal`** (file or `auto` plugin) and a **separate forwarder/recursion** configuration for the root zone. Follow that software’s “split DNS” or “stub zone” documentation; bind the daemon to your **`200:`** address the same way.

**Zone file** (BIND / CoreDNS `file` plugin — minimal example):

```dns
$ORIGIN yourmesh.internal.
@   3600 IN SOA ns1.yourmesh.internal. hostmaster.yourmesh.internal. (
            1 ; serial
            3600 ; refresh
            1800 ; retry
            604800 ; expire
            3600 ; minimum
        )
    3600 IN NS ns1.yourmesh.internal.

ns1     3600 IN AAAA 200:1111:2222:3333:4444:5555:6666:7777
www     3600 IN AAAA 200:aaaa:bbbb:cccc:dddd:eeee:ffff:1111
git     3600 IN AAAA 200:aaaa:bbbb:cccc:dddd:eeee:ffff:2222
```

Replace addresses with real overlay IPs from **`uqdactl getSelf`** on each service host.

**Unbound** users often combine **`stub-zone:`** for `yourmesh.internal` with **`forward-zone:`** for `.` — same idea.

---

## 5. Secondary / high availability (optional)

- Run a **second** DNS VM on another `200:` address.
- **Zone transfer (AXFR/IXFR)** from primary to secondary if you use classic BIND-style zones, or use a **git-backed** CoreDNS file synced to both, or a small database — pick one operational model.
- Put **both** resolver addresses in client config (first preferred, second fallback).

---

## 6. Point clients at the unified resolver

### Linux (systemd-resolved)

Drop-in, e.g. **`/etc/systemd/resolved.conf.d/uqda-dns.conf`**:

```ini
[Resolve]
DNS=200:1111:2222:3333:4444:5555:6666:7777
Domains=~yourmesh.internal
```

Then `systemctl restart systemd-resolved`.  
`Domains=~` makes **`*.yourmesh.internal`** use this resolver; other names still use default routing policy (adjust to taste).

### Linux (static `resolv.conf`)

```text
nameserver 200:1111:2222:3333:4444:5555:6666:7777
```

Some distros overwrite **`/etc/resolv.conf`**; use **resolved**, **NetworkManager**, or **netplan** DNS settings instead of fighting the OS.

### Windows

**Settings → Network → (adapter) → DNS server assignment → Manual** → IPv6 DNS = your resolver’s `200:…`.

### LAN behind an Uqda gateway

If you already use **radvd** for prefix advertisement ([Configuration](configuration.md#advertising-a-prefix)), add **RDNSS** (Recursive DNS Server option) in **`radvd.conf`** so LAN devices auto-learn the resolver:

```
RDNSS 200:1111:2222:3333:4444:5555:6666:7777 { };
```

(Exact syntax depends on radvd version — see `man radvd.conf`.)

---

## 7. Security and abuse prevention

| Risk | Mitigation |
|------|------------|
| **Open resolver on clearnet** | Bind **only** to the **Uqda** address; firewall so **only** `200::/7` (or your member list) can reach **:53**. |
| **DNS amplification** | Do not forward to the Internet from untrusted sources; rate-limit if you must expose broader than the mesh. |
| **Poisoning / integrity** | For high trust, consider **DNSSEC signing** your zone (operational overhead) or internal **DoT** (DNS-over-TLS) between stub and resolver — advanced. |

---

## 8. HTTPS with your unified names

Once **`www.yourmesh.internal`** resolves to a **`200:`** address, you can issue **private CA** certificates for that hostname, or use **HTTP** on the mesh according to your threat model. Public CAs still will not sign arbitrary **`*.internal`** for the public web PKI — see [Hosting a website on Uqda](hosting-on-uqda.md).

---

## 9. Checklist

- [ ] Stable **`200:`** for the DNS host; **`uqda`** running and peered.  
- [ ] DNS software listens on **UDP+TCP 53** on that address.  
- [ ] Zone records for all **mesh services** you want named.  
- [ ] Forwarding policy for **non-mesh** names agreed in the community.  
- [ ] **Firewall** rules on the DNS host.  
- [ ] **Client** configuration (manual, resolved, NM, radvd RDNSS, MDM).  
- [ ] Document the **official resolver IP(s)** and **zone suffix** for your members.

---

## See also

- [Hosting a website on Uqda](hosting-on-uqda.md) — binding web servers and URLs with IPv6 literals.  
- [FAQ](faq.md) — firewalls and reachability.  
- [Network capabilities & full config guide](network-services-and-config.md) — Uqda config (unchanged by DNS).  
- [Documentation hub](README.md).
