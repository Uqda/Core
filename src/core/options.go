package core

import (
	"crypto/ed25519"
	"fmt"
	"net"
	"net/url"
	"strings"
)

func (c *Core) _applyOption(opt SetupOption) (err error) {
	switch v := opt.(type) {
	case Peer:
		u, err := url.Parse(v.URI)
		if err != nil {
			return fmt.Errorf("unable to parse peering URI: %w", err)
		}
		err = c.links.add(u, v.SourceInterface, linkTypePersistent)
		switch err {
		case ErrLinkAlreadyConfigured:
			// Don't return this error, otherwise we'll panic at startup
			// if there are multiple of the same peer configured
			return nil
		default:
			return err
		}
	case ListenAddress:
		c.config._listeners[v] = struct{}{}
	case PeerFilter:
		c.config.peerFilter = v
	case NodeInfo:
		c.config.nodeinfo = v
	case NodeInfoPrivacy:
		c.config.nodeinfoPrivacy = v
	case AllowedPublicKey:
		pk := [32]byte{}
		copy(pk[:], v)
		c.config._allowedPublicKeys[pk] = struct{}{}
	case PrivateNetworkAllowedKey:
		pk := [32]byte{}
		copy(pk[:], v)
		c.config._privateNetworkAllowedKeys[pk] = struct{}{}
	case PeerDialNetwork:
		s := strings.TrimSpace(string(v))
		switch s {
		case "", "ip", "ip4", "ip6":
			if s == "" {
				s = "ip"
			}
			c.config.peerDialNetwork = s
		default:
			return fmt.Errorf("PeerDialNetwork must be empty, ip, ip4, or ip6, got %q", s)
		}
	case PreferIPv4:
		c.config.preferIPv4 = bool(v)
	}
	return
}

type SetupOption interface {
	isSetupOption()
}

type ListenAddress string
type Peer struct {
	URI             string
	SourceInterface string
}
type NodeInfo map[string]interface{}
type NodeInfoPrivacy bool
type AllowedPublicKey ed25519.PublicKey

// PrivateNetworkAllowedKey restricts inbound peering when combined with global
// AllowedPublicKey rules (union with global allowed set; see link handler).
type PrivateNetworkAllowedKey ed25519.PublicKey

type PeerFilter func(net.IP) bool

// PeerDialNetwork selects which address family DNS returns when resolving peer hostnames: "ip" (default), "ip4", or "ip6".
type PeerDialNetwork string

// PreferIPv4 requests that IPv4 addresses from DNS be tried before IPv6 when both are present.
type PreferIPv4 bool

func (a ListenAddress) isSetupOption()    {}
func (a Peer) isSetupOption()             {}
func (a NodeInfo) isSetupOption()         {}
func (a NodeInfoPrivacy) isSetupOption()  {}
func (a AllowedPublicKey) isSetupOption()         {}
func (a PrivateNetworkAllowedKey) isSetupOption() {}
func (a PeerFilter) isSetupOption()       {}
func (a PeerDialNetwork) isSetupOption()  {}
func (a PreferIPv4) isSetupOption()       {}
