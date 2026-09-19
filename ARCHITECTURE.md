# Architecture (current state)

This document describes the codebase **as it exists today** — a pristine checkout
of upstream `github.com/yggdrasil-network/yggdrasil-go` (module path, binary
names, and all identifiers are unmodified "yggdrasil"; see the baseline commit
`27dd82b`). It is a factual audit, not a design proposal. It exists to give any
contributor or AI agent a map of the codebase before any Uqda rebrand or
hardening work begins.

For the naming/rebrand plan, see [NAMING.md](NAMING.md). For the target folder
layout and phased roadmap, see [RESTRUCTURING.md](RESTRUCTURING.md).

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

- Daemon: [cmd/yggdrasil/main.go](cmd/yggdrasil/main.go), with privilege-drop
  helpers split by platform: `chuser_unix.go` / `chuser_other.go`
  (+ `chuser_unix_test.go`).
- Control CLI (admin client): [cmd/yggdrasilctl/main.go](cmd/yggdrasilctl/main.go)
  and `cmd_line_env.go`.
- Key generator: [cmd/genkeys/main.go](cmd/genkeys/main.go).

## Core subsystems

| Subsystem | Location | Notes |
|---|---|---|
| Config loading | `src/config/config.go`, `defaults.go`, `defaults_{darwin,freebsd,linux,openbsd,windows,other}.go` | HJSON or JSON |
| Identity / keys | `src/core/core.go`, `src/core/tls.go` | ed25519 keypair held in `Core{secret, public}`; derived from a TLS certificate in `New()` |
| Address derivation | `src/address/address.go` | `AddrForKey` / `SubnetForKey` derive IPv6 addresses/subnets from ed25519 public keys |
| Peer / transport management | `src/core/link.go` + `link_tcp.go` (`_darwin`/`_linux`/`_other`), `link_tls.go`, `link_quic.go`, `link_ws.go`/`link_wss.go`, `link_socks.go`, `link_unix.go` | One generic link manager, per-transport implementations |
| Routing | External module `github.com/Arceliar/ironwood` | No local routing engine; wired into `src/core` via `New()` / `api.go` / `proto.go` |
| Sessions | Ironwood-internal | Only exposed to operators via `src/admin/getsessions.go` |
| Multicast / local discovery | `src/multicast/` (`multicast.go`, `advertisement.go` + test, `admin.go`, `options.go`, OS variants) | |
| TUN integration | `src/tun/` (`tun.go`, `iface.go`, `admin.go`, `options.go`, OS variants) and `src/ipv6rwc/` (`ipv6rwc.go` + test, `icmpv6.go`) | `ipv6rwc` is the shim between TUN and Core |
| Admin / local API | `src/admin/admin.go` | Listens on unix or tcp socket; command handlers: `addpeer.go`, `removepeer.go`, `getpeers.go`, `getself.go`, `getsessions.go`, `getpaths.go`, `gettree.go`, `error.go`, `options.go` |
| Logging | Wired in `cmd/yggdrasil/main.go` via `github.com/gologme/log` and `github.com/hashicorp/go-syslog` | |
| Update mechanism | **None exists.** | No auto-update / self-update code anywhere in the tree — relevant if a verified-update subsystem is planned later |

## Packaging artifacts (`contrib/`)

| Platform/system | Files |
|---|---|
| systemd | `systemd/yggdrasil.service`, `yggdrasil.service.debian`, `yggdrasil-default-config.service(.debian)` |
| Debian | `deb/generate.sh` |
| macOS | `macos/create-pkg.sh`, `macos/yggdrasil.plist` |
| Windows MSI | `msi/build-msi.sh`, `msi/msversion.sh` |
| Docker | `docker/Dockerfile`, `docker/Dockerfile.multiarch`, `docker/entrypoint.sh` (root `Dockerfile` is a 1-line pointer to these) |
| OpenRC | `openrc/yggdrasil` |
| BusyBox init | `busybox-init/S42yggdrasil` |
| FreeBSD rc | `freebsd/yggdrasil` |
| AppArmor | `apparmor/usr.bin.yggdrasil(ctl)` |
| Ansible | `ansible/genkeys.go` |
| Mobile bindings (gomobile) | `mobile/*.go` |
| Versioning helper | `semver/{name,version}.sh` (used by `build`/`build.bat` to set `-ldflags`) |
| Logo | `logo/ygg-neilalexander.svg` |
| Standalone brute-force test tool | `yggdrasil-brute-simple/` (separate C project with its own `LICENSE`) |

## Tests

Sparse unit tests only; **no fuzz suite and no automated integration suite**
(`func Fuzz*` has zero hits in the tree):

- `src/address/address_test.go`
- `src/config/config_test.go`
- `src/core/core_test.go`, `src/core/link_ws_test.go`, `src/core/options_test.go`, `src/core/version_test.go`
- `src/ipv6rwc/ipv6rwc_test.go`
- `src/multicast/advertisement_test.go`
- `cmd/yggdrasil/chuser_unix_test.go`
- `contrib/mobile/mobile_test.go`

Manual integration scripts (not `go test`-driven): `misc/run-twolink-test`,
`misc/run-schannel-netns`.

## Dependencies (`go.mod`)

- Module: `github.com/yggdrasil-network/yggdrasil-go`
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

Only two legal files exist at the repository root today:

- `LICENSE` — LGPLv3, with a special exception permitting static/dynamic
  linking without triggering the LGPL's source-disclosure requirements for
  the linking application.
- `README.md` — project description, no separate copyright/attribution notice.

There is **no `NOTICE`, `SECURITY.md`, or `CONTRIBUTING.md`** in this checkout.
This matters directly for any future rebrand: Uqda's own documentation must
state plainly that it derives from Yggdrasil (and, transitively, from
Ironwood), and must not remove or obscure the existing LGPLv3 license terms —
but there is no pre-existing NOTICE-style attribution text to preserve beyond
the LICENSE file itself. Any Uqda-side attribution note will need to be
written from scratch, not migrated from an existing file.
