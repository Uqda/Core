# Third-party licenses

Uqda Core's own code is licensed under LGPLv3 with a linking exception
([LICENSE](LICENSE)). The release binaries also contain the components below,
each under its own license. The full license text of every Go module is in
that module's source, at the version pinned in [go.mod](go.mod) and
[go.sum](go.sum), and on <https://pkg.go.dev>.

CI checks this list on every change (`contrib/licenses/check.sh`): a new
dependency fails the build until it is listed here with a license from the
allowed set (MIT, BSD-2-Clause, BSD-3-Clause, ISC, Apache-2.0, MPL-2.0, Zlib).

## Go modules compiled into the binaries

| Module | Version | License | Notes |
|---|---|---|---|
| `github.com/Arceliar/ironwood` | pinned snapshot | MPL-2.0 | Vendored in [third_party/ironwood](third_party/ironwood) with a local fix; source of modified files available there ([UQDA-PATCHES](third_party/ironwood/UQDA-PATCHES.md)) |
| `github.com/Arceliar/phony` | v0.0.0-20220903101357 | MPL-2.0 | Unmodified |
| `github.com/VividCortex/ewma` | v1.2.0 | MIT | |
| `github.com/bits-and-blooms/bitset` | v1.24.5 | BSD-3-Clause | |
| `github.com/bits-and-blooms/bloom/v3` | v3.7.1 | BSD-2-Clause | |
| `github.com/cespare/xxhash/v2` | v2.3.0 | MIT | |
| `github.com/cheggaaa/pb/v3` | v3.2.1 | BSD-3-Clause | |
| `github.com/clipperhouse/displaywidth` | v0.10.0 | MIT | |
| `github.com/clipperhouse/uax29/v2` | v2.7.0 | MIT | |
| `github.com/coder/websocket` | v1.8.15 | ISC | |
| `github.com/fatih/color` | v1.19.0 | MIT | |
| `github.com/gologme/log` | v1.3.0 | BSD-3-Clause | |
| `github.com/hashicorp/go-syslog` | v1.0.1 | MIT | |
| `github.com/hjson/hjson-go/v4` | v4.7.1 | MIT | |
| `github.com/kardianos/minwinsvc` | v1.0.2 | Zlib | Windows only |
| `github.com/mattn/go-colorable` | v0.1.15 | MIT | |
| `github.com/mattn/go-isatty` | v0.0.24 | MIT | |
| `github.com/mattn/go-runewidth` | v0.0.27 | MIT | |
| `github.com/olekukonko/cat` | v0.0.0-20250911104152 | MIT | |
| `github.com/olekukonko/errors` | v1.2.0 | MIT | |
| `github.com/olekukonko/ll` | v0.1.6 | MIT | |
| `github.com/olekukonko/tablewriter` | v1.1.5 | MIT | |
| `github.com/quic-go/quic-go` | v0.63.0 | MIT | |
| `github.com/vishvananda/netlink` | v1.3.1 | Apache-2.0 | Linux only |
| `github.com/vishvananda/netns` | v0.0.5 | Apache-2.0 | Linux only |
| `github.com/wlynxg/anet` | v0.0.5 | BSD-3-Clause | Android |
| `golang.org/x/crypto` | v0.58.0 | BSD-3-Clause | |
| `golang.org/x/net` | v0.61.0 | BSD-3-Clause | |
| `golang.org/x/sys` | v0.49.0 | BSD-3-Clause | |
| `golang.org/x/text` | v0.43.0 | BSD-3-Clause | |
| `golang.zx2c4.com/wintun` | v0.0.0-20230126152724 | MIT | Go bindings only; the driver DLL is listed below |
| `golang.zx2c4.com/wireguard` | v0.0.0-20260522210424 | MIT | TUN support |
| `golang.zx2c4.com/wireguard/windows` | v1.1.1 | MIT | Windows only |
| `rsc.io/qr` | v0.2.0 | BSD-3-Clause | |
| `suah.dev/protect` | v1.2.4 | ISC | OpenBSD pledge/unveil |
| Go standard library | toolchain version | BSD-3-Clause | |

## Files shipped next to the binaries

| File | Where | License |
|---|---|---|
| `wintun.dll` 0.14.1 | Windows MSI | WireGuard LLC *Prebuilt Binaries License*, shipped unmodified as `wintun-LICENSE.txt` beside it. It may be redistributed only together with software that uses it through its published API, as Uqda does, and must not be modified. Wintun and WireGuard are trademarks of their owners, used here only to name the component. |

`contrib/uqda-brute-simple` is a separate tool with its own license and is not
part of the release binaries.

## Obligations this list covers

- **MPL-2.0 (Ironwood, Phony):** source of these components, including Uqda's
  changes to Ironwood, is available in this repository under MPL-2.0.
- **Apache-2.0 (netlink, netns):** neither project ships a NOTICE file; their
  license texts apply.
- **BSD, MIT, ISC, Zlib:** copyright notices and license texts are reproduced
  by this file together with the module sources it points to.

This inventory is maintained in good faith and is not legal advice. Where it
and a component's own license differ, the component's license applies.
