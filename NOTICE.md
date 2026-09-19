# Notice and attribution

Uqda Core (عُقدة) is independently maintained and derived from the open-source
Yggdrasil implementation at `github.com/yggdrasil-network/yggdrasil-go`.
No upstream endorsement or affiliation is implied.

## License and upstream origin

[LICENSE](LICENSE) contains the authoritative LGPLv3 terms and linking exception.
Existing copyright notices, license texts and authorship remain applicable.
Yggdrasil's inherited addressing, peering, administration and host networking
capabilities are upstream work. [Upstream release history](docs/upstream-changelog.md)
preserves the inherited public release entries; [CHANGELOG](CHANGELOG.md) covers Uqda.

## Dependencies

`github.com/Arceliar/ironwood` supplies routing and encrypted sessions and is
licensed under MPL-2.0. `github.com/Arceliar/phony` supplies actor concurrency and
is also licensed under MPL-2.0. Their names and authorship are unchanged.

`go.mod` and `go.sum` identify additional direct and transitive dependencies.
Each dependency's license is authoritative for that component. Distributors must
include the applicable notices and meet the terms for their actual dependency
set; this summary is not an exhaustive third-party license inventory.

The standalone `contrib/yggdrasil-brute-simple` tool has its own license.
The upstream logo in `docs/upstream-assets/ygg-neilalexander.svg` is retained
as source provenance and is not presented as Uqda branding.
