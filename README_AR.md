<div dir="rtl" align="right">

<div align="center">
  <img src="contrib/logo/uqda-logo.png" alt="UQDA" width="720">
</div>

# نواة UQDA

**شبكة IPv6 متراكبة، مشفّرة وذاتية التنظيم**

[English](README.md) · **العربية**

> اسم **Uqda** مأخوذ من كلمة **عُقَد** العربية. المشروع موجّه لإنشاء شبكات IPv6 مشفّرة فوق اتصالات IPv4 أو IPv6 الموجودة مسبقًا.

## حالة المشروع

UQDA برنامج تجريبي ولم يخضع لتدقيق أمني مستقل. وهو ليس شبكة لإخفاء الهوية؛ فقد يستطيع النظير المتصل مباشرة معرفة عنوان IP الحقيقي للاتصال. لذلك يُنصح باستخدام جدار حماية IPv6 وعدم كشف الخدمات التي لا تريد إتاحتها لبقية المشاركين.

## المزايا

- تشفير طرف إلى طرف لحركة البيانات بين عُقد UQDA.
- توجيه متعدد القفزات وذاتي التنظيم دون سلطة توجيه مركزية.
- اشتقاق عنوان IPv6 من الهوية التشفيرية للعقدة.
- اتصال بين النظراء عبر TCP وTLS وQUIC وWebSocket وWSS وSOCKS ومقابس Unix.
- اكتشاف محلي اختياري للعُقد بواسطة multicast.
- واجهة TUN تتيح لتطبيقات IPv6 العادية استخدام الشبكة.
- كود وتكاملات لأنظمة Linux وmacOS وWindows وFreeBSD وOpenBSD وOpenWrt، إضافة إلى EdgeRouter وVyOS.
- إدارة محلية بواسطة أداة `uqdactl`.

## آلية العمل

تنشئ كل عقدة هوية تشفيرية وتشتق عنوان IPv6 من المفتاح العام. ينشئ البرنامج واجهة TUN افتراضية، ثم يتصل بالنظراء المضافين في الإعدادات أو المكتشفين على الشبكة المحلية، ويوجّه حزم IPv6 المشفّرة عبر المسارات المتاحة. ويمكن إنشاء روابط النظراء فوق شبكة IPv4 أو IPv6.

## البناء من المصدر

المتطلبات: [Go 1.25 أو أحدث](https://go.dev/dl/) وGit.

</div>

```bash
git clone https://github.com/Uqda/Core.git
cd Core
./build
```

<div dir="rtl" align="right">

ينتج البناء ملفين تنفيذيين:

- `uqda` — برنامج تشغيل العقدة؛
- `uqdactl` — أداة الإدارة المحلية.

## التشغيل

لتشغيل عقدة مؤقتة بإعدادات ومفاتيح مولّدة تلقائيًا:

</div>

```bash
sudo ./uqda -autoconf
```

<div dir="rtl" align="right">

لإنشاء إعدادات دائمة ومراجعتها قبل التشغيل:

</div>

```bash
./uqda -genconf > uqda.conf
$EDITOR uqda.conf
sudo ./uqda -useconffile ./uqda.conf
```

<div dir="rtl" align="right">

يمكن إضافة `-json` إلى أمر إنشاء الإعدادات للحصول على JSON بدل HJSON المشروح. يحتاج إنشاء واجهة TUN عادةً إلى صلاحيات المدير. وعلى Linux يمكن منح البرنامج قدرة الشبكة المطلوبة بدل تشغيله كاملًا بصلاحيات root:

</div>

```bash
sudo setcap CAP_NET_ADMIN=+eip ./uqda
./uqda -useconffile ./uqda.conf
```

<div dir="rtl" align="right">

## التشغيل بواسطة Docker

</div>

```bash
docker build -t uqda-core -f contrib/docker/Dockerfile .
docker run --rm -it \
  --cap-add=NET_ADMIN \
  --device=/dev/net/tun \
  -v uqda-config:/etc/uqda \
  uqda-core
```

<div dir="rtl" align="right">

ينشئ الحاوي إعداداتها داخل وحدة التخزين الدائمة `uqda-config` عند التشغيل الأول.

## تطوير المشروع

</div>

```bash
gofmt -w .
go vet ./...
go test ./...
go build ./...
```

<div dir="rtl" align="right">

اقرأ [دليل المساهمة](CONTRIBUTING.md) قبل إرسال التغييرات، وأبلغ عن الثغرات بشكل خاص وفق [سياسة الأمان](SECURITY.md).

## الترخيص

يُنشر UQDA Core وفق رخصة GNU LGPL الإصدار الثالث، مع الاستثناء الإضافي الموجود في ملف [LICENSE](LICENSE). وتبقى المكونات الخارجية خاضعة لتراخيصها الخاصة.

## الشكر والمصدر الأصلي

**مشروع UQDA Core مبني ومشتق من الشفرة مفتوحة المصدر لمشروع [Yggdrasil Network](https://github.com/yggdrasil-network/yggdrasil-go).** يحتفظ UQDA بجزء مهم من الأفكار والتسلسل البرمجي للمشروع الأصلي، مع تطويره باسم ومستودع مستقلين. مشروع Yggdrasil مشروع مستقل، ولا يعني هذا المستودع وجود اعتماد رسمي أو تأييد من مطوريه. راجع [NOTICE.md](NOTICE.md) لمزيد من تفاصيل الإسناد.

</div>
