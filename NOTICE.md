# Attribution notice

UQDA Core is based on and derived from the open-source **Yggdrasil Network** implementation:

- Upstream project: https://github.com/yggdrasil-network/yggdrasil-go
- Upstream website: https://yggdrasil-network.github.io/

The UQDA codebase retains substantial architectural concepts and implementation lineage from Yggdrasil, including encrypted IPv6 overlay routing, cryptographic node identities, TUN integration, peer transports, administration facilities, packaging, and platform support. UQDA also depends on related routing components such as `github.com/Arceliar/ironwood`.

UQDA is maintained under its own name and repository. It is not the official Yggdrasil distribution, and no endorsement or affiliation with the Yggdrasil maintainers is implied.

The repository's license is provided in [LICENSE](LICENSE). Copyright and license notices belonging to upstream and third-party contributors remain applicable to their respective work.

The pinned Ironwood dependency is mirrored under `third_party/ironwood` with
its original license and copyright files. UQDA carries a minimal local actor
synchronization patch for peer debug snapshots until an equivalent fix is
available upstream.

The EdgeOS/VyOS integration under `contrib/vyatta` is adapted from
[`neilalexander/vyatta-yggdrasil`](https://github.com/neilalexander/vyatta-yggdrasil),
Copyright (C) Neil Alexander T., and is distributed under GPL-3.0. Its original
license and detailed copyright notice are retained in that directory.
