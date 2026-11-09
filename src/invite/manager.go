package invite

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Limits for invites and the redemption listener.
const (
	DefaultTTL       = 15 * time.Minute
	MaxTTL           = 24 * time.Hour
	MaxUses          = 10
	maxActiveInvites = 16
	maxRedeemConns   = 16
	redeemDeadline   = 10 * time.Second
	maxRequestBytes  = 4096
	maxFailuresPerIP = 10
	failureWindow    = time.Minute
	maxTrackedIPs    = 4096
)

// Logger is the subset of the daemon logger the manager uses.
type Logger interface {
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Debugf(string, ...interface{})
}

// GrantFunc builds the grant for a redemption. local is the address of this
// node that the joiner connected to, so listener URIs can use a reachable host.
type GrantFunc func(local net.IP) Grant

type entry struct {
	hash    [sha256.Size]byte
	expires time.Time
	uses    int
	name    string
}

// Manager issues invites and serves redemptions while any are active.
type Manager struct {
	cert      *tls.Certificate
	publicKey string
	grant     GrantFunc
	log       Logger
	port      uint16
	now       func() time.Time
	addrs     func(port int) []string

	mu       sync.Mutex
	invites  []*entry
	listener net.Listener
	timer    *time.Timer
	failures map[string][]time.Time
	slots    chan struct{}
}

// New creates a manager. cert must be the node's self-signed Ed25519
// certificate; port 0 picks a random port each time the listener opens.
func New(cert *tls.Certificate, publicKey []byte, port uint16, grant GrantFunc, log Logger) *Manager {
	return &Manager{
		cert:      cert,
		publicKey: hex.EncodeToString(publicKey),
		grant:     grant,
		log:       log,
		port:      port,
		now:       time.Now,
		addrs:     localAddresses,
		failures:  make(map[string][]time.Time),
		slots:     make(chan struct{}, maxRedeemConns),
	}
}

// Created describes a new invite.
type Created struct {
	Warning   string   `json:"warning,omitempty"`
	Token     string   `json:"token"`
	Expires   string   `json:"expires"`
	Uses      int      `json:"uses"`
	Addresses []string `json:"addresses"`
}

// Create issues a new invite valid for ttl and uses redemptions.
func (m *Manager) Create(ttl time.Duration, uses int, name string) (Created, error) {
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	if ttl > MaxTTL {
		return Created{}, fmt.Errorf("expires must be at most %s", MaxTTL)
	}
	if uses <= 0 {
		uses = 1
	}
	if uses > MaxUses {
		return Created{}, fmt.Errorf("uses must be at most %d", MaxUses)
	}
	if len(name) > maxNameSize {
		return Created{}, fmt.Errorf("name must be at most %d characters", maxNameSize)
	}
	secret := make([]byte, secretSize)
	if _, err := rand.Read(secret); err != nil {
		return Created{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneLocked()
	if len(m.invites) >= maxActiveInvites {
		return Created{}, fmt.Errorf("too many active invites (%d); cancel some first", maxActiveInvites)
	}
	if err := m.openLocked(); err != nil {
		return Created{}, err
	}
	port := m.listener.Addr().(*net.TCPAddr).Port
	addrs := m.addrs(port)
	if len(addrs) == 0 {
		return Created{}, errors.New("no usable network address found for redeeming invites")
	}
	expires := m.now().Add(ttl)
	m.invites = append(m.invites, &entry{hash: sha256.Sum256(secret), expires: expires, uses: uses, name: name})
	m.scheduleLocked()
	t := Token{
		Key:       m.publicKey,
		Addresses: addrs,
		Secret:    base64.RawURLEncoding.EncodeToString(secret),
		Expires:   expires.Unix(),
		Name:      name,
	}
	created := Created{Token: t.Encode(), Expires: expires.Format(time.RFC3339), Uses: uses, Addresses: addrs}
	if preview := m.grant(nil); len(preview.Peers) == 0 {
		created.Warning = "this node has no peers or fixed-port listeners to share, so the new device will only connect if it finds this one on the local network or has its own peers"
	}
	m.log.Infof("Invite created: %d use(s), expires %s, redeemable on port %d", uses, expires.Format(time.RFC3339), port)
	return created, nil
}

// Active reports the number of unexpired invites.
func (m *Manager) Active() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneLocked()
	return len(m.invites)
}

// CancelAll revokes every invite and closes the listener.
func (m *Manager) CancelAll() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := len(m.invites)
	m.invites = nil
	m.closeLocked()
	return n
}

