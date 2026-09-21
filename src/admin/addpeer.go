package admin

import (
	"fmt"
	"net/url"
)

// AddPeerRequest describes a persistent peer to add.
type AddPeerRequest struct {
	Uri   string `json:"uri"`
	Sintf string `json:"interface,omitempty"`
}

// AddPeerResponse is returned after a peer is added.
type AddPeerResponse struct{}

func (a *AdminSocket) addPeerHandler(req *AddPeerRequest, _ *AddPeerResponse) error {
	u, err := url.Parse(req.Uri)
	if err != nil {
		return fmt.Errorf("unable to parse peering URI: %w", err)
	}
	return a.core.AddPeer(u, req.Sintf)
}
