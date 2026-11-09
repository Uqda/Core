#!/usr/bin/env python3
"""Generates the technical poster docs/assets/how-uqda-works.{en,ar}.svg.

One tall page of engineering figures that explains the Uqda network from top
to bottom: topology, addressing, protocol stack, link handshake, routing,
lookups, a packet's path, visibility, groups, firewall and invites.

Run from the repository root: python3 docs/assets/diagram.py
The Arabic page mirrors the layout, except for bit layouts, packet formats and
configuration, which stay left to right as they are read in code.
"""
from xml.sax.saxutils import escape

W = 1200
L, R = 60, 1140  # content edges
SANS = '"Segoe UI","Helvetica Neue",Arial,"Noto Sans Arabic","Noto Sans",Tahoma,sans-serif'
MONO = 'Consolas,"SFMono-Regular","Liberation Mono",Menlo,monospace'

PAPER = "#ffffff"
INK = "#111827"
TEXT = "#374151"
GRAY = "#6b7280"
LINE = "#4b5563"
RULE = "#d1d5db"
FILL = "#f3f4f6"
BLUE = "#1d4ed8"      # data path, sessions
BLUE_BG = "#eff4ff"
ORANGE = "#c2410c"    # keys, cryptography
ORANGE_BG = "#fff4ec"
GREEN = "#15803d"
GREEN_BG = "#effaf3"
RED = "#b91c1c"
RED_BG = "#fdf0f0"
CODE_BG = "#0f172a"

AR_DIGITS = "٠١٢٣٤٥٦٧٨٩"


class Page:
    def __init__(self, lang):
        self.lang = lang
        self.rtl = lang == "ar"
        self.ltr = False  # temporarily disable mirroring
        self.o = []

    @property
    def mirrored(self):
        return self.rtl and not self.ltr

    def X(self, x, w=0):
        return round(W - x - w, 1) if self.mirrored else x

    def n(self, v):
        s = str(v)
        return "".join(AR_DIGITS[int(c)] if c.isdigit() else c for c in s) if self.rtl else s

    def rect(self, x, y, w, h, fill="none", stroke=LINE, sw=1.2, r=2, dash=None):
        d = f' stroke-dasharray="{dash}"' if dash else ""
        st = f' stroke="{stroke}" stroke-width="{sw}"' if stroke else ""
        self.o.append(f'<rect x="{self.X(x, w)}" y="{y}" width="{w}" height="{h}" rx="{r}" fill="{fill}"{st}{d}/>')

    def line(self, x1, y1, x2, y2, stroke=LINE, sw=1.2, dash=None):
        d = f' stroke-dasharray="{dash}"' if dash else ""
        self.o.append(f'<line x1="{self.X(x1)}" y1="{y1}" x2="{self.X(x2)}" y2="{y2}" stroke="{stroke}" stroke-width="{sw}"{d}/>')

    def path(self, pts, stroke=LINE, sw=1.2, dash=None):
        d = f' stroke-dasharray="{dash}"' if dash else ""
        p = " ".join(f"{self.X(a)},{b}" for a, b in pts)
        self.o.append(f'<polyline points="{p}" fill="none" stroke="{stroke}" stroke-width="{sw}" stroke-linejoin="round"{d}/>')

    def circle(self, x, y, r, fill=PAPER, stroke=LINE, sw=1.4):
        self.o.append(f'<circle cx="{self.X(x)}" cy="{y}" r="{r}" fill="{fill}" stroke="{stroke}" stroke-width="{sw}"/>')

    def head(self, x, y, d, color):
        """Arrow head at (x, y). d: 1 right, -1 left, 0 down, 2 up (page coordinates before mirroring)."""
        X = self.X(x)
        if d in (1, -1):
            s = -d if self.mirrored else d
            pts = [(X, y), (X - 9 * s, y - 4.5), (X - 9 * s, y + 4.5)]
        elif d == 0:
            pts = [(X, y), (X - 4.5, y - 9), (X + 4.5, y - 9)]
        else:
            pts = [(X, y), (X - 4.5, y + 9), (X + 4.5, y + 9)]
        self.o.append('<polygon points="' + " ".join(f"{a},{b}" for a, b in pts) + f'" fill="{color}"/>')

    def harrow(self, x1, x2, y, color=LINE, sw=1.4, dash=None, both=False):
        self.line(x1, y, x2, y, color, sw, dash)
        self.head(x2, y, 1 if x2 > x1 else -1, color)
        if both:
            self.head(x1, y, -1 if x2 > x1 else 1, color)

    def text(self, x, y, s, size=15, fill=TEXT, weight=400, anchor="start", mono=False, italic=False):
        a = anchor
        attrs = ""
        if self.mirrored:
            if mono:
                a = {"start": "end", "end": "start"}.get(anchor, anchor)
            else:
                attrs = ' direction="rtl"'
        fam = f" font-family='{MONO}'" if mono else ""
        it = ' font-style="italic"' if italic else ""
        sp = ' xml:space="preserve"' if mono else ""
        self.o.append(f'<text x="{self.X(x)}" y="{y}" font-size="{size}" font-weight="{weight}" fill="{fill}" text-anchor="{a}"{attrs}{fam}{it}{sp}>{escape(s)}</text>')

    def code(self, x, y, s, size=13, fill=INK, anchor="start", weight=400):
        """Monospace text that is never mirrored in its alignment (addresses, fields)."""
        self.o.append(f'<text x="{self.X(x)}" y="{y}" font-size="{size}" font-weight="{weight}" fill="{fill}" text-anchor="{anchor}" font-family=\'{MONO}\' xml:space="preserve">{escape(s)}</text>')

    def lines(self, x, y, items, size=15, fill=TEXT, lh=None, mono=False, weight=400):
        lh = lh or round(size * 1.5)
        for i, s in enumerate(items):
            if s:
                self.text(x, y + i * lh, s, size, fill, weight, mono=mono)
        return y + lh * len(items)

    def figure(self, y, num, title, caption):
        self.line(L, y, R, y, RULE, 1)
        label = f"FIGURE {num}" if not self.rtl else f"الشكل {self.n(num)}"
        self.text(L, y + 34, label, 13, GRAY, 700)
        self.text(L, y + 66, title, 26, INK, 700)
        self.lines(L, y + 98, caption, 16, GRAY, 24)
        return y + 98 + 24 * len(caption) + 14

    def node_box(self, x, y, w, name, addr, accent=None, h=46):
        self.rect(x, y, w, h, PAPER, accent or LINE, 2 if accent else 1.4)
        self.text(x + w / 2, y + 19, name, 14, INK, 600, anchor="middle")
        if addr:
            self.code(x + w / 2, y + 36, addr, 12, GRAY, anchor="middle")

    def lifeline(self, x, y1, y2, label, sub=None):
        h = 46 if sub else 34
        self.rect(x - 75, y1, 150, h, FILL, LINE, 1.2)
        self.text(x, y1 + 22, label, 14, INK, 700, anchor="middle")
        if sub:
            self.code(x, y1 + 39, sub, 11.5, GRAY, anchor="middle")
        self.line(x, y1 + h, x, y2, GRAY, 1, dash="4 4")

    def msg(self, x1, x2, y, label, color=LINE, dash=None, num=None, both=False, mono=False):
        self.harrow(x1, x2, y, color, 1.5, dash, both)
        lab = (self.n(num) + ". " if num else "") + label
        fill = INK if color == LINE else color
        w = len(lab) * (7.6 if mono else 7.0) + 16
        self.rect((x1 + x2) / 2 - w / 2, y - 25, w, 21, PAPER, None, r=0)
        if mono:
            self.code((x1 + x2) / 2, y - 9, lab, 13, fill, anchor="middle")
        else:
            self.text((x1 + x2) / 2, y - 9, lab, 13.5, fill, 500, anchor="middle")

    def note(self, x, y, w, lines, color=ORANGE, bg=ORANGE_BG, size=13):
        h = 14 + len(lines) * 19
        self.rect(x, y, w, h, bg, color, 1.2)
        for i, s in enumerate(lines):
            self.text(x + 12, y + 22 + i * 19, s, size, INK)
        return y + h


