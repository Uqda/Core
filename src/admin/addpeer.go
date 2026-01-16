package admin

import (
	"fmt"
	"net/url"
	"os"

	"github.com/Uqda/Core/src/config"
	"github.com/hjson/hjson-go/v4"
)

type AddPeerRequest struct {
	Uri   string `json:"uri"`
	Sintf string `json:"interface,omitempty"`
}

type AddPeerResponse struct{}

func (a *AdminSocket) addPeerHandler(req *AddPeerRequest, _ *AddPeerResponse) error {
	u, err := url.Parse(req.Uri)
	if err != nil {
		return fmt.Errorf("unable to parse peering URI: %w", err)
	}
	if err := a.core.AddPeer(u, req.Sintf); err != nil {
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

func (a *AdminSocket) savePeersToConfig() error {
	// Get all configured persistent peers from links
	persistentPeers := []string{}
	interfacePeers := make(map[string][]string)

	a.core.GetConfiguredPeers(func(uri string, sintf string) {
		if sintf == "" {
			// Peers without source interface go to Peers array
			persistentPeers = append(persistentPeers, uri)
		} else {
			// Peers with source interface go to InterfacePeers map
			if interfacePeers[sintf] == nil {
				interfacePeers[sintf] = []string{}
			}
			interfacePeers[sintf] = append(interfacePeers[sintf], uri)
		}
	})

	// Read existing config
	cfg := config.GenerateConfig()
	if f, err := os.Open(a.configFilePath); err == nil {
		cfg.ReadFrom(f)
		f.Close()
	}

	// Update peers list
	cfg.Peers = persistentPeers
	if len(interfacePeers) > 0 {
		cfg.InterfacePeers = interfacePeers
	}

	// Write config back to file
	bs, err := hjson.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(a.configFilePath, bs, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
