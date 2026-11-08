# Getting started

This guide takes a device from nothing to a working private network, then
shows how to add more devices and reach your services. Every command is the
same on Linux, macOS and Windows unless noted. On Windows, run commands in an
administrator PowerShell and leave out `sudo`.

## 1. Install

| System | Command or file |
|---|---|
| Linux (Debian, Ubuntu, Fedora; x86-64 or arm64) | `curl -fsSLo uqda-install.sh https://raw.githubusercontent.com/Uqda/Core/main/contrib/install/linux.sh && sudo sh uqda-install.sh install` |
| Debian / Ubuntu package | Download `uqda-VERSION-ARCH.deb` from [Releases](https://github.com/Uqda/Core/releases), then `sudo apt install ./uqda-*.deb` |
| macOS (Intel or Apple Silicon) | `brew install --cask Uqda/tap/uqda` |
| Windows (x64, x86, ARM64) | Download and run `uqda-VERSION-ARCH.msi` from [Releases](https://github.com/Uqda/Core/releases) |
| Umbrel home server | Add the community store `https://github.com/Uqda/umbrel-app-store`, install **Uqda Network**. See [Umbrel](umbrel.md). |
| Docker | See [installation](installation.md#docker) |
| From source | Go 1.26+, then `./build` (or `build.bat`). See [installation](installation.md). |

Check it worked:

```sh
uqda version
```

On Windows, open a **new** PowerShell window *as administrator* after
installing, so it picks up the updated `PATH`. Releases up to 26.1 did not add
Uqda to `PATH`; with those, use the full path instead:
`& "C:\Program Files\Uqda\uqda.exe" status` (or `Program Files (x86)` for the
x86 package).

Every release lists `SHA256SUMS`; the Linux installer checks it for you. Details
for each system, updates and removal are in [installation](installation.md) and
[quick install](quick-install.md).

## 2. Set up the node

```sh
sudo uqda setup country=germany
```

Pick a country near you; `uqda find` lists the available ones. This one command:

- keeps the identity the installer created (or creates one),
- creates a **private network** with a random group password,
- turns on the **inbound firewall**,
- adds the two fastest healthy public peers in that country.

It then prints the command that restarts Uqda on your system, for example
`sudo systemctl restart uqda` on Linux. Run it.

Want the open public network instead? Use `sudo uqda setup mode=public country=germany`.
[Networks](networks.md) explains the difference.

## 3. Check the node

```sh
uqda status
```

| Check | What a pass means |
|---|---|
| Daemon | Uqda is running and its admin socket answers |
| Identity | The node has a valid key and address |
| Peers | At least one peer link is up |
| Interface | The `uqda0` network interface exists |
| Firewall | Unsolicited inbound traffic is blocked |

Each warning comes with the fix. `uqda info` shows your address, such as
`200:1234:5678::1`.

## 4. Add another device

On the node you just set up:

```sh
sudo uqda invite
```

It prints `uqda join uqda-invite-1:...` and the same invite as a QR code. Install
Uqda on the new device, then:

```sh
sudo uqda join uqda-invite-1:...
```

Restart Uqda on the new device as `join` tells you. The two devices are now in
the same private network. The invite works once, expires after 15 minutes and
contains no password. [Invites](invites.md) explains how it stays safe.

Test the connection from either device:

```sh
uqda test 200:1234:5678::1
```

**Windows, if `join` cannot reach the inviting PC:** Windows treats new
networks as *Public*, and the installer's "Uqda (local network)" rule only
applies to *Private* networks. On your own home network only, in an
administrator PowerShell:

```powershell
Get-NetConnectionProfile                 # shows each network and its category
Set-NetConnectionProfile -InterfaceAlias WLAN -NetworkCategory Private
```

Use the `InterfaceAlias` shown for your Wi-Fi or Ethernet. Do not mark café,
hotel or work networks as Private.

## 5. Reach a service

Your devices can now talk over encrypted IPv6, but the firewall still blocks
connections *into* each device. To let your laptop reach SSH on a server, add a
rule to the server's `uqda.conf` (`/etc/uqda/uqda.conf` on Linux and macOS,
`%ProgramData%\Uqda\uqda.conf` on Windows):

```hjson
Firewall: {
  Enabled: true
  AllowPing: true
  Rules: [
    { Sources: ["LAPTOP-ADDRESS"], Protocol: "tcp", Ports: ["22"], Comment: "laptop SSH" }
  ]
}
```

Replace `LAPTOP-ADDRESS` with the laptop's `uqda info` address, restart Uqda,
then from the laptop:

```sh
ssh user@SERVER-ADDRESS
```

The service itself must listen on IPv6 (most do by default). See
[firewall](firewall.md) for more examples.

On Windows, Defender Firewall also filters the Uqda adapter, which Windows
classifies as a *Public* network. Keep it that way, and allow only the service
you are offering, limited to Uqda addresses:

```powershell
New-NetFirewallRule -DisplayName "My service over Uqda" -Direction Inbound -Protocol TCP -LocalPort 8080 -RemoteAddress 200::/7 -Action Allow
```

## Everyday commands

| Command | What it does |
|---|---|
| `uqda` or `uqda status` | Health checks with fixes |
| `uqda info` | Your address, subnet and public key |
| `uqda peers` | Peer links, traffic and latency |
| `uqda firewall` | Firewall state, rules and blocked packets |
| `uqda test ADDRESS` | Reachability and packet loss to another node |
| `uqda find [country=NAME]` | Public peers by country |
| `uqda setup` / `uqda invite` / `uqda join` | Set up, invite, join |
| `uqda help COMMAND` | Full help for one command, without a running daemon |

## If something does not work

| Symptom | Try |
|---|---|
| `Permission denied` on a command | Run it with `sudo` (macOS, Linux) or as administrator (Windows) |
| No connected peers | `uqda setup country=NAME` with another country, or check outbound internet access |
| `join` cannot reach the inviting node | Put both devices on the same network, or set a fixed `InvitePort` and allow it in the inviter's host firewall |
| Ping works but a service does not | Add a firewall rule for that port, and make sure the service listens on IPv6 |
| Anything else | `uqda status`, then the daemon log (`journalctl -u uqda` on Linux) |

Report problems with the [bug report form](https://github.com/Uqda/Core/issues/new?template=bug_report.yml).
