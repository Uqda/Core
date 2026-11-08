# Invites: add a device in one step

Adding a device to a private Uqda network used to mean copying a group
password and peer addresses by hand. An invite does it for you, without
putting any secret inside the invite itself.

## Use it

On a node that is already part of the network:

```sh
sudo uqda invite
```

It prints a line such as `uqda join uqda-invite-1:…` and the same invite as a
QR code. On the new device:

```sh
sudo uqda join uqda-invite-1:…
```

Then restart Uqda on the new device; `join` prints the exact command for your
system. The new device keeps its own identity and receives the group password
and peers. If it had no configuration yet, `join` creates one with the
[inbound firewall](firewall.md) on.

Options:

| Command | Meaning |
|---|---|
| `uqda invite expires=2h` | Valid for 2 hours (default 15 minutes, at most 24 hours). |
| `uqda invite uses=3` | Works for 3 devices (default 1, at most 10). |
| `uqda invite name=Laptop` | A label shown on the joining device. |
| `uqdactl cancelInvites` | Revoke every active invite now. |
| `uqda join INVITE config=PATH` | Write a configuration other than the default path. |

## How it stays safe

1. The invite contains the inviting node's **public key**, the addresses where
   it waits for redemption, and a **random 256-bit one-time secret**. It does
   not contain the group password or peer list.
2. `join` connects to the inviting node over **TLS 1.3** and accepts it only if
   it proves it holds the private key matching that public key. Someone who
   controls the network address but not the node cannot pose as it.
3. The joining device presents the secret. The inviting node compares a hash of
   it in constant time, checks expiry and remaining uses, and only then sends
   the group password and peers, inside the encrypted connection.
4. The redemption port is open only while an invite is active. Ten failed
   attempts from one address lock it out for a minute, and at most 16
   redemptions run at once.
5. Only a hash of each secret is kept, in memory. Restarting the inviting
   node invalidates every invite.

Until it is used or expires, treat an invite like a password: anyone who has it
can join once. Share it over a channel you trust, or show the QR code in person.

## Reachability

The joining device must reach the inviting node directly over IP during
redemption, usually on the same local network. To redeem across the internet,
the inviting node needs a reachable address and port. Set a fixed `InvitePort`
in `uqda.conf` and allow it in the host firewall; the default picks a random
port each time.

After joining, the new device connects through the shared peers. If the
inviting node has no peers and no fixed-port listeners, `uqda invite` warns
you. Devices on the same local network still find each other through
multicast discovery.
