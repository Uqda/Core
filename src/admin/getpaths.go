package admin

import (
	"encoding/hex"
	"net"
	"slices"
	"strings"

	"github.com/Uqda/Core/src/address"
)

// GetPathsRequest requests the routing-path snapshot.
type GetPathsRequest struct {
}

// GetPathsResponse contains the routing-path snapshot.
type GetPathsResponse struct {
	Paths []PathEntry `json:"paths"`
}

// PathEntry describes a route to one public key.
type PathEntry struct {
	IPAddress string   `json:"address"`
	PublicKey string   `json:"key"`
	Path      []uint64 `json:"path"`
	Sequence  uint64   `json:"sequence"`
}

func (a *AdminSocket) getPathsHandler(_ *GetPathsRequest, res *GetPathsResponse) error {
	paths := a.core.GetPaths()
	res.Paths = make([]PathEntry, 0, len(paths))
	for _, p := range paths {
		addr := address.AddrForKey(p.Key)
		if addr == nil {
			continue
		}
		res.Paths = append(res.Paths, PathEntry{
			IPAddress: net.IP(addr[:]).String(),
			PublicKey: hex.EncodeToString(p.Key),
			Path:      p.Path,
			Sequence:  p.Sequence,
		})
	}
	slices.SortStableFunc(res.Paths, func(a, b PathEntry) int {
		return strings.Compare(a.PublicKey, b.PublicKey)
	})
	return nil
}
