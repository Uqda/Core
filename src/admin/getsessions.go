package admin

import (
	"encoding/hex"
	"net"
	"slices"
	"strings"

	"github.com/Uqda/Core/src/address"
)

// GetSessionsRequest requests the encrypted-session snapshot.
type GetSessionsRequest struct{}

// GetSessionsResponse contains the encrypted-session snapshot.
type GetSessionsResponse struct {
	Sessions []SessionEntry `json:"sessions"`
}

// SessionEntry describes one encrypted session.
type SessionEntry struct {
	IPAddress string   `json:"address"`
	PublicKey string   `json:"key"`
	RXBytes   DataUnit `json:"bytes_recvd"`
	TXBytes   DataUnit `json:"bytes_sent"`
	Uptime    float64  `json:"uptime"`
}

func (a *AdminSocket) getSessionsHandler(_ *GetSessionsRequest, res *GetSessionsResponse) error {
	sessions := a.core.GetSessions()
	res.Sessions = make([]SessionEntry, 0, len(sessions))
	for _, s := range sessions {
		addr := address.AddrForKey(s.Key)
		if addr == nil {
			continue
		}
		res.Sessions = append(res.Sessions, SessionEntry{
			IPAddress: net.IP(addr[:]).String(),
			PublicKey: hex.EncodeToString(s.Key[:]),
			RXBytes:   DataUnit(s.RXBytes),
			TXBytes:   DataUnit(s.TXBytes),
			Uptime:    s.Uptime.Seconds(),
		})
	}
	slices.SortStableFunc(res.Sessions, func(a, b SessionEntry) int {
		return strings.Compare(a.PublicKey, b.PublicKey)
	})
	return nil
}
