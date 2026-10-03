# Umbrel remote access over Uqda

## Product goal / الهدف

Use **your Umbrel's Uqda IPv6 address** to reach its existing dashboard and
compatible apps from another private-group device, without requiring Tailscale
for that connection. Uqda supplies the encrypted network path; Umbrel still
supplies the operating system, app management, accounts, permissions and login.
This is not a new Umbrel OS or a replacement for its underlying Docker network.

الفكرة: افتح جهاز Umbrel الخاص بك من جهاز آخر عبر عنوان عقدة، مع إبقاء تسجيل
دخول Umbrel وحماية تطبيقاته. عقدة تؤمّن طريق الاتصال، ولا تستبدل نظام Umbrel
أو تُحوّل كل التطبيقات تلقائيًا إلى خدمات عامة.

## Research evidence

Reviewed on 2026-10-03:

- [Umbrel remote-access support](https://umbrel.com/support/basics/remote-access)
  describes Tailscale as access to the dashboard, apps, shared folders and
  network backups from devices on the same private network.
- [Official Tailscale package](https://github.com/getumbrel/umbrel-apps/blob/master/tailscale/docker-compose.yml)
  uses host networking, kernel TUN and network administration capabilities.
  Uqda's package already uses host TUN, so replacing the existing app gateways
  with an unauthenticated second proxy is not the appropriate starting point.
- [LAN ingress source](https://github.com/getumbrel/umbrel/blob/11e7f5bdfcb5bb8b46f39edd6defb4c37641763a/packages/umbreld/source/modules/lan-ingress/lan-ingress.ts)
  opens host dashboard listeners on 80/443 and app authentication on 2000.
  App gateway listeners use Umbrel-owned inet nftables redirection, preserving
  the gateway between external clients and IPv4/container backends. This is
  evidence for testing direct overlay access, not a guarantee for every app.
- [App gateway source](https://github.com/getumbrel/umbrel/blob/11e7f5bdfcb5bb8b46f39edd6defb4c37641763a/packages/umbreld/source/modules/app-gateway/app-gateway.ts)
  retains login, app authorization and WebSocket checks, strips gateway cookies
  from app upstreams and formats literal IPv6 redirect authorities in brackets.
- [System address discovery](https://github.com/getumbrel/umbrel/blob/11e7f5bdfcb5bb8b46f39edd6defb4c37641763a/packages/umbreld/source/modules/system/system.ts)
  includes tunnel interfaces but selects IPv4 addresses for its IP list.
  LAN certificate generation consequently does not add Uqda's IPv6 literal.
- [HTTPS support](https://umbrel.com/support/basics/accessing-your-umbrel-over-https)
  explains locally generated certificates and HTTPS requirements for some apps.
  The literal overlay IP is not automatically a verified HTTPS identity.
- [Tailscale-specific metadata](https://github.com/getumbrel/umbrel/blob/11e7f5bdfcb5bb8b46f39edd6defb4c37641763a/packages/umbreld/source/modules/system/tailscale.ts)
  obtains a browser hostname from that app. Native Umbrel client integration
  is not an interchangeable generic remote-access provider interface.

Research of current master is pinned above; platform integration tests use the
separate official 2.0.0 commit `9298257b0e904ca8d8270b1702672666343f0b86`.
Do not assume all later releases behave identically.

## Implemented source workflow (not yet distributed)

The dashboard's **Your Umbrel, from another device / جهاز Umbrel من جهازك الآخر**
section generates these addresses from the daemon's own identity:

```text
http://[YOUR_UQDA_IPV6]/
https://[YOUR_UQDA_IPV6]/
```

Addresses appear only with a ready TUN interface and private-group mode.
Changing modes is never automatic. Copy the address onto another device with
Uqda installed and connected to the same private group. Use Umbrel's existing
owner/member login and, where enabled, 2FA. Compatible apps opened from Umbrel
must still pass Umbrel's per-app authorization and any application login.
Opening Uqda's management page on a phone does not install a tunnel on it.

The optional local check tests only TCP ports 80, 443 and 2000 on this node's
own literal overlay IPv6. It does not accept a custom target, publish any
service, contact Docker, forward traffic or probe the internal umbreld port.
Results always say local and never certify remote login or app functionality.
The full service address book remains available for deliberate per-app checks.

This source enhancement is not in the already published wrapper
`26.0.4-umbrel.2`. It needs review, a new immutable image and a pinned store
update before users receive it. The existing release and devices are unchanged.

## Security and compatibility boundaries

- Use a strong unique group secret only on trusted devices. It is shared-group
  access control, not device-level revocation, roles or a replacement for
  Umbrel/app authentication. A public peer may route encrypted traffic without
  belonging to the private group; transport UP is not private-session readiness.
- Do not remove the firewall or expose a public admin interface to obtain
  convenience. This workflow adds no forwarding or firewall rules, changes no
  router settings and leaves the existing LAN connection intact.
- HTTPS requires a matching name/address and a verified certificate/trust
  relationship. Do not disable certificate validation or install an unverified
  CA. Existing Umbrel hostname certificates do not automatically match Uqda's
  IPv6 literal. HTTP is protected by the overlay between Uqda nodes, but is not
  a browser secure context and cannot substitute for TLS-required app features.
- IPv4-only apps outside Umbrel's supported gateway still require separately
  reviewed configuration. Arbitrary UDP/TCP apps, uploads, WebSockets, native
  app clients, SMB mounts and backups need their own actual functional tests;
  TCP listening alone is insufficient.
- Our mobile module is an embedding SDK, **not a ready-made iOS/Android VPN
  app**. A host app must implement the OS tunnel, secure config storage, UI and
  installation/signing. The SDK now passes `GroupPassword` into Core rather
  than silently ignoring it; real-node tests cover matching, different and
  public groups. Previously distributed bindings require rebuilding/updating.
- Umbrel's native mobile/Mac Tailscale discovery and automatic folder handling
  are not replaced by changing an address in this dashboard. A generic-provider
  integration would require upstream collaboration.
- This remains a **community integration**, not official Umbrel certification,
  store inclusion or endorsement. Those require Umbrel's own review/acceptance.

## Validation gates

1. Portable address/policy tests: private mode, TUN readiness, fixed ports,
   forbidden custom targets and no falsely verified remote result.
2. Authenticated HTTP/control tests: same Origin/session proof and no public-mode
   check; native Docker/browser tests verify addresses, local labels and RTL.
3. Official Umbrel source-platform gate: start an independent disposable Uqda
   peer; fetch the actual dashboard over the private overlay, verify the real
   app gateway's IPv6 authentication redirect, enforce owner/app authentication
   and origin-scoped proof, reject a different group and recover after restoring
   the correct one. Cookies/proofs travel over test stdin, never CLI arguments.
4. SDK regression: real encrypted sessions with same/different/empty group
   passwords; then normal Android/Apple consumer build gates.

Do not call these tests physical Umbrel certification. Production OS reboot,
certificate provisioning, every installed app, SMB/backup restore and signed
phone-client behavior remain separate acceptance work. Do not remove a working
Tailscale setup until the desired replacement use cases pass on the user's own
devices; coexistence is possible without changing default internet routes.

## تجربة المستخدم المقصودة

١. ثبّت عقدة على Umbrel وعلى جهاز الوصول، واربطهما بمجموعة خاصة واحدة.

٢. انسخ عنوان Umbrel الظاهر داخل لوحة عقدة وافتحه من الجهاز الآخر، ثم سجّل
دخولك إلى Umbrel كالمعتاد. كلمة مرور المجموعة لا تستبدل كلمة مرور Umbrel.

٣. اختبر التطبيقات التي تستخدمها فعلًا، بما فيها تسجيل الدخول ورفع الملفات
والاتصالات الحية عند الحاجة. لا نفترض أن فحص منفذ يثبت عمل التطبيق بالكامل.

٤. HTTPS والآيفون وتطبيقات Umbrel الأصلية ومشاركة الملفات تحتاج تحققًا خاصًا؛
لن نسوّقها كبديل كامل لكل وظائف Tailscale قبل اجتياز هذه الاختبارات.
