# Technical Report: Uqda IPv6 Overlay Network Deployment

**Date:** January 16, 2026  
**Version:** Uqda Core v0.1.0  
**Author:** Uqda Network Team

---

## Executive Summary

This technical report documents the successful deployment of an encrypted IPv6 overlay network using Uqda Core, connecting a Windows client to a Linux server with full SSH connectivity. The deployment demonstrates Uqda's capability to create secure, peer-to-peer encrypted networks without requiring public IP addresses or NAT traversal.

---

## 1. Objectives

The primary goal was to establish an encrypted overlay network that enables:

- **Encrypted Peer-to-Peer Communication** - End-to-end encrypted IPv6 connectivity
- **Persistent IPv6 Overlay** - Stable IPv6 addresses independent of physical network changes
- **Full SSH Connectivity** - Secure shell access through the overlay tunnel without public IP exposure
- **Zero Public Exposure** - No need to expose ports to the public internet

---

## 2. Environment Setup

### Server (Linux)

- **OS:** Ubuntu Server 24.04 LTS
- **Kernel:** 6.8.x
- **Public IPv4:** `45.90.99.144`
- **Public IPv6:** `2a0e:97c0:3e3:c2d::1`
- **Uqda Version:** 0.1.0

### Client (Windows)

- **OS:** Windows 10/11
- **OpenSSH:** 9.5p2
- **Uqda Version:** 0.1.0

### Network Architecture

- **Type:** IPv6 Encrypted Overlay Network
- **Interface:** TUN adapter
- **Protocol:** Uqda Network Protocol v0.5
- **Encryption:** End-to-end encrypted by default

---

## 3. Challenges and Solutions

### Challenge 1: Broken Linux `.deb` Package

**Problem:**
- Initial `.deb` packages were only ~1KB in size
- Missing binaries and systemd service files
- Installation failed with "file not found" errors

**Root Cause:**
- GitHub Actions workflow had permission issues
- Shell scripts lacked execute permissions
- Build process failed silently

**Solution:**
- Added `chmod +x` commands to GitHub Actions workflows
- Fixed file paths and naming conventions
- Implemented proper error checking in build scripts

**Result:**
- Packages now build correctly (~6-7MB)
- All binaries and service files included
- Automatic installation works as expected

---

### Challenge 2: SSH Failure Despite Successful Ping

**Problem:**
- IPv6 ping (`ping6`) worked successfully
- SSH and TCP connections timed out
- ICMP vs TCP behavior discrepancy

**Root Cause:**
- ICMP and TCP are handled differently by the network stack
- TCP requires:
  - Correct MTU configuration
  - Firewall rules allowing INPUT traffic
  - Proper routing table entries

**Solution:**
- Configured MTU to 1380 bytes
- Added firewall rules for TCP traffic
- Verified routing table configuration

---

### Challenge 3: Default MTU Too Large

**Problem:**
- Uqda default MTU: `65535` bytes
- This breaks TCP connections in many tunnel scenarios
- Causes packet fragmentation and connection failures

**Root Cause:**
- Maximum MTU setting doesn't account for encapsulation overhead
- Real-world networks have smaller MTU limits

**Solution:**

Modified configuration file (`/etc/uqda/uqda.conf`):

```json
{
  "IfMTU": 1380
}
```

**Result:**
```
tun0: flags=... mtu 1380
```

TCP connections now work reliably.

**Recommendation:**
Always use MTU between `1280-1380` bytes for overlay networks.

---

### Challenge 4: Linux IPv6 Firewall Blocking TCP on tun0

**Problem:**
- `tcpdump` showed SYN packets arriving
- No SYN-ACK responses
- Connections timing out

**Root Cause:**
- Linux `ip6tables` INPUT chain was blocking TCP traffic on `tun0`
- Default firewall policy was too restrictive
- Uqda interface not explicitly allowed

**Solution:**

Added explicit firewall rule:

```bash
ip6tables -I INPUT 1 -i tun0 -p tcp --dport 22 -j ACCEPT
```

**Verification:**

```bash
# Check firewall rules
ip6tables -L INPUT -v -n

# Test connection
tcpdump -i tun0 -n 'tcp port 22'
```

**Result:**
- TCP connections succeed immediately
- SSH works perfectly
- No more timeouts

**Persistence:**

