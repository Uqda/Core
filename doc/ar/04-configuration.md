# دليل التكوين - شبكة عقدة

## ملف التكوين الأساسي

### توليد التكوين

```bash
# توليد تكوين بصيغة HJSON (مع تعليقات)
uqda -genconf > uqda.conf

# أو بصيغة JSON (للمعالجة البرمجية)
uqda -genconf -json > uqda.conf
```

### هيكل ملف التكوين

```json
{
  "PrivateKey": "...",
  "Listen": [],
  "Peers": [],
  "InterfacePeers": {},
  "MulticastInterfaces": [],
  "AllowedPublicKeys": [],
  "IfName": "auto",
  "IfMTU": 65535,
  "NodeInfo": {},
  "NodeInfoPrivacy": false
}
```

## الإعدادات الأساسية

### PrivateKey (المفتاح الخاص)

**⚠️ مهم جداً:** احتفظ بهذا المفتاح سرياً!

```json
{
  "PrivateKey": "abc123def456..."
}
```

**توليد مفتاح جديد:**

```bash
go run ./cmd/genkeys
```

### Listen (عناوين الاستماع)

**تعريف المنافذ التي تستمع عليها عقدتك:**

```json
{
  "Listen": [
    "tls://[::]:9001",
    "quic://[::]:9002"
  ]
}
```

**أنواع الاتصالات المدعومة:**

- `tcp://` - TCP عادي (غير مشفر)
- `tls://` - TCP مع TLS (موصى به)
- `quic://` - QUIC (الأسرع)
- `ws://` - WebSocket
- `wss://` - WebSocket Secure
- `unix://` - UNIX socket (محلي فقط)

### Peers (الأقران)

**قائمة الأقران للاتصال بهم:**

```json
{
  "Peers": [
    "tls://peer1.example.com:9001",
    "tls://peer2.example.com:9001",
    "quic://peer3.example.com:9002"
  ]
}
```

## التكوين المتقدم

### Peering مع كلمة مرور

```json
{
  "Peers": [
    "tls://peer.example.com:9001?password=mySecretPassword123"
  ]
}
```

**⚠️ يجب أن يكون نفس كلمة المرور على كلا الجانبين**

### تثبيت مفتاح عام لـ peer

```json
{
  "Peers": [
    "tls://peer.example.com:9001?key=expected-public-key-here"
  ]
}
```

**الاستخدام:** للتحقق من هوية الـ peer

### Peering عبر بروكسي SOCKS

```json
{
  "Peers": [
    "socks://127.0.0.1:9050/hidden.onion:9001",
    "socks://user:pass@proxy.example.com:1080/peer.example.com:9001"
  ]
}
```

**مفيد للاتصال عبر Tor أو بروكسي آخر**

### InterfacePeers (أقران على واجهات محددة)

**ربط peers على واجهات شبكة محددة:**

```json
{
  "InterfacePeers": {
    "eth0": [
      "tls://local-peer.lan:9001"
    ],
    "wlan0": [
      "tls://wifi-peer.lan:9001"
    ]
  }
}
```

**الاستخدام:** للتحكم في أي واجهة تستخدم لأي peer

### MulticastInterfaces (اكتشاف تلقائي)

**تفعيل الاكتشاف التلقائي على الشبكة المحلية:**

```json
{
  "MulticastInterfaces": [
    {
      "Regex": "eth.*",
      "Beacon": true,
      "Listen": true,
      "Port": 9001,
      "Priority": 1,
      "Password": ""
    }
  ]
}
```

**المعاملات:**

- `Regex`: نمط اسم الواجهة (مثل `eth.*` أو `wlan0`)
- `Beacon`: إرسال إشارات الاكتشاف (true/false)
- `Listen`: الاستماع لاكتشافات الآخرين (true/false)
- `Port`: المنفذ للاستخدام
- `Priority`: الأولوية (أقل = أعلى أولوية)
- `Password`: كلمة مرور اختيارية

### AllowedPublicKeys (الأقران المسموح بهم)

**تقييد الاتصالات الواردة:**

```json
{
  "AllowedPublicKeys": [
    "0abc123def456...",
    "789xyz123abc..."
  ]
}
```

**الاستخدام:** لإنشاء شبكة خاصة - فقط هذه المفاتيح العامة يمكنها الاتصال بك

