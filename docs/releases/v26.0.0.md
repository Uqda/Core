# Uqda 26

Uqda 26 is the first general release of Uqda Core.

## What is included

- Encrypted IPv6 networking interoperable with the Yggdrasil v0.5.14 wire
  protocol, the `uqda` daemon, `uqdactl`, and strict configuration checking.
- Debian packages for amd64, i386, arm64, armhf, armel, mips, and mipsel;
  Windows MSI packages for x64, x86, and ARM64; and macOS packages for Intel
  and Apple Silicon.
- Portable Linux archives for amd64 and arm64, used by the quick installer on
  supported systemd distributions. The installer verifies `SHA256SUMS`, keeps
  the node identity on ordinary updates/removal, refuses unmanaged installs,
  and requires explicit confirmation to purge that identity.
- Android AAR, Apple XCFramework, and a vendored source archive.

The Homebrew Cask is maintained in
[Uqda/homebrew-tap](https://github.com/Uqda/homebrew-tap). Update its version
and checksums to these release assets before announcing the new Homebrew
command. The general-release container image is a separate publication step.

## Installation and updates

Follow the [installation guide](https://github.com/Uqda/Core/blob/v26.0.0/docs/installation.md)
and [quick-install guide](https://github.com/Uqda/Core/blob/v26.0.0/docs/quick-install.md).
On Linux, download the
installer from the reviewed `main` branch, inspect it, then run its `install`,
`update`, `status`, or `uninstall` action as appropriate. On macOS, use
`brew install --cask Uqda/tap/uqda` once the tap points to `v26.0.0`.

## Validation and limitations

The release candidate passed cross-platform Go tests, vulnerability and
dependency scanning, native package builds, and package lifecycle smoke tests.
Separate two-host field tests passed on Ubuntu 24.04 and Fedora 44 with a
checksummed candidate archive. The published assets and their upgrade paths
must be tested again after the final tag; CI results alone do not establish
that every environment is free of defects.

The intermittent upstream idle-session delay was not conclusively reproduced
or fixed. Native installers remain unsigned unless signing is added before
publication; verify the attached `SHA256SUMS` and install only from the
official release. The administration API has no authentication: keep its
socket private and do not expose its TCP endpoint to untrusted networks.
This release does not claim anonymity or an external security audit. Report
security issues privately via the [security policy](https://github.com/Uqda/Core/blob/main/SECURITY.md).
