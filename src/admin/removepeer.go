package admin

import (
	"fmt"
	"net/url"
)

// RemovePeerRequest identifies a persistent peer to remove.
type RemovePeerRequest struct {
	Uri   string `json:"uri"`
	Sintf string `json:"interface,omitempty"`
}

// RemovePeerResponse is returned after a peer is removed.
type RemovePeerResponse struct{}

func (a *AdminSocket) removePeerHandler(req *RemovePeerRequest, _ *RemovePeerResponse) error {
	u, err := url.Parse(req.Uri)
	if err != nil {
		return fmt.Errorf("unable to parse peering URI: %w", err)
	}
	return a.core.RemovePeer(u, req.Sintf)
}
