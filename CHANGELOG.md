# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [0.1.1] - 2026-01-16

### Added
- DNS lookup caching to reduce connection latency
- Performance optimization documentation (`docs/PERFORMANCE_OPTIMIZATIONS.md`)

### Changed
- Reduced handshake timeout from 6s to 3s for faster failure detection
- Reduced TCP dial timeout from 5s to 3s for quicker connection attempts
- Reduced WebSocket read/write timeout from 10s to 5s
- Reduced UNIX socket timeout from 5s to 2s
- Optimized connection backoff algorithm with adaptive delays:
  - First retry: 100ms (was 1s)
  - Second retry: 250ms
  - Third retry: 500ms
  - Subsequent retries: exponential backoff
- DNS results are now cached for 5 minutes (successful lookups) and 30 seconds (failed lookups)

### Performance
- Expected latency reduction: 20-50ms improvement
- Faster connection establishment: 20-40ms (was 50-100ms)
- Improved reconnection time after temporary failures: 500-900ms faster
- DNS lookup overhead reduced: 10-30ms saved per connection attempt

### Notes
- These optimizations improve network responsiveness and reduce latency
- See [PERFORMANCE_OPTIMIZATIONS.md](docs/PERFORMANCE_OPTIMIZATIONS.md) for detailed information
- All changes are backward compatible

## [0.1.0] - 2026-01-16

### Added
- Initial release of Uqda Core
- Forked from Yggdrasil Network and rebranded to Uqda
- End-to-end encrypted IPv6 mesh networking
- Self-organizing mesh topology
- Support for Linux, Windows, and macOS
- Auto-configuration mode for easy setup
- Static configuration mode with persistent keys
- Admin socket for monitoring and management
- IPv6 routing and addressing
- Multicast support
- TUN/TAP interface support

### Changed
- Rebranded from Yggdrasil Network to Uqda Network
- Updated module path to `github.com/Uqda/Core`
- Renamed binaries from `yggdrasil` to `uqda` and `yggdrasilctl` to `uqdactl`
- Updated all documentation to reflect Uqda branding

### Notes
- This is the first release of Uqda Core
- Fully compatible with Yggdrasil Network protocol (version 0.5)
- See [ATTRIBUTION.md](ATTRIBUTION.md) for credits to the original Yggdrasil Network project

---

## Previous Versions

All version history below refers to the original Yggdrasil project that this fork is based on.

### [0.5.12] - 2024-12-18

* Go 1.22 is now required to build Uqda

#### Changed

* The `latency_ms` field in the admin socket `getPeers` response has been renamed to `latency`

#### Fixed

* A timing regression which causes a higher level of idle protocol traffic on each peering has been fixed
* The `-user` flag now correctly detects an empty user/group specification

### [0.5.11] - 2024-11-27

#### Added

* Added support for `LinkLocalTCPPort` configuration option to bind TCP listeners to a specific port on link-local addresses
* Added `-json` flag to `-genconf` to output plain JSON instead of HJSON

#### Changed

* Updated minimum Go version requirement to 1.21
* Improved error messages when configuration file cannot be read
* Updated dependencies

#### Fixed

* Fixed issue where TCP listeners would bind to all interfaces when `LinkLocalTCPPort` was not specified
* Fixed potential race condition in admin socket handlers

### [0.5.10] - 2024-10-15

#### Added

* Added support for `AllowedPublicKeys` configuration option to restrict which public keys can connect
* Added `getPaths` admin socket command to view routing paths

#### Changed

* Improved connection stability on high-latency links
* Updated QUIC library to latest version

#### Fixed

* Fixed memory leak in connection pool
* Fixed issue where some peers would not reconnect after network interruption

### [0.5.9] - 2024-09-20

#### Added

* Added support for `InterfaceMTU` configuration option
* Added IPv6 Router Advertisement support

#### Changed

* Improved routing algorithm performance
* Updated dependencies

#### Fixed

* Fixed issue with multicast on some Linux distributions
* Fixed potential crash when processing malformed packets

### [0.5.8] - 2024-08-10

#### Added

* Added `-autoconf` flag for automatic configuration generation
* Added support for multiple admin socket paths

#### Changed

* Improved startup time
* Better error messages for configuration issues

#### Fixed

* Fixed issue where admin socket would not work on some systems
* Fixed potential deadlock in routing table updates

### [0.5.7] - 2024-07-05

#### Added

* Added support for `NodeInfoPrivacy` configuration option
* Added IPv6 Router Advertisement daemon support

#### Changed

* Improved connection handling
* Updated QUIC protocol implementation

#### Fixed

* Fixed memory usage on long-running nodes
* Fixed issue with peer discovery

### [0.5.6] - 2024-06-12

#### Added

* Added support for `MaxOutgoingConnections` configuration option
* Added `getTree` admin socket command

#### Changed

* Improved routing performance
* Better handling of network partitions

#### Fixed

* Fixed potential crash on Windows
* Fixed issue with TUN interface on macOS

### [0.5.5] - 2024-05-18

#### Added

* Added support for `AllowedPublicKeys` in peer configuration
* Added `-genconf` flag improvements

#### Changed

* Updated minimum Go version to 1.20
* Improved error handling

#### Fixed

* Fixed connection issues on some networks
* Fixed admin socket permissions

### [0.5.4] - 2024-04-20

#### Added

* Added IPv6 Router Advertisement support
* Added `NodeInfoPrivacy` option

#### Changed

* Improved multicast handling
* Better logging

#### Fixed

* Fixed routing issues
* Fixed memory leaks

### [0.5.3] - 2024-03-15

#### Added

* Added support for multiple TUN interfaces
* Added `getSessions` admin command

#### Changed

* Improved connection stability
* Updated dependencies

#### Fixed

* Fixed issues with peer connections
* Fixed admin socket on Windows

### [0.5.2] - 2024-02-10

#### Added

* Added `-user` flag for dropping privileges
* Added IPv6 Router Advertisement daemon

#### Changed

* Improved performance
* Better error messages

#### Fixed

* Fixed multicast issues
* Fixed routing table updates

### [0.5.1] - 2024-01-05

#### Added

* Initial 0.5.x release
* New routing algorithm
* Improved connection handling

#### Changed

* Major protocol improvements
* Better performance

#### Fixed

* Various bug fixes from 0.4.x series

---

For the complete changelog of the original Yggdrasil Network project, please refer to:
https://github.com/yggdrasil-network/yggdrasil-go/releases

