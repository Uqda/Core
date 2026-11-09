#!/usr/bin/env python3
"""Generates the poster docs/assets/how-uqda-works.{en,ar}.svg.

One tall image that explains the whole Uqda network from top to bottom.
Run from the repository root: python3 docs/assets/diagram.py
The Arabic poster mirrors the layout for right-to-left reading.
"""
import math
import random
from xml.sax.saxutils import escape

W = 1200
FONT = '"Segoe UI",system-ui,-apple-system,"Noto Sans Arabic","Noto Sans",Tahoma,sans-serif'
MONO = 'ui-monospace,SFMono-Regular,Consolas,"Liberation Mono",monospace'

# Palette (fixed dark poster, readable on light and dark pages)
BG1, BG2 = "#0a0f1f", "#101a35"
CARD, CARDLINE = "#121b34", "#243259"
INK, MUTED, DIM = "#eef2fb", "#a3aecb", "#6b7799"
TEAL, AMBER, VIOLET, ROSE, GREEN = "#2dd4bf", "#f5b041", "#8b7cf6", "#f87171", "#34d399"

DIGITS_AR = "٠١٢٣٤٥٦٧٨٩"


class Poster:
    def __init__(self, lang):
        self.lang = lang
        self.rtl = lang == "ar"
        self.o = []

    # -- coordinates
    def X(self, x, w=0):
        return round(W - x - w, 1) if self.rtl else x

    def num(self, n):
        s = str(n)
        return "".join(DIGITS_AR[int(c)] for c in s) if self.rtl else s

    # -- primitives
    def raw(self, s):
        self.o.append(s)

    def rect(self, x, y, w, h, fill, stroke=None, r=12, sw=1, op=1, dash=None):
        st = f' stroke="{stroke}" stroke-width="{sw}"' if stroke else ""
        d = f' stroke-dasharray="{dash}"' if dash else ""
        o = f' opacity="{op}"' if op != 1 else ""
        self.o.append(f'<rect x="{self.X(x, w)}" y="{y}" width="{w}" height="{h}" rx="{r}" fill="{fill}"{st}{d}{o}/>')

    def circle(self, x, y, r, fill="none", stroke=None, sw=2, op=1, filt=None, dash=None):
        st = f' stroke="{stroke}" stroke-width="{sw}"' if stroke else ""
        f = f' filter="url(#{filt})"' if filt else ""
        d = f' stroke-dasharray="{dash}"' if dash else ""
        o = f' opacity="{op}"' if op != 1 else ""
        self.o.append(f'<circle cx="{self.X(x)}" cy="{y}" r="{r}" fill="{fill}"{st}{f}{d}{o}/>')

    def ellipse(self, x, y, rx, ry, stroke, sw=2):
        self.o.append(f'<ellipse cx="{self.X(x)}" cy="{y}" rx="{rx}" ry="{ry}" fill="none" stroke="{stroke}" stroke-width="{sw}"/>')

    def line(self, x1, y1, x2, y2, stroke, sw=2, dash=None, op=1, cap="round"):
        d = f' stroke-dasharray="{dash}"' if dash else ""
        o = f' opacity="{op}"' if op != 1 else ""
        self.o.append(f'<line x1="{self.X(x1)}" y1="{y1}" x2="{self.X(x2)}" y2="{y2}" stroke="{stroke}" stroke-width="{sw}" stroke-linecap="{cap}"{d}{o}/>')

    def poly(self, pts, fill="none", stroke=None, sw=2, closed=True, op=1, dash=None):
        p = " ".join(f"{self.X(a)},{b}" for a, b in pts)
        st = f' stroke="{stroke}" stroke-width="{sw}" stroke-linejoin="round" stroke-linecap="round"' if stroke else ""
        d = f' stroke-dasharray="{dash}"' if dash else ""
        o = f' opacity="{op}"' if op != 1 else ""
        tag = "polygon" if closed else "polyline"
        self.o.append(f'<{tag} points="{p}" fill="{fill}"{st}{d}{o}/>')

    def arrow(self, x1, y1, x2, y2, color=MUTED, sw=2.5, dash=None):
        self.line(x1, y1, x2, y2, color, sw, dash)
        X1, X2 = self.X(x1), self.X(x2)
        a = math.atan2(y2 - y1, X2 - X1)
        h = 11
        pts = [(X2, y2), (X2 - h * math.cos(a - .42), y2 - h * math.sin(a - .42)),
               (X2 - h * math.cos(a + .42), y2 - h * math.sin(a + .42))]
        self.o.append('<polygon points="' + " ".join(f"{p:.1f},{q:.1f}" for p, q in pts) + f'" fill="{color}"/>')

    def text(self, x, y, s, size=16, fill=INK, weight=400, anchor="start", mono=False, ls=None):
        xx = self.X(x)
        a = anchor
        attrs = ""
        if self.rtl:
            if mono:
                a = {"start": "end", "end": "start"}.get(anchor, anchor)
            else:
                attrs = ' direction="rtl"'
        fam = f' font-family=\'{MONO}\'' if mono else ""
        l = f' letter-spacing="{ls}"' if ls else ""
        self.o.append(f'<text x="{xx}" y="{y}" font-size="{size}" font-weight="{weight}" fill="{fill}" text-anchor="{a}"{attrs}{fam}{l}>{escape(s)}</text>')

    def para(self, x, y, lines, size=16, fill=MUTED, lh=None, weight=400, mono=False):
        lh = lh or round(size * 1.45)
        for i, l in enumerate(lines):
            if l:
                self.text(x, y + i * lh, l, size, fill, weight, mono=mono)
        return y + len(lines) * lh

    # -- composites
    def node(self, x, y, r=12, color=TEAL, ring=None, glow=True):
        if glow:
            self.circle(x, y, r * 2.2, fill=color, op=.18, filt="blur")
        self.circle(x, y, r, fill="url(#nodeFill)", stroke=color, sw=2.5)
        self.circle(x, y, r * .38, fill=color)
        if ring:
            self.circle(x, y, r + 6, stroke=ring, sw=2.5)

    def badge(self, x, y, n, color=TEAL, r=24):
        self.circle(x, y, r + 8, fill=color, op=.15)
        self.circle(x, y, r, fill=color)
        self.text(x, y + 9, self.num(n), 24, BG1, 800, anchor="middle")

    def chip(self, x, y, label, color=TEAL, mono=False, size=15, pad=16, w=None):
        w = w or int(len(label) * size * (.62 if mono else .56) + 2 * pad)
        self.rect(x, y, w, 36, color, r=18, op=.14)
        self.rect(x, y, w, 36, "none", stroke=color, r=18, sw=1.2, op=.6)
        self.text(x + w / 2, y + 24, label, size, color, 600, anchor="middle", mono=mono)
        return w

    def card(self, x, y, w, h, stroke=CARDLINE, fill="url(#cardFill)", r=18):
        self.rect(x, y, w, h, fill, stroke=stroke, r=r, sw=1.2)

    def section(self, y, h, n, title, sub):
        self.rect(40, y, 1120, h, "url(#sectionFill)", stroke=CARDLINE, r=28, sw=1.2)
        self.rect(40, y, 6, h, TEAL, r=3, op=.0)
        self.badge(100, y + 62, n)
        self.text(146, y + 72, title, 32, INK, 700)
        self.para(146, y + 106, sub, 18, MUTED, 26)

    # -- icons (all drawn from mirrored primitives)
    def icon_phone(self, x, y, c):
        self.rect(x - 18, y - 32, 36, 64, "none", stroke=c, r=7, sw=2.5)
        self.line(x - 6, y + 23, x + 6, y + 23, c, 2.5)

    def icon_laptop(self, x, y, c):
        self.rect(x - 32, y - 28, 64, 42, "none", stroke=c, r=5, sw=2.5)
        self.poly([(x - 42, y + 24), (x + 42, y + 24), (x + 34, y + 14), (x - 34, y + 14)], stroke=c, sw=2.5)

    def icon_desktop(self, x, y, c):
        self.rect(x - 34, y - 32, 68, 46, "none", stroke=c, r=5, sw=2.5)
        self.line(x, y + 14, x, y + 26, c, 2.5)
        self.line(x - 16, y + 28, x + 16, y + 28, c, 2.5)

    def icon_server(self, x, y, c):
        for i in range(3):
            yy = y - 30 + i * 21
            self.rect(x - 30, yy, 60, 17, "none", stroke=c, r=4, sw=2.5)
            self.circle(x + 20, yy + 8.5, 2.5, fill=c)

    def icon_box(self, x, y, c):
        self.rect(x - 32, y - 12, 64, 36, "none", stroke=c, r=8, sw=2.5)
        self.line(x - 18, y - 12, x - 26, y - 32, c, 2.5)
        self.line(x + 18, y - 12, x + 26, y - 32, c, 2.5)
        for i in range(3):
            self.circle(x - 14 + i * 14, y + 6, 3, fill=c)

    def icon_lock(self, x, y, c):
        self.rect(x - 16, y - 4, 32, 26, c, r=5, op=.9)
        self.rect(x - 10, y - 22, 20, 26, "none", stroke=c, r=10, sw=3.5)
        self.circle(x, y + 8, 3.5, fill=BG1)

    def icon_key(self, x, y, c):
        self.circle(x - 12, y, 10, stroke=c, sw=3)
        self.line(x - 2, y, x + 22, y, c, 3)
        self.line(x + 14, y, x + 14, y + 8, c, 3)
        self.line(x + 21, y, x + 21, y + 7, c, 3)

    def icon_globe(self, x, y, c):
        self.circle(x, y, 22, stroke=c, sw=2.5)
        self.ellipse(x, y, 10, 22, c, 2)
        self.line(x - 22, y, x + 22, y, c, 2)

    def icon_chat(self, x, y, c):
        self.rect(x - 24, y - 18, 48, 32, "none", stroke=c, r=8, sw=2.5)
        self.poly([(x - 10, y + 14), (x - 16, y + 24), (x + 2, y + 14)], stroke=c, sw=2.5)

    def icon_folder(self, x, y, c):
        self.poly([(x - 24, y - 16), (x - 8, y - 16), (x - 2, y - 10), (x + 24, y - 10), (x + 24, y + 18), (x - 24, y + 18)], stroke=c, sw=2.5)

    def icon_house(self, x, y, c):
        self.poly([(x - 24, y - 2), (x, y - 24), (x + 24, y - 2)], stroke=c, sw=2.5, closed=False)
        self.rect(x - 18, y - 4, 36, 26, "none", stroke=c, r=3, sw=2.5)
        self.rect(x - 5, y + 8, 10, 14, c, r=2)

    def icon_game(self, x, y, c):
        self.rect(x - 28, y - 14, 56, 30, "none", stroke=c, r=14, sw=2.5)
        self.line(x - 16, y + 1, x - 6, y + 1, c, 2.5)
        self.line(x - 11, y - 4, x - 11, y + 6, c, 2.5)
        self.circle(x + 10, y - 2, 3, fill=c)
        self.circle(x + 17, y + 5, 3, fill=c)

    def icon_envelope(self, x, y, c, w=46, h=32):
        self.rect(x - w / 2, y - h / 2, w, h, c, r=4, op=.95)
        self.poly([(x - w / 2, y - h / 2), (x, y + 2), (x + w / 2, y - h / 2)], stroke=BG1, sw=2.5, closed=False)