# ---------------------------------------------------------------------------
# Text

T = {
"en": {
 "meta": "UQDA CORE 26.1  ·  TECHNICAL OVERVIEW",
 "title": "How the Uqda network works",
 "sub": ["Uqda is an encrypted IPv6 overlay network built only from its users' devices. Each device runs a",
         "node; nodes link to nearby nodes and route each other's traffic. There is no central server."],
 "specs": [("Addresses", "200::/7, derived from Ed25519 public keys"), ("Routing", "Ironwood: spanning tree, greedy forwarding"),
           ("Sessions", "X25519 + XSalsa20-Poly1305, ratcheted keys"), ("Network", "compatible with Yggdrasil")],

 "f1": ("Topology: nodes, peer links and sessions", ["Every box is a device running Uqda. Peer links use whatever network exists;",
        "an encrypted session runs end to end across any number of them."]),
 "home": "Home network 192.168.1.0/24", "inet": "Internet", "remote": "Friend's network",
 "n_phone": "Phone", "n_laptop": "Laptop", "n_srv": "Home server", "n_pa": "Public peer A", "n_pb": "Public peer B", "n_friend": "Friend's PC",
 "multicast": "LAN peers found by multicast; works without internet",
 "session": "end-to-end session: Laptop ↔ Friend's PC",
 "lg": ["peer link (TCP, TLS, QUIC, WebSocket)", "encrypted session", "physical network"],

 "f2": ("Addressing: the address is computed from the key", ["The public key is inverted bit by bit; its leading 1-bits are counted and dropped. Nobody can",
        "use an address without its private key, so the source address of a packet identifies the sender."]),
 "priv": ["Private key", "Ed25519, 32 bytes", "never leaves the device"],
 "pub": ["Public key", "32 bytes", "sent to peers"],
 "derive": ["invert bits,", "count leading 1s"],
 "example": "Example",
 "ex_a": "200:1a2b:3c4d::7", "ex_s": "300:1a2b:3c4d::/64",
 "addr_row": "Node address, 128 bits", "sub_row": "Routed /64 subnet for devices behind the node",
 "fields_a": [("0x02", "prefix 200::/7", "bits 0-7"), ("N", "count of leading 1s", "bits 8-15"), ("key bits after the first 0", "", "bits 16-127")],
 "fields_s": [("0x03", "prefix 300::/8", "bits 0-7"), ("N", "count of leading 1s", "bits 8-15"), ("key bits", "", "bits 16-63"), ("your devices", "", "bits 64-127")],

 "f3": ("Protocol stack and encapsulation", ["Programs see an ordinary IPv6 interface. Below it, Uqda filters, encrypts and routes,",
        "then hands frames to a normal network connection."]),
 "stack": [("Application", "HTTP, SSH, games, chat", 0), ("Transport", "TCP, UDP, ICMPv6", 0),
           ("IPv6 on uqda0", "200::/7, 300::/8", 1), ("Inbound firewall", "stateful, allow rules", 1),
           ("Session", "end-to-end encryption", 1), ("Routing (Ironwood)", "coordinates, lookups", 1),
           ("Link", "TCP, TLS, QUIC, WS, SOCKS", 1), ("Network", "Ethernet, Wi-Fi, mobile", 0)],
 "brace": "Uqda",
 "enc_title": "One outgoing packet",
 "enc": [("from the app", [("IPv6 header", 1.2, FILL), ("TCP", .8, FILL), ("data", 2.2, PAPER)]),
         ("encrypt", [("nonce", .8, ORANGE_BG), ("ciphertext: the whole IPv6 packet", 3.0, ORANGE_BG), ("tag", .6, ORANGE_BG)]),
         ("route", [("routing header", 1.4, BLUE_BG), ("encrypted session payload", 3.0, ORANGE_BG)]),
         ("send", [("link framing / TLS", 1.4, FILL), ("Ironwood frame", 3.0, BLUE_BG)])],

 "f4": ("Peering: the link handshake", ["Two nodes open a connection and prove their keys. Between Uqda nodes the proof covers fresh",
        "random values, so a recorded handshake cannot be replayed to impersonate a node."]),
 "hs": [("connect over TCP, TLS, QUIC or WebSocket", "both", LINE),
        ("metadata: version, public key, nonce (field 0x7571), signature", "both", LINE),
        ("Ed25519 signature over (B's nonce, A's nonce, both keys)", "ab", ORANGE),
        ("Ed25519 signature over (A's nonce, B's nonce, both keys)", "ba", ORANGE)],
 "hs_note": "Both signatures verify: the link is up. A replayed handshake fails, because the nonce is new on every connection.",
 "hs_after": "tree announcements and bloom filters (gossip)",
 "hs_foot": "Yggdrasil nodes and older Uqda releases skip steps 3 and 4; RequireFreshHandshake: true rejects them.",

 "f5": ("Routing: coordinates on a spanning tree", ["Nodes agree on a tree by gossip. A node's coordinates are its path from the root. Each hop goes",
        "to the peer with the smallest tree distance to the destination, so extra links become shortcuts."]),
 "root": "root",
 "dist": "distance(a, b) = depth(a) + depth(b) - 2 x common_prefix(a, b)",
 "tbl_h": ["hop", "at node", "next hop", "distance"],
 "tbl": [("1", "[1,2,1]", "[1,2]", "5"), ("2", "[1,2]", "[2,1]", "1"), ("3", "[2,1]", "[2,1,1]", "0")],
 "shortcut": "shortcut",
 "tbl_note": ["Along the tree: 6 hops.", "Greedy, using the shortcut: 3 hops.", "No node stores a routing table", "for the whole network."],
 "src": "source", "dst": "destination",

 "f6": ("Finding a node: lookups and repair", ["The destination address holds only part of the key. A lookup travels towards nodes whose",
        "bloom filters may contain it; the answer carries the full key and current coordinates."]),
 "lk_life": ["Node A", "Relay", "Relay", "Node B"],
 "lk1": "lookup(partial key of 200:f3…)",
 "lk_bloom": "forwarded only on links whose bloom filter matches",
 "lk2": "response: full key and coordinates",
 "lk3": "traffic, by greedy routing",
 "lk4": "link fails: \"path broken\" back to the sender",
 "lk5": "A looks B up again; traffic takes the new path.",

 "f7": ("One packet, end to end", ["SSH from the laptop to the friend's PC. Relays carry ciphertext they cannot read or change."]),
 "j_life": [("Laptop", "ssh"), ("Laptop", "uqda"), ("Relay", "uqda"), ("Relay", "uqda"), ("Friend's PC", "uqda"), ("Friend's PC", "sshd")],
 "j1": "IPv6 to 200:f3…:22",
 "j2": ["encrypt: X25519 session keys,", "XSalsa20-Poly1305; group secret", "mixed in for private groups"],
 "j3": "ciphertext", "j4": "to the closest peer", "j5": "to the closest peer",
 "j6": ["decrypt; drop replays", "and non-members"],
 "j7": ["firewall: allow tcp/22", "from 200:a2… (laptop)"],
 "j8": "from 200:a2…",
 "j9": "reply on the same flow, allowed back automatically",
 "relay_sees": ["Relays see the keys in the routing header, size and timing.", "They cannot read, modify or replay the content."],

 "f8": ("Who can see what", ["Uqda encrypts content end to end. It does not hide that you use it, or whom you talk to."]),
 "vh": ["", "your IP address", "your Uqda address", "content", "size, timing"],
 "vr": [("Your internet provider", "y", "n1", "n", "y"),
        ("Your direct peers", "y", "y", "n", "y"),
        ("Relay nodes on the path", "n", "y", "n", "y"),
        ("The node you talk to", "n2", "y", "y", "y"),
        ("Devices on your LAN (multicast on)", "y", "y", "n", "n"),
        ("The Uqda developers", "n", "n", "n", "n")],
 "yes": "yes", "no": "no",
 "vfoot": ["¹  Only on plain TCP links. TLS and QUIC links hide it from your provider.",
           "²  Unless that node is also one of your direct peers. Uqda has no telemetry: nothing is sent to its developers."],

 "f9": ("Public network and private groups", ["Both use the same overlay and the same relays. A group secret limits who can open a session."]),
 "g_pub": ("GroupPassword empty: public", ["Any node can open a session with yours.", "The firewall decides what it may reach."]),
 "g_priv": ("GroupPassword set: private group", ["Only members can open sessions. Other nodes", "still relay the members' ciphertext."]),
 "member": "member", "outsider": "other node", "relays": "relays only", "no_session": "no session",

 "f10": ("Inbound firewall", ["Applied to traffic that arrives over Uqda, before it reaches your programs. On for new nodes."]),
 "fw_h": ["#", "source", "protocol", "ports", "comment", "action"],
 "fw": [("-", "reply to a flow this node opened", "any", "any", "stateful", "allow"),
        ("1", "200:a2… (laptop)", "tcp", "22", "laptop SSH", "allow"),
        ("2", "*", "icmp", "-", "AllowPing", "allow"),
        ("-", "anything else", "", "", "counted", "drop")],
 "fw_cfg_t": "uqda.conf",
 "fw_cfg": ["Firewall: {", "  Enabled: true", "  AllowPing: true", "  Rules: [", "    { Sources: [\"200:a2…\"],", "      Protocol: \"tcp\", Ports: [\"22\"],", "      Comment: \"laptop SSH\" }", "  ]", "}"],
 "fw_eval_t": "Order of evaluation",
 "fw_eval": ["1. part of a flow this node opened: allow", "2. first matching rule: allow", "3. anything else: drop and count"],

 "f11": ("Joining a private group with an invite", ["The invite is a one-time code without a password. The group secret is sent only to a device",
        "that reaches the inviter over TLS 1.3 pinned to the inviter's key and presents the code's secret."]),
 "inv_life": [("New device", "uqda join"), ("Inviter", "uqda invite")],
 "inv_pre": ["uqda setup: key pair, group secret", "(192 bits), firewall, peers"],
 "inv_tok": ["token: inviter key, up to 4 addresses,", "one-time secret, expiry (15 minutes)"],
 "inv_oob": "token shown as text and QR code; the user copies it",
 "inv_tls": "TLS 1.3 to an address in the token; certificate key must equal token key",
 "inv_sec": "one-time secret",
 "inv_chk": ["secret valid, not expired,", "not used before: mark used"],
 "inv_grant": "grant: group secret and peers",
 "inv_done": ["writes uqda.conf (backup kept);", "after a restart it is a member"],

 "f12": ("Building on Uqda", ["Anything that listens on IPv6 works unchanged. The source address of a connection is authenticated."]),
 "b_left": ["Host a site or service at home, without port forwarding.", "Use the client's 200:… address as its login.",
            "Read node state through the local admin API (JSON).", "Embed a node in apps with the Android and Apple SDKs."],
 "b_code": ["$ python3 -m http.server 8080 --bind ::", "", "# uqda.conf", "Rules: [ { Sources: [\"*\"], Protocol: \"tcp\",", "          Ports: [\"8080\"] } ]", "", "# any Uqda user can now open", "http://[200:1a2b:3c4d::7]:8080/"],

 "limits_t": "Limits",
 "limits": ["Not an anonymity network: peers see your IP address, relays see who talks to whom.",
            "Experimental: tested extensively, but no independent security audit yet.",
            "Paths can be longer than a direct connection, so some routes are slower.",
            "No telemetry, no accounts, no central servers."],
 "foot": "Uqda Core 26.1  ·  free software, LGPLv3  ·  compatible with the Yggdrasil network  ·  github.com/Uqda/Core",
},
"ar": {
 "meta": "Uqda Core 26.1  ·  نظرة تقنية",
 "title": "كيف تعمل شبكة عُقدة",
 "sub": ["عُقدة شبكة IPv6 مشفّرة فوق الإنترنت، مكوّنة فقط من أجهزة مستخدميها. كل جهاز يشغّل عُقدة،",
         "والعُقد تتصل بالعُقد القريبة منها وتوجّه اتصالات بعضها. لا يوجد خادم مركزي."],
 "specs": [("العناوين", "200::/7، محسوبة من مفاتيح Ed25519 العامة"), ("التوجيه", "Ironwood: شجرة ممتدة وتمرير جشِع"),
           ("الجلسات", "X25519 + XSalsa20-Poly1305 مع تجديد المفاتيح"), ("الشبكة", "متوافقة مع Yggdrasil")],

 "f1": ("البنية: العُقد وروابط الجيران والجلسات", ["كل مربع جهاز يشغّل عُقدة. روابط الجيران تستخدم أي شبكة متاحة،",
        "والجلسة المشفّرة تمتد من الطرف إلى الطرف عبر أي عدد منها."]),
 "home": "شبكة المنزل 192.168.1.0/24", "inet": "الإنترنت", "remote": "شبكة الصديق",
 "n_phone": "هاتف", "n_laptop": "حاسوب محمول", "n_srv": "خادم المنزل", "n_pa": "عُقدة عامة A", "n_pb": "عُقدة عامة B", "n_friend": "حاسوب الصديق",
 "multicast": "جيران الشبكة المحلية يُكتشفون تلقائيًا، دون حاجة للإنترنت",
 "session": "جلسة من الطرف إلى الطرف: الحاسوب المحمول ↔ حاسوب الصديق",
 "lg": ["رابط جيران (TCP وTLS وQUIC وWebSocket)", "جلسة مشفّرة", "شبكة فعلية"],

 "f2": ("العنونة: العنوان محسوب من المفتاح", ["تُعكس بتات المفتاح العام، وتُعدّ بتات 1 في بدايته وتُحذف. لا يستطيع أحد استخدام",
        "عنوان دون مفتاحه الخاص، لذلك يدل عنوان المصدر في الرسالة على مرسلها."]),
 "priv": ["المفتاح الخاص", "Ed25519، بطول 32 بايت", "لا يغادر الجهاز"],
 "pub": ["المفتاح العام", "32 بايت", "يُرسل إلى الجيران"],
 "derive": ["عكس البتات،", "وعدّ بتات 1 الأولى"],
 "example": "مثال",
 "ex_a": "200:1a2b:3c4d::7", "ex_s": "300:1a2b:3c4d::/64",
 "addr_row": "Node address, 128 bits  ·  عنوان العُقدة", "sub_row": "Routed /64 subnet  ·  شبكة فرعية للأجهزة خلف العُقدة",
 "fields_a": [("0x02", "prefix 200::/7", "bits 0-7"), ("N", "leading 1s", "bits 8-15"), ("key bits after the first 0", "", "bits 16-127")],
 "fields_s": [("0x03", "prefix 300::/8", "bits 0-7"), ("N", "leading 1s", "bits 8-15"), ("key bits", "", "bits 16-63"), ("your devices", "", "bits 64-127")],

 "f3": ("طبقات البروتوكول والتغليف", ["ترى البرامج واجهة IPv6 عادية. تحتها تفلتر عُقدة الاتصال وتشفّره وتوجّهه،",
        "ثم تسلّمه إلى اتصال شبكة عادي."]),
 "stack": [("التطبيق", "HTTP وSSH والألعاب", 0), ("النقل", "TCP وUDP وICMPv6", 0),
           ("IPv6 على uqda0", "200::/7 و300::/8", 1), ("الجدار الناري للوارد", "يتتبّع الحالة، قواعد سماح", 1),
           ("الجلسة", "تشفير من الطرف إلى الطرف", 1), ("التوجيه (Ironwood)", "إحداثيات وبحث", 1),
           ("الرابط", "TCP وTLS وQUIC وWS", 1), ("الشبكة", "Ethernet وWi-Fi والجوال", 0)],
 "brace": "عُقدة",
 "enc_title": "One outgoing packet  ·  رسالة صادرة واحدة",
 "enc": [("from the app", [("IPv6 header", 1.2, FILL), ("TCP", .8, FILL), ("data", 2.2, PAPER)]),
         ("encrypt", [("nonce", .8, ORANGE_BG), ("ciphertext: the whole IPv6 packet", 3.0, ORANGE_BG), ("tag", .6, ORANGE_BG)]),
         ("route", [("routing header", 1.4, BLUE_BG), ("encrypted session payload", 3.0, ORANGE_BG)]),
         ("send", [("link framing / TLS", 1.4, FILL), ("Ironwood frame", 3.0, BLUE_BG)])],

 "f4": ("الاتصال بالجيران: مصافحة الرابط", ["تفتح عُقدتان اتصالًا وتُثبت كل منهما مفتاحها. بين عُقد Uqda يشمل الإثبات قيمًا",
        "عشوائية جديدة، فلا يمكن إعادة مصافحة مسجّلة لانتحال عُقدة."]),
 "hs": [("اتصال عبر TCP أو TLS أو QUIC أو WebSocket", "both", LINE),
        ("بيانات: الإصدار والمفتاح العام وnonce (الحقل 0x7571) وتوقيع", "both", LINE),
        ("توقيع Ed25519 على (nonce الطرف B، وnonce الطرف A، والمفتاحين)", "ab", ORANGE),
        ("توقيع Ed25519 على (nonce الطرف A، وnonce الطرف B، والمفتاحين)", "ba", ORANGE)],
 "hs_note": "يتحقق الطرفان من التوقيعين فيعمل الرابط. المصافحة المُعادة تفشل لأن الـ nonce جديد في كل اتصال.",
 "hs_after": "إعلانات الشجرة ومرشّحات بلوم (تناقل)",
 "hs_foot": "عُقد Yggdrasil وإصدارات Uqda القديمة تتخطى الخطوتين ٣ و٤، والخيار RequireFreshHandshake: true يرفضها.",

 "f5": ("التوجيه: إحداثيات على شجرة ممتدة", ["تتفق العُقد على شجرة بالتناقل، وإحداثيات العُقدة هي طريقها من الجذر. كل قفزة تذهب",
        "إلى الجار الأقل مسافة على الشجرة نحو الهدف، فتصبح الروابط الإضافية طرقًا مختصرة."]),
 "root": "الجذر",
 "dist": "distance(a, b) = depth(a) + depth(b) - 2 x common_prefix(a, b)",
 "tbl_h": ["القفزة", "عند العُقدة", "القفزة التالية", "المسافة"],
 "tbl": [("1", "[1,2,1]", "[1,2]", "5"), ("2", "[1,2]", "[2,1]", "1"), ("3", "[2,1]", "[2,1,1]", "0")],
 "shortcut": "طريق مختصر",
 "tbl_note": ["على الشجرة: ٦ قفزات.", "التمرير الجشِع مع المختصر: ٣ قفزات.", "لا تحتفظ أي عُقدة بجدول توجيه", "للشبكة كلها."],
 "src": "المصدر", "dst": "الهدف",

 "f6": ("إيجاد عُقدة: البحث والإصلاح", ["عنوان الهدف يحمل جزءًا من المفتاح فقط. يسير البحث نحو العُقد التي قد تحتويه",
        "مرشّحات بلوم لديها، ويحمل الرد المفتاح الكامل والإحداثيات الحالية."]),
 "lk_life": ["العُقدة A", "وسيط", "وسيط", "العُقدة B"],
 "lk1": "lookup(partial key of 200:f3…)",
 "lk_bloom": "يُمرَّر فقط على الروابط التي يطابقها مرشّح بلوم",
 "lk2": "الرد: المفتاح الكامل والإحداثيات",
 "lk3": "الاتصال، بالتمرير الجشِع",
 "lk4": "تعطّل رابط: إشعار \"path broken\" إلى المرسل",
 "lk5": "تبحث A عن B مجددًا، ويسلك الاتصال الطريق الجديد.",

 "f7": ("رسالة واحدة من الطرف إلى الطرف", ["SSH من الحاسوب المحمول إلى حاسوب الصديق. الوسطاء يحملون نصًا مشفّرًا لا يستطيعون قراءته أو تغييره."]),
 "j_life": [("الحاسوب", "ssh"), ("الحاسوب", "uqda"), ("وسيط", "uqda"), ("وسيط", "uqda"), ("حاسوب الصديق", "uqda"), ("حاسوب الصديق", "sshd")],
 "j1": "IPv6 to 200:f3…:22",
 "j2": ["تشفير: مفاتيح جلسة X25519،", "XSalsa20-Poly1305، ويُمزج سرّ", "المجموعة في الشبكات الخاصة"],
 "j3": "نص مشفّر", "j4": "إلى الجار الأقرب", "j5": "إلى الجار الأقرب",
 "j6": ["فكّ التشفير، ورفض المُعاد", "وغير الأعضاء"],
 "j7": ["الجدار: السماح بـ tcp/22", "من 200:a2… (الحاسوب)"],
 "j8": "from 200:a2…",
 "j9": "الرد على الاتصال نفسه، يُسمح به تلقائيًا",
 "relay_sees": ["يرى الوسطاء المفاتيح في ترويسة التوجيه، والحجم، والتوقيت.", "ولا يستطيعون قراءة المحتوى أو تغييره أو إعادة إرساله."],

 "f8": ("من يرى ماذا", ["تشفّر عُقدة المحتوى من الطرف إلى الطرف، لكنها لا تُخفي أنك تستخدمها، ولا مع من تتحدث."]),
 "vh": ["", "عنوان IP", "عنوانك في عُقدة", "المحتوى", "الحجم والتوقيت"],
 "vr": [("مزوّد الإنترنت", "y", "n1", "n", "y"),
        ("جيرانك المباشرون", "y", "y", "n", "y"),
        ("العُقد الوسيطة على الطريق", "n", "y", "n", "y"),
        ("العُقدة التي تحادثها", "n2", "y", "y", "y"),
        ("أجهزة شبكتك المحلية (مع الاكتشاف)", "y", "y", "n", "n"),
        ("مطوّرو عُقدة", "n", "n", "n", "n")],
 "yes": "نعم", "no": "لا",
 "vfoot": ["¹  فقط على روابط TCP العادية؛ روابط TLS وQUIC تخفيه عن المزوّد.",
           "²  إلا إن كانت تلك العُقدة أيضًا من جيرانك المباشرين. لا يوجد تتبّع: لا يُرسل شيء إلى المطوّرين."],

 "f9": ("الشبكة العامة والمجموعات الخاصة", ["كلاهما يستخدم الشبكة نفسها والوسطاء أنفسهم، وسرّ المجموعة يحدّد من يستطيع فتح جلسة."]),
 "g_pub": ("GroupPassword فارغ: شبكة عامة", ["أي عُقدة تستطيع فتح جلسة مع عُقدتك،", "والجدار الناري يقرّر ما تصل إليه."]),
 "g_priv": ("GroupPassword محدّد: مجموعة خاصة", ["الأعضاء وحدهم يفتحون الجلسات، وتبقى العُقد", "الأخرى تمرّر نصّهم المشفّر."]),
 "member": "عضو", "outsider": "عُقدة أخرى", "relays": "تمرير فقط", "no_session": "لا جلسة",

 "f10": ("الجدار الناري للوارد", ["يُطبَّق على الاتصال القادم عبر عُقدة قبل أن يصل إلى برامجك. مفعّل للعُقد الجديدة."]),
 "fw_h": ["#", "المصدر", "البروتوكول", "المنافذ", "ملاحظة", "الإجراء"],
 "fw": [("-", "ردّ على اتصال فتحته العُقدة", "any", "any", "تتبّع الحالة", "سماح"),
        ("1", "200:a2… (الحاسوب)", "tcp", "22", "laptop SSH", "سماح"),
        ("2", "*", "icmp", "-", "AllowPing", "سماح"),
        ("-", "كل ما عدا ذلك", "", "", "يُحتسب", "رفض")],
 "fw_cfg_t": "uqda.conf",
 "fw_cfg": ["Firewall: {", "  Enabled: true", "  AllowPing: true", "  Rules: [", "    { Sources: [\"200:a2…\"],", "      Protocol: \"tcp\", Ports: [\"22\"],", "      Comment: \"laptop SSH\" }", "  ]", "}"],
 "fw_eval_t": "ترتيب التقييم",
 "fw_eval": ["١. جزء من اتصال فتحته العُقدة: سماح", "٢. أول قاعدة مطابقة: سماح", "٣. كل ما عدا ذلك: رفض واحتساب"],

 "f11": ("الانضمام إلى مجموعة خاصة بدعوة", ["الدعوة رمز لمرة واحدة دون كلمة مرور. لا يُرسل سرّ المجموعة إلا لجهاز يصل إلى",
        "الداعي عبر TLS 1.3 مثبّت على مفتاح الداعي، ويقدّم سرّ الرمز."]),
 "inv_life": [("الجهاز الجديد", "uqda join"), ("الداعي", "uqda invite")],
 "inv_pre": ["uqda setup: زوج مفاتيح، سرّ مجموعة", "(192 بت)، جدار ناري، جيران"],
 "inv_tok": ["الرمز: مفتاح الداعي، حتى ٤ عناوين،", "سرّ لمرة واحدة، انتهاء بعد ١٥ دقيقة"],
 "inv_oob": "يظهر الرمز نصًا ورمز QR، وينسخه المستخدم",
 "inv_tls": "TLS 1.3 إلى عنوان من الرمز؛ يجب أن يطابق مفتاح الشهادة مفتاح الرمز",
 "inv_sec": "السرّ لمرة واحدة",
 "inv_chk": ["السرّ صحيح وغير منتهٍ ولم", "يُستخدم من قبل: يُعلَّم مستخدمًا"],
 "inv_grant": "المنحة: سرّ المجموعة والجيران",
 "inv_done": ["يكتب uqda.conf (مع نسخة احتياطية)،", "وبعد إعادة التشغيل يصبح عضوًا"],

 "f12": ("البناء على عُقدة", ["كل ما يستمع على IPv6 يعمل دون تغيير، وعنوان المصدر في كل اتصال موثَّق."]),
 "b_left": ["استضف موقعًا أو خدمة من المنزل، دون إعادة توجيه منافذ.", "استخدم عنوان العميل 200:… لتسجيل دخوله.",
            "اقرأ حالة العُقدة عبر واجهة الإدارة المحلية (JSON).", "ضمّن عُقدة في التطبيقات بحزم Android وApple."],
 "b_code": ["$ python3 -m http.server 8080 --bind ::", "", "# uqda.conf", "Rules: [ { Sources: [\"*\"], Protocol: \"tcp\",", "          Ports: [\"8080\"] } ]", "", "# any Uqda user can now open", "http://[200:1a2b:3c4d::7]:8080/"],

 "limits_t": "الحدود",
 "limits": ["ليست شبكة إخفاء هوية: جيرانك يرون عنوان IP، والوسطاء يرون من يتحدث مع من.",
            "مشروع تجريبي: مُختبر على نطاق واسع، لكن دون تدقيق أمني مستقل بعد.",
            "قد تكون الطرق أطول من الاتصال المباشر، فبعضها أبطأ.",
            "لا تتبّع، ولا حسابات، ولا خوادم مركزية."],
 "foot": "Uqda Core 26.1  ·  برنامج حر، LGPLv3  ·  متوافقة مع شبكة Yggdrasil  ·  github.com/Uqda/Core",
},
}


