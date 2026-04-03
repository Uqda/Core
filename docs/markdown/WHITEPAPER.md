# Uqda Network – Technical Whitepaper

**Version 1.0**  
January 2026

---

## Executive Summary

Uqda Network is a decentralized routing protocol designed for building resilient, self-organizing multi-hop mesh networks with end-to-end encryption, minimal configuration, and independence from centralized infrastructure.

Built upon proven cryptographic primitives and innovative routing algorithms, Uqda enables direct peer-to-peer connectivity across dynamic network topologies while maintaining stable, location-independent addressing through IPv6.

This whitepaper presents the technical architecture, design principles, and practical optimizations that distinguish Uqda as a pragmatic solution for modern decentralized networking challenges.

---

## 1. Introduction

### 1.1 Motivation

Contemporary network infrastructure faces fundamental challenges:

- **Centralization**: Heavy reliance on ISPs, BGP routing hierarchies, and autonomous system architectures
- **Configuration complexity**: Manual setup requirements that scale poorly
- **Topology rigidity**: Poor adaptation to dynamic or ad-hoc network conditions
- **Infrastructure dependence**: Inability to operate without existing Internet connectivity

Emerging use cases demand different properties:

- Community-operated mesh networks
- Edge computing with direct device-to-device communication
- Resilient networks in infrastructure-limited environments
- Rapid deployment scenarios (emergency response, temporary installations)
- Privacy-preserving peer-to-peer applications

Uqda addresses these requirements through a fundamentally different approach to network routing and organization.

### 1.2 Design Philosophy

Uqda is built on three core principles:

1. **Decentralization by default**: No central authorities, coordinators, or single points of failure
2. **Simplicity of operation**: Networks should form automatically with minimal human intervention
3. **Practical performance**: Theoretical elegance must yield to real-world usability

### 1.3 Relationship to Yggdrasil Network

Uqda Core is a **derivative work** of the Yggdrasil Network protocol (v0.5), maintained as an independent project with focus on:

- **Performance optimizations** for production deployments
- **Operational improvements** based on real-world usage
- **Enhanced tooling** for monitoring and diagnostics

**Protocol Compatibility**: Uqda nodes are protocol-compatible with Yggdrasil v0.5 nodes and can peer with them seamlessly.

**Governance**: Uqda operates as an independent project with its own release cycle, prioritization, and development roadmap.

Full attribution and licensing information: [ATTRIBUTION.md](../ATTRIBUTION.md)

---

## 2. Architecture Overview

### 2.1 Network Model

Uqda operates as an **overlay network** running atop existing IP infrastructure, though the protocol design does not inherently require overlay operation.

Each network node consists of:

- **Userspace router process**: Single-process software router
- **Virtual TUN interface**: IPv6 packet injection/extraction to host OS
- **Peering layer**: Connection management with neighboring nodes
- **Routing engine**: Path discovery and packet forwarding logic
- **Administrative interface**: Monitoring and control socket

### 2.2 Node Equality

All nodes are **functionally equivalent**. There are no designated:

- Root nodes (though spanning tree construction produces transient roots)
- Supernodes or privileged coordinators
- Gateway nodes (unless explicitly configured for specific services)

Each node can simultaneously:

- Originate traffic
- Receive traffic destined for it
- Forward traffic on behalf of other nodes

---

## 3. Identity and Addressing

### 3.1 Cryptographic Identity

Each node possesses a **public/private key pair** (Ed25519):

- The **public key** serves as the node's true identity
- The private key remains secret and is used for signing protocol messages

### 3.2 IPv6 Address Generation

A stable IPv6 address is **deterministically derived** from the public key:

```
IPv6 address = truncate(hash(public_key))
```

This provides:

- **Location independence**: Address is tied to identity, not network position
- **Mobility**: Nodes retain their address as they move through the network
- **No central allocation**: No DHCP, no address registries, no coordination required

### 3.3 Application Compatibility

Most IPv6-capable applications work over Uqda **without modification**, as the network presents standard IPv6 addresses to the host operating system.