def defs():
    return f"""<defs>
<linearGradient id="bg" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="{BG1}"/><stop offset=".5" stop-color="{BG2}"/><stop offset="1" stop-color="{BG1}"/></linearGradient>
<linearGradient id="sectionFill" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#141e3b"/><stop offset="1" stop-color="#0f1730"/></linearGradient>
<linearGradient id="cardFill" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#1a2547"/><stop offset="1" stop-color="#151f3c"/></linearGradient>
<linearGradient id="tealFill" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#134e4a"/><stop offset="1" stop-color="#0f2f3a"/></linearGradient>
<linearGradient id="amberFill" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#4a3312"/><stop offset="1" stop-color="#2b2214"/></linearGradient>
<linearGradient id="violetFill" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#2e2860"/><stop offset="1" stop-color="#1c1b40"/></linearGradient>
<linearGradient id="roseFill" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#4a1d24"/><stop offset="1" stop-color="#2a1720"/></linearGradient>
<linearGradient id="titleGrad" x1="0" y1="0" x2="1" y2="0"><stop offset="0" stop-color="{TEAL}"/><stop offset="1" stop-color="{VIOLET}"/></linearGradient>
<radialGradient id="nodeFill"><stop offset="0" stop-color="#1f2c52"/><stop offset="1" stop-color="#0f1730"/></radialGradient>
<radialGradient id="heroGlow" cx=".5" cy=".5" r=".5"><stop offset="0" stop-color="{TEAL}" stop-opacity=".22"/><stop offset="1" stop-color="{TEAL}" stop-opacity="0"/></radialGradient>
<filter id="blur" x="-1" y="-1" width="3" height="3"><feGaussianBlur stdDeviation="6"/></filter>
</defs>"""


# --------------------------------------------------------------------------
# Content

