# Security policy

Uqda Core protects network sessions using the Yggdrasil-compatible protocol.
It does not promise anonymity, application security or independent audit coverage.
See the [security model](docs/security.md) for trust boundaries and limitations.

## Supported versions

Uqda 26 Beta 1 is the currently supported prerelease. Security fixes target the
default branch; backports to this beta are considered case by case.

## Reporting a vulnerability

Do not disclose vulnerabilities in public issues. Use GitHub private vulnerability
reporting if enabled for `Uqda/Core`, or an established private maintainer contact.
Include the affected revision, reproduction steps, impact and relevant logs with
secrets removed. If upstream Yggdrasil is also affected, coordinate reporting with
its maintainers as well.

Coordinated disclosure is preferred. Response timing is best effort. Confirmed
fixes should include regression tests where practical. Preserve confidentiality
until maintainers and the reporter agree on disclosure or a reasonable response
window has elapsed.
