# Uqda Core 26.0.5

A new patch release. Earlier tags, downloads and container images are retained.

## Changes

- Fix the mobile embedding SDK silently ignoring `GroupPassword`. Android and
  Apple bindings now enforce the configured private group. Upgrade previously
  distributed bindings before relying on private-group isolation.
- Add the Umbrel service address book and private-group dashboard-access workflow
  in source, with Arabic/English guidance and bounded local-only TCP checks.
  Distribution of this UI uses the separately versioned Umbrel wrapper/store.
- Preserve Umbrel's existing login and app gateway. Do not automatically expose
  apps, change firewall/routes, bypass authentication or delete Tailscale.
- Generalize immutable Umbrel wrapper version validation for this new Core line.

## Validation and limits

Real mobile sessions test matching, different and public groups; mutation testing
confirmed the unfixed implementation fails the isolation regression. Source gates
cover Linux/macOS/Windows, race detection, fuzzing, vulnerability checks and mobile
consumer builds. Native installer lifecycle and checksum/archive-tampering tests
are rerun for these release packages.

The official Umbrel 2.0 development platform verified the real dashboard over a
private overlay from an independent node, app-gateway authentication, wrong-group
rejection/recovery and install/restart/update/uninstall lifecycle. This is not
physical-device certification or proof that every Umbrel app works.

The mobile artifacts are embedding SDKs, not ready-made phone VPN applications.
HTTPS certificate/name matching, native Umbrel client integration, SMB and backup
restoration require separate acceptance tests. Installers remain unsigned; Windows
ARM64 is packaging/build validation rather than native installation testing.

## Install or update

macOS: `brew update`, then `brew install --cask Uqda/tap/uqda` or
`brew upgrade --cask Uqda/tap/uqda` for an existing installation.

Linux/systemd:

```sh
curl -fsSLo /tmp/uqda-install.sh https://raw.githubusercontent.com/Uqda/Core/main/contrib/install/linux.sh
sudo sh /tmp/uqda-install.sh install
# Existing quick installation:
sudo sh /tmp/uqda-install.sh update
```

Windows: use the matching MSI from this release. Verify downloaded assets against
`SHA256SUMS`. Normal updates retain identity/configuration. Explicit purge/zap
permanently removes identity; back up privately first. Do not use the quick Linux
installer over a package-managed installation.

## بالعربي

هذا إصدار جديد للنواة، ويصلح تطبيق كلمة مرور المجموعة في SDK الهاتف؛ الملفات
ليست تطبيق VPN جاهزًا للآيفون أو أندرويد. تحسينات واجهة Umbrel توزّع بحزمة مستقلة
ومحدّثة في متجرنا المجتمعي، مع الحفاظ على تسجيل دخول Umbrel وصلاحيات تطبيقاته.
لا نعلن اعتمادًا رسميًا من Umbrel أو بديلًا كاملًا لكل وظائف Tailscale.
