# Uqda

[العربية](README.ar.md) · [Documentation](docs/README.md) · [Releases](https://github.com/Uqda/Core/releases) · [Report a problem](https://github.com/Uqda/Core/issues/new?template=bug_report.yml)

**Uqda is an encrypted network for your devices that no company runs.** There is
no account to create, no coordination server, and no subscription. Each device
gets a permanent IPv6 address derived from its own key, and devices reach each
other over end-to-end encrypted sessions, across the internet or straight over
your local network.

Uqda (عُقدة, "knot" or "node") is an independently maintained implementation
compatible with the existing [Yggdrasil](https://yggdrasil-network.github.io/)
network. It does not create a separate network and does not imply upstream
endorsement.

> **Experimental project.** Uqda is under active development. It is tested
> extensively but has not had an independent security audit. Do not rely on it
> where a failure would put people or critical systems at risk.

## Why Uqda

| | Uqda | Typical mesh VPN (Tailscale, ZeroTier, NetBird) |
|---|---|---|
| Account required | No | Yes |
| Central coordination server | None: nothing to shut down, block or reprice | Required (hosted, or self-hosted) |
| Works on a local network without internet | Yes (local discovery) | Limited |
| Add a device | `uqda invite`, then `uqda join` (one-time, expiring, no password inside) | Sign in to the vendor |
| Who may reach each service | Built-in firewall, rules by device key | ACLs on the vendor's server |
| Replay-protected peer handshake | Yes, between Uqda nodes | Yes |
| Direct connections through NAT | Through peer nodes (relays) | Yes, hole punching |

The last row is where the others are stronger today; see the [roadmap](docs/roadmap.md).

## Two kinds of network

- **Private group:** only devices that share your group secret can talk to
  you. Use it for your own devices, a family or a team. This is the default.
- **Public network:** reach and be reached by any node on the worldwide
  overlay, with the firewall deciding what is open.

[Networks](docs/networks.md) explains both, and when to use each.

## Install and start in three commands

Install for your system ([all options](docs/installation.md)):

| System | Install |
|---|---|
| Linux | `curl -fsSLo uqda-install.sh https://raw.githubusercontent.com/Uqda/Core/main/contrib/install/linux.sh && sudo sh uqda-install.sh install` |
| macOS | `brew install --cask Uqda/tap/uqda` |
| Windows | The `.msi` for your system from [Releases](https://github.com/Uqda/Core/releases) |
| Debian / Ubuntu | The `.deb` from [Releases](https://github.com/Uqda/Core/releases) |
| Umbrel | Community store `https://github.com/Uqda/umbrel-app-store` → **Uqda Network** |

Then:

```sh
sudo uqda setup country=germany   # private network, firewall and two fast peers
sudo systemctl restart uqda       # setup prints the right command for your system
sudo uqda invite                  # on each new device: sudo uqda join <invite>
```

`uqda find` lists countries with public peers. The full walk-through, including
how to reach a service on another device, is in [getting started](docs/getting-started.md).

## Everyday commands

| Command | What it does |
|---|---|
| `uqda status` | Health checks, each with its fix |
| `uqda info` | Your address and public key |
| `uqda peers` | Peer links, traffic and latency |
| `uqda firewall` | Firewall rules and blocked packets |
| `uqda test ADDRESS` | Reachability and packet loss to another node |
| `uqda help COMMAND` | Help for any command |

## Security in brief

- **Identity:** Ed25519 keys. Your address is derived from your key and cannot be claimed without it.
- **Sessions:** end-to-end encrypted. Relays carry traffic but cannot read it.
- **Private groups:** the group secret is stretched with Argon2id. Nodes outside the group cannot open sessions with you.
- **Inbound firewall:** on by default for new nodes. Only replies to your own connections, and the services you allow, get in.
- **Peer handshake:** Uqda nodes prove their key on every connection, so a recorded handshake cannot be replayed. Floods of half-open connections are capped.
- **Invites:** TLS 1.3 pinned to the inviting node's key. They are one-time, expire and are rate-limited.

Uqda does **not** provide anonymity, and the admin API has no authentication
(it stays local). Read the [security model](docs/security.md), and report
vulnerabilities privately as described in [SECURITY](SECURITY.md).

## Project

- Two branches: `main` is stable and `beta` gets new work first; see [branching](docs/branching.md).
- Every change runs the full test suite on Linux, macOS and Windows. That covers race detection,
  fuzzing, packaging on every supported system, and a compatibility gate against the official
  Yggdrasil implementation; see [testing](docs/testing.md).
- Contributions are welcome: see [CONTRIBUTING](CONTRIBUTING.md). Changes are listed in the [CHANGELOG](CHANGELOG.md).

## License and attribution

Uqda Core is licensed under LGPLv3 with the linking exception in [LICENSE](LICENSE).
Routing and encrypted sessions come from Ironwood. See [NOTICE](NOTICE.md) for
Yggdrasil, Ironwood and other upstream attribution.

---

*Uqda is an experimental project, provided as is and without warranty.*
