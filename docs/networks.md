# Public network and private groups

Every Uqda node is part of one worldwide overlay, shared with Yggdrasil. On top
of it you choose who your node may talk to: everyone (public), or only members
of your group (private). The choice is one setting, `GroupPassword`.

## Two layers

| Layer | What it is | Who you connect to |
|---|---|---|
| **Peer links** | Direct TCP, TLS, QUIC or WebSocket connections to nearby nodes. They carry routed traffic. | Your configured `Peers`, nodes that connect to your `Listen` addresses, and local nodes found by multicast. |
| **Sessions** | End-to-end encrypted conversations between two nodes, carried over any number of peer links. | Any node you exchange traffic with. Relays in between cannot read it. |

A private group limits **sessions**, not peer links. Group members can use
public peers as transport, and a public relay can carry their traffic, but
it can neither read it nor start a session with a member.

## Public network

`GroupPassword` is empty.

- Your node can reach, and be reached by, every node on the overlay.
- Useful for public services, joining the wider Yggdrasil community, and
  experiments.
- Anyone on the network can try to connect to your address. The
  [inbound firewall](firewall.md) is what keeps your services private, so keep
  it on and open only what you mean to share.

Set up: `sudo uqda setup mode=public country=NAME`

## Private group

`GroupPassword` is a long random secret shared by every member.

- Sessions only form between nodes with the same secret. Nodes outside the
  group cannot reach your address at all, even with the firewall off.
- You can still use public peers for transport.
- Useful for your own devices, a family, a team or a small organisation.
- The secret is stretched with Argon2id at startup, but it is still a shared
  secret: anyone who has it is a member. `uqda setup` generates a 192-bit one.
  Use `uqda invite` to hand it to new devices instead of copying it around.

Set up: `sudo uqda setup country=NAME` (private is the default).

Inside a group the firewall still matters: it decides which members may reach
which services. For example, every device can ping, but only your laptop can
SSH to the server.

## Choosing

| You want to | Use |
|---|---|
| Connect your own devices, or a family or team | Private group |
| Offer a service to anyone on the overlay | Public network, with firewall rules for that service only |
| Both | Two nodes: a private one for your devices, a public one for the shared service |

## Changing later

- Join an existing group: `sudo uqda join INVITE` on the new device.
- Make a public node private: `sudo uqda setup mode=private`.
- Make a private node public: `sudo uqda setup mode=public confirm=yes`. This
  removes the group protection, so check your firewall rules first.

Restart Uqda after any change. All members of a group must use the same
secret. Groups created before Uqda 26.0.2 must upgrade every member together,
because the password derivation changed.

## What neither mode gives you

- **Anonymity.** Peers see your IP address, and relays see traffic volume and
  timing. Uqda is not Tor.
- **Protection for exposed services.** An allowed connection reaches your
  service directly, so the service still needs its own authentication.
