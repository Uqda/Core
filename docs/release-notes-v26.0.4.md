# Uqda Core 26.0.4

A new patch release adding read-only public TLS peer suggestions. Previous
release tags and downloads remain unchanged.

## Changes

- `uqda find` lists countries from https://publicpeers.neilalexander.dev/ without
  contacting peers, reading node configuration or querying geolocation services.
- `uqda find country=germany limit=3` checks up to eight online TLS candidates,
  then ranks reachable TCP ports by connection time. Text and `--json` output
  include the source timestamp and clear validation limits.
- `uqda help find` works offline. No daemon or administrator privileges are
  required for discovery; existing administration commands are unchanged.
- Bounded parsing, HTTPS redirect restrictions, stale-snapshot rejection and
  public-IP-only DNS validation protect the discovery path. Credentials and
  password-bearing URIs are not suggested. Three workers, three-second candidate
  deadlines and at most two IP attempts limit network activity.
- Catalog fuzz testing joins the existing cross-platform, race, security and
  installer lifecycle gates. Native macOS upgrade tests now start from 26.0.3.
- The two-daemon regression harness reserves its admin and peer ports together,
  preventing duplicate port selections before subprocess startup.

```sh
uqda find
uqda find country=germany
uqda find country=austria limit=2 --json
uqda help find
```

## Safety and compatibility

Discovery does not connect the overlay, add peers, restart services or change
`GroupPassword`, allowlists or node identity. Public peers do not normally
authenticate into a private group; do not disable group protection to use a
suggestion. The existing closed network remains closed.

A successful TCP check is not a TLS handshake, key verification, authenticated
Uqda session, trust guarantee or overlay ping. Its timing is not ping RTT. The
catalog is a third-party status source, not a security endorsement. Operators
see your source IP during probes. Only online TLS candidates are suggested.
See [public peer discovery documentation](https://github.com/Uqda/Core/blob/v26.0.4/docs/public-peers.md).

There is no wire-format or private-group authentication change from 26.0.3.
Normal updates retain identity and configuration. No suite proves the absence
of every bug; cross-compilation is not a native installation test. Windows ARM64
packaging is build-only. Installers remain unsigned.

## Installation and update

macOS:

```sh
brew update
brew install --cask Uqda/tap/uqda
# Existing installation:
brew upgrade --cask Uqda/tap/uqda
```

Linux/systemd:

```sh
curl -fsSLo /tmp/uqda-install.sh https://raw.githubusercontent.com/Uqda/Core/main/contrib/install/linux.sh
sudo sh /tmp/uqda-install.sh install
# Existing quick installation:
sudo sh /tmp/uqda-install.sh update
```

Windows: install the matching new MSI from this release. Do not use the Linux
quick installer over a package-managed installation. Verify release downloads
against `SHA256SUMS`. Normal removal retains configuration; explicit purge/zap
permanently deletes identity and should only follow a protected backup.
