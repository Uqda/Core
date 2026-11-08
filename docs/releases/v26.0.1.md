# Uqda 26.0.1

This patch release adds two operational checks: `uqdactl doctor` (also
`status`) for local health, and `uqdactl test <other-node IPv6>` for real
ICMPv6 reachability across the overlay. `test` reports received probes, loss,
first-probe completion time and median/p95 completion time. It accepts
`count=1` through `count=20` and an optional `idle=75s` interval (up to five
minutes). Use `-json` before the command for machine-readable output. Timing
includes process startup; it is not raw network RTT.

The Linux quick installer keeps the existing node identity during a normal
uninstall, reinstall or update. Native package versions have been advanced so
MSI and macOS packages can upgrade from 26.0.0. Verify downloads against the
release's `SHA256SUMS` before installation. Packages are not code-signed.

The release candidate passed the repository's cross-platform CI. Two live
nodes (Ubuntu 24.04 and Fedora 44) passed bidirectional 20/20 probe and 1 MiB
TCP transfer tests. A probe launched immediately after one service update lost
2/3 replies during reconnection, then returned to 20/20. This release does not
promise instant reachability after service startup, zero future packet loss,
anonymity or an independent security audit. Keep the admin API private.

See [installation](https://github.com/Uqda/Core/blob/v26.0.1/docs/installation.md)
and [quick installation](https://github.com/Uqda/Core/blob/v26.0.1/docs/quick-install.md).