C = {
"en": {
 "kicker": "UQDA  ·  عُقدة  ·  A NETWORK MADE OF NODES",
 "title": "How the Uqda network works",
 "sub": ["Every device that installs Uqda becomes a node. The nodes connect to each",
         "other and carry each other's traffic, encrypted. Together, they are the network."],
 "chips": ["No company", "No account", "No central server", "Free software"],

 "s1": ("Every device becomes a node", ["Uqda means node. Install it on anything, and that device joins the network as an equal."]),
 "devices": ["Phone", "Laptop", "Desktop", "Server", "Home server"],
 "roles": [("Gets an identity", ["A key pair made on the device.", "Nobody issues it, nobody can", "take it away."]),
           ("Connects to peers", ["A few nearby nodes: on your", "local network or across", "the internet."]),
           ("Becomes a router", ["It forwards encrypted traffic", "for others when it lies on the", "best path between them."])],

 "s2": ("An identity nobody can take", ["Your address is calculated from your key. Without the key, nobody can use it."]),
 "id_priv": ("Private key", ["Ed25519, stays on", "the device"]),
 "id_pub": ("Public key", ["shared with", "other nodes"]),
 "id_hash": "hash",
 "id_addr": ("Your address", "200:1a2b:3c4d::7"),
 "id_sub": ("Your subnet", "300:1a2b:3c4d::/64"),
 "id_sub_note": "for devices behind the node",
 "id_note": ["A connection that arrives from 200:1a2b:3c4d::7 really comes from the holder of that key.",
             "Programs can use the address as a login, with no passwords."],

 "s3": ("Inside one node", ["Your programs see ordinary IPv6. Uqda does the rest, layer by layer."]),
 "layers": [("Your apps", "Browser, SSH, games, chat, your own service"),
            ("uqda0 interface", "Plain IPv6: 200::/7 addresses, nothing to learn"),
            ("Inbound firewall", "Replies pass; new connections only if a rule allows them"),
            ("Encrypted sessions", "End to end, forward secrecy, optional group secret"),
            ("Routing (Ironwood)", "Spanning tree, greedy forwarding, bloom-filter lookups"),
            ("Peer links", "TCP, TLS, QUIC, WebSocket; every link proves its key")],
 "send": "send", "recv": "receive",
 "side": [("Identity", AMBER, ["Ed25519 key pair", "Address = hash of the", "public key"]),
          ("Admin API", VIOLET, ["Local socket only", "uqda status · uqda peers", "uqdactl …"]),
          ("No telemetry", GREEN, ["Nothing phones home.", "Logs stay on the device."])],

 "s4": ("Peers: how nodes find each other", ["A peer is a direct link between two nodes. A node needs only a few."]),
 "peers": [("Local network", ["Found automatically by", "multicast. Works even", "without internet."], TEAL),
           ("Peers you choose", ["Your own server, a", "friend's node, any", "address you trust."], VIOLET),
           ("Public peers", ["uqda find and AutoPeers", "pick the fastest and", "replace failed ones."], AMBER),
           ("Invites", ["uqda invite and uqda join", "add your group's devices", "safely."], GREEN)],
 "you": "your node",
 "transports": ["TCP", "TLS", "QUIC", "WebSocket", "SOCKS", "Unix socket"],
 "handshake": ["Every link starts with a handshake. Each side proves it holds its key over a fresh random",
               "value, so a recorded handshake cannot be replayed to impersonate a node."],

 "s5": ("A network that builds itself", ["Nobody plans the map. The nodes agree on one by talking to their neighbours."]),
 "root": "root",
 "tree_note": "Coordinates = the path from the root. No node knows the whole network.",
 "tree_cards": [("Gossip builds a tree", ["Nodes exchange small signed", "messages and agree on one", "tree that spans everyone."]),
                ("Coordinates", ["Each node's position is its", "path from the root, like", "[1, 2] or [2, 1, 1]."]),
                ("Tiny memory", ["A node keeps a few records", "about itself and its peers,", "not the whole network."]),
                ("Grows and heals", ["New nodes add roads. When", "a link breaks, the tree", "re-forms around it."])],

 "s6": ("How a packet finds its way", ["Three simple rules, applied by every node, route traffic across the world."]),
 "route": [("Find the key", ["A small lookup travels along", "the tree. Each link carries a", "bloom filter of the keys behind", "it, so lookups go the right way."]),
           ("Hop to the closest", ["Each node hands the packet to", "the peer whose coordinates", "are closest to the target.", "Shortcuts are used when they exist."]),
           ("Repair", ["When a link breaks, the sender", "is told, looks the target up", "again, and traffic takes a", "new path."])],
 "target": "target",

 "s7": ("What anyone in the middle sees", ["Traffic is encrypted end to end. Nodes in the middle carry it but cannot open it."]),
 "sender": "You", "receiver": "Friend", "relay": "relay",
 "sees": "Relays see", "sees_l": ["that traffic passes,", "its size and timing"],
 "cannot": "Relays cannot", "cannot_l": ["read it, change it,", "or replay it"],
 "crypto": ["X25519 key agreement", "XSalsa20-Poly1305", "Ephemeral keys + ratchet", "Replay protection"],
 "crypto_note": ["Forward secrecy: old traffic stays protected even if a key leaks later. TLS or QUIC on a link",
                 "also hides the protocol from your internet provider."],

 "s8": ("Public network or private group", ["Everyone shares the same roads. One setting decides who may talk to you."]),
 "pub": ("Public", ["Any node can open an encrypted", "session with yours.", "", "The firewall keeps your services", "private: open only what you share.", "", "For public services and the wider", "Yggdrasil community."]),
 "priv": ("Private group", ["Members share a secret, stretched", "with Argon2id and mixed into", "session encryption.", "Outsiders cannot open a session", "with you at all, even with the", "firewall off. Public nodes can still", "relay your encrypted traffic.", "For your devices, family or team."]),
 "outsider": "outsider",
 "setup_pub": "sudo uqda setup mode=public", "setup_priv": "sudo uqda setup",

 "s9": ("A packet's journey, step by step", ["From one app on your laptop to an app on your friend's computer."]),
 "journey": [("Your app sends", "A program sends to your friend's 200:… address, like any IPv6 address.", "ssh user@[200:…]"),
             ("uqda0 takes it", "The operating system routes all of 200::/7 to the uqda0 interface.", "200::/7 → uqda0"),
             ("Find the key", "The address points to a key; a lookup follows bloom filters to it.", "lookup"),
             ("Encrypt", "Session keys from X25519, ratcheted; the group secret is mixed in.", "XSalsa20-Poly1305"),
             ("Hop by hop", "Each node forwards it to the peer closest to the target.", "greedy routing"),
             ("Decrypt and check", "Only your friend's node can open it. Replays and outsiders are dropped.", "verify"),
             ("Firewall, then the app", "Rules decide. The app sees a normal connection from your address.", "allow tcp/22")],

 "s10": ("The firewall decides what comes in", ["On by default for new nodes. Rules match a key, address, subnet, protocol and port."]),
 "fw": [(True, "Reply to a connection you opened", "allowed, always"),
        (True, "Laptop 200:aa…  →  tcp/22", "allowed by rule \"laptop SSH\""),
        (True, "Any node  →  ping", "allowed by AllowPing"),
        (False, "Unknown node 200:ff…  →  tcp/8080", "dropped, no rule"),
        (False, "Unknown node 200:ff…  →  udp/53", "dropped, no rule")],

 "s11": ("Join in three commands", ["From nothing to a private network of your own devices."]),
 "join": [("sudo uqda setup country=germany", ["Keeps or creates the identity", "Private group, 192-bit secret", "Firewall on", "Fastest healthy public peers,", "   kept healthy by AutoPeers"]),
          ("sudo uqda invite", ["A one-time code, also as a QR", "Expires after 15 minutes", "Contains no password", "Redeemed on the local network"]),
          ("sudo uqda join <invite>", ["Checks the inviter over TLS 1.3", "   against its key", "Proves it holds the secret", "Receives the group secret", "   and peers"])],
 "term": ["$ sudo uqda setup country=germany", "$ sudo uqda invite                 # first device", "$ sudo uqda join uqda-invite-1:…   # new device", "$ uqda status                      # all checks pass"],

 "s12": ("Build anything on it", ["Anything that listens on IPv6 already works. The sender's address is a login you can trust."]),
 "build": [("Websites", ["Host at home,", "no domain needed"]), ("Chat and voice", ["Matrix, XMPP,", "Mumble"]),
           ("Files", ["Syncthing, rsync,", "Nextcloud"]), ("Smart home", ["Reach home without", "a vendor cloud"]),
           ("Games", ["Friends join by", "address"]), ("Apps", ["Android and Apple", "SDKs"])],

 "foot_title": "Good to know",
 "foot": [("Not anonymous", ["Your peers see your IP address.", "Uqda hides content, not who", "talks to whom."], ROSE),
          ("Experimental", ["Tested extensively, but no", "independent security audit", "yet."], AMBER),
          ("No telemetry", ["The developers run no servers", "that see your traffic and", "receive no data from you."], GREEN)],
 "brand": "Uqda Core 26.1 · compatible with the Yggdrasil network · free software, LGPLv3 · github.com/Uqda/Core",
},
"ar": {
 "kicker": "عُقدة  ·  UQDA  ·  شبكة مكوّنة من العُقد",
 "title": "كيف تعمل شبكة عُقدة",
 "sub": ["كل جهاز يثبّت عُقدة يصبح عُقدة في الشبكة. تتصل العُقد ببعضها",
         "وتحمل اتصالات بعضها مشفّرة. ومعًا، هي الشبكة نفسها."],
 "chips": ["لا شركة", "لا حساب", "لا خادم مركزي", "برنامج حر"],

 "s1": ("كل جهاز يصبح عُقدة", ["ثبّت عُقدة على أي جهاز، فينضم إلى الشبكة عضوًا مساويًا لغيره."]),
 "devices": ["هاتف", "حاسوب محمول", "حاسوب مكتبي", "خادم", "خادم منزلي"],
 "roles": [("يحصل على هوية", ["زوج مفاتيح يُنشأ على الجهاز.", "لا تصدره جهة، ولا يستطيع", "أحد انتزاعه."]),
           ("يتصل بجيرانه", ["بضع عُقد قريبة: على شبكتك", "المحلية أو عبر الإنترنت."]),
           ("يصبح موجّهًا", ["يمرّر الاتصالات المشفّرة", "للآخرين حين يقع على أفضل", "طريق بينهم."])],

 "s2": ("هوية لا يستطيع أحد انتزاعها", ["عنوانك محسوب من مفتاحك. دون المفتاح لا يستطيع أحد استخدامه."]),
 "id_priv": ("المفتاح الخاص", ["Ed25519، يبقى", "على الجهاز"]),
 "id_pub": ("المفتاح العام", ["يُشارك مع", "العُقد الأخرى"]),
 "id_hash": "بصمة",
 "id_addr": ("عنوانك", "200:1a2b:3c4d::7"),
 "id_sub": ("شبكتك الفرعية", "300:1a2b:3c4d::/64"),
 "id_sub_note": "للأجهزة خلف العُقدة",
 "id_note": ["الاتصال القادم من 200:1a2b:3c4d::7 يأتي فعلًا من صاحب ذلك المفتاح.",
             "تستطيع البرامج استخدام العنوان لتسجيل الدخول، دون كلمات مرور."],

 "s3": ("داخل العُقدة الواحدة", ["ترى برامجك IPv6 عاديًا، وتتولى عُقدة الباقي طبقةً بعد طبقة."]),
 "layers": [("تطبيقاتك", "المتصفح وSSH والألعاب والمحادثة وخدمتك الخاصة"),
            ("واجهة uqda0", "IPv6 عادي: عناوين 200::/7، لا شيء جديد لتتعلمه"),
            ("الجدار الناري للوارد", "الردود تمرّ؛ الاتصالات الجديدة فقط إن سمحت بها قاعدة"),
            ("جلسات مشفّرة", "من الطرف إلى الطرف، سرّية أمامية، سرّ مجموعة اختياري"),
            ("التوجيه (Ironwood)", "شجرة ممتدة، تمرير جشِع، بحث بمرشّحات بلوم"),
            ("روابط الجيران", "TCP وTLS وQUIC وWebSocket؛ كل رابط يُثبت مفتاحه")],
 "send": "إرسال", "recv": "استقبال",
 "side": [("الهوية", AMBER, ["زوج مفاتيح Ed25519", "العنوان = بصمة", "المفتاح العام"]),
          ("واجهة الإدارة", VIOLET, ["مقبس محلي فقط", "uqda status · uqda peers", "uqdactl …"]),
          ("لا تتبّع", GREEN, ["لا شيء يتصل بالمطوّرين.", "السجلات تبقى على الجهاز."])],

 "s4": ("الجيران: كيف تجد العُقد بعضها", ["الجار رابط مباشر بين عُقدتين، وتكفي العُقدةَ بضعةُ جيران."]),
 "peers": [("الشبكة المحلية", ["تُكتشف تلقائيًا، وتعمل", "حتى دون إنترنت."], TEAL),
           ("جيران تختارهم", ["خادمك، أو عُقدة صديق،", "أو أي عنوان تثق به."], VIOLET),
           ("عُقد عامة", ["uqda find وAutoPeers", "تختار الأسرع وتستبدل", "المتعطّل."], AMBER),
           ("الدعوات", ["uqda invite وuqda join", "تضيف أجهزة مجموعتك", "بأمان."], GREEN)],
 "you": "عُقدتك",
 "transports": ["TCP", "TLS", "QUIC", "WebSocket", "SOCKS", "Unix socket"],
 "handshake": ["يبدأ كل رابط بمصافحة يُثبت فيها كل طرف امتلاكه مفتاحه على قيمة عشوائية جديدة،",
               "فلا يمكن إعادة مصافحة مسجّلة لانتحال عُقدة."],

 "s5": ("شبكة تبني نفسها", ["لا أحد يرسم الخريطة. تتفق العُقد عليها بالتحدث مع جيرانها."]),
 "root": "الجذر",
 "tree_note": "الإحداثيات = الطريق من الجذر. لا توجد عُقدة تعرف الشبكة كلها.",
 "tree_cards": [("الشجرة تُبنى بالتناقل", ["تتبادل العُقد رسائل صغيرة", "موقّعة وتتفق على شجرة", "واحدة تشمل الجميع."]),
                ("الإحداثيات", ["موقع كل عُقدة هو طريقها", "من الجذر، مثل [1, 2]", "أو [2, 1, 1]."]),
                ("ذاكرة صغيرة", ["تحتفظ العُقدة بسجلات قليلة", "عن نفسها وجيرانها، لا عن", "الشبكة كلها."]),
                ("تنمو وتُصلح نفسها", ["العُقد الجديدة تضيف طرقًا،", "وحين ينقطع رابط تُعاد", "الشجرة حوله."])],

 "s6": ("كيف تجد الرسالة طريقها", ["ثلاث قواعد بسيطة تطبّقها كل عُقدة، فتوجّه الاتصالات عبر العالم."]),
 "route": [("إيجاد المفتاح", ["بحث صغير يسير على الشجرة.", "يحمل كل رابط مرشّح بلوم", "بالمفاتيح التي خلفه، فيسير", "البحث في الاتجاه الصحيح."]),
           ("القفز إلى الأقرب", ["تسلّم كل عُقدة الرسالة للجار", "الذي إحداثياته أقرب إلى الهدف،", "وتُستخدم الطرق المختصرة", "حين توجد."]),
           ("الإصلاح", ["حين ينقطع رابط يُبلَّغ المرسل،", "فيبحث عن الهدف مجددًا،", "ويسلك الاتصال طريقًا", "جديدًا."])],
 "target": "الهدف",

 "s7": ("ماذا يرى من في المنتصف", ["الاتصال مشفّر من الطرف إلى الطرف. العُقد الوسيطة تحمله ولا تستطيع فتحه."]),
 "sender": "أنت", "receiver": "صديقك", "relay": "وسيط",
 "sees": "يرى الوسطاء", "sees_l": ["أن هناك اتصالًا يمرّ،", "وحجمه وتوقيته"],
 "cannot": "لا يستطيع الوسطاء", "cannot_l": ["قراءته أو تغييره", "أو إعادة إرساله"],
 "crypto": ["X25519 لتبادل المفاتيح", "XSalsa20-Poly1305", "مفاتيح مؤقتة تتجدّد", "حماية من الإعادة"],
 "crypto_note": ["السرّية الأمامية: يبقى الاتصال القديم محميًا حتى لو تسرّب مفتاح لاحقًا. واستخدام TLS أو QUIC",
                 "على الرابط يخفي البروتوكول أيضًا عن مزوّد الإنترنت."],

 "s8": ("شبكة عامة أو مجموعة خاصة", ["الجميع يتشاركون الطرق نفسها، وإعداد واحد يقرّر من يستطيع محادثتك."]),
 "pub": ("عامة", ["أي عُقدة تستطيع فتح جلسة", "مشفّرة مع عُقدتك.", "", "الجدار الناري يُبقي خدماتك", "خاصة: افتح فقط ما تشاركه.", "", "للخدمات العامة ومجتمع", "Yggdrasil الأوسع."]),
 "priv": ("مجموعة خاصة", ["يتشارك الأعضاء سرًّا يُقوّى", "بـ Argon2id ويُمزج في", "تشفير الجلسات.", "لا يستطيع الغرباء فتح جلسة", "معك أبدًا، حتى مع إيقاف", "الجدار. وتستطيع العُقد العامة", "تمرير اتصالك المشفّر.", "لأجهزتك أو عائلتك أو فريقك."]),
 "outsider": "غريب",
 "setup_pub": "sudo uqda setup mode=public", "setup_priv": "sudo uqda setup",

 "s9": ("رحلة الرسالة خطوة بخطوة", ["من تطبيق على حاسوبك إلى تطبيق على حاسوب صديقك."]),
 "journey": [("تطبيقك يرسل", "يرسل برنامج إلى عنوان صديقك 200:… كأي عنوان IPv6.", "ssh user@[200:…]"),
             ("uqda0 تستلمها", "يوجّه نظام التشغيل كل عناوين 200::/7 إلى واجهة uqda0.", "200::/7 → uqda0"),
             ("إيجاد المفتاح", "العنوان يدل على مفتاح، والبحث يتبع مرشّحات بلوم إليه.", "lookup"),
             ("التشفير", "مفاتيح جلسة X25519 متجدّدة، ويُمزج سرّ المجموعة.", "XSalsa20-Poly1305"),
             ("قفزة بعد قفزة", "تمرّرها كل عُقدة إلى الجار الأقرب إلى الهدف.", "greedy routing"),
             ("فكّ التشفير والتحقق", "عُقدة صديقك وحدها تفتحها، وتُرفض المُعادة والغريبة.", "verify"),
             ("الجدار ثم التطبيق", "القواعد تقرّر، ويرى التطبيق اتصالًا عاديًا من عنوانك.", "allow tcp/22")],

 "s10": ("الجدار الناري يقرّر ما يدخل", ["مفعّل افتراضيًا للعُقد الجديدة. القواعد تطابق المفتاح والعنوان والشبكة والبروتوكول والمنفذ."]),
 "fw": [(True, "ردّ على اتصال فتحته أنت", "مسموح دائمًا"),
        (True, "الحاسوب 200:aa…  ←  tcp/22", "مسموح بقاعدة \"laptop SSH\""),
        (True, "أي عُقدة  ←  ping", "مسموح بـ AllowPing"),
        (False, "عُقدة غريبة 200:ff…  ←  tcp/8080", "مرفوض، لا قاعدة"),
        (False, "عُقدة غريبة 200:ff…  ←  udp/53", "مرفوض، لا قاعدة")],

 "s11": ("انضم بثلاثة أوامر", ["من لا شيء إلى شبكة خاصة تجمع أجهزتك."]),
 "join": [("sudo uqda setup country=germany", ["يحتفظ بالهوية أو ينشئها", "مجموعة خاصة بسرّ 192 بت", "تفعيل الجدار الناري", "أسرع العُقد العامة السليمة،", "   ويحافظ عليها AutoPeers"]),
          ("sudo uqda invite", ["رمز لمرة واحدة، ورمز QR", "ينتهي بعد 15 دقيقة", "لا يحتوي على كلمة مرور", "يُستخدم على الشبكة المحلية"]),
          ("sudo uqda join <invite>", ["يتحقق من الداعي عبر TLS 1.3", "   مقابل مفتاحه", "يُثبت امتلاكه السرّ", "يستلم سرّ المجموعة", "   والجيران"])],
 "term": ["$ sudo uqda setup country=germany", "$ sudo uqda invite                 # first device", "$ sudo uqda join uqda-invite-1:…   # new device", "$ uqda status                      # all checks pass"],

 "s12": ("ابنِ عليها ما تشاء", ["كل ما يستمع على IPv6 يعمل مباشرة، وعنوان المرسل تسجيل دخول تستطيع الوثوق به."]),
 "build": [("مواقع", ["استضافة من المنزل،", "دون نطاق"]), ("محادثة وصوت", ["Matrix وXMPP", "وMumble"]),
           ("ملفات", ["Syncthing وrsync", "وNextcloud"]), ("منزل ذكي", ["الوصول للمنزل دون", "سحابة شركة"]),
           ("ألعاب", ["الأصدقاء ينضمون", "بالعنوان"]), ("تطبيقات", ["حزم Android", "وApple"])],

 "foot_title": "من المهم أن تعرف",
 "foot": [("ليست أداة إخفاء هوية", ["جيرانك يرون عنوان IP الخاص بك.", "عُقدة تخفي المحتوى، لا من", "يتحدث مع من."], ROSE),
          ("مشروع تجريبي", ["مُختبر على نطاق واسع، لكن", "دون تدقيق أمني مستقل", "بعد."], AMBER),
          ("لا تتبّع", ["لا يشغّل المطوّرون خوادم ترى", "اتصالك، ولا تصلهم أي", "بيانات منك."], GREEN)],
 "brand": "Uqda Core 26.1 · متوافقة مع شبكة Yggdrasil · برنامج حر، LGPLv3 · github.com/Uqda/Core",
},
}


