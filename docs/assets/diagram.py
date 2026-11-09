#!/usr/bin/env python3
"""Generates docs/assets/how-uqda-works.{en,ar}.svg.

Run from the repository root: python3 docs/assets/diagram.py
The Arabic version mirrors the layout for right-to-left reading.
"""
from xml.sax.saxutils import escape

W, H = 1200, 1190

T = {
"en": {
 "title": "How Uqda works",
 "subtitle": "Every device that installs Uqda becomes a node. The network is nothing but its nodes.",
 "a": "1 · Inside one node",
 "layers": [
  ("Your apps", "Browser, SSH, games, your service: plain IPv6"),
  ("uqda0 interface", "Address 200:1a2b:…::7 and subnet 300:1a2b:…/64"),
  ("Inbound firewall", "Replies pass; new traffic only if a rule allows it"),
  ("Encrypted sessions", "End to end, forward secrecy, optional group secret"),
  ("Routing (Ironwood)", "Spanning tree, greedy hops, bloom-filter lookups"),
  ("Peer links", "TCP, TLS, QUIC, WebSocket; each link proves the key"),
 ],
 "out": "send", "in": "receive",
 "id": ("Identity", ["Ed25519 key pair,", "made on the device,", "never leaves it.", "", "Address = hash of", "the public key, so", "nobody else can", "use it."]),
 "admin": ("Admin API", ["Local socket only", "uqda status", "uqda peers", "uqdactl …"]),
 "tele": ("No telemetry", ["Nothing phones home"]),
 "b": "2 · A network made of nodes",
 "nodes": {"R": "tree root", "P1": "public node", "P2": "public node", "H": "home server", "Q": "relay", "O": "office PC", "S": "relay", "ph": "phone", "lt": "laptop", "fr": "friend"},
 "lan": "Local network: found by multicast, works without internet",
 "path": "Your packet: relays forward it but cannot read it",
 "legend": ["node", "private group member", "tree link", "extra peer link", "your packet"],
 "c": "3 · A packet's journey",
 "steps": [
  ("App sends", ["A program sends to", "the friend's 200:…", "address like any", "IPv6 address."]),
  ("Find the key", ["The address maps to", "a key. A lookup", "follows bloom filters", "to the right branch."]),
  ("Encrypt", ["X25519 session keys,", "ratcheted. A group", "secret is mixed in", "for private groups."]),
  ("Hop by hop", ["Each node passes it", "to the peer closest", "to the target. Relays", "see size and time,", "never content."]),
  ("Decrypt and check", ["Only the target can", "open it. Replays", "and non-members", "are dropped."]),
  ("Firewall, then app", ["Rules decide. The app", "sees a normal IPv6", "connection from an", "address it can trust."]),
 ],
 "d": "4 · Joining and growing",
 "join": [
  ("sudo uqda setup country=…", ["Keeps or creates the identity", "Private group with a 192-bit secret", "Inbound firewall on", "Fastest healthy peers, kept by AutoPeers"]),
  ("sudo uqda invite", ["One-time code, also as a QR code", "Expires after 15 minutes", "Contains no password", "Works on the local network"]),
  ("sudo uqda join <invite>", ["Checks the inviter over TLS 1.3", "against its key", "Proves it holds the invite secret", "Receives the group secret and peers"]),
 ],
 "grow": "Every new node adds one more road for everyone.",
 "foot1": "Not anonymous: your peers see your IP address  ·  Experimental, no independent audit yet  ·  No telemetry",
 "foot2": "Uqda Core 26.1 · compatible with the Yggdrasil network · LGPLv3",
},
"ar": {
 "title": "كيف تعمل عُقدة",
 "subtitle": "كل جهاز يثبّت عُقدة يصبح عُقدة في الشبكة. الشبكة ليست إلا عُقدها.",
 "a": "١ · داخل العُقدة الواحدة",
 "layers": [
  ("تطبيقاتك", "المتصفح وSSH والألعاب وخدمتك: IPv6 عادي"),
  ("واجهة uqda0", "العنوان 200:1a2b:…::7 والشبكة الفرعية 300:…/64"),
  ("الجدار الناري للوارد", "الردود تمرّ؛ الجديد فقط إن سمحت به قاعدة"),
  ("جلسات مشفّرة", "من الطرف إلى الطرف، سرّية أمامية، سرّ مجموعة اختياري"),
  ("التوجيه (Ironwood)", "شجرة ممتدة، قفزات جشِعة، بحث بمرشّحات بلوم"),
  ("روابط الجيران", "TCP وTLS وQUIC وWebSocket؛ كل رابط يُثبت المفتاح"),
 ],
 "out": "إرسال", "in": "استقبال",
 "id": ("الهوية", ["زوج مفاتيح Ed25519", "يُنشأ على الجهاز", "ولا يغادره أبدًا.", "", "العنوان = بصمة", "المفتاح العام، فلا", "يستطيع أحد غيرك", "استخدامه."]),
 "admin": ("واجهة الإدارة", ["مقبس محلي فقط", "uqda status", "uqda peers", "uqdactl …"]),
 "tele": ("لا تتبّع", ["لا شيء يتصل بالمطوّرين"]),
 "b": "٢ · شبكة مكوّنة من العُقد",
 "nodes": {"R": "جذر الشجرة", "P1": "عُقدة عامة", "P2": "عُقدة عامة", "H": "خادم المنزل", "Q": "وسيط", "O": "حاسوب المكتب", "S": "وسيط", "ph": "هاتف", "lt": "حاسوب محمول", "fr": "صديق"},
 "lan": "الشبكة المحلية: اكتشاف تلقائي، تعمل دون إنترنت",
 "path": "رسالتك: الوسطاء يمرّرونها ولا يستطيعون قراءتها",
 "legend": ["عُقدة", "عضو في مجموعة خاصة", "رابط الشجرة", "رابط إضافي", "رسالتك"],
 "c": "٣ · رحلة الرسالة",
 "steps": [
  ("التطبيق يرسل", ["برنامج يرسل إلى", "عنوان الصديق 200:…", "كأي عنوان IPv6", "عادي."]),
  ("إيجاد المفتاح", ["العنوان يدلّ على", "مفتاح. البحث يتبع", "مرشّحات بلوم نحو", "الفرع الصحيح."]),
  ("التشفير", ["مفاتيح جلسة X25519", "تتجدّد باستمرار،", "ويُمزج سرّ المجموعة", "في الشبكات الخاصة."]),
  ("قفزة بعد قفزة", ["كل عُقدة تمرّرها", "للجار الأقرب إلى", "الهدف. الوسطاء يرون", "الحجم والتوقيت فقط،", "لا المحتوى أبدًا."]),
  ("فكّ التشفير والتحقق", ["الهدف وحده يفتحها.", "الرسائل المُعادة", "وغير الأعضاء", "تُرفض."]),
  ("الجدار ثم التطبيق", ["القواعد تقرّر. يرى", "التطبيق اتصال IPv6", "عاديًا من عنوان", "يمكنه الوثوق به."]),
 ],
 "d": "٤ · الانضمام ونموّ الشبكة",
 "join": [
  ("sudo uqda setup country=…", ["يحتفظ بالهوية أو ينشئها", "مجموعة خاصة بسرّ من 192 بت", "تفعيل الجدار الناري", "أسرع الجيران السليمين، يحافظ عليهم AutoPeers"]),
  ("sudo uqda invite", ["رمز لمرة واحدة، وأيضًا رمز QR", "ينتهي بعد 15 دقيقة", "لا يحتوي على كلمة مرور", "يعمل على الشبكة المحلية"]),
  ("sudo uqda join <invite>", ["يتحقق من الداعي عبر TLS 1.3", "مقابل مفتاحه", "يُثبت امتلاكه سرّ الدعوة", "يستلم سرّ المجموعة والجيران"]),
 ],
 "grow": "كل عُقدة جديدة تضيف طريقًا آخر للجميع.",
 "foot1": "ليست أداة إخفاء هوية: جيرانك يرون عنوان IP  ·  مشروع تجريبي دون تدقيق مستقل بعد  ·  لا تتبّع",
 "foot2": "Uqda Core 26.1 · متوافقة مع شبكة Yggdrasil · LGPLv3",
},
}