---

## 4. Routing Protocol

### 4.1 Design Objectives

The routing system must satisfy:

- **Minimal state**: Nodes should not maintain full routing tables
- **Automatic convergence**: Paths discovered without manual configuration
- **Fast recovery**: Quick adaptation to topology changes
- **Scalability**: Efficient operation as network size grows

### 4.2 Routing Approach

Uqda employs a **hybrid routing scheme**:

1. **Spanning tree construction** for coordination and bootstrap messaging
2. **Greedy routing in keyspace** using distributed hash table principles
3. **Source routing** for optimized paths when available

#### Spanning Tree

- Provides global synchronization
- Enables coordinate allocation for nodes
- Used as fallback routing mechanism
- Root selection based on public key (lowest wins)

#### Keyspace Routing

- Nodes exchange bloom filters describing reachable key ranges
- Intermediate nodes populate routing tables with keyspace paths
- Packets forwarded toward destination's position in keyspace

#### Path Optimization

- Tree routing establishes initial connectivity
- Shorter paths discovered through direct keyspace routing
- Source routes used when more efficient than keyspace routing
- Automatic fallback to tree routing if optimized path fails

### 4.3 Protocol Security

All routing messages are **cryptographically signed** to prevent:

- Forgery
- Tampering
- Route injection attacks

---

## 5. Peering and Connectivity

### 5.1 Peering Establishment

Nodes establish peerings through two mechanisms:

#### Static Peering
- Manual configuration of peer addresses
- Persistent across restarts
- Can operate over any IP transport (LAN, WAN, Internet)

#### Multicast Discovery
- Automatic peer discovery on local networks
- Link-local multicast advertisements
- Zero-configuration "plug-and-play" operation

### 5.2 Transport

Peering connections support:

- **TCP** (with optional TLS)
- **QUIC** (in experimental implementations)
- **IPv4 or IPv6** transport

### 5.3 NAT Traversal

Uqda handles NAT environments gracefully:

- Peering connections are **bidirectional** once established
- No port forwarding required for outbound peerings
- Nodes behind NAT remain fully reachable within the Uqda network
- Effective for scenarios with carrier-grade NAT (CGNAT)

---

## 6. Security Model

### 6.1 Threat Model

Uqda assumes:

- **Network adversaries**: Any node or peering link may be controlled by an attacker
- **Passive surveillance**: Traffic may be observed at any point
- **Active interference**: Packets may be dropped, delayed, or reordered

Uqda does **not** attempt to hide:

- Peering relationships (who connects to whom)
- Traffic patterns or timing
- Source and destination identities

### 6.2 Cryptographic Protection

All traffic is **end-to-end encrypted**:

- Encryption uses destination's public key
- Forward secrecy through ephemeral key exchange
- Intermediate nodes cannot decrypt payload
- Authentication prevents impersonation

### 6.3 What Uqda Is Not

**Uqda is not an anonymity network.** It does not provide:

- Traffic anonymization
- Onion routing
- Hidden services
- Unlinkability of communications

Direct peers can observe IP addresses and may infer location or identity.

### 6.4 Security Audit Status

**Current status**: Uqda has **not undergone independent security audit**.

The codebase inherits cryptographic implementations from:
- Go standard library (audited)
- Yggdrasil Network (community-reviewed since 2017)

**Recommendations**:
- Do not use for high-security applications without independent review
- Assume presence of undiscovered vulnerabilities
- Defense-in-depth: Use application-layer encryption for sensitive data

**Planned**: Community-funded audit scheduled for Q2 2026.

---

## 7. Performance Optimizations

Uqda incorporates practical optimizations based on real-world deployment experience:

### 7.1 Connection Establishment

- **Reduced handshake timeout**: 5s (optimized from 6s)
- **Faster TCP dial**: 3s timeout
- **Adaptive backoff algorithm**:
  - First retry: 100ms
  - Second retry: 250ms
  - Third retry: 500ms
  - Exponential thereafter

### 7.2 DNS Caching

