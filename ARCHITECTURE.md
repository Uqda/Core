# Architecture (current state)

This document describes the codebase **as it exists today**. It started as an
audit of the pristine upstream `github.com/yggdrasil-network/yggdrasil-go`
checkout (baseline commit `27dd82b`, before any Uqda-specific change) and has
been kept up to date as the product rename and hardening work described in
[NAMING.md](NAMING.md) actually happened - the module path, binary names, and
file/directory layout below are the current, renamed ones
(`github.com/Uqda/Core`, `uqda`/`uqdactl`), not the original upstream ones.
Where a path or name below is still `yggdrasil`-branded, that's because it is
either an unmodified upstream dependency (Ironwood), a third-party/standalone
tool kept under its own name (`contrib/yggdrasil-brute-simple`), or something
tracked as not-yet-renamed in
[docs/IMPLEMENTATION_STATUS.md](docs/IMPLEMENTATION_STATUS.md) - not an
oversight.

It is a factual audit, not a design proposal. It exists to give any
contributor or AI agent a map of the codebase. For the naming policy, see
[NAMING.md](NAMING.md). For the target folder layout and phased roadmap
(now substantially executed - see IMPLEMENTATION_STATUS.md for what remains),
see [RESTRUCTURING.md](RESTRUCTURING.md).

## Top-level layout

| Path | Contents |
|---|---|
| `cmd/` | The three Go binaries: daemon, control CLI, key generator |
| `src/` | All core library packages |
| `contrib/` | Packaging and platform integration scripts/files |
| `misc/` | Manual (non-`go test`) integration test scripts |
| `build`, `build.bat`, `clean` | Top-level build scripts (POSIX + Windows) |
| `.github/workflows/` | CI: `ci.yml`, `docker.yml`, `pkg.yml` |

## Entry points

- Daemon: [cmd/uqda/main.go](cmd/uqda/main.go), with privilege-drop
  helpers split by platform: `chuser_unix.go` / `chuser_other.go`
  (+ `chuser_unix_test.go`).
- Control CLI (admin client): [cmd/uqdactl/main.go](cmd/uqdactl/main.go)
  and `cmd_line_env.go`.
- Key generator: [cmd/genkeys/main.go](cmd/genkeys/main.go).

## Core subsystems

| Subsystem | Location | Notes |
|---|---|---|
| Config loading | `src/config/config.go`, `defaults.go`, `defaults_{darwin,freebsd,linux,openbsd,windows,other}.go` | HJSON or JSON |
| Identity / keys | `src/core/core.go`, `src/identity/tls.go` | ed25519 keypair held in `Core{secret, public}`; TLS config generation lives in `src/identity` (extracted from `src/core/tls.go`) |
| Address derivation | `src/address/address.go` | `AddrForKey` / `SubnetForKey` derive IPv6 addresses/subnets from ed25519 public keys |
| Peer / transport management | `src/core/link.go` + `link_tcp.go` (`_darwin`/`_linux`/`_other`), `link_tls.go`, `link_quic.go`, `link_ws.go`/`link_wss.go`, `link_socks.go`, `link_unix.go` | One generic link manager, per-transport implementations |
| Routing | External module `github.com/Arceliar/ironwood` | No local routing engine; wired into `src/core` via `New()` / `api.go` / `proto.go` |
| Sessions | Ironwood-internal | Only exposed to operators via `src/admin/getsessions.go` |
| Multicast / local discovery | `src/multicast/` (`multicast.go`, `advertisement.go` + test, `admin.go`, `options.go`, OS variants) | |
| TUN integration | `src/tun/` (`tun.go`, `iface.go`, `admin.go`, `options.go`, OS variants) and `src/ipv6rwc/` (`ipv6rwc.go` + test, `icmpv6.go`) | `ipv6rwc` is the shim between TUN and Core |
| Admin / local API | `src/admin/admin.go` | Listens on unix or tcp socket; command handlers: `addpeer.go`, `removepeer.go`, `getpeers.go`, `getself.go`, `getsessions.go`, `getpaths.go`, `gettree.go`, `error.go`, `options.go` |
| Logging | Wired in `cmd/uqda/main.go` via `github.com/gologme/log` and `github.com/hashicorp/go-syslog` | |
| Update mechanism | **None exists.** | No auto-update / self-update code anywhere in the tree — relevant if a verified-update subsystem is planned later |

