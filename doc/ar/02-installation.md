# دليل التثبيت - شبكة عقدة

## متطلبات النظام

- **Go:** 1.22 أو أحدث
- **Git:** لأجل استنساخ المستودع
- **صلاحيات root/sudo:** لإنشاء واجهة TUN/TAP

## التثبيت من Go

### الطريقة السريعة

```bash
go install github.com/Uqda/Core/cmd/uqda@latest
go install github.com/Uqda/Core/cmd/uqdactl@latest
```

### التحقق من التثبيت

```bash
uqda -version
uqdactl -version
```

## التثبيت من المصدر

### Linux (Ubuntu/Debian)

```bash
# تثبيت المتطلبات
sudo apt update
sudo apt install -y golang-go git build-essential

# استنساخ المستودع
git clone https://github.com/Uqda/Core.git
cd Core

# البناء
./build

# نسخ الملفات التنفيذية
sudo cp uqda /usr/local/bin/
sudo cp uqdactl /usr/local/bin/

# التحقق
uqda -version
```

### Linux (Fedora/RHEL/CentOS)

```bash
# تثبيت المتطلبات
sudo dnf install -y golang git gcc

# استنساخ المستودع
git clone https://github.com/Uqda/Core.git
cd Core

# البناء
./build

# نسخ الملفات التنفيذية
sudo cp uqda /usr/local/bin/
sudo cp uqdactl /usr/local/bin/
```

### Linux (Arch Linux)

```bash
# تثبيت المتطلبات
sudo pacman -S go git base-devel

# استنساخ المستودع
git clone https://github.com/Uqda/Core.git
cd Core

# البناء
./build

# نسخ الملفات التنفيذية
sudo cp uqda /usr/local/bin/
sudo cp uqdactl /usr/local/bin/
```

### macOS

```bash
# تثبيت Go (إذا لم يكن مثبتاً)
brew install go git

# أو من golang.org
# https://golang.org/dl/

# استنساخ المستودع
git clone https://github.com/Uqda/Core.git
cd Core

# البناء
./build

# نسخ الملفات التنفيذية
sudo cp uqda /usr/local/bin/
sudo cp uqdactl /usr/local/bin/
```

### Windows

#### باستخدام PowerShell

```powershell
# تثبيت Go من golang.org
# https://golang.org/dl/

# استنساخ المستودع
git clone https://github.com/Uqda/Core.git
cd Core

# البناء
go build -o uqda.exe ./cmd/uqda
go build -o uqdactl.exe ./cmd/uqdactl

# الملفات التنفيذية ستكون في نفس المجلد
```

#### باستخدام WSL (Windows Subsystem for Linux)

```bash
# اتبع تعليمات Linux أعلاه داخل WSL
```

### FreeBSD

```bash
# تثبيت Go
sudo pkg install go git

# استنساخ المستودع
git clone https://github.com/Uqda/Core.git
cd Core

# البناء
./build

# نسخ الملفات التنفيذية
sudo cp uqda /usr/local/bin/
sudo cp uqdactl /usr/local/bin/
```

### OpenBSD

```bash
# تثبيت Go
doas pkg_add go git

# استنساخ المستودع
git clone https://github.com/Uqda/Core.git
cd Core

# البناء
./build

# نسخ الملفات التنفيذية
doas cp uqda /usr/local/bin/
doas cp uqdactl /usr/local/bin/
```

### Android (Termux)

```bash
# تثبيت المتطلبات
pkg install golang git

# استنساخ المستودع
git clone https://github.com/Uqda/Core.git
cd Core

# البناء
./build

# نسخ الملفات التنفيذية
cp uqda ~/../usr/bin/
cp uqdactl ~/../usr/bin/
```

### OpenWrt

```bash
# تثبيت المتطلبات
opkg update
opkg install golang git

# استنساخ المستودع
git clone https://github.com/Uqda/Core.git
cd Core

# البناء
./build

# نسخ الملفات التنفيذية
cp uqda /usr/bin/
cp uqdactl /usr/bin/
```

