# Private network between exactly two people

This guide is for **two Uqda nodes** that should **only peer with each other** — not with the public mesh, not with random neighbours on the same Wi‑Fi, and not via multicast discovery unless you deliberately configure it.

Uqda traffic between the two of you is **encrypted end‑to‑end** for overlay IPv6. This is **not** the same as hiding your real IP addresses from each other: you still connect over normal TLS/TCP/QUIC between your hosts’ addresses.

---

## What you configure (mental model)

| Piece | Role |
|--------|------|
| **`Peers`** | “I open an outbound connection **to** this URI (the other person’s listener).” |
| **`Listen`** | “I accept **inbound** peerings on this URI (the other person puts this in their **`Peers`**).” |
| **`AllowedPublicKeys`** | “Only accept **incoming** peerings if the remote node’s **public key** is in this list.” |
| **`MulticastInterfaces`** | Link‑local discovery on LAN. **Turn it off** for a strict two‑node setup (see below). |

Important limitation from the config reference: **`AllowedPublicKeys` does not restrict link‑local peers discovered via multicast.** So for “only us two”, you should **disable multicast** (empty list) **or** use a setup where multicast cannot reach anyone else (rare in practice). Safer default: **`MulticastInterfaces: []`**.

---

## Step 0 — Decide who listens (NAT / firewall)

- **Best case:** one of you has a **public IPv4 or IPv6** (or a stable port forward) and can open a **TCP port** in the OS firewall for Uqda’s **`Listen`**.
- **Both behind home NAT** with **no** port forwarding on either side: a **direct** two‑node link over the Internet usually **will not work** without changing the network (forward a port on at least one router, use a VPS with a public IP as *your* listener — that third host is only transport, not “on the Uqda mesh”, or use another VPN underlay). **Same LAN** (same Wi‑Fi / switch) is easier: use **link‑local** or **private LAN** addresses in **`Peers`** / **`Listen`**.

---

## Step 1 — Each generates a config (keep keys secret)

On **both** machines:

```sh
uqda -genconf > uqda.conf
chmod 600 uqda.conf
```

Never share **`PrivateKey`** or the full file if it contains it. You **will** share **public keys** in the next step.

---

## Step 2 — Exchange **public keys** (hex)

**Alice** runs:

```sh
uqda -useconffile uqda.conf -publickey
```

**Bob** runs the same on his file. They send each other the printed hex string **out of band** (signal, in person, etc.).

Each person adds the other’s key to **`AllowedPublicKeys`** in their own `uqda.conf` (array of strings, hex only, no `0x` prefix unless your file format already uses that style — match **`uqda -genconf`** output style).

Example snippet (replace with the real 64‑hex‑character key from **`uqda -publickey`**):

```hjson
AllowedPublicKeys: [
  "<paste_the_other_persons_public_key_hex_here>"
]
```

Use **one** entry per remote person you allow to complete an **inbound** peering to you.

---

## Step 3 — Disable multicast (recommended for “only us”)

In **both** configs set:

```hjson
MulticastInterfaces: []
```

With an empty list, the multicast module does not start (`src/multicast/multicast.go`: no beacon/listen enabled).

---

## Step 4 — Listener + peer (typical: one listener, one dialer)

Assume **Alice** can receive inbound TCP on port **12345** (example).

**Alice** adds to `uqda.conf`:

```hjson
Listen: [
  "tls://0.0.0.0:12345"
]
Peers: []
```

If you prefer IPv6‑only listening:

```hjson
Listen: [
  "tls://[::]:12345"
]
```

**Bob** does **not** need `Listen` for the minimum setup; he only dials Alice. He sets (replace with Alice’s **reachable** IP or DNS name):

```hjson
Listen: []
Peers: [
  "tls://ALICE_PUBLIC_IP_OR_DNS:12345"
]
```

Optional shared secret on the **link** (in addition to normal TLS); **both** the listener URI and the peer URI must use the **same** `password=` value:

**Alice:**

```hjson
Listen: [
  "tls://0.0.0.0:12345?password=YOUR_LONG_RANDOM_SHARED_SECRET"
]
```

**Bob:**

```hjson
Peers: [
  "tls://ALICE_PUBLIC_IP_OR_DNS:12345?password=YOUR_LONG_RANDOM_SHARED_SECRET"
]
```

See peer URI examples in [Configuration reference](configuration-reference.md).

**Firewall:** on Alice’s host (and router if applicable), allow **inbound TCP 12345** (or whatever you chose).

---

## Step 5 — Symmetric setup (both can dial each other)

If **both** have public listeners, each sets **`Listen`** on a chosen port **and** **`Peers`** to the other’s URI. Still keep **`AllowedPublicKeys`** mutual and **`MulticastInterfaces: []`** if you want no LAN discovery.

---

## Step 6 — Same LAN (no Internet path)

If both are on the **same layer‑2 network**, you can often point **`Peers`** at the other laptop’s **global or ULA IPv6** or **link‑local** address. Link‑local URIs need a **zone id** (interface name) on Unix, e.g.:

```text
tls://[fe80::xxxx:xxxx:xxxx:xxxx%en0]:12345
```

(`en0` is an example; use `ip -6 addr` / `ifconfig` / Windows adapter name as appropriate.)

Alternatively, use **private LAN** routable addresses if both have stable addresses on that subnet.

---

## Step 7 — Run and verify

Start Uqda on both (often requires privileges for TUN):

```sh
sudo uqda -useconffile /path/to/uqda.conf
```

On each host:

```sh
uqdactl getPeers
uqdactl getSelf
```

You should see **one** direct peer (the friend). **`getSelf`** shows your overlay address; you can `ping6` the friend’s Uqda IPv6 if ICMP is allowed by both OS firewalls.

---

## What this does **not** guarantee

- **Anonymity** — see [FAQ](faq.md).
- **OS‑level firewall** — anyone who can reach services on your machine over Uqda IPv6 still hits **your** firewall rules; **`AllowedPublicKeys`** only filters **who may complete a peering**, not who can port‑scan your TCP services if you expose them.
- **Third‑party relays** — if you introduce a public relay or VPS, that machine sees **metadata** (IPs, timing); payload between Uqda nodes remains encrypted between Uqda nodes **after** the peering is established according to normal Uqda rules.

---

## See also

- [Configuration](configuration.md) — editing `uqda.conf`, normalising, keys.
- [Configuration reference](configuration-reference.md) — **`Peers`**, **`Listen`**, **`AllowedPublicKeys`**, URI options.
- [Advanced peerings](advanced-peerings.md) — priorities, Tor, multicast firewalls (if you later relax the “only two” model).
- [Key rotation after a leak](key-rotation.md) — if a key or config file was exposed.
