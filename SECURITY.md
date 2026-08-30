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

New nodes remain compatible with protocol 0.5 peers. Secure transcript confirmation is negotiated automatically when both sides support it. For sensitive or controlled overlays, append `?secure=required` to both peering and listener URIs to reject legacy handshakes and prevent downgrade.

Example: `tls://peer.example:9001?secure=required`.
