# macOS release signing and notarization

When Apple Developer credentials are available, stable macOS installer packages
are signed with Apple Developer ID certificates and accepted by Apple's
notarization service before publication. Without those paid credentials, the
workflow publishes an explicitly named `-unsigned.pkg` and the release notes
must disclose that Finder will show a Gatekeeper warning.

## Required GitHub Actions secrets

Configure these repository or organization secrets:

- `APPLE_DEVELOPER_ID_APPLICATION_P12_BASE64`: Developer ID Application
  certificate and private key exported as a password-protected `.p12`, then
  Base64 encoded.
- `APPLE_DEVELOPER_ID_INSTALLER_P12_BASE64`: Developer ID Installer certificate
  and private key exported the same way.
- `APPLE_DEVELOPER_ID_P12_PASSWORD`: password used for both `.p12` exports.
- `APPLE_NOTARY_KEY_P8_BASE64`: App Store Connect API private key (`.p8`),
  Base64 encoded.
- `APPLE_NOTARY_KEY_ID`: App Store Connect API key ID.
- `APPLE_NOTARY_ISSUER_ID`: App Store Connect API issuer ID.

On macOS, encode each binary secret without line wrapping:

```sh
base64 -i developer-id-application.p12 | pbcopy
base64 -i developer-id-installer.p12 | pbcopy
base64 -i AuthKey_EXAMPLE.p8 | pbcopy
```

Never commit certificates, private keys, passwords, or decoded copies.

## Release gate

When signing credentials are configured, for each `amd64` and `arm64` package
the stable release workflow:

1. imports both Developer ID identities into an ephemeral keychain;
2. signs `uqda` and `uqdactl` with hardened runtime and a secure timestamp;
3. signs the installer with the Developer ID Installer identity;
4. submits the package to Apple with `notarytool --wait`;
5. staples and validates the notarization ticket;
6. requires Gatekeeper's install assessment to accept the package; and
7. deletes the temporary keychain even if a preceding step fails.

Before publishing, inspect the logs for `Accepted` from `notarytool` and verify
that `pkgutil --check-signature` names the expected Developer ID Installer team.

Without credentials, users should install through the checksum-verifying
`install.sh` path and must not be instructed to disable Gatekeeper globally.
