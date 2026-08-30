# UQDA for EdgeOS and VyOS

This directory contains the UQDA adaptation of the Vyatta integration from
[`neilalexander/vyatta-yggdrasil`](https://github.com/neilalexander/vyatta-yggdrasil).
The integration files remain licensed under GPL-3.0; see
`LICENSE.upstream.md` and `debian/copyright`.

The package adds:

- `set interfaces uqda tunX` configuration commands;
- one `uqda@tunX.service` instance per configured interface;
- `restart uqda tunX` operational commands;
- optional EdgeOS firewall and IPv6 masquerade hooks;
- isolated configuration and admin sockets for each interface.

Build from the repository root:

```sh
PKGARCH=mipsel PKGTARGET=edgeos2x sh contrib/vyatta/generate.sh
PKGARCH=mips PKGTARGET=edgeos2x sh contrib/vyatta/generate.sh
PKGARCH=amd64 PKGTARGET=vyos13 sh contrib/vyatta/generate.sh
PKGARCH=i386 PKGTARGET=vyos13 sh contrib/vyatta/generate.sh
```

These packages target Vyatta-derived systems. For ordinary Debian/Ubuntu,
use `contrib/deb/generate.sh` instead.
