package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Uqda/Core/internal/autopeer"
	"github.com/Uqda/Core/internal/publicpeers"
	"github.com/Uqda/Core/src/config"
)

// autoPeersConfig validates the AutoPeers section and converts it.
func autoPeersConfig(c config.AutoPeersConfig) (autopeer.Config, error) {
	out := autopeer.Config{Country: strings.ToLower(strings.TrimSpace(c.Country)), Count: c.Count}
	if !c.Enabled {
		return out, nil
	}
	if !publicpeers.ValidCountry(out.Country) {
		return out, fmt.Errorf("AutoPeers.Country must be a catalog country name such as germany (run uqda find)")
	}
	if out.Count == 0 {
		out.Count = autopeer.DefaultCount
	}
	if out.Count < 1 || out.Count > autopeer.MaxCount {
		return out, fmt.Errorf("AutoPeers.Count must be 1 to %d", autopeer.MaxCount)
	}
	out.Interval = autopeer.DefaultInterval
	if c.Interval != "" {
		d, err := time.ParseDuration(c.Interval)
		if err != nil || d < autopeer.MinInterval || d > autopeer.MaxInterval {
			return out, fmt.Errorf("AutoPeers.Interval must be a duration from %s to %s, for example 30m", autopeer.MinInterval, autopeer.MaxInterval)
		}
		out.Interval = d
	}
	return out, nil
}

// startAutoPeers runs the manager until the returned function is called.
func startAutoPeers(cfg autopeer.Config, node autopeer.Node, log autopeer.Logger) (*autopeer.Manager, context.CancelFunc) {
	m := autopeer.New(cfg, node, autopeer.CatalogSource, log)
	ctx, cancel := context.WithCancel(context.Background())
	go m.Run(ctx)
	log.Infof("Auto peers enabled: keeping %d healthy peer(s) in %s", cfg.Count, cfg.Country)
	return m, cancel
}
