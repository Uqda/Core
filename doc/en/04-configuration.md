# Configuration Guide - Uqda Network

## Basic Configuration File

### Generate Configuration

```bash
# Generate config in HJSON format (with comments)
uqda -genconf > uqda.conf

# Or in JSON format (for programmatic manipulation)
uqda -genconf -json > uqda.conf
```

### Configuration File Structure

```json
{
  "PrivateKey": "...",
  "Listen": [],
  "Peers": [],
  "InterfacePeers": {},
  "MulticastInterfaces": [],
  "AllowedPublicKeys": [],
  "IfName": "auto",
  "IfMTU": 65535,
  "NodeInfo": {},
  "NodeInfoPrivacy": false
}
```

## Basic Settings

### PrivateKey

**⚠️ Very Important:** Keep this key secret!

```json
{
  "PrivateKey": "abc123def456..."
}
```

**Generate new key:**

```bash
go run ./cmd/genkeys
```

### Listen (Listen Addresses)

**Define ports your node listens on:**

```json
{
  "Listen": [
    "tls://[::]:9001",
    "quic://[::]:9002"
  ]
}
```

**Supported connection types:**

- `tcp://` - Plain TCP (unencrypted)
- `tls://` - TCP with TLS (recommended)
- `quic://` - QUIC (fastest)
- `ws://` - WebSocket
- `wss://` - WebSocket Secure
- `unix://` - UNIX socket (local only)

### Peers

**List of peers to connect to:**

```json
{
  "Peers": [
    "tls://peer1.example.com:9001",
    "tls://peer2.example.com:9001",
    "quic://peer3.example.com:9002"
  ]
}
```

## Advanced Configuration

### Peering with Password

```json
{
  "Peers": [
    "tls://peer.example.com:9001?password=mySecretPassword123"
  ]
}
```

**⚠️ Must be same password on both sides**

### Pin Public Key for Peer

```json
{
  "Peers": [
    "tls://peer.example.com:9001?key=expected-public-key-here"
  ]
}
```

**Usage:** To verify peer identity

### Peering via SOCKS Proxy

```json
{
  "Peers": [
    "socks://127.0.0.1:9050/hidden.onion:9001",
    "socks://user:pass@proxy.example.com:1080/peer.example.com:9001"
  ]
}
```

**Useful for connecting through Tor or other proxy**

### InterfacePeers (Peers on Specific Interfaces)

**Connect peers on specific network interfaces:**

```json
{
  "InterfacePeers": {
    "eth0": [
      "tls://local-peer.lan:9001"
    ],
    "wlan0": [
      "tls://wifi-peer.lan:9001"
    ]
  }
}
```

**Usage:** To control which interface to use for which peer

### MulticastInterfaces (Automatic Discovery)

**Enable automatic discovery on local network:**

```json
{
  "MulticastInterfaces": [
    {
      "Regex": "eth.*",
      "Beacon": true,
      "Listen": true,
      "Port": 9001,
      "Priority": 1,
      "Password": ""
    }
  ]
}
```

**Parameters:**

- `Regex`: Interface name pattern (e.g., `eth.*` or `wlan0`)
- `Beacon`: Send discovery beacons (true/false)
- `Listen`: Listen for others' discoveries (true/false)
- `Port`: Port to use
- `Priority`: Priority (lower = higher priority)
- `Password`: Optional password

### AllowedPublicKeys (Allowed Peers)

**Restrict incoming connections:**

```json
{
  "AllowedPublicKeys": [
    "0abc123def456...",
    "789xyz123abc..."
  ]
}
```

**Usage:** To create a private network - only these public keys can connect to you

### IfName (Interface Name)

**Name the TUN/TAP interface:**

```json
{
  "IfName": "uqda0"
}
```

**Values:**

- `auto` - Automatic selection
- `uqda0` - Specific name
- `tun0` - Use existing interface

### IfMTU (Maximum Transmission Unit)

```json
{
  "IfMTU": 65535
}
```

**Recommended values:**

- `1500` - For normal networks
- `9000` - For high-performance networks (Jumbo frames)
- `65535` - Maximum

### NodeInfo (Node Information)

**General information about your node:**

```json
{
  "NodeInfo": {
    "name": "My Uqda Node",
    "location": "Damascus, Syria",
    "description": "Personal node"
  }
}
```

**⚠️ This information is visible to everyone on the network**

### NodeInfoPrivacy (Node Info Privacy)

```json
{
  "NodeInfoPrivacy": false
}
```

**Values:**

- `false` - Share node information
- `true` - Hide node information

## Configuration Examples

### Example 1: Simple Node

```json
{
  "Listen": [
    "tls://[::]:9001"
  ],
  "Peers": [
    "tls://public-peer.example.com:9001"
  ]
}
```

### Example 2: Node with Auto-Discovery

```json
{
  "Listen": [
    "tls://[::]:9001"
  ],
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

### Example 3: Private Network

```json
{
  "Listen": [
    "tls://[::]:9001"
  ],
  "Peers": [
    "tls://trusted-peer1.example.com:9001",
    "tls://trusted-peer2.example.com:9001"
  ],
  "AllowedPublicKeys": [
    "key1...",
    "key2..."
  ]
}
```

### Example 4: Multi-Interface Node

```json
{
  "Listen": [
    "tls://[::]:9001"
  ],
  "InterfacePeers": {
    "eth0": [
      "tls://wired-peer.lan:9001"
    ],
    "wlan0": [
      "tls://wifi-peer.lan:9001"
    ]
  },
  "MulticastInterfaces": [
    {
      "Regex": "wlan.*",
      "Beacon": true,
      "Listen": true,
      "Port": 9001
    }
  ]
}
```

## Configuration Management

### Normalize Configuration

```bash
# Normalize config (remove comments, order fields)
uqda -useconffile uqda.conf -normaliseconf > uqda-normalized.conf
```

### Export Private Key

```bash
# Export key in PEM format
uqda -useconffile uqda.conf -exportkey > private-key.pem
```

### Show Address

```bash
# Show your IPv6 address
uqda -useconffile uqda.conf -address
```

### Show Subnet

```bash
# Show your subnet (/64)
uqda -useconffile uqda.conf -subnet
```

### Show Public Key

```bash
# Show public key
uqda -useconffile uqda.conf -publickey
```

## Best Practices

### 1. Use TLS/QUIC

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

### 2. Protect Private Key

```bash
# File permissions
chmod 600 uqda.conf

# File ownership
chown root:root uqda.conf
```

### 3. Use AllowedPublicKeys for Private Networks

```json
{
  "AllowedPublicKeys": [
    "trusted-key-1",
    "trusted-key-2"
  ]
}
```

### 4. Monitor Logs

```bash
# On Linux
journalctl -u uqda -f

# Or
tail -f /var/log/uqda.log
```

---

[Previous: Quick Start ←](03-quickstart.md) | [Next: Technical Concepts →](05-concepts.md)

