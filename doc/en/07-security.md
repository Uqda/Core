# Security and Privacy - Uqda Network

## Security Overview

### What Uqda Protects

✅ **Protects:**

- **Data Content** - End-to-end encrypted
- **Data Integrity** - Cannot be modified
- **Node Identity** - Public key verification
- **Perfect Forward Secrecy** - Past sessions stay safe

❌ **Does NOT Protect:**

- **Real IP Address** - Direct peers can see it
- **Metadata** - Packet size, timing
- **Traffic Analysis** - Connection patterns observable
- **Open Services** - Need firewall

## Encryption

### Cryptographic Primitives Used

#### 1. Ed25519

- **Usage:** Digital signatures and public/private keys
- **Level:** Military-grade
- **Advantages:** Fast and secure

#### 2. Curve25519

- **Usage:** Key Exchange
- **Level:** Military-grade
- **Advantages:** Secure ECDH

#### 3. ChaCha20-Poly1305

- **Usage:** Data encryption
- **Level:** Military-grade
- **Advantages:** Fast and secure

#### 4. SHA-512

- **Usage:** Hashing
- **Level:** Secure
- **Advantages:** Collision-resistant

#### 5. HKDF

- **Usage:** Key derivation
- **Level:** Secure
- **Advantages:** Secure key derivation

### Encryption Process

```
1. Key Exchange (Curve25519 ECDH)
        ↓
2. Session Key Derivation (HKDF)
        ↓
3. Data Encryption (ChaCha20-Poly1305)
        ↓
4. Key Rotation (every time period)
```

## Security Best Practices

### 1. Use a Firewall

#### Linux (ip6tables)

```bash
# Allow established connections
ip6tables -A INPUT -i uqda0 -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT

# Drop everything else
ip6tables -A INPUT -i uqda0 -j DROP

# Save rules
ip6tables-save > /etc/ip6tables.rules
```

#### Linux (ufw)

```bash
# Block incoming on Uqda interface
sudo ufw deny in on uqda0 proto ipv6
```

#### FreeBSD (ipfw)

```bash
# Allow established connections
ipfw add allow ipv6-icmp from any to me via uqda0

# Drop everything else
ipfw add deny ipv6 from any to me via uqda0
```

### 2. Whitelist Trusted Peers

```json
{
  "AllowedPublicKeys": [
    "trusted-peer-1-public-key",
    "trusted-peer-2-public-key"
  ]
}
```

**Usage:** To create a private network - only these keys can connect to you

### 3. Use TLS/QUIC

**✅ Good:**

```json
{
  "Listen": ["tls://[::]:9001"],
  "Peers": ["tls://peer.example.com:9001"]
}
```

**❌ Bad:**

```json
{
  "Listen": ["tcp://[::]:9001"],
  "Peers": ["tcp://peer.example.com:9001"]
}
```

### 4. Protect Your Private Key

```bash
# File permissions
chmod 600 uqda.conf

# File ownership
chown root:root uqda.conf

# Or store in separate file
uqda -useconffile uqda.conf -exportkey > /etc/uqda/private.key
chmod 600 /etc/uqda/private.key
```

### 5. Use Passwords for Peers

```json
{
  "Peers": [
    "tls://peer.example.com:9001?password=strongPassword123"
  ]
}
```

**⚠️ Must be same password on both sides**

### 6. Pin Public Keys

```json
{
  "Peers": [
    "tls://peer.example.com:9001?key=expected-public-key"
  ]
}
```

**Usage:** To verify peer identity - connection fails if key doesn't match

### 7. Monitor Your Node

```bash
# Check active connections
uqdactl getPeers

# Check sessions
uqdactl getSessions

# Monitor logs
journalctl -u uqda -f
```

## Privacy

### ⚠️ Uqda is NOT an Anonymous Network

**What this means:**

- Direct peers can see your real IP address
- Not like Tor or I2P
- Goal is **security** not **anonymity**

### If You Need Anonymity

**Use Tor:**

```json
{
  "Peers": [
    "socks://127.0.0.1:9050/hidden.onion:9001"
  ]
}
```

**Tor Configuration:**

```
# /etc/tor/torrc
IsolateSOCKSAuth 1
```

### Protecting Metadata

**Visible Metadata:**

- Packet size
- Packet timing
- Connection patterns
- Number of connections

**For Protection:**

- Use Tor for connections
- Use traditional VPN
- Avoid connecting to untrusted peers

## Potential Threats

### 1. Man-in-the-Middle Attacks

**Protection:**

- ✅ Use TLS/QUIC
- ✅ Pin public keys
- ✅ Verify keys manually

### 2. DDoS Attacks

**Protection:**

- ✅ Use Firewall
- ✅ Specify AllowedPublicKeys
- ✅ Rate limiting (future)

### 3. Private Key Leakage

**Protection:**

- ✅ Protect file with proper permissions
- ✅ Don't share the key
- ✅ Encrypted backups

### 4. Traffic Analysis

**Protection:**

- ✅ Use Tor
- ✅ Additional encryption at application layer
- ✅ Avoid connecting to untrusted peers

## Security Recommendations by Scenario

### Home Network

```json
{
  "AllowedPublicKeys": [
    "device-1-key",
    "device-2-key",
    "device-3-key"
  ],
  "NodeInfoPrivacy": true
}
```

### Public Network

```json
{
  "NodeInfoPrivacy": false,
  "Listen": ["tls://[::]:9001"]
}
```

**⚠️ Be careful:** Connecting to public peers makes your network visible

### Emergency Network

```json
{
  "MulticastInterfaces": [
    {
      "Regex": ".*",
      "Beacon": true,
      "Listen": true,
      "Port": 9001,
      "Password": "emergency-password"
    }
  ]
}
```

## Security Auditing

### Current Status

- ⚠️ **No full external security audit yet**
- ✅ **Open source code** - Reviewable
- ✅ **Uses known cryptographic standards**
- ⚠️ **In Alpha stage** - Changes may occur

### Contributing to Security

- 🐛 Report vulnerabilities: Uqda@proton.me
- 🔍 Review code
- 🧪 Security testing
- 📝 Improve documentation

---

[Previous: Use Cases ←](06-use-cases.md) | [Next: Troubleshooting →](08-troubleshooting.md)