// Stop is CancelAll for shutdown.
func (m *Manager) Stop() { m.CancelAll() }

func (m *Manager) pruneLocked() {
	now := m.now()
	kept := m.invites[:0]
	for _, e := range m.invites {
		if now.Before(e.expires) && e.uses > 0 {
			kept = append(kept, e)
		}
	}
	m.invites = kept
	if len(m.invites) == 0 {
		m.closeLocked()
	}
}

// scheduleLocked closes the listener when the last invite expires.
func (m *Manager) scheduleLocked() {
	if m.timer != nil {
		m.timer.Stop()
	}
	var last time.Time
	for _, e := range m.invites {
		if e.expires.After(last) {
			last = e.expires
		}
	}
	if last.IsZero() {
		return
	}
	m.timer = time.AfterFunc(time.Until(last)+time.Second, func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		m.pruneLocked()
	})
}

func (m *Manager) openLocked() error {
	if m.listener != nil {
		return nil
	}
	ln, err := net.Listen("tcp", net.JoinHostPort("", strconv.Itoa(int(m.port))))
	if err != nil {
		return fmt.Errorf("open invite listener: %w", err)
	}
	m.listener = ln
	go m.serve(ln)
	return nil
}

func (m *Manager) closeLocked() {
	if m.timer != nil {
		m.timer.Stop()
		m.timer = nil
	}
	if m.listener != nil {
		_ = m.listener.Close()
		m.listener = nil
		m.log.Debugf("Invite listener closed")
	}
}

func (m *Manager) serve(ln net.Listener) {
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*m.cert},
		MinVersion:   tls.VersionTLS13,
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		select {
		case m.slots <- struct{}{}:
			go func() {
				defer func() { <-m.slots }()
				m.handle(tls.Server(conn, tlsConfig))
			}()
		default:
			_ = conn.Close()
		}
	}
}

func (m *Manager) handle(conn *tls.Conn) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(redeemDeadline))
	source := sourceKey(conn.RemoteAddr())
	reply := func(r redeemResponse) { _ = json.NewEncoder(conn).Encode(r) }

	// Read the request before any reply. Closing a socket with unread data
	// makes Windows reset the connection, and the client would lose the reply.
	var req redeemRequest
	if err := json.NewDecoder(io.LimitReader(conn, maxRequestBytes)).Decode(&req); err != nil {
		m.fail(source)
		return
	}
	if m.limited(source) {
		reply(redeemResponse{Error: "too many failed attempts; wait a minute"})
		return
	}
	secret, err := base64.RawURLEncoding.DecodeString(req.Secret)
	if err != nil || len(secret) != secretSize {
		m.fail(source)
		reply(redeemResponse{Error: "invalid or expired invite"})
		return
	}
	name, ok := m.consume(sha256.Sum256(secret))
	if !ok {
		m.fail(source)
		reply(redeemResponse{Error: "invalid or expired invite"})
		return
	}
	var local net.IP
	if a, ok := conn.LocalAddr().(*net.TCPAddr); ok {
		local = a.IP
	}
	g := m.grant(local)
	g.Name = name
	reply(redeemResponse{OK: true, Grant: &g})
	m.log.Infof("Invite redeemed from %s", conn.RemoteAddr())
}

// consume uses one redemption of the invite matching hash.
func (m *Manager) consume(hash [sha256.Size]byte) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.now()
	var match *entry
	for _, e := range m.invites {
		// Compare every entry in constant time; do not stop early.
		if subtle.ConstantTimeCompare(e.hash[:], hash[:]) == 1 && now.Before(e.expires) && e.uses > 0 {
			match = e
		}
	}
	if match == nil {
		return "", false
	}
	match.uses--
	name := match.name
	m.pruneLocked()
	return name, true
}