- Successful lookups cached for 5 minutes
- Failed lookups cached for 30 seconds
- Reduces latency by 10-30ms per connection

### 7.3 Protocol Timeouts

- WebSocket: 5s read/write timeout
- UNIX socket: 2s timeout
- Optimized for responsiveness without sacrificing reliability

### 7.4 Performance Comparison

| Metric | Baseline (Yggdrasil 0.5.12) | Uqda 0.1.1 | Improvement |
|--------|------------------------------|------------|-------------|
| Handshake timeout | 6s | 5s | -16.7% |
| TCP dial timeout | 5s | 3s | -40% |
| First reconnect delay | 1s | 100ms | -90% |
| Connection establishment | 50-100ms | 20-40ms | ~60% faster |
| DNS overhead (uncached) | 30-50ms | <10ms | ~75% reduction |
| Reconnection time | 1-1.5s | 500-900ms | ~50% faster |

**Real-world impact**: In a 10-hop network with 3 temporary link failures per hour, these optimizations reduce average end-to-end latency by 35-55ms and improve user-perceived responsiveness significantly.

---

## 8. Operational Characteristics

### 8.1 Bandwidth Consumption

Nodes with multiple peerings will **route traffic for others**:

- Protocol overhead is minimal (bloom filter exchanges, tree maintenance)
- Data forwarding proportional to network position and connectivity
- Fair packet queuing prevents single-node saturation

### 8.2 MTU Handling

Uqda uses **high MTU values** (up to 65535) on TUN interfaces:

- Stream-based peerings eliminate fragmentation issues
- Reduces system call overhead
- Improves throughput and reduces CPU usage
- Mitigates TCP-over-TCP amplification

### 8.3 Firewall Recommendations

As a **public network**, Uqda nodes are globally reachable by default.

**Strongly recommended**: Deploy IPv6 firewall rules to:

- Block unexpected incoming connections
- Prevent service exposure
- Implement principle of least privilege

### 8.4 Deployment Best Practices

#### Firewall Configuration

**Minimum security posture** (all platforms):
- Block all incoming IPv6 traffic on Uqda interface by default
- Allow only established/related connections
- Explicitly whitelist required services

#### Peer Selection

**For stable networks:**
- Prefer geographically close peers to minimize latency
- Maintain 2-4 static peers for redundancy
- Avoid excessive peering (diminishing returns beyond 5-6 peers)

**For mobile/dynamic scenarios:**
- Enable multicast discovery
- Use public peers as fallback
- Accept higher latency for convenience

#### Monitoring

Essential metrics to track:
- Active peer count (`uqdactl getPeers`)
- Routing table size
- Protocol traffic overhead
- Connection stability (reconnect frequency)

---

## 9. Use Cases

### 9.1 Community Mesh Networks

- Neighborhoods building independent infrastructure
- Resilient local communication
- Reduced dependence on ISPs

### 9.2 Edge Computing

- Direct device-to-device communication
- Low-latency local processing
- No cloud intermediary required

### 9.3 Emergency Networks

- Rapid deployment without infrastructure
- Resilience to partial connectivity
- Communication when Internet unavailable

### 9.4 Research and Development

- Protocol experimentation
- Distributed systems testing
- Network topology research

### 9.5 Decentralized Applications

- Peer-to-peer services
- Censorship-resistant communication
- Privacy-preserving applications

---

## 10. Limitations and Non-Goals

### 10.1 Not a VPN Replacement

While Uqda can be used for private networks, it is **not designed as a traditional VPN**:

- Connecting to public peers bridges networks together
- No built-in Internet gateway functionality
- No exit node concept

### 10.2 Not an Anonymity System

Uqda **does not provide anonymity**:

- Peers know each other's IP addresses
- Traffic patterns may be observable
- No protection against traffic analysis

### 10.3 Current Routing Metric Limitations

**Why hop-count only?**

The protocol prioritizes **verifiable, non-gameable metrics**. Interactive link quality measurements (RTT, bandwidth probes) can be:

