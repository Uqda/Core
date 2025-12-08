# Troubleshooting - Uqda Network

## Common Issues and Solutions

### 1. No Connections

#### Symptoms

```bash
uqdactl getPeers
# Output: (empty)
```

#### Solutions

**Check Configuration:**

```bash
# Check peers in config
cat uqda.conf | grep -i peer

# Check listen addresses
cat uqda.conf | grep -i listen
```

**Check Internet Connection:**

```bash
# Test connection
ping peer.example.com

# Test port
nc -zv peer.example.com 9001
```

**Check Logs:**

```bash
# On Linux
journalctl -u uqda -f

# Or
tail -f /var/log/uqda.log
```

**Restart:**

```bash
sudo systemctl restart uqda
```

### 2. "Connection refused"

#### Symptoms

```
Error: Connection refused
```

#### Possible Causes

1. Peer is not available
2. Firewall blocking connection
3. Wrong address

#### Solutions

**Test Connection:**

```bash
# Test TCP
telnet peer.example.com 9001

# Or
nc -zv peer.example.com 9001
```

**Check Firewall:**

```bash
# On Linux
sudo iptables -L
sudo ip6tables -L

# Check open ports
sudo netstat -tulpn | grep 9001
```

**Check Configuration:**

```bash
# Verify address is correct
cat uqda.conf | grep peer
```

### 3. "Cannot bind to address"

#### Symptoms

```
Error: Cannot bind to address [::]:9001
```

#### Solutions

**Check Used Ports:**

```bash
sudo netstat -tulpn | grep 9001
```

**Change Port:**

```json
{
  "Listen": ["tls://[::]:9002"]
}
```

**Stop Other Service:**

```bash
# Find process
sudo lsof -i :9001

# Stop it
sudo kill <PID>
```

### 4. "Permission denied" when creating TUN

#### Symptoms

```
Error: Permission denied
Cannot create TUN interface
```

#### Solutions

**On Linux:**

```bash
# Grant CAP_NET_ADMIN capability
sudo setcap cap_net_admin+eip /usr/local/bin/uqda

# Or run as root
sudo uqda -useconffile uqda.conf
```

**Load TUN Module:**

```bash
sudo modprobe tun
```

**Check TUN Exists:**

```bash
ls /dev/net/tun
```

### 5. "Command not found"

#### Symptoms

```
bash: uqda: command not found
```

#### Solutions

**Check Path:**

```bash
which uqda
which uqdactl
```

**Add to PATH:**

```bash
export PATH=$PATH:/usr/local/bin

# Or add to ~/.bashrc
echo 'export PATH=$PATH:/usr/local/bin' >> ~/.bashrc
source ~/.bashrc
```

**Copy Files:**

```bash
sudo cp uqda /usr/local/bin/
sudo cp uqdactl /usr/local/bin/
```

### 6. Connections Keep Dropping

#### Symptoms

```
Peers connect then disconnect repeatedly
```

#### Solutions

**Check Connection Stability:**

```bash
# Monitor peers
watch -n 1 'uqdactl getPeers'
```

**Check Logs:**

```bash
journalctl -u uqda -f | grep -i error
```

**Increase Timeout:**

```json
{
  "Peers": [
    "tls://peer.example.com:9001?timeout=30s"
  ]
}
```

**Check Firewall:**

```bash
# Firewall may be dropping connections
sudo iptables -L -v
```

### 7. Slow Performance

#### Symptoms

```
Slow data transfer
High latency
```

#### Solutions

**Use QUIC:**

```json
{
  "Listen": ["quic://[::]:9002"],
  "Peers": [
    "quic://peer.example.com:9002"
  ]
}
```

**Check Routes:**

```bash
# Show routes
uqdactl getPaths

# Choose shortest path
```

**Increase MTU:**

```json
{
  "IfMTU": 9000
}
```

**Check Resources:**

```bash
# CPU usage
top -p $(pgrep uqda)

# Memory usage
ps aux | grep uqda
```

### 8. IPv6 Address Not Working

#### Symptoms

```
Cannot ping IPv6 address
Cannot access services
```

#### Solutions

**Check Address:**

```bash
uqdactl getSelf | grep "IPv6 address"
```

**Test Connection:**

```bash
# Ping yourself
ping6 $(uqdactl getSelf | grep "IPv6 address" | awk '{print $3}')
```

**Check TUN Interface:**

```bash
# Show interface
ip addr show uqda0

# Or
ifconfig uqda0
```

**Enable IPv6:**

```bash
# On Linux
sudo sysctl -w net.ipv6.conf.all.disable_ipv6=0
```

### 9. Multicast Discovery Not Working

#### Symptoms

```
No automatic peer discovery
```

#### Solutions

**Check Configuration:**

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

**Check Interfaces:**

```bash
# Show available interfaces
ip link show

# Or
ifconfig
```

**Check Permissions:**

```bash
# May need special permissions for multicast
sudo setcap cap_net_raw+eip /usr/local/bin/uqda
```

### 10. Issues on Windows

#### Symptoms

```
TUN interface creation fails
```

#### Solutions

**Install TAP-Windows:**

```powershell
# Download from openvpn.org
# Install TAP-Windows adapter
```

**Run as Administrator:**

```powershell
# Run PowerShell as administrator
# Then run uqda
```

**Check Wintun:**

```powershell
# May need Wintun driver
# Download from wintun.net
```

## Diagnostic Tools

### 1. Show Node Information

```bash
uqdactl getSelf
```

### 2. Show Peers

```bash
uqdactl getPeers
```

### 3. Show Sessions

```bash
uqdactl getSessions
```

### 4. Show Routes

```bash
uqdactl getPaths
```

### 5. Show Tree

```bash
uqdactl getTree
```

### 6. Test Connection

```bash
# ping
ping6 <ipv6-address>

# traceroute
traceroute6 <ipv6-address>
```

### 7. Monitor Logs

```bash
# Linux systemd
journalctl -u uqda -f

# Linux syslog
tail -f /var/log/syslog | grep uqda

# macOS
log stream --predicate 'process == "uqda"'
```

### 8. Network Inspection

```bash
# Show interfaces
ip addr show

# Show routing tables
ip -6 route show

# Show connections
netstat -tulpn | grep uqda
```

## Getting Help

### 1. GitHub Issues

https://github.com/Uqda/Core/issues

### 2. Email

Uqda@proton.me

### 3. Logs

When reporting an issue, include:

- Full logs
- Configuration (without private key!)
- System information
- Steps to reproduce

---

[Previous: Security ←](07-security.md) | [Next: Development →](09-development.md)