STYLE = """
.bg{fill:#f7f5f0}.card{fill:#ffffff;stroke:#d9d4c7}.panel{fill:#efece4;stroke:#ddd8cb}
.ink{fill:#1f2328}.muted{fill:#5b616b}.acc{fill:#0f766e}.acc2{fill:#b45309}
.layer{fill:#ffffff;stroke:#c9c3b4}.layerhi{fill:#e6f2f0;stroke:#0f766e}
.edge{stroke:#8b8f97;stroke-width:2}.mesh{stroke:#a9adb4;stroke-width:1.5;stroke-dasharray:5 5}
.pathl{stroke:#0f766e;stroke-width:5;stroke-linecap:round;stroke-linejoin:round;fill:none;opacity:.85}
.arrow{stroke:#5b616b;stroke-width:2;fill:none}.arrowh{fill:#5b616b}
.node{fill:#ffffff;stroke:#1f2328;stroke-width:2}.ring{fill:none;stroke:#b45309;stroke-width:3}
.lan{fill:#fff8ec;stroke:#b45309;stroke-dasharray:6 4}.num{fill:#0f766e}.numt{fill:#ffffff}
.mono{font-family:ui-monospace,SFMono-Regular,Consolas,"Liberation Mono",monospace}
@media (prefers-color-scheme: dark){
.bg{fill:#14171c}.card{fill:#1d2128;stroke:#343a44}.panel{fill:#191d23;stroke:#2c323b}
.ink{fill:#e8eaed}.muted{fill:#a3a9b3}.acc{fill:#2dd4bf}.acc2{fill:#f59e0b}
.layer{fill:#1d2128;stroke:#3a414c}.layerhi{fill:#123330;stroke:#2dd4bf}
.edge{stroke:#7d848f}.mesh{stroke:#5c636e}.pathl{stroke:#2dd4bf}
.arrow{stroke:#a3a9b3}.arrowh{fill:#a3a9b3}.node{fill:#1d2128;stroke:#e8eaed}.ring{stroke:#f59e0b}
.lan{fill:#2a2116;stroke:#f59e0b}.num{fill:#2dd4bf}.numt{fill:#14171c}}
"""


