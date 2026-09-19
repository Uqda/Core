# Packaging

All packages install `uqda` and `uqdactl` and use `uqda.conf`. The version helper
reads `src/version/VERSION`; artifact names retain the canonical machine version.
Debian control metadata replaces the prerelease hyphen with `~` for correct
ordering. MSI and macOS require numeric installer versions: Beta N maps to
`YY.minor.N` (N from 1 to 999), general release to `YY.minor.1000`. Product titles
still use the annual human name. Keep numeric mappings monotonic within a generation.

## Identity lifecycle

Unix installers and the container entrypoint use `contrib/packaging/install-config.sh`.
It validates existing Uqda configuration without rewriting it. On first install,
it validates and copies a single legacy configuration or explicitly generates
one when no source exists. Ambiguous sources, invalid keys or failed validation
abort. Temporary files are private and publication refuses to overwrite an
existing destination. The original configuration remains available as a backup.
External PEM key paths are preserved and require operator-managed migration.

Debian removal stops the service but retains configuration and keys, including
on purge. macOS configuration is created by the install script, outside the
package payload. MSI configuration is outside tracked components and has no
RemoveFile action. Container identity persists only when its volume is retained.
Service-only installations require an existing validated configuration.

## Windows MSI maintainers

The UpgradeCode and component GUIDs in `contrib/msi/build-msi.sh` identify Uqda
as a distinct Windows product. They must not be replaced with upstream Yggdrasil
GUIDs. MajorUpgrade applies only to the Uqda product line, not to an upstream MSI.
Keep these identifiers stable across Uqda upgrades. x64 and ARM64 intentionally
share the 64-bit product line; they are alternative installations, not side-by-side
products. x86 has its own UpgradeCode.

WiX v3 `candle` and `light` are required. The deferred configuration action runs
with installer privileges, limits the new identity directory to Administrators
and SYSTEM, checks errors and aborts rather than generating a key after a failed
migration. Stop any old daemon before starting Uqda with the migrated identity.
Validate clean install, Uqda upgrade, upstream coexistence/migration, invalid
legacy config, external key permissions and uninstall on Windows before shipping.

## Platform validation

Debian uses `dpkg-deb`; macOS uses native `pkgbuild` and `productbuild`; Windows
uses WiX and the checksummed Wintun download. Docker builds use the real root
Dockerfile or the multiarchitecture definition under `contrib/docker`.
Workflow success and native install/upgrade/uninstall execution must be recorded
in release review. Shell syntax checks or cross-compilation alone do not certify
installers. Router-specific third-party packaging is not supplied by this repository.
