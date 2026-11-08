// Package firewall implements Uqda's inbound overlay firewall.
//
// It sits between decrypted overlay traffic and the TUN interface. Traffic
// this node sends is always allowed and remembered as a flow, so replies come
// back. Unsolicited inbound traffic is dropped unless a rule allows its source,
// protocol and destination port. Sources are authenticated: by the time a
// packet reaches the firewall, ipv6rwc has checked that its source address
// belongs to the public key of the session that delivered it.
package firewall

import (
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Uqda/Core/src/address"
	"github.com/Uqda/Core/src/config"
)

// IP protocol numbers used by the firewall.
const (
	protoAny    = 0
	protoTCP    = 6
	protoUDP    = 17
	protoICMPv6 = 58
)

// Flow idle timeouts. A flow is refreshed by traffic in either direction.
const (
	tcpFlowTimeout  = 30 * time.Minute
	udpFlowTimeout  = 3 * time.Minute
	icmpFlowTimeout = 30 * time.Second
	maxFlows        = 65536
)

type keyArray [ed25519.PublicKeySize]byte

type portRange struct{ lo, hi uint16 }

type rule struct {
	text    string
	anySrc  bool
	keys    map[keyArray]struct{}
	addrs   map[address.Address]struct{}
	subnets map[address.Subnet]struct{}
	proto   uint8 // protoAny matches every protocol
	ports   []portRange
}

type flowKey struct {
	remote     [16]byte
	proto      uint8
	localPort  uint16 // ICMP echo: the identifier
	remotePort uint16 // ICMP echo: the identifier
}

// Firewall filters inbound overlay packets. A nil *Firewall allows everything.
type Firewall struct {
	allowPing bool
	rules     []rule

	mu        sync.Mutex
	flows     map[flowKey]time.Time // flow -> expiry
	recent    map[[16]byte]time.Time
	lastSweep time.Time
	now       func() time.Time
	allowed   atomic.Uint64
	dropped   atomic.Uint64
	untracked atomic.Uint64
}

// New compiles cfg. It returns nil, with no error, when the firewall is disabled.
func New(cfg config.FirewallConfig) (*Firewall, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	f := &Firewall{
		allowPing: cfg.AllowPing,
		flows:     make(map[flowKey]time.Time),
		recent:    make(map[[16]byte]time.Time),
		now:       time.Now,
	}
	for i, r := range cfg.Rules {
		compiled, err := compileRule(r)
		if err != nil {
			return nil, fmt.Errorf("firewall rule %d: %w", i+1, err)
		}
		f.rules = append(f.rules, compiled)
	}
	return f, nil
}

// Validate reports every problem in cfg without building a firewall.
func Validate(cfg config.FirewallConfig) []string {
	var problems []string
	for i, r := range cfg.Rules {
		if _, err := compileRule(r); err != nil {
			problems = append(problems, fmt.Sprintf("Firewall rule %d: %s", i+1, err))
		}
	}
	return problems
}

func compileRule(r config.FirewallRule) (rule, error) {
	out := rule{
		keys:    map[keyArray]struct{}{},
		addrs:   map[address.Address]struct{}{},
		subnets: map[address.Subnet]struct{}{},
	}
	if len(r.Sources) == 0 {
		return out, fmt.Errorf("no Sources given; use [\"*\"] to allow any node")
	}
	for _, src := range r.Sources {
		src = strings.TrimSpace(src)
		switch {
		case src == "*":
			out.anySrc = true
		case len(src) == hex.EncodedLen(ed25519.PublicKeySize):
			raw, err := hex.DecodeString(src)
			if err != nil {
				return out, fmt.Errorf("source %q is not a valid public key", src)
			}
			var k keyArray
			copy(k[:], raw)
			out.keys[k] = struct{}{}
		case strings.HasSuffix(src, "/64"):
			ip, _, err := net.ParseCIDR(src)
			if err != nil || ip.To16() == nil || ip.To4() != nil {
				return out, fmt.Errorf("source %q is not a valid Uqda /64 subnet", src)
			}
			var sn address.Subnet
			copy(sn[:], ip.To16())
			if !sn.IsValid() {
				return out, fmt.Errorf("source %q is not inside the Uqda subnet range (300::/8)", src)
			}
			out.subnets[sn] = struct{}{}
		default:
			ip := net.ParseIP(src)
			if ip == nil || ip.To4() != nil {
				return out, fmt.Errorf("source %q is not a public key, Uqda address, /64 subnet or \"*\"", src)
			}
			var a address.Address
			copy(a[:], ip.To16())
			if !a.IsValid() {
				return out, fmt.Errorf("source %q is not inside the Uqda address range (200::/7)", src)
			}
			out.addrs[a] = struct{}{}
		}
	}
	switch strings.ToLower(strings.TrimSpace(r.Protocol)) {
	case "", "any":
		out.proto = protoAny
	case "tcp":
		out.proto = protoTCP
	case "udp":
		out.proto = protoUDP
	case "icmp", "icmpv6", "icmp6":
		out.proto = protoICMPv6
	default:
		return out, fmt.Errorf("protocol %q must be tcp, udp, icmp or any", r.Protocol)
	}
	for _, p := range r.Ports {
		pr, err := parsePortRange(p)
		if err != nil {
			return out, err
		}
		out.ports = append(out.ports, pr)
	}
	if len(out.ports) > 0 && out.proto != protoTCP && out.proto != protoUDP {
		return out, fmt.Errorf("ports need protocol tcp or udp")
	}
	out.text = describeRule(r)
	return out, nil
}

