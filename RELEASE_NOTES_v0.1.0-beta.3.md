# UQDA Core v0.1.0-beta.3

This prerelease contains the malformed-handshake safety fix and verified
cross-platform installation work from beta.2, plus a corrected update channel.

## Security

- Rejects missing or invalid Ed25519 public keys before signature verification.
- Adds regression and fuzz coverage for malformed hello and confirmation data.
- Verifies every downloaded installer and package against release SHA-256 sums.

## Installation and updates

- `install.sh` selects native packages for systemd Linux, macOS, EdgeOS and
  VyOS, or a matching portable archive for supported Linux/BSD targets.
- `updater.sh` reads the release channel from `VERSION`, then follows the same
  checksum-verified native upgrade path while preserving configuration.

This is experimental software and has not received an independent security
audit. Router packages are cross-built and inspected in CI; they have not been
validated here on every physical router model. macOS packages are not yet
notarized and release assets are not yet signed with a separate project key.