# ---------------------------------------------------------------------------
# Figures

def header(p, t):
    p.text(L, 64, t["meta"], 13, GRAY, 700)
    p.text(L, 120, t["title"], 44, INK, 800)
    p.lines(L, 164, t["sub"], 18, TEXT, 28)
    y = 236
    p.rect(L, y, R - L, 96, FILL, None, r=0)
    for i, (k, v) in enumerate(t["specs"]):
        col, row = i % 2, i // 2
        x = L + 24 + col * 540
        p.text(x, y + 38 + row * 32, k, 14, INK, 700)
        p.text(x + 110, y + 38 + row * 32, v, 14, TEXT)
    return y + 96 + 44


def fig_topology(p, t, y):
    y = p.figure(y, 1, *t["f1"])
    top = y + 34
    p.rect(L, top + 50, 340, 250, "none", GRAY, 1.2, dash="6 5")
    p.text(L + 14, top + 74, t["home"], 13, GRAY, 600)
    p.rect(440, top + 30, 320, 300, "none", GRAY, 1.2, r=36, dash="6 5")
    p.text(600, top + 60, t["inet"], 13, GRAY, 600, anchor="middle")
    p.rect(840, top + 50, 300, 250, "none", GRAY, 1.2, dash="6 5")
    p.text(854, top + 74, t["remote"], 13, GRAY, 600)
    nodes = {
        "phone": (L + 18, top + 100, 145, t["n_phone"], "200:a1…"),
        "laptop": (L + 178, top + 100, 145, t["n_laptop"], "200:a2…"),
        "srv": (L + 98, top + 186, 145, t["n_srv"], "200:a3…"),
        "pa": (470, top + 110, 150, t["n_pa"], "tls://…:443"),
        "pb": (580, top + 220, 150, t["n_pb"], "quic://…:443"),
        "friend": (920, top + 165, 150, t["n_friend"], "200:f3…"),
    }

    def c(k, side):
        x, yy, w, *_ = nodes[k]
        return {"top": (x + w / 2, yy), "bot": (x + w / 2, yy + 46), "l": (x, yy + 23), "r": (x + w, yy + 23)}[side]
    bus = top + 262
    p.line(L + 30, bus, L + 310, bus, LINE, 3)
    for k in ("phone", "laptop", "srv"):
        bx, by = c(k, "bot")
        p.line(bx, by, bx, bus, LINE, 1.2)
    p.text(L + 170, bus + 24, t["multicast"], 12, GRAY, anchor="middle", italic=True)
    links = [(c("srv", "r"), c("pa", "l"), "TLS"), (c("pa", "bot"), c("pb", "top"), "TCP"), (c("pb", "r"), c("friend", "l"), "QUIC")]
    for (x1, y1), (x2, y2), lab in links:
        p.line(x1, y1, x2, y2, LINE, 2)
        lx2, ly2 = x1 + (x2 - x1) * .5, y1 + (y2 - y1) * .5
        p.rect(lx2 - 20, ly2 - 9, 40, 18, PAPER, None, r=0)
        p.code(lx2, ly2 + 4, lab, 12, GRAY, anchor="middle", weight=700)
    lx, ly = c("laptop", "top")
    fx, fy = c("friend", "top")
    p.path([(lx, ly), (lx, top + 8), (fx, top + 8), (fx, fy)], BLUE, 2.2, dash="7 5")
    p.head(fx, fy, 0, BLUE)
    p.text((lx + fx) / 2, top, t["session"], 13.5, BLUE, 600, anchor="middle")
    for k, (x, yy, w, name, addr) in nodes.items():
        p.node_box(x, yy, w, name, addr, accent=BLUE if k in ("laptop", "friend") else None)
    ly2 = top + 360
    p.line(L, ly2, L + 40, ly2, LINE, 2)
    p.text(L + 52, ly2 + 5, t["lg"][0], 13, TEXT)
    p.line(L + 400, ly2, L + 440, ly2, BLUE, 2.2, dash="7 5")
    p.text(L + 452, ly2 + 5, t["lg"][1], 13, TEXT)
    p.rect(L + 640, ly2 - 9, 40, 18, "none", GRAY, 1.2, dash="6 5")
    p.text(L + 692, ly2 + 5, t["lg"][2], 13, TEXT)
    return ly2 + 30


