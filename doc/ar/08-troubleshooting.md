# استكشاف الأخطاء - شبكة عقدة

## مشاكل شائعة وحلولها

### 1. لا توجد اتصالات (No Connections)

#### الأعراض

```bash
uqdactl getPeers
# Output: (empty)
```

#### الحلول

**التحقق من التكوين:**

```bash
# فحص الأقران في التكوين
cat uqda.conf | grep -i peer

# فحص عناوين الاستماع
cat uqda.conf | grep -i listen
```

**التحقق من الاتصال بالإنترنت:**

```bash
# اختبار الاتصال
ping peer.example.com

# اختبار المنفذ
nc -zv peer.example.com 9001
```

**التحقق من السجلات:**

```bash
# على Linux
journalctl -u uqda -f

# أو
tail -f /var/log/uqda.log
```

**إعادة التشغيل:**

```bash
sudo systemctl restart uqda
```

### 2. "Connection refused"

#### الأعراض

```
Error: Connection refused
```

#### الأسباب المحتملة

1. الـ peer غير متاح
2. Firewall يمنع الاتصال
3. عنوان خاطئ

#### الحلول

**اختبار الاتصال:**

```bash
# اختبار TCP
telnet peer.example.com 9001

# أو
nc -zv peer.example.com 9001
```

**التحقق من Firewall:**

```bash
# على Linux
sudo iptables -L
sudo ip6tables -L

# فحص المنافذ المفتوحة
sudo netstat -tulpn | grep 9001
```

**التحقق من التكوين:**

```bash
# التأكد من صحة العنوان
cat uqda.conf | grep peer
```

### 3. "Cannot bind to address"

#### الأعراض

```
Error: Cannot bind to address [::]:9001
```

#### الحلول

**التحقق من المنافذ المستخدمة:**

```bash
sudo netstat -tulpn | grep 9001
```

**تغيير المنفذ:**

```json
{
  "Listen": ["tls://[::]:9002"]
}
```

**إيقاف الخدمة الأخرى:**

```bash
# إيجاد العملية
sudo lsof -i :9001

# إيقافها
sudo kill <PID>
```

### 4. "Permission denied" عند إنشاء TUN

#### الأعراض

```
Error: Permission denied
Cannot create TUN interface
```

#### الحلول

**على Linux:**

```bash
# إعطاء صلاحيات CAP_NET_ADMIN
sudo setcap cap_net_admin+eip /usr/local/bin/uqda

# أو تشغيل كـ root
sudo uqda -useconffile uqda.conf
```

**تحميل وحدة TUN:**

```bash
sudo modprobe tun
```

**التحقق من وجود TUN:**

```bash
ls /dev/net/tun
```

### 5. "Command not found"

#### الأعراض

```
bash: uqda: command not found
```

#### الحلول

**التحقق من المسار:**

```bash
which uqda
which uqdactl
```

**إضافة إلى PATH:**

```bash
export PATH=$PATH:/usr/local/bin

# أو إضافة إلى ~/.bashrc
echo 'export PATH=$PATH:/usr/local/bin' >> ~/.bashrc
source ~/.bashrc
```

**نسخ الملفات:**

```bash
sudo cp uqda /usr/local/bin/
sudo cp uqdactl /usr/local/bin/
```

### 6. الاتصالات تنقطع باستمرار

#### الأعراض

```
Peers connect then disconnect repeatedly
```

#### الحلول

**التحقق من استقرار الاتصال:**

```bash
# مراقبة الأقران
watch -n 1 'uqdactl getPeers'
```

**التحقق من السجلات:**

```bash
journalctl -u uqda -f | grep -i error
```

**زيادة timeout:**

```json
{
  "Peers": [
    "tls://peer.example.com:9001?timeout=30s"
  ]
}
```

**التحقق من Firewall:**

```bash
# قد يكون Firewall يقطع الاتصالات
sudo iptables -L -v
```

### 7. الأداء البطيء

#### الأعراض

```
Slow data transfer
High latency
```

#### الحلول

**استخدام QUIC:**

```json
{
  "Listen": ["quic://[::]:9002"],
  "Peers": [
    "quic://peer.example.com:9002"
  ]
}
```

**التحقق من المسار:**

```bash
# عرض المسارات
uqdactl getPaths

# اختيار أقصر مسار
```

**زيادة MTU:**

```json
{
  "IfMTU": 9000
}
```

**التحقق من الموارد:**

```bash
# استخدام CPU
top -p $(pgrep uqda)

# استخدام الذاكرة
ps aux | grep uqda
```

### 8. العنوان IPv6 لا يعمل

#### الأعراض

```
Cannot ping IPv6 address
Cannot access services
```

#### الحلول

**التحقق من العنوان:**

```bash
uqdactl getSelf | grep "IPv6 address"
```

**اختبار الاتصال:**

```bash
# ping نفسك
ping6 $(uqdactl getSelf | grep "IPv6 address" | awk '{print $3}')
```

**التحقق من واجهة TUN:**

```bash
# عرض الواجهة
ip addr show uqda0

# أو
ifconfig uqda0
```

**تفعيل IPv6:**

```bash
# على Linux
sudo sysctl -w net.ipv6.conf.all.disable_ipv6=0
```

### 9. Multicast Discovery لا يعمل

#### الأعراض

```
No automatic peer discovery
```

#### الحلول

**التحقق من التكوين:**

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

**التحقق من الواجهات:**

```bash
# عرض الواجهات المتاحة
ip link show

# أو
ifconfig
```

**التحقق من الصلاحيات:**

```bash
# قد تحتاج صلاحيات خاصة للـ multicast
sudo setcap cap_net_raw+eip /usr/local/bin/uqda
```

### 10. المشاكل على Windows

#### الأعراض

```
TUN interface creation fails
```

#### الحلول

**تثبيت TAP-Windows:**

```powershell
# تحميل من openvpn.org
# تثبيت TAP-Windows adapter
```

**تشغيل كمسؤول:**

```powershell
# تشغيل PowerShell كمسؤول
# ثم تشغيل uqda
```

**التحقق من Wintun:**

```powershell
# قد تحتاج Wintun driver
# تحميل من wintun.net
```

## أدوات التشخيص

### 1. عرض معلومات العقدة

```bash
uqdactl getSelf
```

### 2. عرض الأقران

```bash
uqdactl getPeers
```

### 3. عرض الجلسات

```bash
uqdactl getSessions
```

### 4. عرض المسارات

```bash
uqdactl getPaths
```

### 5. عرض الشجرة

```bash
uqdactl getTree
```

### 6. اختبار الاتصال

```bash
# ping
ping6 <ipv6-address>

# traceroute
traceroute6 <ipv6-address>
```

### 7. مراقبة السجلات

```bash
# Linux systemd
journalctl -u uqda -f

# Linux syslog
tail -f /var/log/syslog | grep uqda

# macOS
log stream --predicate 'process == "uqda"'
```

### 8. فحص الشبكة

```bash
# عرض الواجهات
ip addr show

# عرض الجداول
ip -6 route show

# عرض الاتصالات
netstat -tulpn | grep uqda
```

## الحصول على المساعدة

### 1. GitHub Issues

https://github.com/Uqda/Core/issues

### 2. البريد الإلكتروني

Uqda@proton.me

### 3. السجلات

عند الإبلاغ عن مشكلة، أرفق:

- السجلات الكاملة
- التكوين (بدون المفتاح الخاص!)
- معلومات النظام
- خطوات إعادة الإنتاج

---

[السابق: الأمان ←](07-security.md) | [التالي: التطوير →](09-development.md)

