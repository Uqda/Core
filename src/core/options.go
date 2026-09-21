package core

import (
	"crypto/ed25519"
	"fmt"
	"net"
	"net/url"
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
			// Repeated configuration entries are idempotent.
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
	case GroupPassword:
		c.config.groupPassword = string(v)
	}
	return
}

// SetupOption configures a Core during construction.
type SetupOption interface {
	isSetupOption()
}

// ListenAddress adds an inbound listener URI.
type ListenAddress string

// Peer configures a persistent outbound peer.
type Peer struct {
	URI             string
	SourceInterface string
}

// NodeInfo contains metadata shared with remote nodes.
type NodeInfo map[string]interface{}

// NodeInfoPrivacy controls whether default node metadata is shared.
type NodeInfoPrivacy bool

// AllowedPublicKey permits an inbound peer identity.
type AllowedPublicKey ed25519.PublicKey

// PeerFilter decides whether a remote IP is eligible for peering.
type PeerFilter func(net.IP) bool

// GroupPassword separates encrypted session traffic into a private group.
type GroupPassword string

func (a ListenAddress) isSetupOption()    {}
func (a Peer) isSetupOption()             {}
func (a NodeInfo) isSetupOption()         {}
func (a NodeInfoPrivacy) isSetupOption()  {}
func (a AllowedPublicKey) isSetupOption() {}
func (a PeerFilter) isSetupOption()       {}
func (a GroupPassword) isSetupOption()    {}
