# Uqda — رؤية المشروع / Project vision

## العربية

**عُقدة (Uqda)** شبكة تراكب IPv6 مشفّرة من طرف إلى طرف، مبنية على أفكار بروتوكول موجّه نحو التوسّع والمرونة. الهدف هو أن يملك أي شخص أو منظمة طبقة شبكة خاصة تعمل فوق الإنترنت أو الشبكات المحلية، دون سجّل مركزي ودون اعتماد على مزوّد واحد للهوية أو العناوين.

### الفكرة الأساسية

- **هوية = عنوان**: عنوان IPv6 الثابت (`200::/7`) مشتق من مفتاحك العام (Ed25519)، فيبقى عنوانك ثابتاً طالما احتفظت بالمفتاح.
- **تشفير افتراضي**: حركة البيانات في الطبقة المشفّرة محمية؛ الموجّهات تر ven فقط المسارات لا محتوى التطبيقات.
- **شبكة ذاتية الإصلاح**: اكتشاف الجيران والمسارات متعدد القفزات يتم تلقائياً داخل نفس النسيج.
- **توافق مع Yggdrasil 0.5**: يمكن للعقد أن تتعايش مع شبكات Yggdrasil عند تكوين الأقران والنقل بشكل متوافق.

### ما يقدّمه المشروع عملياً

- عقدة قابلة للتشغيل على **Linux وWindows وmacOS** مع واجهة TUN اختيارية.
- **إدارة عبر مقبس admin**، وأداة سطر أوامر `uqdactl`، وواجهة ويب اختيارية (`UIListen`).
- **جلسات متعددة النقل**: TCP، TLS، QUIC، WebSocket، SOCKS، UNIX، وغيرها حسب الإعداد.
- **سياسات أمان تشغيلية**: مفاتيح مسموحة للأقران، شبكات خاصة بدعوات، مصادقة admin، وفحص الثغرات في سير عمل CI.

### القيم

- البرمجيات الحرة (LGPL-3.0)، بدون قفل بائع، بدون اشتراك إلزامي.
- الشفافية في الاعتماد على مشروع Yggdrasil الأصلي (راجع `ATTRIBUTION.md`).

---

## English

**Uqda** (from Arabic *ʿuqda*, “knot / node”) is an end-to-end encrypted IPv6 overlay mesh. The goal is practical decentralized connectivity: stable addresses derived from public keys, self-organizing routes, and no central registry for identities or topology.

### Core ideas

- **Identity-as-address**: A `/128` in `200::/7` and a `/64` subnet are derived from your Ed25519 key; keep the key, keep your addresses.
- **Encryption by default**: The encrypted routing layer protects mesh traffic; it is not a substitute for application-layer TLS when you need origin authentication for arbitrary services.
- **Self-healing paths**: Multi-hop paths are discovered and repaired without manual static routing for typical mesh use.
- **Yggdrasil v0.5 protocol compatibility** when peers and transports are configured accordingly.

### What this repository delivers

- Production **node binary (`uqda`)** and **control tool (`uqdactl`)** for major desktop/server OS targets.
- **Admin socket API**, optional **Web UI**, optional **Prometheus metrics** proxy path.
- **Multiple transports** (TCP/TLS/QUIC/WS/SOCKS/UNIX, etc.) and operational features: **AllowedPublicKeys**, **private networks / invites**, DNS tuning for IPv4 underlays while the overlay stays IPv6.
- **Security-oriented workflow**: dependency updates, `govulncheck`, golangci-lint in CI.

### Values

- Free software (LGPL-3.0), no mandatory SaaS, no single gatekeeper for participation.
- Clear **attribution** to the upstream **Yggdrasil Network** project that this code descends from.

---

For install and build instructions, see [README.md](README.md). For security reporting, see [SECURITY.md](SECURITY.md).
