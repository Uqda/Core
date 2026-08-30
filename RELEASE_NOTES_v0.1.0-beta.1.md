# UQDA Core v0.1.0-beta.1

This is the first public beta of UQDA Core. It is intended for testing and
evaluation. It has not received an independent security audit and must not be
treated as an anonymity system or as production-ready security software.

## Highlights

- backward-compatible hardened handshake negotiation for protocol 0.5 peers;
- strict `secure=required` mode for controlled links;
- Windows MSI installers for x64, x86 and ARM64;
- macOS packages for Intel and Apple Silicon;
- Debian packages for amd64, i386, ARM, ARM64 and MIPS;
- integrated EdgeOS packages for EdgeRouter X and EdgeRouter Lite;
- integrated VyOS packages for amd64 and i386;
- portable Linux, FreeBSD and OpenBSD archives;
- English and Arabic documentation.

## Verification

The release workflow builds every attached file from the release commit,
validates native package metadata, and publishes `SHA256SUMS`. CI runs unit
tests on Linux, Windows and macOS with Go 1.25 and 1.26, cross-builds FreeBSD
and OpenBSD, runs lint and CodeQL, scans reachable dependencies, and fuzzes the
handshake parser.

Cross-built router, FreeBSD and OpenBSD artifacts still require testing on real
hardware or virtual machines. Successful compilation is not a guarantee of
runtime or driver compatibility on every device revision.

## Compatibility

New nodes remain compatible with protocol 0.5 peers. Hardened negotiation is
used automatically when supported by both peers. Add `?secure=required` to both
the peer and listener URI to reject legacy handshakes on managed links.

See `SECURITY.md`, `README.md`, `README_AR.md` and `CHANGELOG.md` before use.
