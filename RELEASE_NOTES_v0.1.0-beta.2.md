# UQDA Core v0.1.0-beta.2

This prerelease fixes a remotely reachable malformed-handshake panic found by
the CI fuzz smoke test and adds checksum-verified installation and update tools.

## Security

- Rejects missing or invalid Ed25519 public keys before signature verification.
- Adds regression tests for malformed hello and confirmation messages.
- Keeps protocol 0.5 compatibility by negotiating the hardened confirmation
  extension only when both peers advertise support.

## Installation

- `install.sh` selects native packages for systemd Linux, macOS, EdgeOS and
  VyOS, or a matching portable archive for supported Linux/BSD targets.
- `updater.sh` preserves the selected release channel and reuses the same native
  package upgrade path.
- Every downloaded installer and package is verified against `SHA256SUMS`.

This is experimental software and has not received an independent security
audit. Router packages are cross-built and inspected in CI; they have not been
validated here on every physical router model.
