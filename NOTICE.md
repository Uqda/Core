# Notice and attribution

Uqda Core (عُقدة) is independently maintained and derived from the open-source
Yggdrasil implementation at `github.com/yggdrasil-network/yggdrasil-go`.
No upstream endorsement or affiliation is implied.

## Origin of this fork

On 2026-09-19 the `develop` branch of `github.com/yggdrasil-network/yggdrasil-go`
was imported unchanged as commit `27dd82b` ("Baseline import"). Every later
change is in the Git history of this repository, so the exact difference to
upstream can be reproduced with `git diff 27dd82b`.

Uqda's main modifications, all released under the same license:

- renamed binaries, packages, paths and service names (`uqda`, `uqdactl`,
  `/etc/uqda`), with migration from existing Yggdrasil configurations;
- an inbound firewall, private networks with group passwords, one-time invites,
  fresh-handshake key proofs, handshake and admin limits;
- `uqda setup`, `find`, `test`, `status` and automatic public peer selection;
- packaging for Debian, Windows MSI, macOS (Homebrew), Docker and Umbrel;
- a vendored, patched Ironwood (see below), documentation and tests.

## Names and trademarks

"Yggdrasil" is the name of an independent open-source project. Uqda uses it only
to describe network compatibility and the origin of this code. Uqda is not a
product of, or endorsed by, the Yggdrasil developers. "WireGuard" and "Wintun"
are trademarks of their owners and are used only to name the components Uqda
ships. "Uqda" and "عُقدة" are the names of this project.

## License and upstream origin

[LICENSE](LICENSE) contains the authoritative LGPLv3 terms and linking exception.
Existing copyright notices, license texts and authorship remain applicable.
Yggdrasil's inherited addressing, peering, administration and host networking
capabilities are upstream work. [Upstream release history](docs/upstream/changelog.md)
preserves the inherited public release entries; [CHANGELOG](CHANGELOG.md) covers Uqda.

## Dependencies

`github.com/Arceliar/ironwood` supplies routing and encrypted sessions and is
licensed under MPL-2.0. `github.com/Arceliar/phony` supplies actor concurrency and
is also licensed under MPL-2.0. Their names and authorship are unchanged.

`third_party/ironwood` preserves the pinned upstream snapshot, its copyright and
license, with a local fix for concurrent peer-latency inspection. Changes and
provenance are listed in [UQDA-PATCHES](third_party/ironwood/UQDA-PATCHES.md).

Every component compiled into or shipped with the release binaries, including
`wintun.dll` in the Windows installer, is listed with its license in
[THIRD_PARTY_LICENSES](THIRD_PARTY_LICENSES.md); CI keeps that list complete.
Each dependency's license is authoritative for that component. Release packages
install `LICENSE`, this file and the third-party list next to the binaries.

The standalone `contrib/uqda-brute-simple` tool has its own license.
The upstream logo in `docs/upstream/assets/upstream-logo-neilalexander.svg` is retained
as source provenance and is not presented as Uqda branding.
