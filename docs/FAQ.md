# Frequently Asked Questions (FAQ)

## General

### What is Uqda?

Uqda Network is a decentralized mesh networking protocol optimized for real-world performance. It enables direct peer-to-peer connectivity with end-to-end encryption, minimal configuration, and independence from centralized infrastructure.

### How is Uqda different from Yggdrasil?

Uqda is based on Yggdrasil Network v0.5 with significant performance optimizations:

- **Faster connection establishment** (20-40ms vs 50-100ms)
- **Reduced timeouts** for better responsiveness
- **DNS caching** to minimize lookup overhead
- **Adaptive backoff algorithm** for faster reconnection

Uqda nodes are **100% protocol-compatible** with Yggdrasil v0.5 nodes and can peer with them seamlessly. See [../ATTRIBUTION.md](../ATTRIBUTION.md) for details.

### Is Uqda production-ready?

**Yes, for non-critical workloads.** Uqda is reliable for:
- Community networks
- Edge computing scenarios
- Development and research
- Experimental deployments

**Not recommended** for:
- Mission-critical systems
- Life-safety applications
- High-security environments (without independent audit)

### What does "Uqda" mean?

"Uqda" (عُقدة) is Arabic for "node" or "knot", representing the interconnected nature of mesh networks.

---

## Technical

### Can Uqda nodes peer with Yggdrasil nodes?

**Yes!** Uqda is protocol-compatible with Yggdrasil v0.5. Nodes from both networks can peer and communicate seamlessly.

### Does Uqda provide anonymity?

**No.** Uqda does not provide anonymity:

- Direct peers can see your IP address
- Traffic patterns may be observable
- No protection against traffic analysis

If you need anonymity, use **Tor** or **I2P** instead.

### How do I configure a firewall?

Uqda nodes are globally reachable by default. **Strongly recommended**: Configure IPv6 firewall rules.

**Linux (iptables/ip6tables):**
```bash
# Block all incoming on Uqda interface
ip6tables -A INPUT -i uqda0 -j DROP

# Allow established connections
ip6tables -A INPUT -i uqda0 -m state --state ESTABLISHED,RELATED -j ACCEPT

# Allow specific services (example: SSH)
ip6tables -A INPUT -i uqda0 -p tcp --dport 22 -j ACCEPT
```

**Windows (PowerShell):**
```powershell
# Block all inbound on Uqda interface
New-NetFirewallRule -DisplayName "Block Uqda Inbound" `
    -Direction Inbound -InterfaceAlias "Uqda" -Action Block

# Allow specific service (example: SSH)
New-NetFirewallRule -DisplayName "Allow SSH on Uqda" `
    -Direction Inbound -InterfaceAlias "Uqda" -Protocol TCP -LocalPort 22 -Action Allow
```

**macOS (pfctl):**
```bash
# Add to /etc/pf.conf:
block in on uqda0 all
pass in on uqda0 proto tcp from any to any port 22
```

See [../SECURITY.md](../SECURITY.md) for more details.

### How does addressing work?

Each node has a **permanent IPv6 address** derived from its public key:

```
IPv6 address = truncate(hash(public_key))
```

This means:
- Your address **never changes** (unless you regenerate keys)
- Address is **location-independent** (works anywhere)
- **No central allocation** needed (no DHCP, no registries)

Addresses use the `0200::/7` IPv6 range (IETF-deprecated, no conflict with ULA).

### What transport protocols are supported?

- **TCP** (with optional TLS)
- **QUIC** (experimental)
- **WebSocket/WebSocket Secure** (for browser compatibility)
- **IPv4 or IPv6** transport

### Does Uqda work behind NAT?

**Yes!** Uqda handles NAT environments gracefully:

- Peering connections are **bidirectional** once established
- No port forwarding required for outbound peerings
- Nodes behind NAT remain fully reachable within the Uqda network
- Effective even with carrier-grade NAT (CGNAT)

---

## Performance

### Why is Uqda faster?

Uqda includes several performance optimizations:

1. **Reduced timeouts**:
   - Handshake: 5s (vs 6s)
   - TCP dial: 3s (vs 5s)
   - WebSocket: 5s (vs 10s)

2. **DNS caching**:
   - Successful lookups cached for 5 minutes
   - Failed lookups cached for 30 seconds
   - Reduces latency by 10-30ms per connection

