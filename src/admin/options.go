package admin

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"net"
	"slices"
	"sync"
	"time"

	"github.com/Arceliar/ironwood/network"

	"github.com/Uqda/Core/src/address"
)

const (
	lookupRetention  = time.Hour
	maxLookupRecords = 4096
)

type lookupKey [ed25519.PublicKeySize]byte

type lookupInfo struct {
	path []uint64
	time time.Time
}

func recordLookup(infos map[lookupKey]lookupInfo, key lookupKey, path []uint64, now time.Time) {
	for existingKey, recorded := range infos {
		if now.Sub(recorded.time) > lookupRetention {
			delete(infos, existingKey)
		}
	}
	if _, exists := infos[key]; !exists && len(infos) >= maxLookupRecords {
		var oldestKey lookupKey
		var oldestTime time.Time
		oldestSet := false
		for existingKey, recorded := range infos {
			if !oldestSet || recorded.time.Before(oldestTime) {
				oldestKey = existingKey
				oldestTime = recorded.time
				oldestSet = true
			}
		}
		delete(infos, oldestKey)
	}
	infos[key] = lookupInfo{path: slices.Clone(path), time: now}
}

func (c *AdminSocket) _applyOption(opt SetupOption) {
	switch v := opt.(type) {
	case ListenAddress:
		c.config.listenaddr = v
	case LogLookups:
		c.logLookups()
	}
}

// SetupOption configures an AdminSocket during construction.
type SetupOption interface {
	isSetupOption()
}

// ListenAddress sets the admin listener URI.
type ListenAddress string

func (a ListenAddress) isSetupOption() {}

// LogLookups enables the bounded lookup-history admin endpoint.
type LogLookups struct{}

func (l LogLookups) isSetupOption() {}

func (c *AdminSocket) logLookups() {
	type resi struct {
		Address string   `json:"addr"`
		Key     string   `json:"key"`
		Path    []uint64 `json:"path"`
		Time    int64    `json:"time"`
	}
	type res struct {
		Infos []resi `json:"infos"`
	}
	infos := make(map[lookupKey]lookupInfo)
	var m sync.Mutex
	c.core.PacketConn.PacketConn.Debug.SetDebugLookupLogger(func(l network.DebugLookupInfo) {
		var k lookupKey
		copy(k[:], l.Key[:])
		now := time.Now()
		m.Lock()
		recordLookup(infos, k, l.Path, now)
		m.Unlock()
	})
	_ = c.AddHandler(
		"lookups", "Dump a record of lookups received in the past hour", []string{},
		func(in json.RawMessage) (interface{}, error) {
			m.Lock()
			rs := make([]resi, 0, len(infos))
			now := time.Now()
			for k, v := range infos {
				if now.Sub(v.time) > lookupRetention {
					delete(infos, k)
					continue
				}
				addr := address.AddrForKey(ed25519.PublicKey(k[:]))
				if addr == nil {
					continue
				}
				ip := net.IP(addr[:]).String()
				rs = append(rs, resi{Address: ip, Key: hex.EncodeToString(k[:]), Path: v.path, Time: v.time.Unix()})
			}
			m.Unlock()
			return &res{Infos: rs}, nil
		},
	)
}