def fig_address(p, t, y):
    y = p.figure(y, 2, *t["f2"])
    top = y + 10
    p.rect(L, top, 250, 92, ORANGE_BG, ORANGE, 1.4)
    p.text(L + 16, top + 30, t["priv"][0], 15, ORANGE, 700)
    p.lines(L + 16, top + 54, t["priv"][1:], 13.5, INK, 21)
    p.harrow(L + 250, L + 320, top + 46, ORANGE, 1.6)
    p.rect(L + 320, top, 220, 92, PAPER, LINE, 1.4)
    p.text(L + 336, top + 30, t["pub"][0], 15, INK, 700)
    p.lines(L + 336, top + 54, t["pub"][1:], 13.5, TEXT, 21)
    p.harrow(L + 540, L + 700, top + 46, LINE, 1.6)
    p.text(L + 620, top + 30, t["derive"][0], 12.5, GRAY, anchor="middle")
    p.text(L + 620, top + 72, t["derive"][1], 12.5, GRAY, anchor="middle")
    p.rect(L + 700, top, R - L - 700, 92, BLUE_BG, BLUE, 1.4)
    p.text(L + 716, top + 30, t["example"], 12.5, GRAY, 600)
    ax = L + 716  # addresses aligned to the box's reading start
    if p.mirrored:
        p.code(ax, top + 58, t["ex_a"], 19, BLUE, anchor="end", weight=700)
        p.code(ax, top + 80, t["ex_s"], 14, INK, anchor="end", weight=600)
    else:
        p.code(ax, top + 58, t["ex_a"], 19, BLUE, weight=700)
        p.code(ax, top + 80, t["ex_s"], 14, INK, weight=600)

    p.ltr = True  # bit layouts always read left to right

    def layout(yy, title, fields, widths, colors):
        p.o.append(f'<text x="{L}" y="{yy}" font-size="14" font-weight="700" fill="{INK}">{escape(title)}</text>')
        x = L
        for (a, b, bits), w, col in zip(fields, widths, colors):
            p.rect(x, yy + 12, w, 54, col, LINE, 1.2, r=0)
            fam = f" font-family='{MONO}'" if a.startswith("0x") or a == "N" else ""
            p.o.append(f'<text x="{x + w / 2}" y="{yy + (36 if b else 44)}" font-size="14" font-weight="700" fill="{INK}" text-anchor="middle"{fam}>{escape(a)}</text>')
            if b:
                p.o.append(f'<text x="{x + w / 2}" y="{yy + 56}" font-size="12" fill="{GRAY}" text-anchor="middle">{escape(b)}</text>')
            p.o.append(f'<text x="{x + 4}" y="{yy + 82}" font-size="11" fill="{GRAY}" font-family=\'{MONO}\'>{escape(bits)}</text>')
            x += w
    layout(top + 136, t["addr_row"], t["fields_a"], [170, 170, 740], [ORANGE_BG, FILL, PAPER])
    layout(top + 246, t["sub_row"], t["fields_s"], [170, 170, 370, 370], [ORANGE_BG, FILL, PAPER, BLUE_BG])
    p.ltr = False
    return top + 336


