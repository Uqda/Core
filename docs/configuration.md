# Configuration

Uqda can run in one of two modes: **with a configuration file**, or in **autoconfigure** mode.

## Static configuration (recommended for most users)

A static configuration file lets you keep the same keypair (and therefore the same IPv6 address), maintain a list of peers, and tune listeners and multicast. Most Uqda packages generate **`/etc/uqda.conf`** (or a platform-specific path) automatically; you usually still need to add **Peers** before the node is useful on a wide area network. See [Manually connecting to peers](#manually-connecting-to-peers) below.

## Autoconfigure mode

Quick start with defaults:

```sh
uqda -autoconf
```

In this mode, Uqda tries to peer with other nodes on the **same subnet** via multicast; it does **not** connect to arbitrary public peers by default. Keys are random on each start, so the address changes every time.

## HJSON and JSON

Uqda accepts **HJSON** (comments, relaxed syntax) or strict **JSON**. HJSON is the usual default for hand-edited files.

For **what the network offers** and a **full walkthrough of every config field**, URI parameter, and related **`uqda` / `uqdactl`** commands, see **[Network capabilities and complete configuration guide](network-services-and-config.md)**.

## Configuration reference

Option names, defaults, and semantics are documented in:

- The output of **`uqda -genconf`** (annotated template).
- The implementation in **`src/config/`** in this repository.
- The tabular **[Configuration reference](configuration-reference.md)** for **NodeConfig** fields.

## Generating a new config file

If you installed from a package, a default file may already exist, commonly:

- **`/etc/uqda.conf`**, or
- **`/etc/uqda/uqda.conf`**

depending on the platform.

Otherwise:

```sh
# HJSON (typical)
sudo uqda -genconf > /etc/uqda.conf

# JSON
sudo uqda -genconf -json > /etc/uqda.conf
```

## Using a config file

**Stdin:**

```sh
uqda -useconf < /etc/uqda.conf
```

**Path:**

```sh
uqda -useconffile /etc/uqda.conf
```

## Normalising a config file

Convert formats or fill in missing keys with defaults:

```sh
# HJSON → JSON
uqda -normaliseconf -useconffile /etc/uqda.conf -json

# JSON → HJSON
uqda -normaliseconf -useconffile /etc/uqda.conf
```

Normalising is useful after upgrades when new options appear. Some packages run normalisation during upgrade.

## Exporting keys to external files

```sh
uqda -useconffile /etc/uqda.conf -exportkey > uqda.key
```

Then remove **`PrivateKey`** from the config and set **`PrivateKeyPath`** to the exported file.

If the key or full config may have been **stolen or copied**, do not keep using that identity: follow **[Key rotation after a leak](key-rotation.md)**.

## Manually connecting to peers

Add URIs under **`Peers`**. At startup, Uqda opens connections to them.

For a **closed** link between **exactly two** nodes (no public mesh, no multicast strangers), see **[Private network: two people only](private-two-nodes.md)**.

Examples:

| Type | Example |
|------|---------|
| TCP | `tcp://hostname:port` |
| TCP+TLS | `tls://hostname:port` |
| QUIC | `quic://hostname:port` |
| Via SOCKS | `socks://proxyhostname:proxyport/hostname:port` |

By default, **link-local auto-peering** (multicast) is enabled for devices on the same LAN / L2 segment.

Remote peerings use ordinary TCP (or TLS, QUIC, etc.) over IPv4 or IPv6; NAT or firewalls may require **port forwarding** for **inbound** listeners. For discovery on the public mesh, community **public peer** lists exist—use only sources you trust.

## Advertising a prefix

Each node has a **routed /64** in addition to its primary address. If your address is `200:1111:2222:3333:4444:5555:6666:7777`, your subnet is typically `300:1111:2222:3333::/64` (first byte **`200` → `300`** for the prefix /8; the rest of the first 64 bits match).

Find yours with:

```sh
uqdactl getSelf
```

It is also printed at startup.

### Linux example with radvd

1. Enable IPv6 forwarding, e.g. `sysctl -w net.ipv6.conf.all.forwarding=1`.
2. Assign a router address on the LAN interface, e.g.  
   `ip addr add 300:1111:2222:3333::1/64 dev eth0`  
   using your **`300:…::/64`** from **`uqdactl getSelf`**.
3. **`/etc/radvd.conf`** sketch:

```
interface eth0
{
        AdvSendAdvert on;
        prefix 300:1111:2222:3333::/64 {
            AdvOnLink on;
            AdvAutonomous on;
        };
        route 200::/7 {};
};
```

A /64 prefix has less entropy than a full 128-bit address; **do not treat `300::/8` addresses as strong identity proofs**.

## Stronger addresses and prefixes

If you advertise a prefix or want a harder-to-collide key/address, use the key search tool:

```sh
go run ./cmd/genkeys
```

It prints progressively “better” keys; paste the chosen keys into your configuration.

## See also

- [Key rotation after a leak](key-rotation.md)
- [Advanced peerings](advanced-peerings.md)
- [Installation guides](install-linux-manual.md) (index in the repository [README](../README.md))
