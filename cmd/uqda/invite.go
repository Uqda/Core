package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"net"
	"net/url"

	"github.com/Uqda/Core/src/address"
	"github.com/Uqda/Core/src/config"
	"github.com/Uqda/Core/src/invite"
)

// inviteGrant builds what a redeemed invite shares: the group password, this
// node's configured peers, and this node's own fixed-port listeners on the
// address the joining device used to reach it.
func inviteGrant(cfg *config.NodeConfig, groupPassword string, publicKey ed25519.PublicKey) invite.GrantFunc {
	return func(local net.IP) invite.Grant {
		peers := append([]string{}, cfg.Peers...)
		if local != nil && !local.IsUnspecified() {
			for _, raw := range cfg.Listen {
				u, err := url.Parse(raw)
				if err != nil || u.Port() == "" || u.Port() == "0" {
					continue // random ports cannot be shared ahead of time
				}
				switch u.Scheme {
				case "tcp", "tls", "quic":
				default:
					continue
				}
				peers = append(peers, u.Scheme+"://"+net.JoinHostPort(local.String(), u.Port()))
			}
		}
		return invite.Grant{
			GroupPassword:  groupPassword,
			Peers:          peers,
			InviterKey:     hex.EncodeToString(publicKey),
			InviterAddress: net.IP(address.AddrForKey(publicKey)[:]).String(),
		}
	}
}
