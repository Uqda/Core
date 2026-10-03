# Uqda Network for Umbrel 26.0.4-umbrel.2

This is a new, separately versioned Umbrel integration release. The underlying
Uqda Core remains the verified 26.0.4 release; native Linux/macOS installers and
Homebrew are not changed by this wrapper release. The previous image is retained.

## Changes

- Require an origin-scoped, tab-held random proof for authenticated reads and
  changes. Cookies ignore ports, so another co-hosted app receiving the cookie
  cannot bootstrap this proof or read node status with the cookie alone.
- Retain authentication on reload of the same tab; clear proof and draft secrets
  on logout or expiry. A new tab or blocked session storage may require login.
- Unlock settings immediately after save/restart, including failure paths, and
  prevent conflicting operations or stale polls from changing the displayed state.
- Preserve and redact advanced peer credentials, including encoded query names.
- Explain trusted private groups, separate end-to-end/service checks and backups
  in Arabic and English, with responsive RTL/mobile layout.
- Export only approved, portable store files with both image references pinned
  to the published multiarchitecture digest; never include keys or backups.

## Validation and limits

24 real-daemon/authentication/configuration/export tests passed, alongside native
amd64 and arm64 Docker/TUN, actual browser and two-independent-node TCP/UDP gates.
Wrong private-group passwords were rejected and correct-password recovery passed.
The pinned official Umbrel 2.0 development app manager passed installation,
restart, previous-wrapper-to-new-source upgrade with retained identity, and
uninstall/fresh-install identity replacement. Release packaging reruns these
gates, including installation of the actual new published image.

Evidence and detailed review: https://github.com/Uqda/Core/pull/22

These are software-development tests, not certification of production OS boot,
backup restoration, full owner-browser gateway flow, external storage or Raspberry
Pi firmware. Plain HTTP does not protect browser passwords in transit; use HTTPS
when available. Services require IPv6 listening and firewall permission. Other
Docker apps are not exposed automatically; Uqda does not promise anonymity.

## Install or update

Add `https://github.com/Uqda/umbrel-app-store` in Umbrel's Community App Stores,
then install **Uqda Network**. Sign in using the app password shown by Umbrel.
Add a trusted peer and the same strong group password as your other nodes.
New installations start isolated until you configure these settings.

For existing installations, back up app data first and update through Umbrel.
The node identity persists across updates and restarts. Uninstalling app data
removes the identity; never restore the same identity onto two active nodes.

## بالعربي

هذا إصدار جديد لتكامل Umbrel، وليس إصدارًا جديدًا لنواة Uqda أو Homebrew.
حسّنّا حماية الجلسة، وحفظ الإعدادات، ومسح الأسرار عند الخروج، وإخفاء بيانات
الاتصال الحساسة. الواجهة عربية وإنجليزية وتعمل على شاشة الجوال.

أضف رابط المتجر أعلاه، وثبّت Uqda Network، ثم استخدم كلمة مرور التطبيق التي
يعرضها Umbrel. أضف عقدة موثوقة وكلمة مرور مشتركة قوية لتربط أجهزتك بشبكة IPv6
مشفّرة. اختبر الاتصال والخدمة المطلوبة بشكل منفصل؛ التطبيقات لا تُكشف تلقائيًا.

للتحديث: انسخ بيانات التطبيق احتياطيًا ثم حدّث من Umbrel. الحذف الكامل يمسح
هوية العقدة. اختبارات Docker ومدير Umbrel الرسمي نجحت، لكن إقلاع النظام
الإنتاجي واستعادة النسخ الاحتياطية لم يُختبرا، ولا يوجد ضمان بانعدام كل المشاكل.