# --------------------------------------------------------------------------
# Sections. Each takes the poster, the content and a top y, and returns the
# bottom y.

def hero(p, t, y):
    p.rect(0, y, W, 600, "url(#heroGlow)", r=0)
    p.text(W / 2, y + 86, t["kicker"], 15, TEAL, 700, anchor="middle", ls=3)
    p.raw(f'<text x="{W/2}" y="{y+160}" font-size="60" font-weight="800" fill="url(#titleGrad)" text-anchor="middle"{" direction=\"rtl\"" if p.rtl else ""}>{escape(t["title"])}</text>')
    for i, l in enumerate(t["sub"]):
        p.text(W / 2, y + 214 + i * 32, l, 21, MUTED, 400, anchor="middle")
    # mesh of glowing nodes
    rnd = random.Random(11)
    pts = []
    while len(pts) < 46:
        x, yy = rnd.uniform(70, 1130), rnd.uniform(y + 300, y + 500)
        if all((x - a) ** 2 + (yy - b) ** 2 > 62 ** 2 for a, b in pts):
            pts.append((x, yy))
    edges = set()
    for i, (x, yy) in enumerate(pts):
        near = sorted(range(len(pts)), key=lambda j: (pts[j][0] - x) ** 2 + (pts[j][1] - yy) ** 2)[1:4]
        for j in near:
            edges.add(tuple(sorted((i, j))))
    for i, j in edges:
        p.line(*pts[i], *pts[j], "#2a3a6a", 1.6)
    hl = [5, 12, 20, 27, 33]
    order = sorted(range(len(pts)), key=lambda k: pts[k][0])
    path = [order[k] for k in (2, 9, 17, 26, 34, 43)]
    for a, b in zip(path, path[1:]):
        p.line(*pts[a], *pts[b], TEAL, 3.5, op=.9)
    for k, (x, yy) in enumerate(pts):
        c = TEAL if k in path else (VIOLET if k in hl else "#5c6fa8")
        p.node(x, yy, 7 if k not in path else 9, c, glow=k in path or k in hl)
    # chips
    widths = [int(len(c) * 15 * .6 + 40) for c in t["chips"]]
    total = sum(widths) + 14 * (len(widths) - 1)
    x = (W - total) / 2
    for c, w in zip(t["chips"], widths):
        p.chip(x, y + 535, c, TEAL, w=w)
        x += w + 14
    return y + 610


