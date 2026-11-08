package invite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/Uqda/Core/src/admin"
)

// CreateRequest holds createInvite arguments.
type CreateRequest struct {
	Expires string `json:"expires,omitempty"` // Go duration, default 15m, max 24h
	Uses    string `json:"uses,omitempty"`    // default 1, max 10
	Name    string `json:"name,omitempty"`
}

// CancelResponse reports how many invites were revoked.
type CancelResponse struct {
	Cancelled int `json:"cancelled"`
}

// SetupAdminHandlers registers createInvite and cancelInvites.
func SetupAdminHandlers(a *admin.AdminSocket, m *Manager) {
	_ = a.AddHandler(
		"createInvite", "Create a one-time, expiring invite for another device", []string{"expires", "uses", "name"},
		func(in json.RawMessage) (interface{}, error) {
			var req CreateRequest
			dec := json.NewDecoder(bytes.NewReader(in))
			dec.DisallowUnknownFields()
			if err := dec.Decode(&req); err != nil {
				return nil, fmt.Errorf("invalid arguments: use expires=15m uses=1 name=TEXT")
			}
			ttl := DefaultTTL
			if req.Expires != "" {
				d, err := time.ParseDuration(req.Expires)
				if err != nil || d <= 0 {
					return nil, fmt.Errorf("expires must be a duration such as 15m, 2h or 24h")
				}
				ttl = d
			}
			uses := 1
			if req.Uses != "" {
				n, err := strconv.Atoi(req.Uses)
				if err != nil || n < 1 {
					return nil, fmt.Errorf("uses must be a whole number from 1 to %d", MaxUses)
				}
				uses = n
			}
			return m.Create(ttl, uses, req.Name)
		},
	)
	_ = a.AddHandler(
		"cancelInvites", "Revoke every active invite", []string{},
		func(in json.RawMessage) (interface{}, error) {
			return CancelResponse{Cancelled: m.CancelAll()}, nil
		},
	)
}
