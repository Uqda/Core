# 🌐 شبكة عُقدة

<div align="center">

[![Build Status](https://github.com/Uqda/Core/actions/workflows/ci.yml/badge.svg)](https://github.com/Uqda/Core/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-LGPL--3.0-blue.svg)](../LICENSE)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.25-00ADD8.svg)](https://golang.org)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20Windows%20%7C%20macOS-lightgrey.svg)](#منصات-مدعومة)

**مشفرة من طرف إلى طرف • ذاتية الإصلاح • بدون إعدادات**

[البدء السريع](#-البدء-السريع) • [الوثائق](#-الوثائق) • [تحميل](https://github.com/Uqda/Core/releases) • [المجتمع](#-المجتمع)

</div>

---

## 📌 ما هي عُقدة؟

**شبكة عُقدة** (من العربية **عُقدة** بمعنى "عُقدة") هي بروتوكول توجيه لامركزي لبناء شبكات mesh متعددة القفزات مرنة وذاتية التنظيم مع تشفير من طرف إلى طرف.

### الميزات الرئيسية

- 🔒 **مشفرة من طرف إلى طرف** - جميع حركة المرور مشفرة افتراضياً (ChaCha20-Poly1305)
- 🌐 **متوافقة مع البروتوكول** - تعمل بسلاسة مع عقد Yggdrasil v0.5
- ⚡ **محسّنة للأداء** - تحسين زمن الاستجابة بمقدار 20-50ms عن الأساسي
- 🔄 **شبكة Mesh ذاتية الإصلاح** - اكتشاف المسار واستعادته تلقائياً
- 🎯 **مستقلة عن الموقع** - عنوان IPv6 دائم مشتق من هويتك
- 🪶 **بدون إعدادات** - تتشكل الشبكات تلقائياً
- 💰 **مجانية للأبد** - بدون تكلفة، بدون تسجيل، بدون سلطة مركزية

---

## ⚡ البدء السريع

### التثبيت

**Linux (Debian/Ubuntu)**
```bash
# تحميل حزمة .deb من الإصدارات
sudo dpkg -i uqda-debian-amd64.deb
```

**Windows**
```powershell
# تحميل وتشغيل مثبت .msi
# أو عبر سطر الأوامر:
msiexec /i uqda-windows-x64.msi
```

**macOS**
```bash
# تحميل وفتح مثبت .pkg
# أو البناء من المصدر (انظر أدناه)
```

**من المصدر**
```bash
# المتطلبات: Go 1.25.8+
git clone https://github.com/Uqda/Core.git
cd Core
./build
```

### تشغيل عُقدة

**الإعداد التلقائي (موصى به)**
```bash
sudo ./uqda -autoconf
```

**مع ملف الإعدادات**
```bash
# توليد الإعدادات
./uqda -genconf > uqda.conf

# تعديل uqda.conf حسب الحاجة، ثم:
sudo ./uqda -useconffile uqda.conf
```

> **ملاحظة:** صلاحيات Root/Administrator مطلوبة لإنشاء واجهات شبكة افتراضية.

---

## 🖥️ المنصات المدعومة

| المنصة | المعمارية | تنسيق الحزمة |
|----------|--------------|----------------|
| **Linux** | x86_64, ARM64 | `.deb` |
| **Windows** | x86_64, ARM64 | `.msi` |
| **macOS** | Intel, Apple Silicon | `.pkg` |

قم بتحميل الحزم المبنية مسبقاً من صفحة [الإصدارات](https://github.com/Uqda/Core/releases).

---

## 🏗️ البناء من المصدر

### المتطلبات
- Go 1.25.8 أو أحدث
- Git

### أوامر البناء
```bash
# استنساخ المستودع
git clone https://github.com/Uqda/Core.git
cd Core

# البناء لمنصتك
./build

# أمثلة التجميع المتقاطع
GOOS=windows GOARCH=amd64 ./build    # Windows 64-bit
GOOS=linux GOARCH=arm64 ./build      # Linux ARM64
GOOS=darwin GOARCH=arm64 ./build     # macOS Apple Silicon
```

---

## 📚 الوثائق

- **[الوثيقة التقنية](WHITEPAPER.md)** - الوثائق التقنية الكاملة
- **[الملخص التنفيذي](EXECUTIVE_SUMMARY.md)** - نظرة عامة في صفحة واحدة
- **[الأسئلة الشائعة](FAQ.md)** - الأسئلة المتكررة
- **[سياسة الأمان](../SECURITY.md)** - الإبلاغ عن الثغرات وأفضل الممارسات
- **[سجل التغييرات](../CHANGELOG.md)** - تاريخ الإصدارات والتغييرات
- **[الاعتراف](../ATTRIBUTION.md)** - الاعتمادات ومعلومات الترخيص

للحصول على وثائق مفصلة، زر [Wiki](https://github.com/Uqda/Core/wiki).

---

## 🔧 الإعدادات

يمكن لـ عُقدة أن تعمل في وضعين:

### وضع الإعداد التلقائي
يولد مفاتيح تشفير عشوائية عند كل بدء. مثالي للاختبار:
```bash
sudo uqda -autoconf
```

### وضع الإعدادات الثابتة
يستخدم ملف إعدادات مستمر:
```bash
# توليد الإعدادات
uqda -genconf > uqda.conf

# تعديل الملف لإضافة أقران، ثم:
sudo uqda -useconffile uqda.conf
```

أمثلة أقران لإضافتها إلى إعداداتك:
```conf
{
  Peers: [
    tcp://[2001:db8::1]:12345,
    tcp://example.com:12345
  ]
}
```

---

## 🤝 المجتمع

- **GitHub Discussions** - [اطرح الأسئلة وشارك الأفكار](https://github.com/Uqda/Core/discussions)
- **متتبع المشاكل** - [أبلغ عن الأخطاء](https://github.com/Uqda/Core/issues)
- **البريد الإلكتروني** - uqda@proton.me

نرحب بالمساهمات! راجع [CONTRIBUTING.md](../CONTRIBUTING.md) للإرشادات.

---

## 🔒 الأمان

وجدت ثغرة أمنية؟ من فضلك **لا** تفتح issue عام.

أرسل لنا بريداً خاصاً على: **uqda@proton.me**

---

## 📄 الترخيص

مرخصة تحت **GNU Lesser General Public License v3.0** مع استثناء توزيع ثنائي.

```
Copyright (C) 2025-2026 شبكة عُقدة
```

راجع [LICENSE](../LICENSE) للتفاصيل الكاملة.

---

## 🙏 الاعترافات

شبكة عُقدة مبنية على مشروع [Yggdrasil Network](https://yggdrasil-network.github.io/).

نشكر فريق Yggdrasil—Neil Alexander، Arceliar، وجميع المساهمين—على عملهم الرائد في الشبكات المشفرة اللامركزية.

للاعتراف الكامل، راجع [ATTRIBUTION.md](../ATTRIBUTION.md).

---

<div align="center">

**صُنع بـ ❤️ للشبكات اللامركزية**

[⬆ العودة للأعلى](#-شبكة-عُقدة)

</div>