def fig_stack(p, t, y):
    y = p.figure(y, 3, *t["f3"])
    top = y + 10
    sx, lw, lh = L + 44, 440, 46
    for i, (name, desc, uq) in enumerate(t["stack"]):
        yy = top + i * lh
        p.rect(sx, yy, lw, lh, BLUE_BG if uq else FILL, LINE, 1.1, r=0)
        p.text(sx + 16, yy + 29, name, 15, INK, 700)
        p.text(sx + lw - 14, yy + 29, desc, 13, GRAY, anchor="end")
    b1, b2 = top + 2 * lh, top + 7 * lh
    p.path([(sx - 8, b1 + 3), (sx - 20, b1 + 3), (sx - 20, b2 - 3), (sx - 8, b2 - 3)], BLUE, 2)
    cx, cy = p.X(sx - 32), (b1 + b2) / 2
    p.o.append(f'<text x="{cx}" y="{cy}" font-size="13" font-weight="700" fill="{BLUE}" text-anchor="middle" transform="rotate(-90 {cx} {cy})">{escape(t["brace"])}</text>')

    p.ltr = True  # packet formats always read left to right
    ex = 600 if not p.rtl else L
    total = 540
    p.o.append(f'<text x="{ex}" y="{top + 14}" font-size="15" font-weight="700" fill="{INK}">{escape(t["enc_title"])}</text>')
    yy = top + 32
    for k, (label, parts) in enumerate(t["enc"]):
        unit = total / sum(w for _, w, _ in parts)
        x = ex
        for name, w, col in parts:
            ww = w * unit
            p.rect(x, yy, ww, 40, col, LINE, 1.1, r=0)
            p.o.append(f'<text x="{x + ww / 2}" y="{yy + 25}" font-size="12.5" fill="{INK}" text-anchor="middle" font-family=\'{MONO}\'>{escape(name)}</text>')
            x += ww
        if k < len(t["enc"]) - 1:
            p.line(ex + 26, yy + 40, ex + 26, yy + 70, GRAY, 1.2)
            p.head(ex + 26, yy + 70, 0, GRAY)
            p.o.append(f'<text x="{ex + 40}" y="{yy + 60}" font-size="12.5" fill="{GRAY}" font-family=\'{MONO}\'>{escape(t["enc"][k + 1][0])}</text>')
        yy += 76
    p.ltr = False
    return top + 8 * lh + 20