class SVG:
    def __init__(self, rtl):
        self.rtl = rtl
        self.out = []

    def X(self, x, w=0):
        return W - x - w if self.rtl else x

    def rect(self, x, y, w, h, cls, r=10):
        self.out.append(f'<rect x="{self.X(x, w)}" y="{y}" width="{w}" height="{h}" rx="{r}" class="{cls}"/>')

    def text(self, x, y, s, size=13, cls="ink", weight=400, anchor="start", mono=False):
        """x is the reading-start edge in LTR coordinates."""
        if anchor == "middle":
            xx, a = self.X(x), "middle"
        else:
            xx, a = self.X(x), "start"
        d = ' direction="rtl"' if self.rtl and not mono else ""
        if self.rtl and mono:
            a = "end"  # Latin commands stay left-to-right, aligned to the right edge
        c = cls + (" mono" if mono else "")
        self.out.append(f'<text x="{xx}" y="{y}" font-size="{size}" font-weight="{weight}" text-anchor="{a}" class="{c}"{d}>{escape(s)}</text>')

    def line(self, x1, y1, x2, y2, cls):
        self.out.append(f'<line x1="{self.X(x1)}" y1="{y1}" x2="{self.X(x2)}" y2="{y2}" class="{cls}"/>')

    def arrow(self, x1, y1, x2, y2):
        self.line(x1, y1, x2, y2, "arrow")
        import math
        X1, X2 = self.X(x1), self.X(x2)
        ang = math.atan2(y2 - y1, X2 - X1)
        p = [(X2, y2), (X2 - 10 * math.cos(ang - .45), y2 - 10 * math.sin(ang - .45)),
             (X2 - 10 * math.cos(ang + .45), y2 - 10 * math.sin(ang + .45))]
        self.out.append('<polygon points="' + " ".join(f"{a:.1f},{b:.1f}" for a, b in p) + '" class="arrowh"/>')

    def circle(self, x, y, r, cls):
        self.out.append(f'<circle cx="{self.X(x)}" cy="{y}" r="{r}" class="{cls}"/>')

    def polyline(self, pts, cls):
        self.out.append('<polyline points="' + " ".join(f"{self.X(a)},{b}" for a, b in pts) + f'" class="{cls}"/>')


