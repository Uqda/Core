# Uqda Network for Umbrel 26.0.5-umbrel.1

A new immutable Umbrel integration release, based on Uqda Core 26.0.5. Existing
tags and images are retained. Update through the Uqda community app store:
https://github.com/Uqda/umbrel-app-store

## Practical access

- A private-group section provides your Umbrel's own Uqda IPv6 dashboard address
  and instructions for another Uqda-connected device. Umbrel retains its login,
  permissions and app gateway; no second unauthenticated proxy is introduced.
- Save up to twelve existing HTTPS/HTTP/SSH/TCP service addresses and copy bounded
  diagnostic commands. Entries do not install apps, open ports or publish services.
- Check only this node's own TCP listeners. Results distinguish local listening
  from actual remote access, authentication, application health and valid TLS.
- Arabic/English guidance, responsive RTL layout and guarded asynchronous UI state.
- Core 26.0.5 includes the mobile SDK private-group fix. SDK artifacts are not
  ready-made iOS/Android VPN apps.

## Installation and use

Add the community store above to Umbrel, install Uqda Network, and use the app
password shown by Umbrel. Add a trusted peer and the same strong private-group
secret used on your other Uqda devices. Open the Umbrel IPv6 address from another
device with a working Uqda tunnel and sign in to Umbrel normally.

Back up app data privately before updating. Updates/restarts preserve identity
and settings. Removing app data deletes the identity; never run two active nodes
from one identity backup. Public-mode nodes are not offered the private-access
workflow. This does not itself block an already public service or alter a firewall.

## Validation and boundaries

Release gates cover native amd64/arm64 Docker/TUN, real browser UI, independent
nodes exchanging TCP/UDP and HTTP, private-group rejection/recovery, and the pinned
official Umbrel 2.0 development app manager. Platform tests fetch the actual built
Umbrel dashboard, check IPv6 app-auth redirects and owner/UI-proof enforcement,
and exercise install/restart/previous-public-image upgrade/uninstall/reinstall.
See the linked release evidence before treating any gate as passed.

This is a community integration, not official Umbrel certification or endorsement.
It is not yet a full substitute for every Tailscale feature. Native Umbrel clients,
SMB, backups, production OS reboot and every third-party app require separate tests.
HTTPS needs a matching identity and verified certificate: the existing Umbrel
certificate does not automatically cover an overlay IPv6 literal. Do not disable
certificate checks or install unverified CAs. HTTP over the encrypted overlay is
not a browser secure context. Keep a working alternative until your use cases pass.

## بالعربي

الفائدة الجديدة: افتح واجهة Umbrel وتطبيقاته المتوافقة عبر عنوان عقدة من جهاز
آخر ضمن مجموعتك الخاصة، مع بقاء تسجيل دخول Umbrel وصلاحياته. أضفنا دليل خدمات
وفحوص TCP محلية وإرشادات عربية وإنجليزية. لا يفتح التطبيق منافذ أو يثبت خدمات
نيابةً عنك، ونجاح فحص منفذ لا يعني أن الخدمة تعمل من الخارج.

حدّث من متجر Umbrel بعد أخذ نسخة احتياطية خاصة. هوية العقدة تبقى عند التحديث،
والحذف الكامل يمسحها. الهاتف وHTTPS والملفات والنسخ الاحتياطي لها حدود واختبارات
إضافية؛ لا نعلن أن جميع ميزات Tailscale استُبدلت أو أن كل تطبيق خالٍ من المشاكل.
