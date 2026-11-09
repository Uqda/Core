# Security policy

Uqda Core is an **experimental** project. It encrypts traffic between nodes
using the Yggdrasil-compatible protocol, but it does not promise anonymity or
application security, and it has not had an independent security audit. See
the [security model](docs/security.md) for trust boundaries and limits.

## Supported versions

| Version | Security fixes |
|---|---|
| Latest stable release (26.1.x) | Yes |
| Latest beta | Yes, fixed in the next beta |
| Older releases | No. Please update |

## Reporting a vulnerability

**Do not report vulnerabilities in public issues.** Use either:

- GitHub private vulnerability reporting:
  <https://github.com/Uqda/Core/security/advisories/new>, or
- email: <Uqda@proton.me>

Please include the affected version or commit, steps to reproduce, the impact
you expect, and logs with keys and passwords removed. If upstream Yggdrasil or
Ironwood is also affected, say so; we coordinate with their maintainers before
anything is published.

## What happens next

These are targets for a volunteer project, not guarantees:

| Step | Target |
|---|---|
| Acknowledge your report | within 7 days |
| First assessment and severity | within 14 days |
| Fix or mitigation for serious issues | as soon as practical, usually within 90 days |
| Public advisory | after a fix is released, or after 90 days by agreement |

We credit reporters in the advisory unless you prefer not to be named.

When a vulnerability is actively exploited or seriously affects users, the
maintainers will publish an advisory and a fixed release as quickly as
possible, and inform the security contacts of affected downstream projects.

## Safe harbour

We will not pursue legal action against anyone who, in good faith:

- tests only their own nodes or nodes whose owners agreed,
- avoids harming other users, their data and the network (no denial-of-service
  testing on public peers),
- reports the issue to us privately and gives us reasonable time to fix it.

## Abuse

Uqda is decentralised software. The maintainers do not operate the network,
cannot see its traffic and cannot block nodes. If you believe a node is used
for illegal activity, contact the operator of that node or the competent
authority. Reports about this repository, its releases or its documentation go
to <Uqda@proton.me>; we answer them as quickly as we can.