3. **Adaptive backoff**:
   - First retry: 100ms (vs 1s)
   - Faster reconnection after temporary failures

See the [Technical Whitepaper](WHITEPAPER.md) for performance details.

### What's the latency improvement?

In typical scenarios:
- **Connection establishment**: 20-40ms (vs 50-100ms baseline)
- **End-to-end latency**: 20-50ms improvement
- **Reconnection time**: 500-900ms faster

In a 10-hop network with 3 temporary link failures per hour, these optimizations reduce average end-to-end latency by **35-55ms**.

### How much bandwidth does Uqda use?

Protocol overhead is minimal:
- Bloom filter exchanges for routing
- Spanning tree maintenance messages
- Data forwarding proportional to network position

Nodes with multiple peerings will route traffic for others (this is how mesh networks work).

---

## Security

### Has Uqda been audited?

**Not yet.** Uqda has not undergone independent security audit.

The codebase inherits cryptographic implementations from:
- Go standard library (audited)
- Yggdrasil Network (community-reviewed since 2017)

**Recommendations**:
- Do not use for high-security applications without independent review
- Assume presence of undiscovered vulnerabilities
- Defense-in-depth: Use application-layer encryption for sensitive data

**Planned**: Community-funded audit scheduled for Q2 2026.

### Is it safe to use?

**Yes, with appropriate precautions:**

1. **Configure firewalls** (see above)
2. **Don't expose sensitive services** without additional security
3. **Use application-layer encryption** for highly sensitive data
4. **Monitor your node** for unexpected traffic

See [../SECURITY.md](../SECURITY.md) for detailed security guidance.

### How do I report a security vulnerability?

**Please do NOT open a public issue.**

Email us privately at: **uqda@proton.me**

Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

We will acknowledge receipt within 48 hours and work with you on responsible disclosure.

---

## Use Cases

### What can I use Uqda for?

- **Community mesh networks** - Neighborhoods building independent infrastructure
- **Edge computing** - Direct device-to-device communication
- **Emergency networks** - Rapid deployment without infrastructure
- **Research and development** - Protocol experimentation
- **Decentralized applications** - Peer-to-peer services

### What should I NOT use Uqda for?

- **Anonymity** - Use Tor/I2P instead
- **VPN replacement** - Uqda is not designed as a traditional VPN
- **Mission-critical systems** - Not yet production-ready for critical applications
- **High-security environments** - Without independent audit

---

## Troubleshooting

### My node won't connect to peers

1. **Check firewall rules** - Make sure you're not blocking outbound connections
2. **Verify peer addresses** - Ensure peer URLs are correct
3. **Check network connectivity** - Can you reach the peer's IP address?
4. **Review logs** - Check for error messages

### Connection is slow

1. **Check peer distance** - Prefer geographically close peers
2. **Monitor routing table** - Use `uqdactl getPeers` to see active connections
3. **Review performance metrics** - See the [Technical Whitepaper](WHITEPAPER.md)

### How do I add/remove peers?

**Using uqdactl:**
```bash
# Add peer
uqdactl addPeer tcp://peer.example.com:12345

# Remove peer
uqdactl removePeer tcp://peer.example.com:12345

# List peers
uqdactl getPeers
```

**Using configuration file:**
Edit `uqda.conf` and add/remove entries from the `Peers` array.

---

## Community

### How can I contribute?

See [../CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines on:
- Code contributions
- Documentation improvements
- Bug reports
- Feature requests

### Where can I get help?

- **GitHub Discussions** - [Ask questions & share ideas](https://github.com/Uqda/Core/discussions)
- **Issue Tracker** - [Report bugs](https://github.com/Uqda/Core/issues)
- **Email** - uqda@proton.me

### Is there a community chat?

Currently, we use GitHub Discussions for community interaction. This may expand in the future.

---

## Licensing

### What license is Uqda under?

**GNU Lesser General Public License v3.0 (LGPLv3)** with binary distribution exception.

See [../LICENSE](../LICENSE) for full details.

### Can I use Uqda commercially?

**Yes!** LGPLv3 allows commercial use. See the license for details.

---

## More Questions?

- Check the [Technical Whitepaper](WHITEPAPER.md) for in-depth technical details
- See the [Executive Summary](EXECUTIVE_SUMMARY.md) for a quick overview
- Visit [GitHub Discussions](https://github.com/Uqda/Core/discussions) to ask the community
- Email us at **uqda@proton.me**