def s_devices(p, t, y):
    h = 470
    p.section(y, h, 1, *t["s1"])
    xs = [190, 395, 600, 805, 1010]
    icons = [p.icon_phone, p.icon_laptop, p.icon_desktop, p.icon_server, p.icon_box]
    for i, (x, f) in enumerate(zip(xs, icons)):
        p.circle(x, y + 205, 56, fill="url(#cardFill)", stroke=CARDLINE, sw=1.2)
        f(x, y + 205, INK)
        p.node(x + 40, y + 165, 9, TEAL)
        p.text(x, y + 290, t["devices"][i], 17, INK, 600, anchor="middle")
    for a, b in zip(xs, xs[1:]):
        p.line(a + 60, y + 205, b - 60, y + 205, TEAL, 2, dash="2 8", op=.7)
    cols = [TEAL, VIOLET, AMBER]
    for i, (title, lines) in enumerate(t["roles"]):
        x = 70 + i * 360
        p.card(x, y + 322, 340, 122)
        p.circle(x + 30, y + 352, 14, fill=cols[i])
        p.text(x + 30, y + 358, p.num(i + 1), 15, BG1, 800, anchor="middle")
        p.text(x + 54, y + 358, title, 19, INK, 700)
        p.para(x + 22, y + 388, lines, 15, MUTED, 21)
    return y + h


