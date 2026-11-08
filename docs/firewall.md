# Inbound firewall

Every Uqda node has a public IPv6 address on the overlay. Without a firewall,
any service listening on that address (SSH, file sharing, a web dashboard) can
be reached by any node that can route to you. The inbound firewall closes that
by default and lets you open exactly what you need.

## How it works

- **Outbound is always allowed.** When this node opens a connection, the
  firewall remembers the flow, and the replies come back.
- **Inbound is dropped unless a rule allows it.** Rules match the sender,
  the protocol and the destination port.
- **Senders are authenticated.** A packet's source address has already been
  checked against the public key of the encrypted session that carried it,
  so a rule for a key or address cannot be satisfied by spoofing.
- **Ping is a separate switch** (`AllowPing`). ICMPv6 errors such as
  *Packet Too Big* are always allowed, because path MTU discovery needs them.
- It filters only traffic arriving over Uqda. Your LAN and the internet are
  unaffected; use your operating system firewall for those.

New configurations created with `uqda -genconf` (and by the installers) have
the firewall on. **Existing configurations without a `Firewall` section keep
it off**, so upgrading never cuts off a working service; `uqda status` warns
until you turn it on.

## Configuration

```hjson
Firewall: {
  Enabled: true
  AllowPing: true
  Rules: [
    # SSH, only from my laptop (by its Uqda address)
    { Sources: ["200:1234:5678::1"], Protocol: "tcp", Ports: ["22"], Comment: "laptop SSH" }

    # A web app for two trusted nodes, identified by public key
    { Sources: ["9f2c...64 hex characters...", "41ab...64 hex characters..."], Protocol: "tcp", Ports: ["8080"] }

    # Everything from my own other machine's routed /64
    { Sources: ["300:1234:5678:9abc::/64"], Protocol: "any" }

    # A public service anyone on the network may use
    { Sources: ["*"], Protocol: "tcp", Ports: ["443"] }
  ]
}
```

| Field | Meaning |
|---|---|
| `Sources` | Who may connect. Public keys (64 hex characters), Uqda addresses (`200::/7`), routed `/64` subnets (`300::/8`), or `"*"` for any node. Required. |
| `Protocol` | `tcp`, `udp`, `icmp` or `any`. Default `any`. |
| `Ports` | Destination ports such as `"22"` or ranges such as `"8000-8100"`. Omit to allow every port. Only for `tcp` and `udp`. |
| `Comment` | Free text shown by `uqda firewall`. |

Find a node's address with `uqda info` on that node, or its public key in the
same output. Validate changes before restarting:

```sh
uqda -useconffile /etc/uqda/uqda.conf -checkconf
```

Rules are read at startup; restart Uqda after editing them.

## Checking it

```sh
uqda firewall
```

shows whether the firewall is on, its rules, active flows, and allowed and
dropped packet counts. `uqda status` includes a firewall check. The admin
equivalent is `uqdactl getFirewall` (JSON with `-json`).

## Limits

- Flows idle out after 30 minutes (TCP), 3 minutes (UDP) and 30 seconds
  (ping). Long-idle connections may need keepalives.
- At most 65,536 flows are tracked. Beyond that new outbound flows still go
  out, but their replies are dropped until older flows expire; the
  `untracked_flows` counter in `uqdactl getFirewall -json` shows when this
  happens.
- On Umbrel the firewall stays off for now: the app's random private group is
  the access boundary, and dashboard controls for rules are planned.
- The firewall is not a substitute for application authentication. An allowed
  node is still talking to your service directly.