## Packaging artifacts (`contrib/`)

| Platform/system | Files |
|---|---|
| systemd | `systemd/uqda.service`, `uqda.service.debian`, `uqda-default-config.service(.debian)` |
| Debian | `deb/generate.sh` |
| macOS | `macos/create-pkg.sh`, `macos/uqda.plist` |
| Windows MSI | `msi/build-msi.sh`, `msi/msversion.sh` |
| Docker | `docker/Dockerfile`, `docker/Dockerfile.multiarch`, `docker/entrypoint.sh` (root `Dockerfile` is a 1-line pointer to these) |
| OpenRC | `openrc/uqda` |
| BusyBox init | `busybox-init/S42uqda` |
| FreeBSD rc | `freebsd/uqda` |
| AppArmor | `apparmor/usr.bin.uqda(ctl)` |
| Ansible | `ansible/genkeys.go` |
| Mobile bindings (gomobile) | `mobile/*.go` |
| Versioning helper | `semver/{name,version}.sh` (used by `build`/`build.bat` to set `-ldflags`) |
| Logo | `logo/ygg-neilalexander.svg` |
| Standalone brute-force test tool | `yggdrasil-brute-simple/` (separate C project with its own `LICENSE`) |

## Tests

Beyond the original upstream unit tests, this repo now also has a black-box
daemon interop test and two fuzz targets (see
[docs/IMPLEMENTATION_STATUS.md](docs/IMPLEMENTATION_STATUS.md) Phase 16 for
what's verified and how):

- `src/address/address_test.go`
- `src/config/config_test.go`, `src/config/permissions_test.go`
- `src/core/core_test.go`, `src/core/link_test.go` (includes
  `FuzzVersionMetadataDecode`), `src/core/link_ws_test.go`,
  `src/core/options_test.go`, `src/core/version_test.go`
- `src/identity/tls_test.go`
- `src/ipv6rwc/ipv6rwc_test.go`
- `src/multicast/advertisement_test.go` (includes
  `FuzzMulticastAdvertisementUnmarshalBinary`)
- `src/admin/admin_test.go`
- `cmd/uqda/chuser_unix_test.go`, `cmd/uqda/main_test.go`,
  `cmd/uqda/checkconf_test.go`
- `contrib/ansible/genkeys_test.go`
- `contrib/mobile/mobile_test.go`
- `tests/interop/daemon_test.go` — builds and runs the real `uqda` binary as
  two separate OS processes; same-codebase today, becomes a genuine
  cross-implementation check once a pinned upstream Yggdrasil fixture is
  added (tracked in IMPLEMENTATION_STATUS.md)

Manual integration scripts (not `go test`-driven): `misc/run-twolink-test`,
`misc/run-schannel-netns`.

## Dependencies (`go.mod`)

- Module: `github.com/Uqda/Core`
- Go directive: `go 1.25.0` — **note:** `README.md` still says "Go 1.22 or later"; this mismatch predates any Uqda work and should be reconciled independently of rebranding.
- Routing: `github.com/Arceliar/ironwood`
- Actor model: `github.com/Arceliar/phony`
- Transports: `github.com/quic-go/quic-go`, `github.com/coder/websocket`
- TUN backends: `golang.zx2c4.com/wireguard`, `wireguard/wintun`, `wireguard/windows`
- Networking utilities: `github.com/vishvananda/netlink`, `github.com/vishvananda/netns`
- Standard extensions: `golang.org/x/{crypto,net,sys,text}`
- Config format: `github.com/hjson/hjson-go/v4`
- Logging: `github.com/gologme/log`, `github.com/hashicorp/go-syslog`
- Windows service: `github.com/kardianos/minwinsvc`
- CLI output: `github.com/olekukonko/tablewriter`
- Sandboxing: `suah.dev/protect`

## Legal / attribution state

- `LICENSE` — LGPLv3, with a special exception permitting static/dynamic
  linking without triggering the LGPL's source-disclosure requirements for
  the linking application. Unmodified since the baseline import.
- `NOTICE.md` — states that Uqda derives from Yggdrasil, and gives verified
  (not assumed) attribution for Ironwood and phony (both MPL-2.0).
- `SECURITY.md` — vulnerability reporting policy and an explicit statement
  of what is and isn't guaranteed.

There is still no `CONTRIBUTING.md` in this checkout.
