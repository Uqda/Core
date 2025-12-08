# التطوير والمساهمة - شبكة عقدة

## البنية التقنية

### اللغة والاعتماديات

**اللغة:** Go (Golang) 1.22+

**الاعتماديات الرئيسية:**

- `golang.org/x/crypto` - التشفير
- `golang.org/x/net` - الشبكات
- `github.com/hjson/hjson-go` - تحليل HJSON
- `github.com/Arceliar/ironwood` - Mesh networking
- `golang.zx2c4.com/wireguard` - TUN interface

### هيكل الكود

```
src/
├── core/           # Core protocol logic
│   ├── core.go     # Main core type
│   ├── link.go     # Peering connections
│   ├── router.go   # Routing logic
│   └── proto.go    # Protocol handlers
├── tun/            # TUN/TAP interface
├── multicast/      # Multicast discovery
├── config/         # Configuration handling
├── admin/          # Admin API
└── address/        # Address derivation
```

## البناء من المصدر

### المتطلبات

```bash
# Go 1.22 أو أحدث
go version

# Git
git --version
```

### البناء

```bash
# استنساخ المستودع
git clone https://github.com/Uqda/Core.git
cd Core

# البناء
./build

# أو يدوياً
go build -o uqda ./cmd/uqda
go build -o uqdactl ./cmd/uqdactl
```

### Cross-Compilation

```bash
# Windows
GOOS=windows GOARCH=amd64 ./build

# macOS
GOOS=darwin GOARCH=amd64 ./build

# ARM (Raspberry Pi)
GOOS=linux GOARCH=arm GOARM=7 ./build

# MIPS (OpenWrt)
GOOS=linux GOARCH=mipsle ./build
```

## الاختبار

### تشغيل الاختبارات

```bash
# جميع الاختبارات
go test ./...

# اختبارات محددة
go test ./src/core/...

# مع verbose
go test -v ./...

# مع coverage
go test -cover ./...
```

### اختبارات التكامل

```bash
# اختبارات الشبكة
cd misc
./run-twolink-test
./run-schannel-netns
```

## المساهمة

### كيفية المساهمة

1. **Fork المستودع**
2. **إنشاء branch جديد**
3. **إجراء التغييرات**
4. **إضافة الاختبارات**
5. **إرسال Pull Request**

### معايير الكود

```bash
# تنسيق الكود
go fmt ./...

# فحص الأخطاء
go vet ./...

# golangci-lint
golangci-lint run
```

### رسائل Commit

```
type(scope): subject

body

footer
```

**الأمثلة:**

```
feat(core): add new routing algorithm

fix(config): resolve BOM encoding issue

docs(readme): update installation instructions
```

## هيكل المشروع

### المجلدات الرئيسية

```
Uqda/Core/
├── cmd/              # الأوامر التنفيذية
│   ├── uqda/         # البرنامج الرئيسي
│   └── uqdactl/      # أداة التحكم
├── src/              # الكود المصدري
│   ├── core/         # البروتوكول الأساسي
│   ├── config/       # التكوين
│   ├── tun/          # TUN/TAP
│   └── multicast/    # الاكتشاف
├── contrib/          # مساهمات إضافية
│   ├── systemd/      # ملفات systemd
│   ├── docker/       # Docker
│   └── ansible/      # Ansible
└── misc/             # أدوات إضافية
```

## API الداخلي

### Core API

```go
// إنشاء core جديد
core, err := core.New(certificate, logger, options...)

// الحصول على العنوان
address := core.Address()

// الحصول على الشبكة الفرعية
subnet := core.Subnet()

// إيقاف core
core.Stop()
```

### Admin API

```go
// إنشاء admin socket
admin, err := admin.New(core, logger, options...)

// إعداد handlers
admin.SetupAdminHandlers()
```

## التطوير المستقبلي

### الميزات المخططة

- [ ] تحسينات الأداء
- [ ] ميزات أمنية إضافية
- [ ] دعم منصات جديدة
- [ ] تحسينات التوجيه

### كيفية المساهمة

1. **اختر issue من GitHub**
2. **أنشئ branch جديد**
3. **اكتب الكود**
4. **أضف الاختبارات**
5. **أرسل PR**

---

[السابق: استكشاف الأخطاء ←](08-troubleshooting.md) | [التالي: الفهرس →](README.md)