def build(lang):
    t = T[lang]
    s = SVG(rtl=(lang == "ar"))
    font = '"Segoe UI",system-ui,-apple-system,"Noto Sans Arabic","Noto Sans",Tahoma,sans-serif'
    s.out.append(f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {W} {H}" width="{W}" height="{H}" font-family=\'{font}\' role="img" aria-labelledby="t d">')
    s.out.append(f'<title id="t">{escape(t["title"])}</title><desc id="d">{escape(t["subtitle"])}</desc>')
    s.out.append(f"<style>{STYLE}</style>")
    s.out.append(f'<rect width="{W}" height="{H}" class="bg"/>')

    s.text(40, 58, t["title"], 34, "ink", 700)
    s.text(40, 88, t["subtitle"], 16, "muted")

    # ---- Panel A: inside a node
    s.rect(30, 110, 560, 470, "panel", 16)
    s.text(55, 145, t["a"], 20, "acc", 700)
    lx, lw, ly, lh, gap = 85, 330, 170, 54, 12
    for i, (title, body) in enumerate(t["layers"]):
        y = ly + i * (lh + gap)
        s.rect(lx, y, lw, lh, "layerhi" if i in (2, 3) else "layer", 8)
        s.text(lx + 14, y + 22, title, 15, "ink", 700)
        s.text(lx + 14, y + 42, body, 12, "muted")
    top, bot = ly + 4, ly + 6 * (lh + gap) - gap - 4
    s.arrow(62, top, 62, bot)
    s.arrow(72, bot, 72, top)
    s.text(48, bot + 18, t["out"] + " ↓   " + t["in"] + " ↑", 11, "muted")
    # side column
    sx, sw = 432, 140
    s.rect(sx, 170, sw, 170, "card", 10)
    s.text(sx + 12, 194, t["id"][0], 15, "acc2", 700)
    for j, l in enumerate(t["id"][1]):
        s.text(sx + 12, 216 + j * 15, l, 12, "muted")
    s.arrow(sx, 263, lx + lw + 2, 263)
    s.rect(sx, 356, sw, 110, "card", 10)
    s.text(sx + 12, 380, t["admin"][0], 15, "ink", 700)
    for j, l in enumerate(t["admin"][1]):
        s.text(sx + 12, 402 + j * 16, l, 12, "muted", mono=(j > 0))
    s.rect(sx, 482, sw, 62, "card", 10)
    s.text(sx + 12, 506, t["tele"][0], 15, "acc", 700)
    s.text(sx + 12, 527, t["tele"][1][0], 12, "muted")

    # ---- Panel B: network of nodes
    s.rect(610, 110, 560, 470, "panel", 16)
    s.text(635, 145, t["b"], 20, "acc", 700)
    N = {"R": (890, 195), "P1": (760, 262), "P2": (1020, 262), "H": (690, 345), "Q": (835, 345),
         "O": (960, 345), "S": (1105, 345), "ph": (668, 436), "lt": (756, 436), "fr": (1055, 440)}
    s.rect(630, 404, 164, 80, "lan", 12)
    tree = [("R", "P1"), ("R", "P2"), ("P1", "H"), ("P1", "Q"), ("P2", "O"), ("P2", "S"), ("H", "ph"), ("H", "lt"), ("O", "fr")]
    mesh = [("Q", "O"), ("H", "Q"), ("S", "fr"), ("ph", "lt")]
    for a, b in mesh:
        s.line(*N[a], *N[b], "mesh")
    for a, b in tree:
        s.line(*N[a], *N[b], "edge")
    s.polyline([N["lt"], N["H"], N["Q"], N["O"], N["fr"]], "pathl")
    for k, (x, y) in N.items():
        r = 17 if k == "R" else 13
        s.circle(x, y, r, "node")
        if k in ("ph", "lt", "fr", "O"):
            s.circle(x, y, r + 5, "ring")
        s.text(x, y + (-26 if k == "R" else 30 if k in ("ph", "lt") else 34), t["nodes"][k], 12, "ink" if k != "R" else "acc", 600, anchor="middle")
    s.text(635, 502, t["lan"], 11, "acc2", 600)
    s.text(635, 527, t["path"], 13, "acc", 700)
    # legend
    lgx, lgy = 640, 552
    s.circle(lgx + 6, lgy - 4, 6, "node"); s.text(lgx + 18, lgy, t["legend"][0], 11, "muted")
    s.circle(lgx + 72, lgy - 4, 6, "node"); s.circle(lgx + 72, lgy - 4, 10, "ring"); s.text(lgx + 88, lgy, t["legend"][1], 11, "muted")
    s.line(lgx + 238, lgy - 4, lgx + 262, lgy - 4, "edge"); s.text(lgx + 268, lgy, t["legend"][2], 11, "muted")
    s.line(lgx + 335, lgy - 4, lgx + 359, lgy - 4, "mesh"); s.text(lgx + 365, lgy, t["legend"][3], 11, "muted")
    s.polyline([(lgx + 452, lgy - 4), (lgx + 476, lgy - 4)], "pathl"); s.text(lgx + 482, lgy, t["legend"][4], 11, "muted")

    # ---- Panel C: packet journey
    s.rect(30, 600, 1140, 240, "panel", 16)
    s.text(55, 635, t["c"], 20, "acc", 700)
    cw, cg, cx0, cy = 168, 16, 52, 655
    for i, (title, lines) in enumerate(t["steps"]):
        x = cx0 + i * (cw + cg)
        s.rect(x, cy, cw, 165, "card", 10)
        s.circle(x + 24, cy + 26, 14, "num")
        s.text(x + 24, cy + 31, str(i + 1) if lang == "en" else "١٢٣٤٥٦"[i], 14, "numt", 700, anchor="middle")
        s.text(x + 46, cy + 31, title, 14, "ink", 700)
        for j, l in enumerate(lines):
            s.text(x + 14, cy + 66 + j * 19, l, 12, "muted")
        if i < 5:
            s.arrow(x + cw + 2, cy + 82, x + cw + cg - 2, cy + 82)

    # ---- Panel D: joining
    s.rect(30, 860, 1140, 245, "panel", 16)
    s.text(55, 895, t["d"], 20, "acc", 700)
    jw, jg, jx0, jy = 340, 35, 52, 912
    for i, (cmd, lines) in enumerate(t["join"]):
        x = jx0 + i * (jw + jg)
        s.rect(x, jy, jw, 150, "card", 10)
        s.text(x + 16, jy + 30, cmd, 14, "acc2", 700, mono=True)
        for j, l in enumerate(lines):
            s.text(x + 16, jy + 60 + j * 21, "• " + l, 13, "ink")
        if i < 2:
            s.arrow(x + jw + 4, jy + 75, x + jw + jg - 4, jy + 75)
    s.text(55, 1088, t["grow"], 14, "acc", 700)

    # ---- footer
    s.text(40, 1140, t["foot1"], 13, "muted", 600)
    s.text(40, 1165, t["foot2"], 12, "muted")
    s.out.append("</svg>")
    return "\n".join(s.out) + "\n"


if __name__ == "__main__":
    import os
    here = os.path.dirname(os.path.abspath(__file__))
    for lang in ("en", "ar"):
        with open(os.path.join(here, f"how-uqda-works.{lang}.svg"), "w", encoding="utf-8", newline="\n") as f:
            f.write(build(lang))
    print("written")