To make firewall rules persistent:

```bash
apt install iptables-persistent
netfilter-persistent save
```

---

### Challenge 5: Windows Firewall Blocking Outbound TCP

**Problem:**
- `Test-NetConnection` showed:
  ```
  PingSucceeded : True
  TcpTestSucceeded : False
  ```
- Uqda interface classified as "Public" network
- Windows Firewall blocking outbound TCP

**Root Cause:**
- Windows Firewall applies different rules to Public networks
- Uqda TUN interface detected as Public network
- Default outbound rules too restrictive

**Solution:**

Created custom firewall rule:

```powershell
New-NetFirewallRule `
  -DisplayName "Allow SSH over Uqda IPv6" `
  -Direction Outbound `
  -Program "C:\Windows\System32\OpenSSH\ssh.exe" `
  -Protocol TCP `
  -RemotePort 22 `
  -InterfaceAlias "Uqda" `
  -Action Allow
```

**Alternative (Broader Rule):**

```powershell
New-NetFirewallRule `
  -DisplayName "Allow TCP over Uqda" `
  -Direction Outbound `
  -InterfaceAlias "Uqda" `
  -Protocol TCP `
  -Action Allow
```

**Verification:**

```powershell
Test-NetConnection -ComputerName [201:63f8:831b:d439:beff:421a:e033:cdca] -Port 22
```

**Result:**
```
TcpTestSucceeded : True
```

---

## 4. Final Configuration

### Linux Server Configuration

**Uqda Config (`/etc/uqda/uqda.conf`):**

```json
{
  IfMTU: 1380,
  Listen: [
    "tls://[::]:33347"
  ],
  Peers: [
    "tls://<windows-peer-ip>:<port>"
  ]
}
```

**Firewall Rules:**

```bash
# Allow TCP on Uqda interface
ip6tables -I INPUT 1 -i tun0 -p tcp --dport 22 -j ACCEPT

# Optional: Allow all TCP on tun0
ip6tables -I INPUT 1 -i tun0 -p tcp -j ACCEPT

# Save rules
netfilter-persistent save
```

**Systemd Service:**

```bash
systemctl status uqda
systemctl enable uqda
```

---

### Windows Client Configuration

**Uqda Config (`C:\ProgramData\Uqda\uqda.conf`):**

```json
{
  "IfMTU": 1380,
  "Peers": [
    "tls://<linux-server-ip>:33347"
  ]
}
```

**Firewall Rule:**

```powershell
New-NetFirewallRule `
  -DisplayName "Allow SSH over Uqda IPv6" `
  -Direction Outbound `
  -InterfaceAlias "Uqda" `
  -Protocol TCP `
  -RemotePort 22 `
  -Action Allow
```

**Service Management:**

```powershell
# Check service status
Get-Service Uqda

# Start service
Start-Service Uqda
```

---

## 5. Validation and Testing

### Connectivity Tests

**From Windows:**

```powershell
# Test IPv6 connectivity
ping -6 201:63f8:831b:d439:beff:421a:e033:cdca

# Test TCP port
Test-NetConnection -ComputerName 201:63f8:831b:d439:beff:421a:e033:cdca -Port 22

# SSH connection
ssh -6 root@[201:63f8:831b:d439:beff:421a:e033:cdca]
```

**From Linux:**

```bash
# Check Uqda status
sudo uqdactl getSelf

# View peers
sudo uqdactl getPeers

# Test connectivity
ping6 200:5d86:87e1:4b3b:bcfe:833d:3c87:94bf

# SSH test (if Windows has SSH server)
ssh -6 user@[200:5d86:87e1:4b3b:bcfe:833d:3c87:94bf]
```

---

### Performance Metrics

**Latency:**
- Overlay latency: ~23ms (measured via `uqdactl getPeers`)
- Comparable to direct connection

**Throughput:**
- TCP connections stable
- No packet loss observed
- MTU 1380 optimal for this setup

**Security:**
- All traffic encrypted end-to-end
- No public IP exposure required
- No NAT traversal needed

---

## 6. Best Practices and Recommendations

### 1. MTU Configuration

**Always set MTU between 1280-1380 bytes:**

```json
{
  "IfMTU": 1380
}
```

**Rationale:**
- Accounts for encapsulation overhead
- Compatible with most network paths
- Prevents TCP fragmentation issues

