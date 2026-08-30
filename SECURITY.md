# Security policy

## Reporting a vulnerability

Please report suspected vulnerabilities through GitHub's **Private vulnerability reporting** feature on the repository's Security tab. Do not open a public issue or disclose exploit details publicly before a fix is available.

Include, when possible:

- the affected version or commit;
- the operating system and architecture;
- steps to reproduce the issue;
- the expected security impact; and
- a minimal proof of concept with secrets and personal data removed.

The maintainers will acknowledge the report, investigate it, and coordinate disclosure and remediation with the reporter. Please allow a reasonable period for a fix before publishing details.

## Supported versions

Security fixes are made on the current `main` branch and included in the next release. Users should run the latest available release.

Protocol 0.6 introduces transcript-bound handshake confirmations and is not wire-compatible with older nodes. Upgrade every peer in a private overlay together; do not mix protocol 0.5 and 0.6 nodes.
