# Security Policy

## Supported Versions

We provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |
| < 0.1   | :x:                |

---

## Reporting a Vulnerability

**Please do NOT open a public GitHub issue for security vulnerabilities.**

### How to Report

Email us privately at: **uqda@proton.me**

Include the following information:

1. **Description** - Clear description of the vulnerability
2. **Steps to Reproduce** - Detailed steps to reproduce the issue
3. **Potential Impact** - What could an attacker do?
4. **Suggested Fix** - If you have ideas for a fix
5. **Proof of Concept** - If applicable, include a minimal PoC

### What to Expect

- **Acknowledgment**: We will acknowledge receipt within **48 hours**
- **Assessment**: We will assess the vulnerability and determine severity
- **Timeline**: We will provide an estimated timeline for fix
- **Disclosure**: We will coordinate with you on responsible disclosure
- **Credit**: We will credit you in security advisories (if desired)

### Severity Levels

We use the following severity levels:

- **Critical**: Remote code execution, authentication bypass, data exfiltration
- **High**: Privilege escalation, denial of service, information disclosure
- **Medium**: Local privilege escalation, information leakage
- **Low**: Minor information disclosure, denial of service (limited scope)

---

## Security Best Practices

### Firewall Configuration

**Uqda nodes are globally reachable by default.** Configure IPv6 firewall rules to protect your node.

#### Linux (iptables/ip6tables)

```bash
# Block all incoming on Uqda interface
ip6tables -A INPUT -i uqda0 -j DROP

# Allow established connections
ip6tables -A INPUT -i uqda0 -m state --state ESTABLISHED,RELATED -j ACCEPT

# Allow specific services (example: SSH on port 22)
ip6tables -A INPUT -i uqda0 -p tcp --dport 22 -j ACCEPT
```

#### Windows (PowerShell)

```powershell
# Block all inbound on Uqda interface
New-NetFirewallRule -DisplayName "Block Uqda Inbound" `
    -Direction Inbound -InterfaceAlias "Uqda" -Action Block

# Allow specific service (example: SSH)
New-NetFirewallRule -DisplayName "Allow SSH on Uqda" `
    -Direction Inbound -InterfaceAlias "Uqda" -Protocol TCP -LocalPort 22 -Action Allow
```

#### macOS (pfctl)

Add to `/etc/pf.conf`:
```
block in on uqda0 all
pass in on uqda0 proto tcp from any to any port 22
```

Then reload:
```bash
sudo pfctl -f /etc/pf.conf
```

### Service Exposure

**Do not expose sensitive services** on the Uqda interface without additional security:

- Use **application-layer encryption** for sensitive data
- Implement **authentication** for all services
- Use **principle of least privilege** - only expose what's necessary
- **Monitor** your node for unexpected connections

### Key Management

- **Protect your private key** - It's your network identity
- **Backup your configuration** - Store securely (encrypted)
- **Rotate keys** if compromised - Generate new config and update peers
- **Don't share private keys** - Each node should have unique keys

### Network Isolation

For sensitive deployments:

- **Separate networks** - Use different Uqda networks for different purposes
- **Access control** - Implement application-layer access control
- **Monitoring** - Monitor network traffic and connections
- **Logging** - Log security-relevant events

---

## Security Audit Status

**Current status**: Uqda has **not undergone independent security audit**.

The codebase inherits cryptographic implementations from:
- **Go standard library** (audited)
- **Yggdrasil Network** (community-reviewed since 2017)

### Recommendations

- **Do not use** for high-security applications without independent review
- **Assume** presence of undiscovered vulnerabilities
- **Defense-in-depth**: Use application-layer encryption for sensitive data
- **Monitor** your deployment for anomalies

### Planned Audit

**Community-funded audit scheduled for Q2 2026.**

If you're interested in contributing to the audit fund, contact us at **uqda@proton.me**.

---

## Known Limitations

### Anonymity

**Uqda does not provide anonymity:**
- Direct peers can see your IP address
- Traffic patterns may be observable
- No protection against traffic analysis

If you need anonymity, use **Tor** or **I2P** instead.

### Protocol Security

- Routing messages are cryptographically signed
- All traffic is end-to-end encrypted
- No protection against active network adversaries (packet dropping, delay, reordering)

