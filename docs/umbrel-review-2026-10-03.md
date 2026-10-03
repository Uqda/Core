# Umbrel integration review — 2026-10-03

## Scope and baseline

Reviewed the public `Uqda/umbrel-app-store` package and Core PR #22 at
`877ef28a0f465b2dea0effa471372a8751d0904a`. The implementation is not on Core's
main branch. The store currently installs the existing immutable
`26.0.4-umbrel.1` image; the local changes below are not in that image.

Official references consulted:

- https://github.com/getumbrel/umbrel-community-app-store — store/app ID prefix,
  directory structure and installation through community stores.
- https://github.com/getumbrel/umbrel/releases/tag/2.0.0 — current official release.
- https://github.com/getumbrel/umbrel/blob/2.0.0/packages/umbreld/source/modules/apps/app.ts
  — declared data-root mounts and lifecycle integration. Source inspection is
  not an actual compatibility test.

The manifest's `uqda-network` ID matches store prefix `uqda`. Configuration
persists separately from the privileged Core runtime; the unprivileged dashboard
uses the restricted control socket. The package pins both services to the same
multiarchitecture image digest and uses the Umbrel app proxy. Existing Linux
container tests cover TUN, permissions and identity retention, but their previous
success does not validate this review's changes or physical Umbrel behavior.

## Local fixes

- Immediately unlock all settings after saving/restarting, including error paths.
  Lock the entire settings form during restart and prevent logout racing changes.
- Clear unsaved passwords, cached identity and configuration on logout/session
  expiry. Ignore delayed status responses after logout or configuration mutation.
- Add an Arabic/English usage guide explaining private-group setup, separate
  end-to-end/service checks, IPv6 requirements and identity/backup safeguards.
- Correct installation documentation to the dedicated community-store URL.
- Use portable ZIP entry paths and update previously pinned icon URLs during export.
- Export only approved package files, never recursively copy configuration,
  private keys or backups. Reject mismatched image versions before exporting.

## Validation performed locally

- Dependency-free UI regression suite passed: save/restart controls, validation
  failure, logout/session expiry cleanup, stale polling and guide translations.
  The same suite intentionally fails against the old implementation at the
  post-save settings-lock assertion, confirming the original regression.
- Four portable Python export tests passed: exact package entries, both image
  digests, icon revision, invalid input, mismatched version and secret exclusion.
- JavaScript syntax checks, digest-required package check and `git diff --check`
  passed.
- Added real-browser regression assertions and wired UI state tests into CI.
  These new browser assertions have not been executed yet.

## Remaining gates before publication

1. Commit/push the reviewed changes and run the complete Linux daemon, Docker/TUN
   and actual browser CI gates. Remote operations were interrupted by an automatic
   approval-review usage-limit failure; no bypass was attempted.
2. Before publishing wrapper changes, increment the wrapper version (do not
   overwrite `26.0.4-umbrel.1`), update manifest and both compose image tags, build
   after successful gates, then pin the newly published digest. Keep old packages
   available for existing installations.
3. Update the dedicated store from that verified export, with release notes and
   screenshots from the actual browser test. Do not label unexecuted visual QA
   as complete.
4. Actual Umbrel installation, device reboot, app update/removal/restoration,
   proxy authentication and encrypted traffic remain unverified. The user chose
   development/GitHub validation and did not supply an Umbrel device.
5. Native arm64 runtime and external/network-storage behavior require separate
   validation; building arm64 images alone does not establish runtime support.

No live server, firewall, private-group configuration, existing release or
published store package was changed during this local review.
