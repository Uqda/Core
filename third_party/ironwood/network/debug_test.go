package network

import (
	"sync"
	"testing"
	"time"

	"github.com/Arceliar/phony"
)

func TestPeerLatencySnapshotConcurrentUpdates(t *testing.T) {
	var c core
	c.peers.init(&c)
	p := &peer{port: 1}
	phony.Block(&c.peers, func() {
		c.peers.peers[p.key] = map[*peer]struct{}{p: {}}
	})
	debug := Debug{c: &c}
	var workers sync.WaitGroup
	workers.Add(2)
	go func() {
		defer workers.Done()
		for i := 0; i < 2000; i++ {
			phony.Block(p, func() {
				p.srst = time.Now()
				p.srrt = p.srst.Add(2 * time.Millisecond)
			})
		}
	}()
	go func() {
		defer workers.Done()
		for i := 0; i < 2000; i++ {
			phony.Block(&c.peers, func() { delete(c.peers.peers[p.key], p) })
			phony.Block(&c.peers, func() { c.peers.peers[p.key][p] = struct{}{} })
		}
	}()
	for i := 0; i < 2000; i++ {
		for _, info := range debug.GetPeers() {
			if info.Port != 1 || (info.Latency != 0 && info.Latency != 2*time.Millisecond) {
				t.Errorf("invalid peer snapshot: %+v", info)
			}
		}
	}
	workers.Wait()
}
