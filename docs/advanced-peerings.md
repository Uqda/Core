# Advanced peerings

This page covers multi-homed setups, listener and multicast priorities, Tor over SOCKS, and firewalls with multicast discovery.

## Multi-homed outbound peerings

On hosts with several interfaces that can reach the same peer (or the Internet), use **`InterfacePeers`** instead of **`Peers`** to pin peerings to an interface.

You can even open **multiple** peerings to the **same** peer over **different** interfaces for high availability:

```hjson
InterfacePeers: {
  "eth0": [
    "tls://a.b.c.d:e"
  ]
  "eth1": [
    "tls://a.b.c.d:e"
  ]
}
```

## Prioritised listeners

With Wi‑Fi and Ethernet, you may want both to accept inbound peerings but **prefer** one path when both exist. Use the **`priority`** query parameter on the peering URI (**lower is better**).

Example: wired `a.b.c.d`, wireless `f.g.h.i`:

```
tls://a.b.c.d:e?priority=1
tls://f.g.h.i:e?priority=2
```

**Note:** `priority` only affects traffic **between two peerings to the same remote node**. It does **not** steer traffic across unrelated nodes or replace global routing policy.

## Prioritised multicast interfaces

For **`MulticastInterfaces`**, set **`Priority`** (lower is better) so discovered peerings on `eth0` are preferred over `eth1` to the same node:

```hjson
MulticastInterfaces: [
  {
    Regex: eth0
    Beacon: true
    Listen: true
    Priority: 1
  },
  {
    Regex: eth1
    Beacon: true
    Listen: true
    Priority: 2
  }
]
```

Same limitation as above: only applies between peerings to the **same** node.

## Multiple outbound Tor circuits

Peering over Tor via SOCKS is possible but often **slow and fragile**; circuits can drop at any time.

Mitigation: enable **`IsolateSOCKSAuth`** in Tor, then use **distinct** SOCKS username/password pairs so each peering uses its own circuit:

```hjson
Peers: [
  "socks://one:one@localhost:9050/a.b.c.d:e"
  "socks://two:two@localhost:9050/a.b.c.d:e"
  "socks://three:three@localhost:9050/a.b.c.d:e"
]
```

See the Tor manual for **`IsolateSOCKSAuth`**.

## Multicast peerings behind a host firewall

If multicast discovery does not work through your firewall:

1. Under **`MulticastInterfaces`**, set **`Port`** to a **fixed** TCP port.
2. Allow **inbound TCP** to that port (other nodes may initiate peerings).
3. Allow **inbound UDP** on port **9001** (optionally to **`ff02::114`**) so beacons are received.

## See also

- [Configuration](configuration.md)
