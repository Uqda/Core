# Installation

Uqda 26 Beta 1 is not published. Build with Go 1.25 or newer using `./build`
or `build.bat`. Run `uqda --version` and `uqdactl version` to check the local
product identity. Both display `Uqda Core 26 Beta 1`.

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
