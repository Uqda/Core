# Release identity

`src/version/VERSION` is authoritative for the product machine version. The Go
package embeds it; `contrib/semver/version.sh` derives packaging and display forms.
Do not derive Uqda versions from upstream tags or wire protocol metadata.

| Meaning | Prepared value |
|---|---|
| Product/release title | Uqda 26 |
| Core display | Uqda Core 26 |
| Machine version | `26.0.0` |
| Git tag | `v26.0.0` (not yet created) |
| GitHub release type | General release (pending validation) |

Beta 1 was published as a GitHub prerelease on 2026-09-27 after the required
validation gates passed. The `26.0.0` version on this branch is a release
candidate, not evidence that the general release has been published. Package
workflows prepare artifacts; create the tag and publish only after review,
merge, artifact verification, and native install/upgrade/remove validation.

The annual generation is 26 for 2026, 27 for 2027 and 28 for 2028. Public names
are `Uqda 26 Beta 1`, `Uqda 26 Beta 2`, `Uqda 26` and `Uqda 26.1`.
Beta tags retain `v26.0-beta.N`; the requested general-release tag is
`v26.0.0`. The trailing zero is a machine-version patch field and is omitted
from the public annual name. The general release must not be tagged until its
version file, packages, changelog and native install/upgrade/remove checks agree.

Package workflows prepare artifacts. Use `version.sh --title`, its default tag
output and `--prerelease` when creating release metadata. Verify the tag matches
the embedded version; use the reviewed changelog rather than a commit dump.
Debian uses `26.0~beta.1` for ordering; MSI/macOS numeric versions use `26.0.1`
for Beta 1 and `26.0.1000` for `v26.0.0`. These are installer encodings, not
alternate product versions. See [packaging](packaging.md).

Require formatting, build, vet, tests, vulnerability scanning, fuzz smoke,
independent upstream interoperability, mobile/native package validation and an
actually successful Linux race CI run before claiming their gates have passed.