def s_identity(p, t, y):
    h = 430
    p.section(y, h, 2, *t["s2"])
    yy = y + 165
    # private key
    p.rect(70, yy, 230, 130, "url(#amberFill)", stroke=AMBER, r=18, sw=1.5)
    p.icon_lock(115, yy + 58, AMBER)
    p.text(150, yy + 48, t["id_priv"][0], 19, AMBER, 700)
    p.para(150, yy + 76, t["id_priv"][1], 14, MUTED, 19)
    p.arrow(305, yy + 65, 355, yy + 65, AMBER)
    # public key
    p.rect(360, yy, 210, 130, "url(#cardFill)", stroke=CARDLINE, r=18, sw=1.5)
    p.icon_key(405, yy + 58, INK)
    p.text(440, yy + 48, t["id_pub"][0], 19, INK, 700)
    p.para(440, yy + 76, t["id_pub"][1], 14, MUTED, 19)
    p.arrow(575, yy + 65, 640, yy + 65, TEAL)
    p.text(607, yy + 52, t["id_hash"], 13, TEAL, 700, anchor="middle")
    # address and subnet
    p.rect(645, yy, 485, 60, "url(#tealFill)", stroke=TEAL, r=14, sw=1.5)
    p.text(665, yy + 38, t["id_addr"][0], 16, TEAL, 700)
    p.text(1110, yy + 39, t["id_addr"][1], 20, INK, 700, anchor="end", mono=True)
    p.rect(645, yy + 70, 485, 60, "url(#violetFill)", stroke=VIOLET, r=14, sw=1.5)
    p.text(665, yy + 98, t["id_sub"][0], 16, VIOLET, 700)
    p.text(665, yy + 118, t["id_sub_note"], 12, MUTED)
    p.text(1110, yy + 109, t["id_sub"][1], 18, INK, 700, anchor="end", mono=True)
    p.card(70, yy + 158, 1060, 82, stroke=CARDLINE)
    p.para(95, yy + 192, t["id_note"], 16, INK, 26)
    return y + h


def s_inside(p, t, y):
    lh, gap = 62, 12
    h = 160 + 6 * (lh + gap) + 40
    p.section(y, h, 3, *t["s3"])
    lx, lw, ly = 125, 610, y + 160
    fills = ["url(#cardFill)", "url(#cardFill)", "url(#tealFill)", "url(#tealFill)", "url(#violetFill)", "url(#cardFill)"]
    strokes = [CARDLINE, CARDLINE, TEAL, TEAL, VIOLET, CARDLINE]
    for i, (title, body) in enumerate(t["layers"]):
        yy = ly + i * (lh + gap)
        p.rect(lx, yy, lw, lh, fills[i], stroke=strokes[i], r=12, sw=1.4)
        p.text(lx + 20, yy + 27, title, 18, INK, 700)
        p.text(lx + 20, yy + 49, body, 14, MUTED)
    top, bot = ly + 6, ly + 6 * (lh + gap) - gap - 6
    p.arrow(88, top, 88, bot, TEAL, 3)
    p.arrow(104, bot, 104, top, VIOLET, 3)
    p.text(80, bot + 28, "↓ " + t["send"], 13, TEAL, 600)
    p.text(80, bot + 46, "↑ " + t["recv"], 13, VIOLET, 600)
    sy = ly
    for title, col, lines in t["side"]:
        hh = 30 + 26 + len(lines) * 21 + 14
        p.rect(770, sy, 360, hh, "url(#cardFill)", stroke=col, r=14, sw=1.4)
        p.circle(796, sy + 30, 7, fill=col)
        p.text(814, sy + 36, title, 18, col, 700)
        p.para(796, sy + 66, lines, 14, MUTED, 21)
        sy += hh + 16
    p.arrow(768, ly + 52, 740, ly + (lh + gap) + 31, AMBER, 2)
    return y + h


def s_peers(p, t, y):
    h = 640
    p.section(y, h, 4, *t["s4"])
    cx, cy = 600, y + 320
    boxes = [(70, y + 160), (790, y + 160), (70, y + 380), (790, y + 380)]
    anchors = [(410, y + 215), (790, y + 215), (410, y + 435), (790, y + 435)]
    for (bx, by), (ax, ay), (title, lines, col) in zip(boxes, anchors, t["peers"]):
        p.line(cx, cy, ax, ay, col, 2.5, dash="6 6", op=.8)
        p.rect(bx, by, 340, 120, "url(#cardFill)", stroke=col, r=16, sw=1.4)
        p.node(bx + 34, by + 36, 9, col, glow=False)
        p.text(bx + 58, by + 42, title, 19, col, 700)
        p.para(bx + 24, by + 72, lines, 14, MUTED, 20)
    p.circle(cx, cy, 80, fill=TEAL, op=.08)
    p.circle(cx, cy, 58, fill="url(#tealFill)", stroke=TEAL, sw=2.5)
    p.node(cx, cy - 8, 12, TEAL)
    p.text(cx, cy + 30, t["you"], 15, INK, 700, anchor="middle")
    widths = [int(len(c) * 15 * .62 + 34) for c in t["transports"]]
    total = sum(widths) + 12 * (len(widths) - 1)
    x = (W - total) / 2
    for c, w in zip(t["transports"], widths):
        p.chip(x, y + 520, c, VIOLET, mono=True, w=w)
        x += w + 12
    for k, l in enumerate(t["handshake"]):
        p.text(W / 2, y + 590 + k * 22, l, 15, MUTED, anchor="middle")
    return y + h


def s_tree(p, t, y):
    h = 600
    p.section(y, h, 5, *t["s5"])
    p.card(70, y + 150, 680, 410)
    N = {"r": (410, y + 205, "[ ]"),
         "a": (250, y + 290, "[1]"), "b": (570, y + 290, "[2]"),
         "c": (160, y + 380, "[1,1]"), "d": (330, y + 380, "[1,2]"), "e": (490, y + 380, "[2,1]"), "f": (660, y + 380, "[2,2]"),
         "g": (115, y + 470, "[1,1,1]"), "h2": (215, y + 470, "[1,1,2]"), "i": (420, y + 470, "[2,1,1]"), "j": (560, y + 470, "[2,1,2]"), "k": (690, y + 470, "[2,2,1]")}
    tree = [("r", "a"), ("r", "b"), ("a", "c"), ("a", "d"), ("b", "e"), ("b", "f"), ("c", "g"), ("c", "h2"), ("e", "i"), ("e", "j"), ("f", "k")]
    mesh = [("d", "e"), ("h2", "i"), ("d", "i")]
    for a, b in mesh:
        p.line(N[a][0], N[a][1], N[b][0], N[b][1], "#4a5a8a", 2, dash="5 6")
    for a, b in tree:
        p.line(N[a][0], N[a][1], N[b][0], N[b][1], "#6b7cb0", 2.5)
    for k, (x, yy, c) in N.items():
        p.node(x, yy, 15 if k == "r" else 11, AMBER if k == "r" else TEAL, glow=k == "r")
        p.text(x, yy + 32, c, 13, MUTED, 600, anchor="middle", mono=True)
    p.text(410, y + 168, t["root"], 14, AMBER, 700, anchor="middle")
    p.text(410, y + 540, t["tree_note"], 14, INK, 600, anchor="middle")
    cy = y + 150
    cols = [TEAL, AMBER, VIOLET, GREEN]
    for (title, lines), col in zip(t["tree_cards"], cols):
        p.rect(775, cy, 355, 94, "url(#cardFill)", stroke=CARDLINE, r=14, sw=1.2)
        p.rect(775, cy + 14, 4, 66, col, r=2)
        p.text(797, cy + 30, title, 17, col, 700)
        p.para(797, cy + 52, lines, 13, MUTED, 18)
        cy += 105
    return y + h


