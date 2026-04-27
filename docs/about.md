# About Uqda

Uqda is an experimental **software router** and routing design for building **multi-hop** IPv6 networks with minimal configuration. The design is **decentralised**: nodes carry only a small amount of state. Routing is predominantly **shortest-path** oriented—the mesh tries to use direct paths when they exist.

Uqda runs on many operating systems and can be used on ordinary computers, servers, and embedded devices.

## How the network behaves

Every participating node is a **router**: it forwards traffic and connects to others using **peerings** over IP (wired, Wi‑Fi, LAN, WAN, or the public Internet). Nearby nodes on the same broadcast domain can often peer **automatically** via **multicast discovery**.

Nodes cooperate to move packets toward their destinations. Even in a **sparse** topology, nodes tend to remain reachable when peerings exist—**NAT** on a node does not prevent **return traffic** over an established outbound peering in the usual way.

The protocol is built to **adapt** when links fail: the mesh **reconverges** when alternate paths exist, which suits **mesh** deployments where topology changes often.

Each node has a **cryptographic identity**. In the current implementation, **stable IPv6 addresses** are derived from that identity, so most IPv6-aware applications work **without modification**. The address is **location-independent** in the sense that it stays with the key material as you move (routing updates as you re-peer).

## Why Uqda?

Many networks today are **hierarchical**, need heavy **manual** configuration, and lean on **centralised** allocation to scale. That makes **ad-hoc** networks slow to stand up, and pushes users toward traditional ISPs.

Uqda aims for **little configuration**: multi-hop connectivity can come up quickly once peerings exist. Nodes do not need a central authority to “assign” an address in the usual sense; they generate keys and keep their identity as they roam. Routing information propagates automatically among peers.

End-to-end reachability between participants can support **edge** workloads and **real-world mesh** experiments. The overlay can also run **without** relying on the global Internet where that matters.

## How Uqda compares to other projects

**Anonymity overlays** (Tor, I2P, similar) target **anonymity** guarantees and different threat models. Uqda **does not** aim to provide or guarantee anonymity. Those systems are overlays **by design**; Uqda is implemented as an overlay today largely because that is a practical way to deploy and study the routing design.

**VPN-style** tools (WireGuard, Tailscale, Nebula, ZeroTier, …) focus on **private** networks or controlled membership. You *can* build private meshes with Uqda, but that is not the only goal. Bridging a **private** island to a **public** peer effectively **merges** reachability—treat that as intentional architecture, not a surprise.

Uqda has **no built-in “exit node”** concept for the clearnet; if you need that, use **proxies** or other tunnels **on top** of Uqda.

## Project status

Uqda Core remains **research-grade / alpha** software: under active development, protocol and wire formats may still evolve, and **not** every failure mode is known.

A future **beta** might promise stronger compatibility expectations; a hypothetical **1.0** would imply broader stability commitments. Until then, avoid **life-safety** or **sole** reliance for critical workloads.

Reasonable outcomes over time include: the design scales and stabilises; the project stabilises without huge adoption; or real-world load exposes design limits—each teaches something for **future versions** or **other implementations**.

For practical setup, see the [installation guides](../README.md#documentation) and [Configuration](configuration.md). For a **single map** of the repository (code layout, binaries, features, and all docs), see the **[documentation hub](README.md)**.
