# Release identity

`src/version/VERSION` is authoritative for the product machine version. The Go
package embeds it; `contrib/semver/version.sh` derives packaging and display forms.
Do not derive Uqda versions from upstream tags or wire protocol metadata.

| Meaning | Prepared value |
|---|---|
| Product/release title | Uqda 26 Beta 1 |
| Core display | Uqda Core 26 Beta 1 |
| Machine version | `26.0-beta.1` |
| Reserved Git tag | `v26.0-beta.1` |
| GitHub release type | Pre-release |

Beta 1 is unpublished. Preparing metadata does not authorize a tag or release.
Community-requested capabilities and final validation must be ready before the
release owner publishes it. No workflow here creates a Git tag or GitHub release.

The annual generation is 26 for 2026, 27 for 2027 and 28 for 2028. Public names
are `Uqda 26 Beta 1`, `Uqda 26 Beta 2`, `Uqda 26` and `Uqda 26.1`.
The initial `.0` appears in machine/tag forms, not normally in product names.
This is a two-component annual version, not three-component semantic versioning.

Package workflows prepare artifacts. Use `version.sh --title`, its default tag
output and `--prerelease` when creating release metadata. Verify the tag matches
the embedded version; use the reviewed changelog rather than a commit dump.
Debian uses `26.0~beta.1` for ordering; MSI/macOS numeric versions use `26.0.1`
for Beta 1 and reserve revision 1000 for the general release. These are installer
encodings, not alternate product versions. See [packaging](packaging.md).

Require formatting, build, vet, tests, vulnerability scanning, fuzz smoke,
independent upstream interoperability, mobile/native package validation and an
actually successful Linux race CI run before claiming their gates have passed.
