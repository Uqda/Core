# البدء السريع - شبكة عقدة

## البداية في 5 دقائق

### الخطوة 1: توليد التكوين

```bash
# توليد ملف تكوين بسيط
uqda -genconf > uqda.conf
```

### الخطوة 2: تعديل التكوين (اختياري)

```bash
# فتح الملف للتعديل
nano uqda.conf
```

**إضافة peers (أقران):**

```json
{
  "Peers": [
    "tls://peer1.example.com:9001",
    "tls://peer2.example.com:9001"
  ]
}
```

### الخطوة 3: تشغيل عقدة

```bash
# تشغيل مع ملف التكوين
sudo uqda -useconffile uqda.conf
```

**أو التشغيل التلقائي (للتجربة):**

```bash
# سينشئ مفاتيح عشوائية ويعمل تلقائياً
sudo uqda -autoconf
```

### الخطوة 4: التحقق من الحالة

**في terminal آخر:**

```bash
# عرض معلومات عقدتك
uqdactl getSelf

# عرض الأقران المتصلين
uqdactl getPeers

# عرض عنوان IPv6 الخاص بك
uqdactl getSelf | grep "IPv6 address"
```

## مثال عملي كامل

### 1. توليد التكوين

```bash
uqda -genconf > /tmp/uqda.conf
```

### 2. عرض التكوين

```bash
cat /tmp/uqda.conf
```

**مثال على المخرجات:**

```hjson
{
  // Private key (keep secret!)
  PrivateKey: abc123def456...
  
  // Listen addresses
  Listen: [
    "tls://[::]:9001"
  ]
  
  // Peers to connect to
  Peers: []
  
  // Interface name
  IfName: auto
}
```

### 3. إضافة peer عام

```bash
# تعديل الملف
nano /tmp/uqda.conf
```

**أضف peer:**

```json
{
  "Peers": [
    "tls://public-peer.example.com:9001"
  ]
}
```

### 4. تشغيل العقدة

```bash
sudo uqda -useconffile /tmp/uqda.conf
```

### 5. اختبار الاتصال

```bash
# في terminal آخر
# الحصول على عنوان IPv6
MY_IP=$(uqdactl getSelf | grep "IPv6 address" | awk '{print $3}')

# اختبار الاتصال
ping6 $MY_IP
```

## أوامر أساسية

### عرض معلومات العقدة

```bash
uqdactl getSelf
```

**المخرجات:**

```
Build name:        Uqda
Build version:     0.1.2
IPv6 address:      200:1234:5678:9abc::1
IPv6 subnet:       300:1234:5678:9abc::/64
Routing table:     42 entries
Public key:        abc123def456...
```

### عرض الأقران المتصلين

```bash
uqdactl getPeers
```

**المخرجات:**

```
URI                          State  Dir  IP Address              Uptime    RTT
tls://peer1.com:9001         Up     Out  200:1111::1            5m30s     12ms
tls://peer2.com:9001         Up     In   200:2222::1            2m15s     8ms
```

### عرض الجلسات النشطة

```bash
uqdactl getSessions
```

### عرض جدول التوجيه

```bash
uqdactl getPaths
```

## سيناريوهات الاستخدام

### السيناريو 1: ربط جهازين في نفس الشبكة

**على الجهاز الأول:**

```bash
# توليد التكوين
uqda -genconf > uqda1.conf

# تعديل التكوين لإضافة عنوان الاستماع
nano uqda1.conf
# أضف: "Listen": ["tls://[::]:9001"]

# تشغيل
sudo uqda -useconffile uqda1.conf

# عرض العنوان العام
uqdactl getSelf | grep "Public key"
```

**على الجهاز الثاني:**

```bash
# توليد التكوين
uqda -genconf > uqda2.conf

# إضافة الجهاز الأول كـ peer
nano uqda2.conf
# أضف في Peers:
# "tls://[IP-الجهاز-الأول]:9001"

# تشغيل
sudo uqda -useconffile uqda2.conf
```

### السيناريو 2: الاتصال بشبكة عامة

```bash
# توليد التكوين
uqda -genconf > uqda.conf

# إضافة peers عامة
nano uqda.conf
```

**أضف peers:**

```json
{
  "Peers": [
    "tls://public-peer1.example.com:9001",
    "tls://public-peer2.example.com:9001",
    "quic://public-peer3.example.com:9002"
  ]
}
```

```bash
# تشغيل
sudo uqda -useconffile uqda.conf
```

### السيناريو 3: اكتشاف تلقائي على الشبكة المحلية

```bash
# توليد التكوين
uqda -genconf > uqda.conf

# تعديل التكوين
nano uqda.conf
```

**أضف multicast:**

```json
{
  "MulticastInterfaces": [
    {
      "Regex": ".*",
      "Beacon": true,
      "Listen": true,
      "Port": 9001
    }
  ]
}
```

```bash
# تشغيل
sudo uqda -useconffile uqda.conf

# سيكتشف الأجهزة الأخرى تلقائياً!
```

## استكشاف الأخطاء

### المشكلة: لا توجد اتصالات

```bash
# التحقق من الأقران
uqdactl getPeers

# إذا كانت فارغة، جرب:
# 1. التحقق من التكوين
cat uqda.conf | grep -i peer

# 2. التحقق من الاتصال بالإنترنت
ping peer.example.com

# 3. التحقق من السجلات
journalctl -u uqda -f
```

### المشكلة: "Connection refused"

**الأسباب المحتملة:**

1. الـ peer غير متاح
2. Firewall يمنع الاتصال
3. عنوان خاطئ

**الحل:**

```bash
# اختبار الاتصال
telnet peer.example.com 9001

# أو
nc -zv peer.example.com 9001
```

### المشكلة: "Cannot bind to address"

**الحل:**

```bash
# التحقق من المنافذ المستخدمة
sudo netstat -tulpn | grep 9001

# تغيير المنفذ في التكوين
# "Listen": ["tls://[::]:9002"]
```

---

[السابق: التثبيت ←](02-installation.md) | [التالي: التكوين المتقدم →](04-configuration.md)

