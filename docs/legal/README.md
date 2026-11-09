# Legal and compliance notes

This page records how the project meets its license and legal obligations, and
which decisions still belong to the maintainers. It is a working checklist, **not
legal advice**. Items marked ⚖️ need a lawyer or tax adviser (in Austria, for
example, a *Rechtsanwalt* or *Steuerberater*) before they are decided.

## Done in the repository

| Obligation | Where |
|---|---|
| LGPLv3 + linking exception kept verbatim from upstream | [LICENSE](../../LICENSE), checked identical to upstream on 2026-11-09 |
| Upstream copyright notices kept | The only upstream source header (`src/tun/tun_windows.go`) is unchanged |
| Attribution, fork point, modifications, non-affiliation | [NOTICE](../../NOTICE.md) |
| Inventory of every third-party component and its license, kept complete by CI | [THIRD_PARTY_LICENSES](../../THIRD_PARTY_LICENSES.md), `contrib/licenses/check.sh` |
| MPL-2.0 source of the modified Ironwood | [third_party/ironwood](../../third_party/ironwood) |
| `wintun.dll` shipped unmodified with its license | Windows MSI installs `wintun-LICENSE.txt` beside it |
| License files in every package | `.deb`: `/usr/share/doc/uqda/`; MSI: install folder; Docker: `/usr/share/licenses/uqda/` plus OCI `source` and `licenses` labels |
| Privacy information, no telemetry | [PRIVACY](../../PRIVACY.md) |
| Experimental status, no anonymity, user responsibility (English and Arabic) | [README](../../README.md), [README.ar](../../README.ar.md) |
| Vulnerability reporting, response targets, safe harbour, abuse contact | [SECURITY](../../SECURITY.md) |
| Privacy-preserving defaults | `NodeInfo` empty; `NodeInfoPrivacy: true` in new configurations; firewall on |
| Release integrity | `SHA256SUMS` on every release; checked by the Linux installer |

## Policies the project follows

- **No telemetry.** Uqda must not phone home. A change that contacts any server
  without the user starting it needs a documented reason, an opt-out and an
  entry in [PRIVACY](../../PRIVACY.md).
- **Descriptive naming only.** "Yggdrasil" appears only to describe compatibility
  or origin, never in a product name, logo or domain.
- **Non-commercial.** Uqda is free software. If donations are accepted, they
  never buy features, early access or a different version: every donor and
  non-donor gets the same releases.
- **Abuse.** The maintainers do not operate the network and cannot see or block
  traffic. Reports about the repository, releases or documentation go to
  <Uqda@proton.me> and are answered as soon as possible.
- **Export control.** Uqda uses standard, published cryptography and its source
  code is publicly available. Publicly available software is generally outside
  EU dual-use export controls, but distributing binaries to sanctioned countries
  or persons can still be restricted. ⚖️ Confirm before targeted distribution
  outside the EU.

## Open items for the maintainers

These need information or decisions that only the project owner can provide.
They are not done by automation on purpose.

| Item | Why | Who |
|---|---|---|
| Site notice (*Offenlegung* under §25 MedienG; *Impressum* under §5 ECG if the site is commercial) for any project website or social account | Austrian media law requires a name and place of residence or business for the publisher | Owner; [template](impressum-template.md) |
| PGP key for <Uqda@proton.me> | Lets reporters encrypt; add its fingerprint to SECURITY.md | Owner |
| Signing releases (Windows Authenticode, Apple notarisation, Sigstore or GPG) and an SBOM | Integrity and the EU Cyber Resilience Act's expectations for software in the EU market | Owner (certificates cost money) |
| Trademark search for "Uqda" / "عُقدة" (EUIPO, WIPO) before registering a mark | Avoids conflicts with existing marks | Owner ⚖️ |
| Legal form (for example an Austrian *Verein*) before accepting donations or signing contracts | Separates personal liability from the project; affects taxes | Owner ⚖️ |
| Role under the Cyber Resilience Act (non-commercial open-source project, or "open-source steward" if a legal entity supports it) | Determines which CRA duties apply from 2027 | Owner ⚖️ |
| Publishing mobile apps in app stores | Store terms and the LGPL's relinking requirements must fit together | Owner ⚖️ |
| A mirror outside GitHub (for example Codeberg) | Keeps the source available if one host fails | Owner |