### IfName (اسم الواجهة)

**تسمية واجهة TUN/TAP:**

```json
{
  "IfName": "uqda0"
}
```

**القيم:**

- `auto` - اختيار تلقائي
- `uqda0` - اسم محدد
- `tun0` - استخدام واجهة موجودة

### IfMTU (حجم الحزمة الأقصى)

```json
{
  "IfMTU": 65535
}
```

**القيم الموصى بها:**

- `1500` - للشبكات العادية
- `9000` - للشبكات عالية الأداء (Jumbo frames)
- `65535` - الحد الأقصى

### NodeInfo (معلومات العقدة)

**معلومات عامة عن عقدتك:**

```json
{
  "NodeInfo": {
    "name": "My Uqda Node",
    "location": "Damascus, Syria",
    "description": "Personal node"
  }
}
```

**⚠️ هذه المعلومات مرئية للجميع على الشبكة**

### NodeInfoPrivacy (خصوصية معلومات العقدة)

```json
{
  "NodeInfoPrivacy": false
}
```

**القيم:**

- `false` - مشاركة معلومات العقدة
- `true` - إخفاء معلومات العقدة

## أمثلة تكوين

### مثال 1: عقدة بسيطة

```json
{
  "Listen": [
    "tls://[::]:9001"
  ],
  "Peers": [
    "tls://public-peer.example.com:9001"
  ]
}
```

### مثال 2: عقدة مع اكتشاف تلقائي

```json
{
  "Listen": [
    "tls://[::]:9001"
  ],
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

### مثال 3: شبكة خاصة

```json
{
  "Listen": [
    "tls://[::]:9001"
  ],
  "Peers": [
    "tls://trusted-peer1.example.com:9001",
    "tls://trusted-peer2.example.com:9001"
  ],
  "AllowedPublicKeys": [
    "key1...",
    "key2..."
  ]
}
```

### مثال 4: عقدة متعددة الواجهات

```json
{
  "Listen": [
    "tls://[::]:9001"
  ],
  "InterfacePeers": {
    "eth0": [
      "tls://wired-peer.lan:9001"
    ],
    "wlan0": [
      "tls://wifi-peer.lan:9001"
    ]
  },
  "MulticastInterfaces": [
    {
      "Regex": "wlan.*",
      "Beacon": true,
      "Listen": true,
      "Port": 9001
    }
  ]
}
```

## إدارة التكوين

### تطبيع التكوين

```bash
# تطبيع التكوين (إزالة التعليقات، ترتيب الحقول)
uqda -useconffile uqda.conf -normaliseconf > uqda-normalized.conf
```

### تصدير المفتاح الخاص

```bash
# تصدير المفتاح بصيغة PEM
uqda -useconffile uqda.conf -exportkey > private-key.pem
```

### عرض العنوان

```bash
# عرض عنوان IPv6 الخاص بك
uqda -useconffile uqda.conf -address
```

### عرض الشبكة الفرعية

```bash
# عرض الشبكة الفرعية (/64)
uqda -useconffile uqda.conf -subnet
```

### عرض المفتاح العام

```bash
# عرض المفتاح العام
uqda -useconffile uqda.conf -publickey
```

## أفضل الممارسات

### 1. استخدام TLS/QUIC

**✅ جيد:**

```json
{
  "Listen": ["tls://[::]:9001"],
  "Peers": ["tls://peer.example.com:9001"]
}
```

**❌ سيء:**

```json
{
  "Listen": ["tcp://[::]:9001"],
  "Peers": ["tcp://peer.example.com:9001"]
}
```

### 2. حماية المفتاح الخاص

```bash
# صلاحيات الملف
chmod 600 uqda.conf

# ملكية الملف
chown root:root uqda.conf
```

### 3. استخدام AllowedPublicKeys للشبكات الخاصة

```json
{
  "AllowedPublicKeys": [
    "trusted-key-1",
    "trusted-key-2"
  ]
}
```

### 4. مراقبة السجلات

```bash
# على Linux
journalctl -u uqda -f

# أو
tail -f /var/log/uqda.log
```

---

[السابق: البدء السريع ←](03-quickstart.md) | [التالي: المفاهيم التقنية →](05-concepts.md)