def fig_handshake(p, t, y):
    y = p.figure(y, 4, *t["f4"])
    top = y + 10
    ax, bx = 250, 950
    end = top + 380
    p.lifeline(ax, top, end, t["lk_life"][0])
    p.lifeline(bx, top, end, t["lk_life"][3])
    yy = top + 84
    for i, (lab, d, col) in enumerate(t["hs"]):
        if d == "both":
            p.msg(ax, bx, yy, lab, col, num=i + 1, both=True)
        elif d == "ab":
            p.msg(ax, bx, yy, lab, col, num=i + 1)
        else:
            p.msg(bx, ax, yy, lab, col, num=i + 1)
        yy += 54
    p.rect(ax - 40, yy - 24, bx - ax + 80, 36, GREEN_BG, GREEN, 1.2)
    p.text(ax - 26, yy - 1, t["hs_note"], 13, INK)
    yy += 64
    p.msg(ax, bx, yy, t["hs_after"], GRAY, dash="5 4", both=True, num=5)
    p.text(L, end + 34, t["hs_foot"], 13, GRAY, italic=True)
    return end + 46


def fig_tree(p, t, y):
    y = p.figure(y, 5, *t["f5"])
    top = y + 24
    N = {"r": (330, top + 20, "[ ]"), "1": (190, top + 110, "[1]"), "2": (470, top + 110, "[2]"),
         "11": (120, top + 200, "[1,1]"), "12": (260, top + 200, "[1,2]"), "21": (400, top + 200, "[2,1]"), "22": (540, top + 200, "[2,2]"),
         "111": (120, top + 290, "[1,1,1]"), "121": (260, top + 290, "[1,2,1]"), "211": (400, top + 290, "[2,1,1]"), "221": (540, top + 290, "[2,2,1]")}
    tree = [("r", "1"), ("r", "2"), ("1", "11"), ("1", "12"), ("2", "21"), ("2", "22"), ("11", "111"), ("12", "121"), ("21", "211"), ("22", "221")]
    for a, b in tree:
        p.line(N[a][0], N[a][1] + 15, N[b][0], N[b][1] - 15, LINE, 1.3)
    # the extra peer link between [1,2] and [2,1], drawn as the greedy hop
    x1, y1 = N["12"][0] + 40, N["12"][1]
    x2 = N["21"][0] - 40
    p.harrow(x1, x2, y1, BLUE, 2.4)
    p.text((x1 + x2) / 2, y1 - 9, t["shortcut"], 11.5, BLUE, 600, anchor="middle")
    # greedy hops up and down
    p.line(N["121"][0] + 18, N["121"][1] - 15, N["12"][0] + 18, N["12"][1] + 15, BLUE, 2.4)
    p.head(N["12"][0] + 18, N["12"][1] + 15, 2, BLUE)
    p.line(N["21"][0] + 18, N["21"][1] + 15, N["211"][0] + 18, N["211"][1] - 15, BLUE, 2.4)
    p.head(N["211"][0] + 18, N["211"][1] - 15, 0, BLUE)
    for k, (x, yy, lab) in N.items():
        on = k in ("121", "12", "21", "211")
        fill = ORANGE_BG if k == "r" else (BLUE_BG if k in ("121", "211") else PAPER)
        stroke = ORANGE if k == "r" else (BLUE if on else LINE)
        p.rect(x - 40, yy - 15, 80, 30, fill, stroke, 1.6 if on or k == "r" else 1.2, r=3)
        p.code(x, yy + 5, lab, 13, INK, anchor="middle", weight=600)
    p.text(N["r"][0] + 52, N["r"][1] + 5, t["root"], 13, ORANGE, 700)
    p.text(N["121"][0], N["121"][1] + 36, t["src"], 12.5, BLUE, 700, anchor="middle")
    p.text(N["211"][0], N["211"][1] + 36, t["dst"], 12.5, BLUE, 700, anchor="middle")
    p.code(L, top + 372, t["dist"], 13, TEXT, anchor="end" if p.mirrored else "start")
    tx, tw = 680, R - 680
    cols = [70, 120, 150, 120]
    p.rect(tx, top - 10, tw, 38, FILL, None, r=0)
    x = tx
    for h, w in zip(t["tbl_h"], cols):
        p.text(x + 12, top + 15, h, 13, INK, 700)
        x += w
    for i, row in enumerate(t["tbl"]):
        yy = top + 28 + i * 42
        p.line(tx, yy, tx + tw, yy, RULE, 1)
        x = tx
        for j, (v, w) in enumerate(zip(row, cols)):
            if j in (1, 2):
                p.code(x + 12, yy + 27, v, 13.5, BLUE if j == 2 else INK, anchor="end" if p.mirrored else "start", weight=600 if j == 2 else 400)
            else:
                p.text(x + 12, yy + 27, p.n(v), 13.5, INK)
            x += w
    p.line(tx, top + 154, tx + tw, top + 154, LINE, 1.2)
    p.lines(tx, top + 196, t["tbl_note"], 14, TEXT, 22)
    return top + 392