- Spoofed by malicious nodes
- Expensive to verify independently
- Difficult to secure in adversarial environments

**Practical impact:**

- Some Internet-based peerings may be suboptimal
- Physical proximity doesn't guarantee routing proximity
- Users should manually tune peerings for performance-critical scenarios

**Future work**: Research into secure, verifiable quality metrics is ongoing.

---

## 11. Project Status and Roadmap

### 11.1 Current Status

- **Maturity**: Active development, production-ready for non-critical workloads
- **Stability**: Reliable for experimental and community deployments
- **Not recommended**: Mission-critical or life-safety applications

### 11.2 Future Development

Planned enhancements include:

- Advanced routing optimizations
- Improved monitoring and diagnostics tools
- Enhanced performance analytics
- Broader platform support
- Potential protocol evolution as requirements emerge

---

## 12. Technical Specifications

### 12.1 Protocol Version

- Current version: **0.5-compatible**
- Backward compatibility maintained across minor versions

### 12.2 Cryptography

- **Signing**: Ed25519
- **Key exchange**: X25519
- **Encryption**: ChaCha20-Poly1305

### 12.3 Address Space

- **IPv6 range**: `0200::/7` (IETF-deprecated range, no conflict with ULA)

### 12.4 Supported Platforms

- Linux (kernel 3.x+)
- Windows (7+)
- macOS (10.12+)
- FreeBSD (experimental)

---

## 13. Comparative Analysis

| Feature | Uqda | Yggdrasil | Tailscale | Nebula | I2P/Tor |
|---------|------|-----------|-----------|--------|---------|
| Decentralized | ✅ | ✅ | ❌ (coordinator) | ❌ (lighthouse) | ✅ |
| Zero-config mesh | ✅ | ✅ | ⚠️ (managed) | ⚠️ (manual PKI) | ❌ |
| Protocol-level encryption | ✅ | ✅ | ✅ | ✅ | ✅ |
| Anonymity | ❌ | ❌ | ❌ | ❌ | ✅ |
| NAT traversal | ✅ (bidirectional) | ✅ | ✅ (STUN/relay) | ⚠️ (limited) | ✅ |
| Performance focus | ✅ | ⚠️ | ✅ | ✅ | ❌ (latency high) |
| Works without Internet | ✅ | ✅ | ❌ | ⚠️ | ❌ |
| Production ready | ⚠️ (non-critical) | ⚠️ (alpha) | ✅ | ✅ | ✅ |

**Legend**: ✅ Full support | ⚠️ Partial/Limited | ❌ Not supported

---

## 14. Conclusion

Uqda represents a **production-focused evolution** of proven decentralized routing principles, optimized for:

- **Real-world performance**: Measurable latency improvements
- **Operational simplicity**: Minimal configuration overhead
- **Practical deployment**: Community networks, edge computing, research

### Not a Silver Bullet

Uqda does **not** solve all networking problems. It is:

- **Not** a replacement for the Internet
- **Not** an anonymity system
- **Not** ready for mission-critical applications (yet)

### What Uqda Enables

- **Community ownership** of network infrastructure
- **Direct connectivity** without intermediaries
- **Resilient communication** in adverse conditions
- **Foundation for decentralized applications**

### Get Involved

- **Documentation**: `https://github.com/Uqda/Core`
- **Source code**: `https://github.com/Uqda/Core`
- **Community**: Join discussions on GitHub

Uqda returns control of network infrastructure to its users.  
**Build the network you need. Own the network you build.**

---

## References

- Original protocol research: Yggdrasil Network Project
- Cryptographic primitives: libsodium, Go standard library
- Network topology research: Various distributed hash table and spanning tree publications

---

## Attribution

Uqda Core is based on the Yggdrasil Network protocol, with significant implementation optimizations and operational improvements. See [ATTRIBUTION.md](../ATTRIBUTION.md) for full credits.

---

**Document Version**: 1.0  
**Last Updated**: January 2026  
**Contact**: uqda@proton.me  
**License**: See LICENSE file in repository

