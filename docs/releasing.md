# Publishing a stable release

Stable releases are intentionally driven by one versioned release-notes file.
No Apple Developer account is required.

## Normal release path

1. Choose a stable SemVer tag such as `v0.1.2`.
2. Finish the code, tests, documentation, and `CHANGELOG.md` changes in a pull
   request.
3. Add `.github/releases/v0.1.2.md` with the user-facing release notes. The
   filename is the release version; there is no second version file to update.
4. Merge only after all required checks pass.
5. The `Stable Release` workflow selects the newest stable release-notes file,
   reruns the release quality gate, builds every platform package, creates and
   signs `SHA256SUMS`, creates provenance attestations, and publishes the GitHub
   release and tag.
6. Verify that the workflow and published release succeeded before announcing
   the version.

The publisher refuses to overwrite an existing release. Editing old notes will
therefore never silently replace published assets.

## Manual run or retry

Open **Actions → Stable Release → Run workflow**. Enter the exact tag, or leave
the tag blank to select the newest stable release-notes file. A retry uses the
same full quality gate and still refuses to overwrite an existing release.

## macOS without a paid Apple account

When Apple credentials are absent, the workflow publishes explicitly named
`*-unsigned.pkg` files and all normal checksum, Sigstore, and provenance
verification still applies. If paid Apple credentials are added later, the
same workflow automatically signs and notarizes the macOS binaries and
installer. Never disable Gatekeeper globally.
