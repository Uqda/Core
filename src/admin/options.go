package admin

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"net"
	"sync"
	"time"

	"github.com/Arceliar/ironwood/network"

	"github.com/Uqda/Core/src/address"
)

func (c *AdminSocket) _applyOption(opt SetupOption) {
	switch v := opt.(type) {
	case ListenAddress:
		c.config.listenaddr = v
	case LogLookups:
		c.logLookups()
	case ConfigFilePath:
		c.configFilePath = string(v)
	case AuthToken:
		c.authToken = string(v)
	}
}

type SetupOption interface {
	isSetupOption()
}

type ListenAddress string

func (a ListenAddress) isSetupOption() {}

type LogLookups struct{}

func (l LogLookups) isSetupOption() {}

type ConfigFilePath string

func (c ConfigFilePath) isSetupOption() {}

// AuthToken is the shared secret required on each admin request when non-empty (JSON field "auth").
type AuthToken string

func (a AuthToken) isSetupOption() {}

type lookupTraceKey [ed25519.PublicKeySize]byte

type lookupTraceEntry struct {
	path []uint64
	at   time.Time
}

// Retention and bounds for debug lookup tracing (LogLookups). Without periodic pruning the map
// would grow with every distinct key seen on the network.
const (
	lookupDebugRetention     = time.Hour
	lookupDebugPruneInterval = 5 * time.Minute
	maxLookupTraceEntries    = 4096
)

func pruneLookupTraceMap(infos map[lookupTraceKey]lookupTraceEntry, now time.Time) {
	cutoff := now.Add(-lookupDebugRetention)
	for k, v := range infos {
		if v.at.Before(cutoff) {
			delete(infos, k)
		}
	}
	for len(infos) > maxLookupTraceEntries {
		for k := range infos {
			delete(infos, k)
			break
		}
	}
}

func (a *AdminSocket) logLookups() {
	type resi struct {
		Address string   `json:"addr"`
		Key     string   `json:"key"`
		Path    []uint64 `json:"path"`
		Time    int64    `json:"time"`
	}
	type res struct {
		Infos []resi `json:"infos"`
	}
	infos := make(map[lookupTraceKey]lookupTraceEntry)
	var m sync.Mutex

	go func() {
		t := time.NewTicker(lookupDebugPruneInterval)
		defer t.Stop()
		for range t.C {
			m.Lock()
			pruneLookupTraceMap(infos, time.Now())
			m.Unlock()
		}
	}()

	a.core.PacketConn.PacketConn.Debug.SetDebugLookupLogger(func(l network.DebugLookupInfo) {
		var k lookupTraceKey
		copy(k[:], l.Key[:])
		m.Lock()
		infos[k] = lookupTraceEntry{path: l.Path, at: time.Now()}
		m.Unlock()
	})
	_ = a.AddHandler(
		"lookups", "Dump a record of lookups received in the past hour", []string{},
		func(in json.RawMessage) (interface{}, error) {
			m.Lock()
			defer m.Unlock()
			pruneLookupTraceMap(infos, time.Now())
			rs := make([]resi, 0, len(infos))
			for k, v := range infos {
				addrKey := address.AddrForKey(ed25519.PublicKey(k[:]))
				addr := net.IP(addrKey[:]).String()
				rs = append(rs, resi{Address: addr, Key: hex.EncodeToString(k[:]), Path: v.path, Time: v.at.Unix()})
			}
			return &res{Infos: rs}, nil
		},
	)
}
