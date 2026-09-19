# Changelog

## Uqda 26 Beta 1

Unreleased. Machine version: `26.0-beta.1`; reserved tag: `v26.0-beta.1`.
This entry describes the intended first public beta and is not a publication notice.

### Highlights

- Uqda Core daemon `uqda`, control utility `uqdactl`, and canonical `uqda.conf`.
- Annual release identity shared by CLI output, build scripts and package metadata.
- Encrypted IPv6 networking using the existing Yggdrasil protocol and Ironwood.

### Security

- Unix warnings for unsafe configuration/key permissions and warnings for
  non-loopback TCP administration sockets.
- Ansible private-key vault files use owner-only permissions.
- Bounds checks and fuzz targets for handshake and multicast metadata parsing.

### Reliability

- Peer removal permits immediate re-addition; reconnect backoff includes jitter.

### Configuration

- `-checkconf` validates keys, peer/listener schemes, admin endpoints and
  multicast interface patterns without starting a node.
- Daemon-loaded configurations require an explicit valid identity; missing or invalid
  keys do not silently generate a replacement node.
- Packaging uses Uqda paths, stops on invalid/ambiguous migration sources, and
  retains persistent configuration on uninstall.

### Compatibility

- Mobile API breaking change: exported `mobile.Yggdrasil` becomes `mobile.Uqda`.
  Regenerate bindings and update clients; no legacy type alias is provided.
- Windows MSI uses a distinct Uqda product identity. Configuration migration is
  separate from Windows Installer upgrades; no in-place upstream MSI upgrade is claimed.

### Developer Experience

- Public architecture, security, compatibility, testing and packaging documentation.
- Black-box daemon tests, pinned upstream v0.5.14 interoperability with real IPv6
  packet delivery, TCP/TLS reconnect and multi-hop checks, and Linux race CI.

### Known Limitations

- The admin API has no authentication; protect its socket and keep it local.
- The independent packet gate uses the real IPv6 packet interface without kernel
  TUN; native OS network integration requires separate platform validation.
- Android/Apple binding artifacts and native installers require their platform
  toolchains and validation before release. Authored CI does not establish a pass.
- No anonymity guarantee or independent external security audit is claimed.

Inherited public upstream release notes are preserved in
[upstream release history](docs/upstream-changelog.md).