def s_route(p, t, y):
    h = 560
    p.section(y, h, 6, *t["s6"])
    cols = [TEAL, VIOLET, AMBER]
    for i, ((title, lines), col) in enumerate(zip(t["route"], cols)):
        x = 70 + i * 360
        p.card(x, y + 150, 340, 380)
        p.rect(x + 20, y + 170, 300, 170, "#0f1730", stroke=CARDLINE, r=12)
        ox, oy = x + 20, y + 170
        if i == 0:
            pts = [(150, 30), (80, 85), (220, 85), (45, 140), (115, 140), (185, 140), (255, 140)]
            for a, b in [(0, 1), (0, 2), (1, 3), (1, 4), (2, 5), (2, 6)]:
                p.line(ox + pts[a][0], oy + pts[a][1], ox + pts[b][0], oy + pts[b][1], "#4a5a8a", 2)
            for a, b in [(3, 1), (1, 0), (0, 2), (2, 6)]:
                p.line(ox + pts[a][0], oy + pts[a][1], ox + pts[b][0], oy + pts[b][1], col, 4)
            for k, (a, b) in enumerate(pts):
                p.node(ox + a, oy + b, 8, AMBER if k == 6 else (col if k in (3, 1, 0, 2) else "#5c6fa8"), glow=False)
            for (a, b) in [(115, 58), (185, 58), (62, 112), (237, 112)]:
                p.rect(ox + a - 11, oy + b - 8, 22, 16, VIOLET, r=3, op=.8)
                for q in range(3):
                    p.line(ox + a - 7 + q * 7, oy + b - 4, ox + a - 7 + q * 7, oy + b + 4, BG1, 1.5)
            p.text(ox + 255, oy + 165, t["target"], 12, AMBER, 700, anchor="middle")
        elif i == 1:
            xs = [35, 100, 165, 230, 275]
            ys = [120, 70, 110, 60, 110]
            for k in range(4):
                p.arrow(ox + xs[k] + 10, oy + ys[k], ox + xs[k + 1] - 12, oy + ys[k + 1], col, 3)
            for k, (a, b) in enumerate(zip(xs, ys)):
                p.node(ox + a, oy + b, 9, AMBER if k == 4 else col, glow=False)
                p.text(ox + a, oy + b + 28, p.num(4 - k), 13, MUTED, 700, anchor="middle")
            p.text(ox + 275, oy + 40, t["target"], 12, AMBER, 700, anchor="middle")
        else:
            a, b, c2, d = (40, 85), (130, 50), (210, 85), (130, 130)
            p.line(ox + a[0], oy + a[1], ox + b[0], oy + b[1], "#4a5a8a", 2)
            p.line(ox + b[0], oy + b[1], ox + c2[0], oy + c2[1], ROSE, 3, dash="4 5")
            p.line(ox + 162, oy + 58, ox + 178, oy + 76, ROSE, 3.5)
            p.line(ox + 178, oy + 58, ox + 162, oy + 76, ROSE, 3.5)
            p.line(ox + a[0], oy + a[1], ox + d[0], oy + d[1], col, 4)
            p.line(ox + d[0], oy + d[1], ox + c2[0], oy + c2[1], col, 4)
            p.line(ox + c2[0], oy + c2[1], ox + 275, oy + 85, col, 4)
            for q in (a, b, c2, d):
                p.node(ox + q[0], oy + q[1], 9, col if q != b else "#5c6fa8", glow=False)
            p.node(ox + 275, oy + 85, 9, AMBER, glow=False)
            p.text(ox + 275, oy + 112, t["target"], 12, AMBER, 700, anchor="middle")
        p.circle(x + 36, y + 380, 15, fill=col)
        p.text(x + 36, y + 386, p.num(i + 1), 16, BG1, 800, anchor="middle")
        p.text(x + 62, y + 387, title, 19, INK, 700)
        p.para(x + 24, y + 420, lines, 14, MUTED, 21)
    return y + h


def s_crypto(p, t, y):
    h = 560
    p.section(y, h, 7, *t["s7"])
    ly = y + 260
    p.rect(70, ly - 60, 200, 120, "url(#tealFill)", stroke=TEAL, r=18, sw=1.5)
    p.icon_laptop(170, ly - 10, INK)
    p.text(170, ly + 42, t["sender"], 17, TEAL, 700, anchor="middle")
    p.rect(930, ly - 60, 200, 120, "url(#tealFill)", stroke=TEAL, r=18, sw=1.5)
    p.icon_desktop(1030, ly - 10, INK)
    p.text(1030, ly + 42, t["receiver"], 17, TEAL, 700, anchor="middle")
    p.line(275, ly, 925, ly, TEAL, 3, dash="2 10")
    for x in (480, 720):
        p.node(x, ly, 16, VIOLET)
        p.text(x, ly + 44, t["relay"], 14, VIOLET, 700, anchor="middle")
    for x in (370, 600, 830):
        p.icon_envelope(x, ly, AMBER, 40, 28)
        p.circle(x + 18, ly - 14, 9, fill=TEAL)
    p.rect(330, y + 150, 260, 66, "url(#cardFill)", stroke=MUTED, r=12, sw=1)
    p.text(350, y + 175, t["sees"], 15, INK, 700)
    p.para(350, y + 197, [" ".join(t["sees_l"])], 13, MUTED, 18)
    p.rect(610, y + 150, 260, 66, "url(#roseFill)", stroke=ROSE, r=12, sw=1)
    p.text(630, y + 175, t["cannot"], 15, ROSE, 700)
    p.para(630, y + 197, [" ".join(t["cannot_l"])], 13, MUTED, 18)
    widths = [int(len(c) * 15 * .58 + 40) for c in t["crypto"]]
    total = sum(widths) + 14 * (len(widths) - 1)
    x = (W - total) / 2
    for c, w in zip(t["crypto"], widths):
        p.chip(x, y + 375, c, GREEN, w=w)
        x += w + 14
    p.card(70, y + 440, 1060, 90)
    p.para(95, y + 477, t["crypto_note"], 15, INK, 25)
    return y + h


