# Use Cases - Uqda Network

## 1. Secure Home Network

### Scenario

Connect all your personal devices securely over the Internet.

```
🌐 Internet
    ↓
📡 Router (running Uqda)
    ↓
├── 💻 Laptop (200:1111::1)
├── 📱 Phone (200:2222::1)
├── 🖥️ Desktop (200:3333::1)
└── 📺 Smart TV (200:4444::1)
```

### Benefits

- ✅ All devices communicate directly
- ✅ Encrypted by default
- ✅ Works even behind NAT
- ✅ Access home network from anywhere

### Configuration

```json
{
  "Listen": ["tls://[::]:9001"],
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

## 2. Global Distributed Network

### Scenario

Connect friends and family worldwide.

```
🇸🇾 Syria ←→ 🇹🇷 Turkey ←→ 🇩🇪 Germany
    ↓                          ↓
🇯🇴 Jordan                 🇳🇱 Netherlands
```

### Benefits

- ✅ Direct connections, no middlemen
- ✅ Bypass censorship
- ✅ Low latency (shortest path routing)
- ✅ Resilient to regional outages

### Configuration

```json
{
  "Listen": ["tls://[::]:9001"],
  "Peers": [
    "tls://friend1.example.com:9001",
    "tls://friend2.example.com:9001",
    "tls://family.example.com:9001"
  ]
}
```

## 3. Community Network

### Scenario

Neighborhood or local community network.

```
🏠 House 1 ←→ 🏢 Apartment ←→ 🏠 House 2
    ↓            ↓              ↓
🏠 House 3 ←→ 🏪 Store ←→ 🏫 School
```

### Benefits

- ✅ Share internet connection
- ✅ Local services (file sharing, games)
- ✅ Emergency communication
- ✅ Community independence

### Configuration

```json
{
  "Listen": ["tls://[::]:9001"],
  "Peers": [
    "tls://neighbor1.local:9001",
    "tls://neighbor2.local:9001",
    "tls://community-gateway.local:9001"
  ],
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

## 4. Emergency Network

### Scenario

When Internet is down or during disasters.

```
📱 Phone 1 ←→ 💻 Laptop ←→ 📱 Phone 2
  (WiFi)      (Ethernet)    (WiFi)
```

### Benefits

- ✅ Works without internet
- ✅ Instant network setup
- ✅ No infrastructure needed
- ✅ Critical for emergencies

### Configuration

```json
{
  "Listen": ["tls://[::]:9001"],
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

## 5. Secure IoT Devices

### Scenario

Connect IoT devices securely.

```
📹 Camera ←→ 🏠 Hub ←→ 💡 Lights
    ↓          ↓         ↓
🌡️ Sensors ←→ 🔐 Lock ←→ 🔊 Speaker
```

### Benefits

- ✅ No cloud dependency
- ✅ End-to-end encryption
- ✅ Local processing
- ✅ Privacy preserved

### Configuration

```json
{
  "Listen": ["tls://[::]:9001"],
  "Peers": [
    "tls://hub.local:9001"
  ],
  "AllowedPublicKeys": [
    "camera-key",
    "sensor-key",
    "lock-key"
  ]
}
```

## 6. Personal VPN

### Scenario

Alternative to Tailscale/Zerotier.

```
💻 Laptop ←→ 🌐 Internet ←→ 🏠 Home Server
```

### Benefits

- ✅ No central servers
- ✅ Encrypted by default
- ✅ Self-healing
- ✅ Completely free

### Configuration

**On Home Server:**

```json
{
  "Listen": ["tls://[::]:9001"],
  "Peers": []
}
```

**On Laptop:**

```json
{
  "Listen": ["tls://[::]:9001"],
  "Peers": [
    "tls://home-server.example.com:9001"
  ]
}
```

## 7. Censorship Circumvention

### Scenario

Access restricted content.

```
🇸🇾 Syria ←→ 🇹🇷 Turkey ←→ 🌐 Open Internet
```

### Benefits

- ✅ Bypass restrictions
- ✅ Decentralized design
- ✅ Hard to block

### Configuration

```json
{
  "Listen": ["tls://[::]:9001"],
  "Peers": [
    "tls://free-peer1.example.com:9001",
    "tls://free-peer2.example.com:9001",
    "quic://free-peer3.example.com:9002"
  ]
}
```

## 8. Edge Computing

### Scenario

Direct connection between devices.

```
🖥️ Server 1 ←→ 🖥️ Server 2 ←→ 🖥️ Server 3
```

### Benefits

- ✅ Direct connection between devices
- ✅ No need for central servers
- ✅ Reduce latency

### Configuration

```json
{
  "Listen": ["quic://[::]:9002"],
  "Peers": [
    "quic://edge-server1.example.com:9002",
    "quic://edge-server2.example.com:9002"
  ]
}
```

## 9. Development Network

### Scenario

Connect development environments.

```
💻 Developer 1 ←→ 🖥️ Dev Server ←→ 💻 Developer 2
```

### Benefits

- ✅ Secure connection between developers
- ✅ Share resources
- ✅ Test network

### Configuration

```json
{
  "Listen": ["tls://[::]:9001"],
  "Peers": [
    "tls://dev-server.example.com:9001"
  ],
  "AllowedPublicKeys": [
    "dev-key-1",
    "dev-key-2"
  ]
}
```

## 10. Gaming Network

### Scenario

Low-latency network for gaming.

```
🎮 Player 1 ←→ 🎮 Player 2 ←→ 🎮 Player 3
```

### Benefits

- ✅ Low latency
- ✅ Direct connection
- ✅ No game servers

### Configuration

```json
{
  "Listen": ["quic://[::]:9002"],
  "Peers": [
    "quic://player1.example.com:9002",
    "quic://player2.example.com:9002"
  ]
}
```

---

[Previous: Technical Concepts ←](05-concepts.md) | [Next: Security →](07-security.md)

