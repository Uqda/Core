# Security Policy

## Scope and honesty statement

Uqda Core is a hardened, actively-developed fork of `yggdrasil-go`. As of
this document, it has **not** undergone an independent external security
audit. Claims in this document and elsewhere in the repository describe
what has actually been done (fuzz targets run, tests written, code
reviewed) — see [docs/THREAT_MODEL.md](docs/THREAT_MODEL.md) for the
specific, current state of each identified threat, including the ones that
remain open. Uqda does not claim to be "impossible to hack," does not
claim to provide anonymity, and does not claim protection it has not
actually verified. If you find a place where documentation overstates what
has been done, that is itself worth reporting.

Uqda protects the networking session between nodes according to the
Yggdrasil-compatible protocol. It does not secure services you choose to
run on top of it — a host reachable over Uqda still needs its own
firewalling, authentication, and hardening.

## Supported versions

This repository is pre-1.0 and under active restructuring (see
[docs/IMPLEMENTATION_STATUS.md](docs/IMPLEMENTATION_STATUS.md)). Until a
first tagged Uqda release exists, the only supported version is the
current `develop`/default branch — there is no backport policy yet.

## Reporting a vulnerability

If you find a security issue:

1. **Do not open a public GitHub issue for it.**
2. Report it privately. If this repository is hosted on GitHub, use
   GitHub's private vulnerability reporting feature (Security tab ->
   "Report a vulnerability") once it is enabled for this repository. Until
   that is set up, contact the maintainer directly through a private
   channel rather than a public tracker.
3. Include, as applicable:
   - Affected component/file and, if known, commit hash
   - Steps to reproduce, or a minimal proof-of-concept
   - Impact you believe this has (crash, information disclosure, identity
     compromise, network-compatibility break, etc.)
   - Whether you believe this also affects upstream `yggdrasil-go` (if so,
     please also consider reporting it upstream, since a fix likely needs
     to happen in both places or the fork will silently diverge from a
     known-safe baseline)

## What to expect

- Acknowledgement of a report should happen promptly; as a small,
  currently single-maintainer project, "promptly" is a best effort, not a
  contractual SLA.
- Coordinated disclosure is preferred: please give a reasonable window for
  a fix before any public disclosure. What's reasonable depends on
  severity — a remotely triggerable crash or identity-compromise bug
  warrants faster turnaround than a low-severity local issue.
- Fixes for confirmed vulnerabilities will include, where practical, a
  regression test (see the repository's engineering rules: "for every
  meaningful bug: reproduce it, write a regression test, fix it, verify
  the test fails before / passes after").

## What this project does and does not guarantee

**Does:**
- Aims to remain byte-for-byte wire-compatible with the Yggdrasil network
  protocol (see [NAMING.md](NAMING.md)) — a compatibility break is treated
  as a release-blocking defect, security-relevant or not.
- Fuzzes untrusted-input parsers it owns (see
  [docs/THREAT_MODEL.md](docs/THREAT_MODEL.md) for exactly which ones, and
  what fuzzing has actually found).
- Uses established cryptographic primitives (ed25519, blake2b) rather than
  custom cryptography, and will continue to.

**Does not:**
- Does not claim formal verification or independent audit coverage unless
  and until one has actually happened, at which point this document will
  say so specifically (auditor, scope, date, report link).
- Does not promise anonymity or resistance to traffic analysis.
- Does not treat "no known vulnerabilities" as "no vulnerabilities exist" —
  see the residual-risk notes throughout
  [docs/THREAT_MODEL.md](docs/THREAT_MODEL.md) for the specific gaps this
  project is aware of and has not yet closed (e.g. the admin socket
  currently has no authentication layer, relying on local-only defaults
  and OS file permissions instead).