func parsePortRange(s string) (portRange, error) {
	s = strings.TrimSpace(s)
	lo, hi, isRange := strings.Cut(s, "-")
	a, err := strconv.ParseUint(strings.TrimSpace(lo), 10, 16)
	if err != nil || a == 0 {
		return portRange{}, fmt.Errorf("port %q must be 1-65535 or a range like 8000-8100", s)
	}
	b := a
	if isRange {
		b, err = strconv.ParseUint(strings.TrimSpace(hi), 10, 16)
		if err != nil || b == 0 || b < a {
			return portRange{}, fmt.Errorf("port range %q is invalid", s)
		}
	}
	return portRange{uint16(a), uint16(b)}, nil
}

func describeRule(r config.FirewallRule) string {
	proto := strings.ToLower(r.Protocol)
	if proto == "" {
		proto = "any"
	}
	ports := "all ports"
	if len(r.Ports) > 0 {
		ports = "port " + strings.Join(r.Ports, ",")
	}
	text := fmt.Sprintf("allow %s %s from %s", proto, ports, strings.Join(r.Sources, ", "))
	if r.Comment != "" {
		text += " (" + r.Comment + ")"
	}
	return text
}

// packetInfo is what the firewall needs from one IPv6 packet.
type packetInfo struct {
	ok        bool
	proto     uint8
	srcPort   uint16
	dstPort   uint16
	icmpType  uint8
	icmpID    uint16
	hasPorts  bool
	laterFrag bool // a non-first fragment, which carries no ports
}

// parse walks the IPv6 extension headers to the upper-layer header.
func parse(pkt []byte) packetInfo {
	var info packetInfo
	if len(pkt) < 40 || pkt[0]>>4 != 6 {
		return info
	}
	nh := pkt[6]
	off := 40
	for i := 0; i < 8; i++ {
		switch nh {
		case 0, 43, 60: // hop-by-hop, routing, destination options
			if off+2 > len(pkt) {
				return info
			}
			nh, off = pkt[off], off+(int(pkt[off+1])+1)*8
		case 51: // authentication header
			if off+2 > len(pkt) {
				return info
			}
			nh, off = pkt[off], off+(int(pkt[off+1])+2)*4
		case 44: // fragment
			if off+8 > len(pkt) {
				return info
			}
			fragOffset := binary.BigEndian.Uint16(pkt[off+2:off+4]) >> 3
			nh, off = pkt[off], off+8
			if fragOffset != 0 {
				info.ok, info.proto, info.laterFrag = true, nh, true
				return info
			}
		default:
			i = 8
			continue
		}
		if off > len(pkt) {
			return info
		}
	}
	info.proto = nh
	switch nh {
	case protoTCP, protoUDP:
		if off+4 > len(pkt) {
			return info
		}
		info.srcPort = binary.BigEndian.Uint16(pkt[off : off+2])
		info.dstPort = binary.BigEndian.Uint16(pkt[off+2 : off+4])
		info.hasPorts = true
	case protoICMPv6:
		if off+1 > len(pkt) {
			return info
		}
		info.icmpType = pkt[off]
		if info.icmpType == 128 || info.icmpType == 129 {
			if off+8 > len(pkt) {
				return info
			}
			info.icmpID = binary.BigEndian.Uint16(pkt[off+4 : off+6])
		}
	}
	info.ok = true
	return info
}

func flowTimeout(proto uint8) time.Duration {
	switch proto {
	case protoTCP:
		return tcpFlowTimeout
	case protoICMPv6:
		return icmpFlowTimeout
	default:
		return udpFlowTimeout
	}
}

// Outbound records a packet this node is sending so replies are allowed.
func (f *Firewall) Outbound(pkt []byte) {
	if f == nil {
		return
	}
	info := parse(pkt)
	if !info.ok || info.laterFrag {
		return
	}
	var key flowKey
	copy(key.remote[:], pkt[24:40])
	key.proto = info.proto
	switch {
	case info.hasPorts:
		key.localPort, key.remotePort = info.srcPort, info.dstPort
	case info.proto == protoICMPv6 && info.icmpType == 128:
		key.localPort, key.remotePort = info.icmpID, info.icmpID
	case info.proto == protoICMPv6:
		return // replies and errors we send do not open flows
	}
	now := f.now()
	expiry := now.Add(flowTimeout(info.proto))
	f.mu.Lock()
	defer f.mu.Unlock()
	if now.Sub(f.lastSweep) > time.Minute {
		f.sweepLocked(now)
	}
	if _, exists := f.flows[key]; !exists && len(f.flows) >= maxFlows {
		f.sweepLocked(now)
		if len(f.flows) >= maxFlows {
			f.untracked.Add(1)
			return
		}
	}
	f.flows[key] = expiry
	f.recent[key.remote] = expiry
}