def fig_lookup(p, t, y):
    y = p.figure(y, 6, *t["f6"])
    top = y + 10
    xs = [190, 460, 730, 1000]
    end = top + 380
    for x, lab in zip(xs, t["lk_life"]):
        p.lifeline(x, top, end, lab)
    yy = top + 92
    for a, b in zip(xs, xs[1:]):
        p.harrow(a, b, yy, BLUE, 1.6)
    p.code((xs[0] + xs[1]) / 2, yy - 9, p.n(1) + ". " + t["lk1"], 12.5, BLUE, anchor="middle")
    p.text((xs[1] + xs[3]) / 2, yy + 22, t["lk_bloom"], 12, GRAY, anchor="middle", italic=True)
    yy += 72
    for a, b in zip(xs[::-1], xs[::-1][1:]):
        p.harrow(a, b, yy, BLUE, 1.6, dash="6 4")
    p.text((xs[2] + xs[3]) / 2, yy - 9, p.n(2) + ". " + t["lk2"], 13, BLUE, 500, anchor="middle")
    yy += 64
    for a, b in zip(xs, xs[1:]):
        p.harrow(a, b, yy, INK, 1.6)
    p.text((xs[0] + xs[1]) / 2, yy - 9, p.n(3) + ". " + t["lk3"], 13, INK, 500, anchor="middle")
    yy += 64
    mid = (xs[2] + xs[3]) / 2
    p.line(xs[2], yy, mid - 14, yy, GRAY, 1.6)
    p.line(mid - 10, yy - 10, mid + 10, yy + 10, RED, 2.4)
    p.line(mid + 10, yy - 10, mid - 10, yy + 10, RED, 2.4)
    yy += 44
    p.harrow(xs[2], xs[0], yy, RED, 1.6, dash="6 4")
    p.text((xs[0] + xs[2]) / 2, yy - 9, p.n(4) + ". " + t["lk4"], 13, RED, 500, anchor="middle")
    yy += 28
    p.rect(xs[0] - 75, yy, 520, 34, BLUE_BG, BLUE, 1.2)
    p.text(xs[0] - 61, yy + 22, p.n(5) + ". " + t["lk5"], 13, INK)
    return end + 20


def fig_journey(p, t, y):
    y = p.figure(y, 7, *t["f7"])
    top = y + 10
    xs = [140, 324, 508, 692, 876, 1060]
    end = top + 640
    for x, (lab, sub) in zip(xs, t["j_life"]):
        p.lifeline(x, top, end, lab, sub)
    p.rect(xs[2] - 75, top + 66, 350, 56, FILL, RULE, 1)
    p.lines(xs[2] - 63, top + 88, t["relay_sees"], 12, TEXT, 19)
    yy = top + 160
    p.msg(xs[0], xs[1], yy, t["j1"], INK, num=1, mono=True)
    yy += 20
    p.note(xs[1] - 75, yy, 300, [p.n(2) + ". " + t["j2"][0]] + t["j2"][1:], ORANGE, ORANGE_BG)
    yy += 112
    p.msg(xs[1], xs[2], yy, t["j3"], ORANGE, num=3)
    yy += 48
    p.msg(xs[2], xs[3], yy, t["j4"], ORANGE, num=4)
    yy += 48
    p.msg(xs[3], xs[4], yy, t["j5"], ORANGE, num=5)
    yy += 20
    p.note(xs[4] - 125, yy, 250, [p.n(6) + ". " + t["j6"][0]] + t["j6"][1:], ORANGE, ORANGE_BG)
    yy += 66
    p.note(xs[4] - 125, yy, 250, [p.n(7) + ". " + t["j7"][0]] + t["j7"][1:], GREEN, GREEN_BG)
    yy += 92
    p.msg(xs[4], xs[5], yy, t["j8"], INK, num=8, mono=True)
    yy += 52
    p.msg(xs[5], xs[0], yy, t["j9"], GRAY, dash="6 4", num=9)
    return end + 20


def fig_visibility(p, t, y):
    y = p.figure(y, 8, *t["f8"])
    top = y + 10
    cols = [380, 175, 175, 175, 175]
    p.rect(L, top, R - L, 44, FILL, None, r=0)
    x = L
    for i, (h, w) in enumerate(zip(t["vh"], cols)):
        if i:
            p.text(x + w / 2, top + 28, h, 14, INK, 700, anchor="middle")
        x += w
    for i, row in enumerate(t["vr"]):
        yy = top + 44 + i * 46
        p.line(L, yy, R, yy, RULE, 1)
        p.text(L + 16, yy + 29, row[0], 14.5, INK, 500)
        x = L + cols[0]
        for v, w in zip(row[1:], cols[1:]):
            yes = v.startswith("y")
            sup = "¹" if v.endswith("1") else ("²" if v.endswith("2") else "")
            p.text(x + w / 2, yy + 29, (t["yes"] if yes else t["no"]) + sup, 14, RED if yes else GREEN, 700 if yes else 500, anchor="middle")
            x += w
    yy = top + 44 + len(t["vr"]) * 46
    p.line(L, yy, R, yy, LINE, 1.2)
    p.lines(L, yy + 30, t["vfoot"], 13, GRAY, 21)
    return yy + 66


