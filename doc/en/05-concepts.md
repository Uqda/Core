# Technical Concepts - Uqda Network

## How the Network Works

### 1. Cryptographic Addressing

Each node in Uqda Network gets an IPv6 address derived from its public key.

#### Address Generation Process:

```
Private Key
    ↓
Public Key (Ed25519)
    ↓
Hash (SHA-512)
    ↓
Truncate + Add Prefix
    ↓
IPv6 Address (200:xxxx:xxxx:xxxx::/64)
```

#### Advantages:

- ✅ **Location Independent** - Address moves with device
- ✅ **No DHCP Needed** - No address allocation server required
- ✅ **Stable** - Same address as long as you use same keys
- ✅ **Secure** - Very difficult to spoof or steal

### 2. Address Ranges

#### Node Addresses: `200::/7`

```
Example: 200:1234:5678:9abc:def0:1234:5678:9abc
```

- Unique address per node
- Derived from public key
- Routable on Uqda Network

#### Subnets: `300::/7`

```
Example: 300:1234:5678:9abc::/64
```

- One /64 subnet per node
- For advertising to other devices
- Can assign addresses to non-Uqda devices

### 3. Smart Routing

#### Spanning Tree Construction

```
         [Root Node]
        /     |     \
    [A]      [B]     [C]
   /  \      |      /  \
 [D]  [E]   [F]   [G]  [H]
```

**How it works:**

1. Each node knows its distance from root
2. Routes calculated based on tree structure
3. Tree rebuilds automatically when topology changes

#### Optimal Path Selection

```
Source → Destination: Multiple paths available

Path 1: A → B → C → D (3 hops) ✅ Chosen
Path 2: A → E → F → G → D (4 hops)
Path 3: A → H → I → J → K → D (5 hops)
```

**Selection criteria:**

- 🎯 Shortest path first (minimum hops)
- ⚡ Lowest latency
- 📊 Link quality
- 🔄 Load balancing

#### Greedy Routing

When direct path isn't available:

```
[Source] → [Closest to destination] → [Next closest] → [Destination]
```

- Each hop brings you closer to target
- Falls back to tree routing if needed
- Hybrid approach for optimal performance

### 4. Self-Healing

#### When Link Fails:

```
Before:  A → B → C → D

           ↓ (B fails)

After:   A → E → F → D
         (automatic reroute)
```

**Advantages:**

- ⚡ Millisecond recovery time
- 🔄 No manual intervention
- 🎯 Always finds best path
- 💪 Resilient to failures

### 5. End-to-End Encryption

#### How it Works:

```
[Your Device] 🔒 ← encrypted → 🔒 [Their Device]
              ↓
      [Intermediate Nodes]
    (can't read the data)
```

#### Encrypted Components:

**1. Key Exchange:**

- Curve25519 ECDH
- Secure key exchange

**2. Session Key Derivation:**

- HKDF
- Unique keys per session

**3. Data Encryption:**

- ChaCha20-Poly1305
- Military-grade encryption

**4. Key Rotation:**

- Every time period
- Perfect Forward Secrecy (PFS)

#### What's Protected:

- ✅ **Confidentiality** - Data content encrypted
- ✅ **Integrity** - Data can't be modified
- ✅ **Authenticity** - Sender verified
- ✅ **Forward Secrecy** - Past sessions stay safe

#### What's NOT Protected:

- ❌ **Anonymity** - Peers can see your IP
- ❌ **Metadata** - Packet timing and size visible
- ❌ **Traffic Analysis** - Connection patterns observable

## Connection Types (Peering)

### 1. TCP

```
tcp://peer.example.com:9001
```

- Plain TCP connection
- ⚠️ Unencrypted
- ✅ Simple and fast

### 2. TLS

```
tls://peer.example.com:9001
```

- TCP with TLS encryption
- ✅ Recommended
- ✅ Secure and encrypted

### 3. QUIC

```
quic://peer.example.com:9002
```

- QUIC protocol
- ✅ Fastest
- ✅ Loss-resistant

### 4. WebSocket

```
ws://peer.example.com:9001
wss://peer.example.com:9001
```

- WebSocket
- ✅ Works behind HTTP proxy
- ✅ Useful for restricted networks

### 5. SOCKS

```
socks://127.0.0.1:9050/hidden.onion:9001
```

- Via SOCKS proxy
- ✅ For connecting through Tor
- ✅ For connecting through other proxy

### 6. UNIX Socket

```
unix:///var/run/uqda.sock
```

- Local connection only
- ✅ Very fast
- ✅ Secure (local)

## Automatic Discovery (Multicast Discovery)

### How it Works:

```
Device A: "Hello! I'm running Uqda!"
Device B: "Me too! Let's connect!"

         ↓

   [Automatic Connection]
```

### Steps:

1. 📢 Broadcast on local network
2. 👂 Listen for other Uqda nodes
3. 🤝 Establish automatic connections
4. ⚡ Zero configuration needed

### When it Works:

- ✅ On same WiFi network
- ✅ On same Ethernet cable
- ✅ On same local network (LAN)

### When it Doesn't Work:

- ❌ Across Internet
- ❌ Through VPN
- ❌ Through complex NAT

## Security and Privacy

### What Uqda Protects:

✅ **Protects:**

- Data content sent (end-to-end encrypted)
- Data integrity (can't be modified)
- Node identity (public key verification)

❌ **Does NOT Protect:**

- Real IP address from direct peers
- Metadata information (packet size, timing)
- Open services on node (need firewall)

### Security Recommendations:

#### 1. Use Firewall

```bash
# On Linux with ip6tables
ip6tables -A INPUT -i uqda0 -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
ip6tables -A INPUT -i uqda0 -j DROP
```

#### 2. Specify Allowed Peers

```json
{
  "AllowedPublicKeys": [
    "trusted-key-1",
    "trusted-key-2"
  ]
}
```

#### 3. Use TLS/QUIC

- ✅ `tls://` instead of `tcp://`
- ✅ `quic://` for best performance
- ✅ `wss://` instead of `ws://`

#### 4. Review Open Services

```bash
# Check open services on your IPv6 address
netstat -tulpn | grep uqda0
```

---

[Previous: Configuration ←](04-configuration.md) | [Next: Use Cases →](06-use-cases.md)

