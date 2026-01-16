package admin

import (
	"fmt"
	"net/url"
)

type RemovePeerRequest struct {
	Uri   string `json:"uri"`
	Sintf string `json:"interface,omitempty"`
}

type RemovePeerResponse struct{}

func (a *AdminSocket) removePeerHandler(req *RemovePeerRequest, _ *RemovePeerResponse) error {
	u, err := url.Parse(req.Uri)
	if err != nil {
		return fmt.Errorf("unable to parse peering URI: %w", err)
	}
	if err := a.core.RemovePeer(u, req.Sintf); err != nil {
		return err
	}
	// Save peers to config file if config file path is set
	if a.configFilePath != "" {
		if err := a.savePeersToConfig(); err != nil {
			a.log.Warnf("Failed to save peers to config file: %v", err)
		}
	}
	return nil
}
