package config

import (
	"encoding/hex"
	"fmt"
)

// PrivateNetwork groups invite-based settings: shared multicast password, allowed
// member keys, and bootstrap peers. Keys are ed25519 public keys as hex strings.
type PrivateNetwork struct {
	Name        string   `json:"name" comment:"Logical name for this private group."`
	Password    string   `json:"password,omitempty" comment:"Multicast password shared by members (hex-encoded random bytes)."`
	Peers       []string `json:"peers,omitempty" comment:"Bootstrap peer URIs for this network."`
	AllowedKeys []string `json:"allowed_keys,omitempty" comment:"Member public keys (hex) allowed for inbound peering for this network."`
	CreatedAt   int64    `json:"created_at,omitempty"`
	IsOwner     bool     `json:"is_owner,omitempty"`
}

// PrivateNetworkAllowedKeyHex returns unique decoded public keys from all PrivateNetworks.
func PrivateNetworkAllowedKeyHex(cfg *NodeConfig) ([][]byte, error) {
	seen := make(map[[32]byte]struct{})
	var out [][]byte
	for _, pn := range cfg.PrivateNetworks {
		for _, h := range pn.AllowedKeys {
			k, err := decodeHexKey(h)
			if err != nil {
				return nil, err
			}
			var fp [32]byte
			copy(fp[:], k)
			if _, ok := seen[fp]; ok {
				continue
			}
			seen[fp] = struct{}{}
			out = append(out, k)
		}
	}
	return out, nil
}

func decodeHexKey(h string) ([]byte, error) {
	k, err := hex.DecodeString(h)
	if err != nil {
		return nil, err
	}
	if len(k) != 32 {
		return nil, fmt.Errorf("public key must be 32 bytes (64 hex chars)")
	}
	return k, nil
}
