# Privacy

Uqda Core is software you run yourself. **It contains no telemetry, no
analytics, no crash reporting and no update check.** The Uqda developers do not
run servers that your traffic passes through, and they receive no data from
your node.

This page explains what your node does share with others when it runs, so
you can decide how to use it.

## What other parties can see

| Who | What they see | When |
|---|---|---|
| Peers you connect to directly | Your IP address and port, your Uqda public key and address, traffic volume and timing | Always, for each peer link |
| Other nodes on the network | Your Uqda address and public key; whatever you put in `NodeInfo` | When they route to you or ask for your node info |
| Devices on your local network | That a Uqda node is running, its public key and listening port | When multicast discovery is on (the default) |
| The public peer catalog (`publicpeers.neilalexander.dev`, run by a third party) | Your IP address | Each time `uqda find`, `uqda setup` or `AutoPeers` fetches the list |
| Public peers measured by `find`, `setup` or `AutoPeers` | Your IP address | When they are tested |
| Services you reach over Uqda | Your Uqda address, which is tied to your node's key | Like any server sees its clients |

Traffic between nodes is end-to-end encrypted: nodes that forward it cannot
read its content. **Uqda does not make you anonymous.** If revealing your IP
address, location or identity would put you at risk, do not rely on Uqda to
protect you.

## What stays on your device

- Your private key and configuration (`uqda.conf`). Never share the file; it
  contains your private key.
- The log, if logging is enabled. It is written locally only and never sent
  anywhere. Remove keys and passwords before sharing a log with anyone.
- Invites: a one-time secret that is valid for 15 minutes and is never sent to
  any server.

## Reducing what you share

- Keep `NodeInfo` empty (the default). Do not put your name, location or
  contact details in it; every node on the network can read it.
- New configurations set `NodeInfoPrivacy: true`, so your platform and Uqda
  version are not shared. Older configurations may have it set to `false`.
- Turn off multicast discovery (`MulticastInterfaces: []`) on networks you do
  not trust.
- Use a private network with a group password ([networks](docs/networks.md))
  and keep the [firewall](docs/firewall.md) on.
- Choose peers you trust, or run your own, instead of the public catalog.

## If you run a public peer

A public peer sees the IP address of everyone who connects to it. If you run
one, you may become responsible under data-protection law (such as the GDPR)
for the connection data your server keeps. Keep logging minimal, keep logs only
as long as you need them, and publish how you handle them.

## Changes and contact

Changes to this page are listed in the [CHANGELOG](CHANGELOG.md). Questions:
<Uqda@proton.me>.

---

## بالعربية

عُقدة برنامج تشغّله بنفسك. **لا يحتوي على أي تتبّع أو تحليلات أو تقارير أعطال أو
فحص للتحديثات.** لا يشغّل مطوّرو عُقدة خوادم يمرّ عبرها اتصالك، ولا تصلهم أي
بيانات من جهازك.

ما يراه الآخرون:

- **من تتصل بهم مباشرة (peers):** عنوان IP الخاص بك، ومفتاحك العام وعنوانك في
  الشبكة، وحجم الاتصال وتوقيته.
- **العُقد الأخرى:** عنوانك في الشبكة ومفتاحك العام، وكل ما تضعه في `NodeInfo`.
- **الأجهزة على شبكتك المحلية:** أن عُقدة تعمل على جهازك، عند تفعيل الاكتشاف
  التلقائي (multicast)، وهو مفعّل افتراضيًا.
- **قائمة العُقد العامة** (`publicpeers.neilalexander.dev`، يشغّلها طرف آخر):
  عنوان IP الخاص بك في كل مرة تُستخدم فيها `uqda find` أو `uqda setup` أو
  `AutoPeers`.

الاتصال بين العُقد مشفّر من الطرف إلى الطرف، فلا تستطيع العُقد التي تمرّره قراءة
محتواه. لكن **عُقدة لا تخفي هويتك**: إن كان كشف عنوانك أو موقعك أو هويتك يعرّضك
للخطر، فلا تعتمد عليها لحمايتك.

يبقى على جهازك فقط: مفتاحك الخاص وملف الإعداد (لا تشاركه أبدًا)، والسجلّ إن
فعّلته، والدعوات.

لتقليل ما تشاركه: اترك `NodeInfo` فارغًا ولا تضع فيه اسمك أو موقعك، وأوقف
الاكتشاف التلقائي على الشبكات غير الموثوقة، واستخدم شبكة خاصة بكلمة مرور مع
تفعيل الجدار الناري.

إن شغّلت عُقدة عامة يتصل بها الآخرون، فقد تصبح مسؤولًا وفق قوانين حماية البيانات
(مثل GDPR) عن بيانات الاتصال التي يحتفظ بها خادمك؛ فقلّل السجلات واحذفها حين لا
تحتاجها.

للتواصل: <Uqda@proton.me>.