def s_groups(p, t, y):
    h = 640
    p.section(y, h, 8, *t["s8"])
    for side in (0, 1):
        x = 70 + side * 540
        col = TEAL if side == 0 else AMBER
        fill = "url(#tealFill)" if side == 0 else "url(#amberFill)"
        title, lines = t["pub"] if side == 0 else t["priv"]
        p.rect(x, y + 150, 520, 460, fill, stroke=col, r=20, sw=1.5)
        p.text(x + 30, y + 196, title, 24, col, 800)
        ox, oy = x + 30, y + 220
        p.rect(ox, oy, 460, 150, "#0c1328", stroke=CARDLINE, r=14, op=.9)
        pts = [(60, 40), (150, 95), (240, 40), (330, 100), (410, 45), (110, 125), (280, 125), (380, 135)]
        for a, b in [(0, 1), (1, 2), (2, 3), (3, 4), (1, 5), (3, 6), (6, 7), (2, 6)]:
            p.line(ox + pts[a][0], oy + pts[a][1], ox + pts[b][0], oy + pts[b][1], "#3a4a7a", 2)
        members = {0, 1, 5, 6} if side else set()
        for k, (a, b) in enumerate(pts):
            p.node(ox + a, oy + b, 9, AMBER if k in members else TEAL, ring=AMBER if k in members else None, glow=False)
        if side == 1:
            q = pts[4]
            p.line(ox + q[0] - 9, oy + q[1] - 30, ox + q[0] + 9, oy + q[1] - 14, ROSE, 3.5)
            p.line(ox + q[0] + 9, oy + q[1] - 30, ox + q[0] - 9, oy + q[1] - 14, ROSE, 3.5)
            p.text(ox + q[0], oy + q[1] + 30, t["outsider"], 12, ROSE, 700, anchor="middle")
        p.para(x + 30, y + 405, lines, 15, INK, 21)
        p.text(x + 30, y + 590, t["setup_pub"] if side == 0 else t["setup_priv"], 15, col, 700, mono=True)
    return y + h


def s_journey(p, t, y):
    step = 104
    h = 170 + len(t["journey"]) * step + 20
    p.section(y, h, 9, *t["s9"])
    x0 = 120
    top = y + 175
    p.line(x0, top, x0, top + (len(t["journey"]) - 1) * step, "#2f3f70", 4)
    cols = [TEAL, TEAL, VIOLET, GREEN, VIOLET, GREEN, TEAL]
    for i, (title, body, code) in enumerate(t["journey"]):
        yy = top + i * step
        p.circle(x0, yy, 26, fill=cols[i], op=.16)
        p.circle(x0, yy, 19, fill=cols[i])
        p.text(x0, yy + 6, p.num(i + 1), 17, BG1, 800, anchor="middle")
        p.rect(165, yy - 36, 965, 78, "url(#cardFill)", stroke=CARDLINE, r=14, sw=1.2)
        p.text(190, yy - 6, title, 19, INK, 700)
        p.text(190, yy + 22, body, 15, MUTED)
        w = int(len(code) * 14 * .62 + 32)
        p.chip(1110 - w, yy - 24, code, cols[i], mono=True, size=14, w=w)
    return y + h


def s_firewall(p, t, y):
    rows = t["fw"]
    h = 170 + len(rows) * 74 + 20
    p.section(y, h, 10, *t["s10"])
    for i, (ok, what, why) in enumerate(rows):
        yy = y + 160 + i * 74
        col = GREEN if ok else ROSE
        p.rect(70, yy, 1060, 60, "url(#cardFill)" if ok else "url(#roseFill)", stroke=col if not ok else CARDLINE, r=12, sw=1.2)
        p.circle(108, yy + 30, 17, fill=col, op=.2)
        if ok:
            p.poly([(99, yy + 31), (106, yy + 38), (118, yy + 23)], stroke=col, sw=3.5, closed=False)
        else:
            p.line(101, yy + 23, 115, yy + 37, col, 3.5)
            p.line(115, yy + 23, 101, yy + 37, col, 3.5)
        p.text(145, yy + 37, what, 17, INK, 600)
        p.text(1105, yy + 37, why, 15, col, 700, anchor="end")
    return y + h


def s_join(p, t, y):
    h = 690
    p.section(y, h, 11, *t["s11"])
    cols = [TEAL, VIOLET, AMBER]
    for i, ((cmd, lines), col) in enumerate(zip(t["join"], cols)):
        x = 70 + i * 360
        p.rect(x, y + 150, 340, 270, "url(#cardFill)", stroke=col, r=18, sw=1.4)
        p.circle(x + 36, y + 190, 17, fill=col)
        p.text(x + 36, y + 197, p.num(i + 1), 17, BG1, 800, anchor="middle")
        p.rect(x + 64, y + 172, 256, 36, "#0b1226", r=8)
        p.text(x + 74, y + 196, cmd, 12.5, col, 700, mono=True)
        for j, l in enumerate(lines):
            cont = l.startswith(" ")
            p.text(x + (43 if cont else 28), y + 246 + j * 30, l.strip() if cont else "• " + l, 15, INK)
        if i < 2:
            p.arrow(x + 344, y + 285, x + 356, y + 285, MUTED, 2.5)
    # terminal
    tx, ty = 70, y + 450
    p.rect(tx, ty, 1060, 205, "#070b17", stroke=CARDLINE, r=16, sw=1.2)
    for k, c in enumerate([ROSE, AMBER, GREEN]):
        p.circle(tx + 26 + k * 20, ty + 22, 6, fill=c)
    term = t["term"]
    # Terminal output is always left-to-right.
    for k, l in enumerate(term):
        yy = ty + 70 + k * 34
        xx = tx + 30
        p.o.append(f'<text x="{xx if not p.rtl else W - tx - 1060 + 30}" y="{yy}" font-size="17" fill="{GREEN if k == 3 else INK}" font-family=\'{MONO}\' xml:space="preserve">{escape(l)}</text>')
    return y + h


def s_build(p, t, y):
    h = 400
    p.section(y, h, 12, *t["s12"])
    icons = [p.icon_globe, p.icon_chat, p.icon_folder, p.icon_house, p.icon_game, p.icon_phone]
    cols = [TEAL, VIOLET, AMBER, GREEN, ROSE, TEAL]
    for i, ((title, lines), f, col) in enumerate(zip(t["build"], icons, cols)):
        x = 70 + i * 180
        p.card(x, y + 150, 165, 215)
        p.circle(x + 82.5, y + 212, 40, fill=col, op=.13)
        f(x + 82.5, y + 212, col)
        p.text(x + 82.5, y + 288, title, 17, INK, 700, anchor="middle")
        for j, l in enumerate(lines):
            p.text(x + 82.5, y + 314 + j * 20, l, 13, MUTED, anchor="middle")
    return y + h


def footer(p, t, y):
    p.text(W / 2, y + 60, t["foot_title"], 28, INK, 700, anchor="middle")
    for i, (title, lines, col) in enumerate(t["foot"]):
        x = 70 + i * 360
        p.rect(x, y + 95, 340, 170, "url(#cardFill)", stroke=col, r=18, sw=1.5)
        p.circle(x + 34, y + 133, 9, fill=col)
        p.text(x + 54, y + 140, title, 20, col, 700)
        p.para(x + 26, y + 178, lines, 15, MUTED, 22)
    p.line(70, y + 310, 1130, y + 310, CARDLINE, 1.2)
    p.node(W / 2, y + 355, 10, TEAL)
    p.text(W / 2, y + 400, t["brand"], 14, DIM, 600, anchor="middle")
    return y + 440


def build(lang):
    t = C[lang]
    p = Poster(lang)
    body = []
    y = 0
    y = hero(p, t, y)
    for f in (s_devices, s_identity, s_inside, s_peers, s_tree, s_route, s_crypto, s_groups, s_journey, s_firewall, s_join, s_build):
        y = f(p, t, y + 34)
    y = footer(p, t, y + 20)
    height = int(y)
    head = (f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {W} {height}" width="{W}" height="{height}" '
            f'font-family=\'{FONT}\' role="img" aria-labelledby="t d">'
            f'<title id="t">{escape(t["title"])}</title><desc id="d">{escape(" ".join(t["sub"]))}</desc>'
            + defs() + f'<rect width="{W}" height="{height}" fill="url(#bg)"/>')
    return head + "\n" + "\n".join(p.o) + "\n</svg>\n"


if __name__ == "__main__":
    import os
    here = os.path.dirname(os.path.abspath(__file__))
    for lang in ("en", "ar"):
        with open(os.path.join(here, f"how-uqda-works.{lang}.svg"), "w", encoding="utf-8", newline="\n") as f:
            f.write(build(lang))
    print("written")
