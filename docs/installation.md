# Installation

New here? [Getting started](getting-started.md) covers install, setup and
adding devices in five steps. This page has the details for each system.

Releases are published on [GitHub Releases](https://github.com/Uqda/Core/releases).
Download an artifact and verify it against that release's `SHA256SUMS`, or build
with Go 1.26 or newer using `./build` or `build.bat`. `uqda version` and
`uqdactl version` show the installed release. For everyday status, run `uqda`;
see the [command guide](commands.md).

## Packages

| System | How | Guide |
|---|---|---|
| Linux with systemd (x86-64, arm64) | Quick installer, with checksum checks and identity-preserving updates | [quick install](quick-install.md#linux-systemd-x86-64-or-arm64) |
| Debian, Ubuntu (amd64, arm64, armhf, armel, i386, mips, mipsel) | `.deb` from Releases | below |
| macOS (Intel, Apple Silicon) | Homebrew cask `Uqda/tap/uqda` | [quick install](quick-install.md#macos-homebrew) |
| Windows (x64, x86, ARM64) | `.msi` from Releases | [quick install](quick-install.md) |
| Umbrel | Community app store | [Umbrel](umbrel.md) |
| Android, iOS apps | SDKs (`.aar`, `.xcframework`) for app developers; no ready-made app yet | [mobile](mobile.md) |

After installing, run `sudo uqda setup country=NAME` (see [getting started](getting-started.md)).

## Docker

```sh
docker build -t uqda-core .
docker run -d --name uqda --cap-add NET_ADMIN --device /dev/net/tun   --network host -v uqda-config:/etc/uqda uqda-core
```

Persist `/etc/uqda`: it holds the node identity, and a container without its
volume gets a new address. Keep the admin endpoint private.

## Details

Source builds require the complete repository checkout: it contains the pinned,
locally patched Ironwood dependency. `go install ...@version` is not a supported
installation method with this local replacement; use the release packages or
the checkout/build commands. Go applications importing Core as a dependency do
not inherit its module replacement automatically; they must explicitly select
the patched dependency in their own main module. Native release binaries and
the published mobile bindings include the patch.

| Integration | Configuration | Build entry point |
|---|---|---|
| Debian/systemd | `/etc/uqda/uqda.conf` | `sh contrib/deb/generate.sh` |
| Windows MSI | `%ProgramData%\Uqda\uqda.conf` | `sh contrib/msi/build-msi.sh x64` |
| macOS launchd | `/etc/uqda/uqda.conf` | `sh contrib/macos/create-pkg.sh` |
| Docker | `/etc/uqda/uqda.conf` | `docker build -t uqda-core .` |
| FreeBSD rc | `/usr/local/etc/uqda/uqda.conf` | `contrib/freebsd/uqda` |

Default admin sockets use `/run/uqda/admin.sock` on Linux and
`/var/run/uqda/admin.sock` on macOS/BSD; Windows uses `tcp://localhost:9001`.

Migration sources are `/etc/yggdrasil/yggdrasil.conf` or `/etc/yggdrasil.conf`
on Unix, `%ProgramData%\Yggdrasil\yggdrasil.conf` on Windows, and `config.conf`
in the mounted container volume. Ambiguous Unix sources require manual selection.
Validation failure aborts migration. Back up and inspect the configuration and
any external PEM key before uninstalling upstream software. Stop the old daemon
before starting the same identity in Uqda. See [compatibility](compatibility.md).

For Docker, persist `/etc/uqda` and supply the TUN device and required network
capabilities when using host networking integration. Keep the admin endpoint
private. Removing a container without preserving its volume loses its identity.

Service-only installations require a pre-existing validated configuration; configuration
generation is an explicit administrative action, not a daemon dependency.
Native installation, upgrade and uninstall checks belong on the target platform.
See [packaging](packaging.md).