### Implementation Security

- Inherits security properties from Yggdrasil v0.5
- No known critical vulnerabilities (as of January 2026)
- Regular security updates as issues are discovered

### Windows installers (MSI) and Smart App Control

**Root cause:** Windows trusts installers that carry a valid **Authenticode** signature from a **publicly trusted** code-signing CA. Without that, **Smart App Control** and **SmartScreen** may block or warn — this is OS policy, not a bug in Uqda.

**What this repo does:** The **Release** workflow (`release.yml`) runs `contrib/msi/sign-msi.ps1` after each MSI is built. If repository secrets are configured, the MSI is signed with **`signtool`** (SHA-256, RFC 3161 timestamp via DigiCert’s TSA) and verified before upload. That is the technical “fix from the root” for Windows trust.

**Maintainers — required setup (one-time):**

1. **Buy or obtain a code-signing certificate** (`.pfx` / PKCS#12) from a CA trusted for Authenticode (e.g. DigiCert, Sectigo, SSL.com). **EV** certificates usually gain SmartScreen reputation faster than OV; both work once Windows trusts the publisher.
2. **Add two GitHub Actions secrets** on the repository (Settings → Secrets and variables → Actions):
   - **`WINDOWS_CODESIGN_PFX_BASE64`** — entire `.pfx` file, **base64-encoded** (not the raw binary).  
     Example (PowerShell):  
     `[Convert]::ToBase64String([IO.File]::ReadAllBytes('path\to\cert.pfx'))`  
     Paste the output string into the secret.
   - **`WINDOWS_CODESIGN_PASSWORD`** — password for that `.pfx`.
3. **Tag a new release** (or run the Release workflow manually). On **`Uqda/Core`**, the Windows job **fails** if these secrets are missing — so **public releases do not ship unsigned MSIs** (the root fix for Smart App Control blocking unknown publishers).

**Forks / other remotes:** The “require secrets” check only applies to `github.repository == Uqda/Core`. Forks can still build without signing (their own Actions), but end users should prefer **official** GitHub Releases for signed installers.

**Protecting the key:** Prefer HSM or cloud signing for production; storing a `.pfx` in GitHub Secrets is a common CI pattern but protect the file and rotate if leaked. Do not commit `.pfx` to git.

**macOS:** Broad distribution without Gatekeeper issues requires **Developer ID** signing plus **notarization** — separate from Windows; not automated in this workflow yet.

**Smart App Control (SAC) after signing:** A valid Authenticode signature is required; SAC can still show prompts for **new** publishers until Microsoft’s reputation systems catch up ([Smart App Control overview](https://learn.microsoft.com/windows/security/application-security/application-control/smart-app-control/)). An **EV** code-signing certificate often reduces this window. If a signed build is still blocked, options are: wait/retry, use a PC without SAC enforcement, or adjust **Windows Security → App & browser control** (only if you accept the risk). **Group Policy** blocks (e.g. “system administrator prevented installation”) need **IT** to allow the installer or the publisher.

**Users on locked-down PCs:** Organizational devices may require IT approval even for signed apps.

---

## Security Updates

Security updates are released as:
- **Patch releases** (e.g., 0.1.1 → 0.1.2) for critical/high severity issues
- **Minor releases** (e.g., 0.1.x → 0.2.0) for medium/low severity issues

**Subscribe to GitHub releases** to be notified of security updates.

---

## Responsible Disclosure

We follow responsible disclosure practices:

1. **Private reporting** - Report vulnerabilities privately first
2. **Timeline coordination** - We coordinate disclosure timeline with reporter
3. **Fix development** - We develop and test fixes
4. **Public disclosure** - After fix is available, we disclose publicly
5. **Credit** - We credit researchers in security advisories

---

## Security Advisories

Security advisories are published:
- In GitHub **Releases** (tagged with `security`)
- In GitHub **Security Advisories** (GHSA)
- Via email to subscribers (if you've reported a vulnerability)

---

## Contact

For security-related inquiries:
- **Email**: uqda@proton.me
- **Subject**: `[SECURITY] Your Subject Here`

---

**Last Updated**: April 2026