def fig_groups(p, t, y):
    y = p.figure(y, 9, *t["f9"])
    top = y + 10
    pts = [(70, 60), (190, 30), (300, 80), (420, 40), (130, 150), (260, 170), (390, 150)]
    links = [(0, 1), (1, 2), (2, 3), (0, 4), (4, 5), (5, 6), (6, 3), (2, 5)]
    for side in (0, 1):
        ox = L + side * 560
        title, lines = t["g_pub"] if side == 0 else t["g_priv"]
        p.rect(ox, top, 520, 290, "none", RULE, 1.2, r=0)
        p.text(ox + 20, top + 32, title, 15, INK, 700)
        gy = top + 56
        for a, b in links:
            p.line(ox + pts[a][0], gy + pts[a][1], ox + pts[b][0], gy + pts[b][1], LINE, 1.3)
        members = {0, 3, 4, 5} if side else set()
        if side == 0:
            p.path([(ox + pts[0][0], gy + pts[0][1] - 16), (ox + pts[1][0], gy + pts[1][1] - 16), (ox + pts[2][0], gy + pts[2][1] - 16), (ox + pts[3][0], gy + pts[3][1] - 16)], BLUE, 2, dash="6 4")
            p.path([(ox + pts[4][0], gy + pts[4][1] + 16), (ox + pts[5][0], gy + pts[5][1] + 16), (ox + pts[6][0], gy + pts[6][1] + 16)], BLUE, 2, dash="6 4")
        else:
            p.path([(ox + pts[0][0], gy + pts[0][1] - 16), (ox + pts[1][0], gy + pts[1][1] - 16), (ox + pts[2][0], gy + pts[2][1] - 16), (ox + pts[3][0], gy + pts[3][1] - 16)], BLUE, 2, dash="6 4")
            p.text(ox + pts[1][0], gy + pts[1][1] + 30, t["relays"], 11.5, GRAY, 600, anchor="middle", italic=True)
            sx, sy = ox + pts[6][0], gy + pts[6][1]
            tx, ty = ox + pts[5][0], gy + pts[5][1]
            p.line(sx - 14, sy + 20, tx + 14, ty + 20, RED, 2, dash="4 3")
            mx = (sx + tx) / 2
            p.line(mx - 7, sy + 13, mx + 7, sy + 27, RED, 2.2)
            p.line(mx + 7, sy + 13, mx - 7, sy + 27, RED, 2.2)
            p.text(mx, sy + 46, t["no_session"], 11.5, RED, 600, anchor="middle")
        for k, (a, b) in enumerate(pts):
            m = k in members
            p.circle(ox + a, gy + b, 11, ORANGE_BG if m else PAPER, ORANGE if m else LINE, 2 if m else 1.3)
        if side == 1:
            p.circle(ox + 30, top + 266, 7, ORANGE_BG, ORANGE, 2)
            p.text(ox + 44, top + 271, t["member"], 12, TEXT)
            p.circle(ox + 140, top + 266, 7, PAPER, LINE, 1.3)
            p.text(ox + 154, top + 271, t["outsider"], 12, TEXT)
        p.lines(ox + 20, top + 326, lines, 14, TEXT, 22)
    return top + 380


def fig_firewall(p, t, y):
    y = p.figure(y, 10, *t["f10"])
    top = y + 10
    cols = [44, 270, 96, 70, 130, 80]
    tw = sum(cols)
    p.rect(L, top, tw, 40, FILL, None, r=0)
    x = L
    for h, w in zip(t["fw_h"], cols):
        p.text(x + 12, top + 26, h, 13, INK, 700)
        x += w
    for i, row in enumerate(t["fw"]):
        yy = top + 40 + i * 44
        drop = i == len(t["fw"]) - 1
        if drop:
            p.rect(L, yy, tw, 44, RED_BG, None, r=0)
        p.line(L, yy, L + tw, yy, RULE, 1)
        x = L
        for j, (v, w) in enumerate(zip(row, cols)):
            if j == 5:
                p.text(x + 12, yy + 28, v, 13.5, RED if drop else GREEN, 700)
            elif j in (2, 3) or (j == 1 and v[:1] in "2*"):
                p.code(x + 12, yy + 28, v, 13, INK, anchor="end" if p.mirrored else "start")
            else:
                p.text(x + 12, yy + 28, p.n(v) if j == 0 else v, 13.5, INK)
            x += w
    yy = top + 40 + len(t["fw"]) * 44
    p.line(L, yy, L + tw, yy, LINE, 1.2)
    p.text(L, yy + 36, t["fw_eval_t"], 14, INK, 700)
    p.lines(L, yy + 62, t["fw_eval"], 14, TEXT, 22)
    cw = R - L - tw - 30
    p.ltr = True
    bx = R - cw if not p.rtl else L
    p.rect(bx, top, cw, 262, CODE_BG, None, r=4)
    p.o.append(f'<text x="{bx + 16}" y="{top + 26}" font-size="12" fill="#94a3b8" font-family=\'{MONO}\'>{escape(t["fw_cfg_t"])}</text>')
    for k, s in enumerate(t["fw_cfg"]):
        p.o.append(f'<text x="{bx + 16}" y="{top + 56 + k * 22}" font-size="13" fill="#e2e8f0" font-family=\'{MONO}\' xml:space="preserve">{escape(s)}</text>')
    p.ltr = False
    return max(yy + 140, top + 290)


def fig_invite(p, t, y):
    y = p.figure(y, 11, *t["f11"])
    top = y + 10
    nx, ix = 240, 900
    end = top + 560
    p.lifeline(nx, top, end, *t["inv_life"][0])
    p.lifeline(ix, top, end, *t["inv_life"][1])
    yy = top + 70
    p.note(ix - 150, yy, 300, t["inv_pre"], LINE, FILL)
    yy += 62
    p.note(ix - 150, yy, 300, t["inv_tok"], ORANGE, ORANGE_BG)
    yy += 100
    p.msg(ix, nx, yy, t["inv_oob"], GRAY, dash="3 4", num=1)
    yy += 54
    p.msg(nx, ix, yy, t["inv_tls"], INK, num=2)
    yy += 52
    p.msg(nx, ix, yy, t["inv_sec"], ORANGE, num=3)
    yy += 20
    p.note(ix - 150, yy, 300, t["inv_chk"], GREEN, GREEN_BG)
    yy += 88
    p.msg(ix, nx, yy, t["inv_grant"], GREEN, num=4)
    yy += 20
    p.note(nx - 150, yy, 300, t["inv_done"], BLUE, BLUE_BG)
    return end + 20


def fig_build(p, t, y):
    y = p.figure(y, 12, *t["f12"])
    top = y + 10
    for i, s in enumerate(t["b_left"]):
        yy = top + 26 + i * 34
        p.rect(L, yy - 10, 6, 6, INK, None, r=0)
        p.text(L + 18, yy, s, 15, TEXT)
    p.ltr = True
    bx = 640 if not p.rtl else L
    p.rect(bx, top, 500, 218, CODE_BG, None, r=4)
    for k, s in enumerate(t["b_code"]):
        col = "#94a3b8" if s.startswith("#") else ("#93c5fd" if s.startswith("http") else "#e2e8f0")
        p.o.append(f'<text x="{bx + 18}" y="{top + 32 + k * 24}" font-size="13" fill="{col}" font-family=\'{MONO}\' xml:space="preserve">{escape(s)}</text>')
    p.ltr = False
    return top + 240


def footer(p, t, y):
    p.line(L, y, R, y, RULE, 1)
    p.text(L, y + 42, t["limits_t"], 20, INK, 700)
    yy = y + 76
    for i, s in enumerate(t["limits"]):
        p.rect(L, yy + i * 28 - 10, 6, 6, INK, None, r=0)
        p.text(L + 18, yy + i * 28, s, 15, TEXT)
    yy += len(t["limits"]) * 28
    p.line(L, yy + 6, R, yy + 6, RULE, 1)
    p.text(W / 2, yy + 40, t["foot"], 13, GRAY, anchor="middle")
    return yy + 70


def build(lang):
    t = T[lang]
    p = Page(lang)
    y = header(p, t)
    for f in (fig_topology, fig_address, fig_stack, fig_handshake, fig_tree, fig_lookup, fig_journey,
              fig_visibility, fig_groups, fig_firewall, fig_invite, fig_build):
        y = f(p, t, y) + 36
    y = footer(p, t, y)
    h = int(y)
    head = (f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {W} {h}" width="{W}" height="{h}" '
            f"font-family='{SANS}' role=\"img\" aria-labelledby=\"t d\">"
            f'<title id="t">{escape(t["title"])}</title><desc id="d">{escape(" ".join(t["sub"]))}</desc>'
            f'<rect width="{W}" height="{h}" fill="{PAPER}"/>')
    return head + "\n" + "\n".join(p.o) + "\n</svg>\n"


if __name__ == "__main__":
    import os
    here = os.path.dirname(os.path.abspath(__file__))
    for lang in ("en", "ar"):
        with open(os.path.join(here, f"how-uqda-works.{lang}.svg"), "w", encoding="utf-8", newline="\n") as f:
            f.write(build(lang))
    print("written")