func (m *Manager) limited(source string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	cutoff := m.now().Add(-failureWindow)
	recent := m.failures[source][:0]
	for _, t := range m.failures[source] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	if len(recent) == 0 {
		delete(m.failures, source)
	} else {
		m.failures[source] = recent
	}
	return len(recent) >= maxFailuresPerIP
}

func (m *Manager) fail(source string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.failures[source]; !ok && len(m.failures) >= maxTrackedIPs {
		m.failures = make(map[string][]time.Time)
	}
	m.failures[source] = append(m.failures[source], m.now())
}

func sourceKey(addr net.Addr) string {
	if a, ok := addr.(*net.TCPAddr); ok {
		if v4 := a.IP.To4(); v4 != nil {
			return v4.String()
		}
		return a.IP.Mask(net.CIDRMask(64, 128)).String()
	}
	return addr.String()
}

// maxInviteAddresses keeps invites, and their QR codes, short. A joining
// device tries the addresses in order, so the most likely one comes first.
const maxInviteAddresses = 4

// virtualInterfaceHints mark adapters of virtual machines, containers and
// other tunnels. Their addresses are rarely reachable from another device on
// the network, and a joining device would waste time trying them.
var virtualInterfaceHints = []string{
	"vethernet", "wsl", "hyper-v", "virtualbox", "vboxnet", "vmware", "vmnet",
	"docker", "br-", "veth", "virbr", "lxc", "lxd", "podman", "cni",
	"zerotier", "tailscale", "wireguard", "wg", "tun", "tap", "utun", "uqda",
	"bluetooth", "loopback", "npcap",
}

func virtualInterface(name string) bool {
	n := strings.ToLower(name)
	for _, h := range virtualInterfaceHints {
		if strings.HasPrefix(n, h) || strings.Contains(n, " "+h) || strings.Contains(n, "("+h) || strings.Contains(n, h+" ") {
			return true
		}
	}
	return false
}

type candidateAddr struct {
	iface string
	ip    net.IP
}

// addressRank orders invite addresses: private IPv4 (the usual home or office
// LAN) first, then IPv6 unique-local, then public IPv4, then global IPv6.
func addressRank(ip net.IP) int {
	switch {
	case ip.To4() != nil && ip.IsPrivate():
		return 0
	case ip.To4() == nil && ip.IsPrivate():
		return 1
	case ip.To4() != nil:
		return 2
	default:
		return 3
	}
}

// pickInviteAddresses filters and orders candidate addresses for an invite.
func pickInviteAddresses(candidates []candidateAddr, port int) []string {
	var keep []candidateAddr
	for _, c := range candidates {
		ip := c.ip
		switch {
		case ip == nil, ip.IsLoopback(), ip.IsLinkLocalUnicast(), ip.IsMulticast(), ip.IsUnspecified():
			continue
		case ip.To4() == nil && ip[0]&0xfe == 0x02:
			continue // Uqda overlay addresses are not reachable before joining
		case virtualInterface(c.iface):
			continue
		}
		keep = append(keep, c)
	}
	sort.SliceStable(keep, func(a, b int) bool { return addressRank(keep[a].ip) < addressRank(keep[b].ip) })
	var out []string
	for _, c := range keep {
		if len(out) == maxInviteAddresses {
			break
		}
		out = append(out, net.JoinHostPort(c.ip.String(), strconv.Itoa(port)))
	}
	return out
}

// localAddresses lists this host's usable addresses for redeeming invites.
func localAddresses(port int) []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var candidates []candidateAddr
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			if ipn, ok := a.(*net.IPNet); ok {
				candidates = append(candidates, candidateAddr{iface: iface.Name, ip: ipn.IP})
			}
		}
	}
	return pickInviteAddresses(candidates, port)
}