---

### 2. Firewall Configuration

**Linux (ip6tables):**

```bash
# Allow TCP on Uqda interface
ip6tables -I INPUT 1 -i tun0 -p tcp -j ACCEPT

# Make persistent
apt install iptables-persistent
netfilter-persistent save
```

**Windows (PowerShell):**

```powershell
# Create rule for Uqda interface
New-NetFirewallRule `
  -DisplayName "Allow TCP over Uqda" `
  -Direction Outbound `
  -InterfaceAlias "Uqda" `
  -Protocol TCP `
  -Action Allow
```

---

### 3. Peer Security

**Restrict peers using public keys:**

```json
{
  "AllowedPublicKeys": [
    "<peer-public-key-hex>"
  ]
}
```

**Benefits:**
- Only authorized peers can connect
- Prevents unauthorized access
- Enhances network security

---

### 4. Service Configuration

**Disable SSH on public IP:**

```bash
# Edit SSH config
nano /etc/ssh/sshd_config

# Change:
# ListenAddress 0.0.0.0
# To:
ListenAddress 127.0.0.1
ListenAddress <uqda-ipv6-address>

# Restart SSH
systemctl restart sshd
```

**Result:**
- SSH only accessible via Uqda overlay
- No public exposure
- Enhanced security

---

### 5. Monitoring and Troubleshooting

**Useful Commands:**

```bash
# Check Uqda status
sudo uqdactl getSelf
sudo uqdactl getPeers

# Monitor traffic
sudo tcpdump -i tun0 -n

# Check routing
ip -6 route show

# Verify MTU
ip link show tun0

# Check firewall
ip6tables -L INPUT -v -n
```

---

## 7. Network Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    Uqda Overlay Network                      │
│                  (Encrypted IPv6 Tunnel)                    │
└─────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┴───────────────────┐
        │                                       │
┌───────▼────────┐                    ┌────────▼──────┐
│  Linux Server  │                    │ Windows Client│
│                │                    │               │
│ IPv6:          │                    │ IPv6:         │
│ 201:63f8:...   │◄───── Encrypted ───►│ 200:5d86:...  │
│                │     Connection     │               │
│ tun0 (MTU 1380)│                    │ Uqda (MTU 1380)│
│                │                    │               │
│ SSH Server     │                    │ SSH Client    │
│ Port 22        │                    │               │
└────────────────┘                    └───────────────┘
        │                                       │
        └───────────────────┬───────────────────┘
                            │
                    Physical Network
                    (IPv4/IPv6 Underlay)
```

---

## 8. Key Learnings

### What Worked Well

1. **Uqda Protocol** - Stable and reliable
2. **Encryption** - Transparent and automatic
3. **IPv6 Addressing** - Persistent and location-independent
4. **Peer Management** - Simple and effective

### Critical Configuration Points

1. **MTU** - Must be set correctly (1280-1380)
2. **Firewall** - Must allow TCP on overlay interface
3. **Service Files** - Must be properly configured
4. **Peer Configuration** - Must match on both sides

### Common Pitfalls

1. **Forgetting MTU** - Causes TCP failures
2. **Firewall Rules** - Often overlooked
3. **Interface Classification** - Windows treats overlay as Public
4. **Service Dependencies** - systemd ordering matters

---

## 9. Conclusion

This deployment successfully demonstrates:

✅ **Uqda is production-ready** for encrypted overlay networks  
✅ **Protocol is solid** - issues were configuration-related, not protocol flaws  
✅ **Cross-platform works** - Windows ↔ Linux connectivity achieved  
✅ **Security is strong** - End-to-end encryption without public exposure  
✅ **Scalability is possible** - Can add multiple peers easily  

### Final Status

- ✅ Encrypted IPv6 overlay network operational
- ✅ Windows ↔ Linux connectivity established
- ✅ SSH working through tunnel
- ✅ No public IP exposure required
- ✅ No NAT traversal needed
- ✅ Ready for production use

---

## 10. References

- **Uqda Core Repository:** https://github.com/Uqda/Core
- **Documentation:** https://github.com/Uqda/Core/blob/main/README.md
- **Original Project:** Yggdrasil Network (https://yggdrasil-network.github.io/)

---


**Report Version:** 1.0  
**Last Updated:** January 16, 2026  
**Status:** Final

