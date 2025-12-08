# Quick Start - Uqda Network

## Get Started in 5 Minutes

### Step 1: Generate Configuration

```bash
# Generate a simple config file
uqda -genconf > uqda.conf
```

### Step 2: Edit Configuration (Optional)

```bash
# Open file for editing
nano uqda.conf
```

**Add peers:**

```json
{
  "Peers": [
    "tls://peer1.example.com:9001",
    "tls://peer2.example.com:9001"
  ]
}
```

### Step 3: Run Node

```bash
# Run with config file
sudo uqda -useconffile uqda.conf
```

**Or auto-configuration mode (for testing):**

```bash
# Will generate random keys and run automatically
sudo uqda -autoconf
```

### Step 4: Check Status

**In another terminal:**

```bash
# Show your node info
uqdactl getSelf

# Show connected peers
uqdactl getPeers

# Show your IPv6 address
uqdactl getSelf | grep "IPv6 address"
```

## Complete Practical Example

### 1. Generate Configuration

```bash
uqda -genconf > /tmp/uqda.conf
```

### 2. View Configuration

```bash
cat /tmp/uqda.conf
```

**Example output:**

```hjson
{
  // Private key (keep secret!)
  PrivateKey: abc123def456...
  
  // Listen addresses
  Listen: [
    "tls://[::]:9001"
  ]
  
  // Peers to connect to
  Peers: []
  
  // Interface name
  IfName: auto
}
```

### 3. Add Public Peer

```bash
# Edit file
nano /tmp/uqda.conf
```

**Add peer:**

```json
{
  "Peers": [
    "tls://public-peer.example.com:9001"
  ]
}
```

### 4. Run Node

```bash
sudo uqda -useconffile /tmp/uqda.conf
```

### 5. Test Connection

```bash
# In another terminal
# Get IPv6 address
MY_IP=$(uqdactl getSelf | grep "IPv6 address" | awk '{print $3}')

# Test connection
ping6 $MY_IP
```

## Basic Commands

### Show Node Information

```bash
uqdactl getSelf
```

**Output:**

```
Build name:        Uqda
Build version:     0.1.2
IPv6 address:      200:1234:5678:9abc::1
IPv6 subnet:       300:1234:5678:9abc::/64
Routing table:     42 entries
Public key:        abc123def456...
```

### Show Connected Peers

```bash
uqdactl getPeers
```

**Output:**

```
URI                          State  Dir  IP Address              Uptime    RTT
tls://peer1.com:9001         Up     Out  200:1111::1            5m30s     12ms
tls://peer2.com:9001         Up     In   200:2222::1            2m15s     8ms
```

### Show Active Sessions

```bash
uqdactl getSessions
```

### Show Routing Table

```bash
uqdactl getPaths
```

## Usage Scenarios

### Scenario 1: Connect Two Devices on Same Network

**On First Device:**

```bash
# Generate config
uqda -genconf > uqda1.conf

# Edit config to add listen address
nano uqda1.conf
# Add: "Listen": ["tls://[::]:9001"]

# Run
sudo uqda -useconffile uqda1.conf

# Show public key
uqdactl getSelf | grep "Public key"
```

**On Second Device:**

```bash
# Generate config
uqda -genconf > uqda2.conf

# Add first device as peer
nano uqda2.conf
# Add in Peers:
# "tls://[FIRST-DEVICE-IP]:9001"

# Run
sudo uqda -useconffile uqda2.conf
```

### Scenario 2: Connect to Public Network

```bash
# Generate config
uqda -genconf > uqda.conf

# Add public peers
nano uqda.conf
```

**Add peers:**

```json
{
  "Peers": [
    "tls://public-peer1.example.com:9001",
    "tls://public-peer2.example.com:9001",
    "quic://public-peer3.example.com:9002"
  ]
}
```

```bash
# Run
sudo uqda -useconffile uqda.conf
```

### Scenario 3: Automatic Discovery on Local Network

```bash
# Generate config
uqda -genconf > uqda.conf

# Edit config
nano uqda.conf
```

**Add multicast:**

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

```bash
# Run
sudo uqda -useconffile uqda.conf

# Will automatically discover other devices!
```

## Troubleshooting

### Issue: No Connections

```bash
# Check peers
uqdactl getPeers

# If empty, try:
# 1. Check config
cat uqda.conf | grep -i peer

# 2. Check internet connection
ping peer.example.com

# 3. Check logs
journalctl -u uqda -f
```

### Issue: "Connection refused"

**Possible causes:**

1. Peer is not available
2. Firewall blocking connection
3. Wrong address

**Solution:**

```bash
# Test connection
telnet peer.example.com 9001

# Or
nc -zv peer.example.com 9001
```

### Issue: "Cannot bind to address"

**Solution:**

```bash
# Check used ports
sudo netstat -tulpn | grep 9001

# Change port in config
# "Listen": ["tls://[::]:9002"]
```

---

[Previous: Installation ←](02-installation.md) | [Next: Configuration →](04-configuration.md)

