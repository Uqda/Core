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
  These browser assertions subsequently passed on native amd64 and arm64 CI.

## Follow-up: software-only validation

The user requested testing without buying or providing a physical device.
Umbrel officially supports an EFI VM and ships a Docker-based `umbrel-dev`
environment. Physical hardware is not a prerequisite for app-manager validation.
The development container shares the runner kernel; it does not establish real
production OS boot/reboot or hardware-specific behavior.

GitHub access recovered. The changes were pushed to the existing draft PR #22.
At `8e9edd649763f90aa4912615b7c22b7379259e46`, both native amd64 and arm64
Docker/TUN/browser gates passed, as did 21 Python lifecycle/auth/export tests,
the UI state regression suite and multiarchitecture wrapper build:
https://github.com/Uqda/Core/actions/runs/37127082266

The expanded network gate creates a second independent node, transfers and
hash-verifies 1 MiB of IPv6 TCP in each direction, exchanges UDP in each direction,
verifies wrong private-group password rejection with a live listening endpoint,
then verifies recovery with the correct password. The temporary peer/listeners are
removed and the original host node retains its identity. It uses disposable CI
containers, not the user's live servers.

The follow-up also fixes percent-encoded credential parameter names in peer URLs.
Credential-bearing/unknown query options disable browser editing and are redacted;
the original on-disk configuration is preserved. Regression tests cover encoded
password/secret and token parameters.

The actual Umbrel 2.0 app-manager gate uses the official release source pinned to
`9298257b0e904ca8d8270b1702672666343f0b86`, in an isolated upstream worktree and
development container. It tests the existing digest-pinned published store image,
not an unpublished replacement. The initial environment built and started, but
its three lifecycle tests failed with `ECONNREFUSED 127.0.0.1:8926`:
https://github.com/Uqda/Core/actions/runs/37126228987

Source inspection identified the harness problem: the upstream test factory sets
`port: 0`, which deliberately skips LAN ingress in Umbrel 2.0. The revised fixture
starts the actual instance on the production internal port before registration,
then waits for the real app gateway's authentication redirect. No replacement
gateway or mocked app manager is introduced. A rerun is required to prove the fix.

The next run built the environment but failed during the upstream HTTPS CA setup:
https://github.com/Uqda/Core/actions/runs/37128409855
OpenSSL tried to read the nonexistent Homebrew `openssl.cnf`; all four app tests
were skipped after that setup error, not passed. The revised CI command explicitly
uses the real Debian `/etc/ssl/openssl.cnf`. It does not bypass TLS verification.
The isolated pre-test development service also has a documented 30-second stop
deadline; it does not certify graceful production shutdown.

The subsequent run passed HTTPS setup but exposed another harness-order issue:
https://github.com/Uqda/Core/actions/runs/37129687909
With `autoStart:false`, the upstream helper constructed API URLs containing an
undefined port; registration failed and four app tests were skipped. The fixture
now first initializes its real API helpers, then stops/restarts the actual server
on the production internal port before registration. No API response is mocked.

That gate distinguishes a manifest-only app update from a binary/image upgrade.
The first fixture isolated dashboard-password tests by disabling the outer auth
layer in temporary test data. The revised test never disables either layer: it
uses the official factory's real owner browser-session cookie jar, checks the
anonymous gateway redirect, checks that owner access alone still requires the
dashboard password, then authenticates the dashboard. It also checks manifest
version changes and uninstall/fresh-install identity replacement. These expanded
assertions passed against the published image in all four lifecycle tests:
https://github.com/Uqda/Core/actions/runs/37130361331

The review additionally found a cookie-replay boundary between co-hosted apps:
cookies do not isolate ports, and the legacy Umbrel network is shared. The source
now requires an origin-scoped, tab-held random proof for session bootstrap and
status, as well as mutations. A cookie alone cannot recover that proof. Added
HTTP regressions and a real-browser second-origin cookie-capture fixture cover
this boundary. They do not claim HTTPS transport protection or isolate all apps'
network traffic. This fix is not in the existing published wrapper image.

The platform workflow now separately tests a current-source wrapper, built and
digest-pinned in a disposable loopback-only registry, against the unchanged real
Umbrel app manager. It does not publish an external image or alter the public
store. Consult its completed result before treating the new source as validated.

At `3af5503`, all four app-manager lifecycle tests passed for both the existing
published image and the freshly rebuilt source, including the new cookie-proof
boundary in the source matrix:
https://github.com/Uqda/Core/actions/runs/37130948499

At `d7d248a`, 24 real-daemon/auth/configuration/export tests, native amd64 and
arm64 Docker/TUN/network/browser tests, UI state tests and the PR's fresh
multiarchitecture build passed:
https://github.com/Uqda/Core/actions/runs/37131151836
The matching push run also passed; its image publication step leaves an existing
tag untouched, so its green build-job label alone is not a fresh-build claim:
https://github.com/Uqda/Core/actions/runs/37131145599

Screenshots from that PR run were downloaded and visually reviewed in English,
Arabic RTL and a 390-pixel mobile viewport. The actual browser also visits a
second local HTTP origin, confirms it receives the port-independent test cookie,
verifies it cannot read the tab-local proof or bootstrap/read with that cookie,
then returns to the dashboard and reloads without losing the authenticated view.

The expanded source matrix switches to the previous published image and then
upgrades to the newly built source digest through the real manager. It verifies
both containers' actual configured image, retained identity and the new read-auth
boundary. Both matrices passed all four lifecycle tests at `d7d248a`:
https://github.com/Uqda/Core/actions/runs/37131145571

## Remaining gates before publication

1. Before publishing wrapper changes, increment the wrapper version (do not
   overwrite `26.0.4-umbrel.1`), update manifest and both compose image tags, build
   after successful gates, then pin the newly published digest. Keep old packages
   available for existing installations.
2. Update the dedicated store from that verified export, with release notes and
   screenshots from the actual browser test. Do not label unexecuted visual QA
   as complete.
3. Production Umbrel VM boot/reboot, backup restoration, full owner-browser proxy
   authentication and external/network-storage behavior remain unverified. A real
   VM can validate production lifecycle without needing physical hardware.
4. Native arm64 application runtime now passes in Docker; that is not a Raspberry
   Pi firmware/kernel or arm64 production Umbrel lifecycle test.

No live server, firewall, private-group configuration, existing release or
published store package was changed during this local review.