## التثبيت كخدمة (Linux - systemd)

### إنشاء ملف الخدمة

```bash
sudo nano /etc/systemd/system/uqda.service
```

```ini
[Unit]
Description=Uqda Network Router
Documentation=https://github.com/Uqda/Core
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/uqda -useconffile /etc/uqda/uqda.conf
ExecReload=/bin/kill -HUP $MAINPID
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### تفعيل الخدمة

```bash
# إنشاء مجلد التكوين
sudo mkdir -p /etc/uqda

# توليد التكوين
sudo uqda -genconf > /etc/uqda/uqda.conf

# تفعيل وتشغيل الخدمة
sudo systemctl daemon-reload
sudo systemctl enable uqda
sudo systemctl start uqda

# التحقق من الحالة
sudo systemctl status uqda
```

## التثبيت على macOS (LaunchDaemon)

### إنشاء ملف LaunchDaemon

```bash
sudo nano /Library/LaunchDaemons/network.uqda.uqda.plist
```

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key>
    <string>network.uqda.uqda</string>
    <key>ProgramArguments</key>
    <array>
      <string>/usr/local/bin/uqda</string>
      <string>-useconffile</string>
      <string>/etc/uqda/uqda.conf</string>
    </array>
    <key>KeepAlive</key>
    <true/>
    <key>RunAtLoad</key>
    <true/>
  </dict>
</plist>
```

### تفعيل الخدمة

```bash
# إنشاء مجلد التكوين
sudo mkdir -p /etc/uqda

# توليد التكوين
sudo uqda -genconf > /etc/uqda/uqda.conf

# تحميل الخدمة
sudo launchctl load /Library/LaunchDaemons/network.uqda.uqda.plist
```

## التثبيت على Windows (خدمة)

### استخدام NSSM (Non-Sucking Service Manager)

```powershell
# تحميل NSSM من nssm.cc

# تثبيت الخدمة
nssm install Uqda "C:\path\to\uqda.exe" "-useconffile C:\path\to\uqda.conf"

# بدء الخدمة
nssm start Uqda
```

## Cross-Compilation (البناء لأنظمة أخرى)

### بناء لـ Windows من Linux

```bash
GOOS=windows GOARCH=amd64 ./build
```

### بناء لـ macOS من Linux

```bash
GOOS=darwin GOARCH=amd64 ./build
```

### بناء لـ ARM (Raspberry Pi)

```bash
GOOS=linux GOARCH=arm GOARM=7 ./build
```

### بناء لـ MIPS (OpenWrt)

```bash
GOOS=linux GOARCH=mipsle ./build
```

## التحقق من التثبيت

```bash
# عرض معلومات الإصدار
uqda -version

# عرض معلومات العقدة (بعد التشغيل)
uqdactl getSelf

# عرض الأقران المتصلين
uqdactl getPeers
```

## استكشاف الأخطاء

### المشكلة: "Permission denied" عند إنشاء TUN

**الحل على Linux:**

```bash
# إعطاء صلاحيات CAP_NET_ADMIN
sudo setcap cap_net_admin+eip /usr/local/bin/uqda
```

### المشكلة: "Command not found"

**الحل:**

```bash
# التأكد من أن الملفات في PATH
which uqda
which uqdactl

# إضافة /usr/local/bin إلى PATH إذا لزم الأمر
export PATH=$PATH:/usr/local/bin
```

### المشكلة: "Cannot create TUN interface"

**الحل:**

```bash
# على Linux: تحميل وحدة TUN
sudo modprobe tun

# على macOS: قد تحتاج صلاحيات إدارية
sudo uqda -useconffile /etc/uqda/uqda.conf
```

---

[السابق: المقدمة ←](01-introduction.md) | [التالي: الإعداد السريع →](03-quickstart.md)