// Inbound reports whether a decrypted packet from the node with public key
// from may be delivered to the host.
func (f *Firewall) Inbound(pkt []byte, from ed25519.PublicKey) bool {
	if f == nil {
		return true
	}
	if f.inbound(pkt, from) {
		f.allowed.Add(1)
		return true
	}
	f.dropped.Add(1)
	return false
}

func (f *Firewall) inbound(pkt []byte, from ed25519.PublicKey) bool {
	info := parse(pkt)
	if !info.ok {
		return false
	}
	var remote [16]byte
	copy(remote[:], pkt[8:24])
	now := f.now()

	if info.laterFrag {
		// Later fragments carry no ports: allow them from nodes we are
		// talking to, or when a rule allows the source for all ports.
		f.mu.Lock()
		exp, ok := f.recent[remote]
		f.mu.Unlock()
		return (ok && now.Before(exp)) || f.ruleAllows(pkt, from, info.proto, 0, false)
	}

	if info.proto == protoICMPv6 {
		switch {
		case info.icmpType < 128:
			// Errors such as Packet Too Big and Destination Unreachable are
			// needed for path MTU discovery and fast failure.
			return true
		case info.icmpType == 128:
			return f.allowPing || f.ruleAllows(pkt, from, protoICMPv6, 0, false)
		case info.icmpType == 129:
			return f.established(flowKey{remote, protoICMPv6, info.icmpID, info.icmpID}, now)
		default:
			return f.ruleAllows(pkt, from, protoICMPv6, 0, false)
		}
	}

	if info.hasPorts && f.established(flowKey{remote, info.proto, info.dstPort, info.srcPort}, now) {
		return true
	}
	return f.ruleAllows(pkt, from, info.proto, info.dstPort, info.hasPorts)
}

// established reports whether key is a live flow, refreshing it if so.
func (f *Firewall) established(key flowKey, now time.Time) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	exp, ok := f.flows[key]
	if !ok {
		return false
	}
	if !now.Before(exp) {
		delete(f.flows, key)
		return false
	}
	f.flows[key] = now.Add(flowTimeout(key.proto))
	return true
}

func (f *Firewall) ruleAllows(pkt []byte, from ed25519.PublicKey, proto uint8, port uint16, hasPorts bool) bool {
	var src address.Address
	var srcNet address.Subnet
	copy(src[:], pkt[8:24])
	copy(srcNet[:], pkt[8:16])
	var key keyArray
	copy(key[:], from)
	for i := range f.rules {
		r := &f.rules[i]
		if r.proto != protoAny && r.proto != proto {
			continue
		}
		if len(r.ports) > 0 {
			if !hasPorts {
				continue
			}
			inRange := false
			for _, pr := range r.ports {
				if port >= pr.lo && port <= pr.hi {
					inRange = true
					break
				}
			}
			if !inRange {
				continue
			}
		}
		if r.anySrc {
			return true
		}
		if _, ok := r.keys[key]; ok {
			return true
		}
		if _, ok := r.addrs[src]; ok {
			return true
		}
		if _, ok := r.subnets[srcNet]; ok {
			return true
		}
	}
	return false
}

func (f *Firewall) sweepLocked(now time.Time) {
	for k, exp := range f.flows {
		if !now.Before(exp) {
			delete(f.flows, k)
		}
	}
	for k, exp := range f.recent {
		if !now.Before(exp) {
			delete(f.recent, k)
		}
	}
	f.lastSweep = now
}

// Status is a point-in-time view of the firewall for administration.
type Status struct {
	Enabled     bool     `json:"enabled"`
	AllowPing   bool     `json:"allow_ping"`
	Rules       []string `json:"rules"`
	ActiveFlows int      `json:"active_flows"`
	Allowed     uint64   `json:"allowed_packets"`
	Dropped     uint64   `json:"dropped_packets"`
	Untracked   uint64   `json:"untracked_flows"`
}

// Status returns current counters. A nil firewall reports itself disabled.
func (f *Firewall) Status() Status {
	if f == nil {
		return Status{Rules: []string{}}
	}
	now := f.now()
	f.mu.Lock()
	if now.Sub(f.lastSweep) > time.Minute {
		f.sweepLocked(now)
	}
	active := len(f.flows)
	f.mu.Unlock()
	rules := make([]string, 0, len(f.rules))
	for _, r := range f.rules {
		rules = append(rules, r.text)
	}
	return Status{
		Enabled:     true,
		AllowPing:   f.allowPing,
		Rules:       rules,
		ActiveFlows: active,
		Allowed:     f.allowed.Load(),
		Dropped:     f.dropped.Load(),
		Untracked:   f.untracked.Load(),
	}
}
