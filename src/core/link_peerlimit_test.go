package core

import (
	"encoding/hex"
	"testing"
	"time"
)

func TestMaxInboundPeers(t *testing.T) {
	hub := newCoreWith(t, MaxInboundPeers(1))
	addr := "tcp://" + listenTCP(t, hub)
	first, second := newCoreWith(t), newCoreWith(t)

	requireNoError(t, first.CallPeer(mustParseURL(t, addr), ""))
	if !waitForCondition(5*time.Second, func() bool { _, ok := hasPeer(hub, first.PublicKey()); return ok }) {
		t.Fatal("first inbound peer was not accepted")
	}
	requireNoError(t, second.CallPeer(mustParseURL(t, addr), ""))
	time.Sleep(500 * time.Millisecond)
	if _, ok := hasPeer(hub, second.PublicKey()); ok {
		t.Fatal("inbound peer beyond MaxInboundPeers was accepted")
	}

	// Freeing the slot admits a new peer.
	first.Stop()
	if !waitForCondition(5*time.Second, func() bool { _, ok := hasPeer(hub, first.PublicKey()); return !ok }) {
		t.Fatal("stopped peer was not removed")
	}
	requireNoError(t, second.CallPeer(mustParseURL(t, addr), ""))
	if !waitForCondition(5*time.Second, func() bool { _, ok := hasPeer(hub, second.PublicKey()); return ok }) {
		t.Fatal("peer was not accepted after a slot was freed")
	}
}

func TestMaxInboundPeersExemptsAllowedKeys(t *testing.T) {
	trusted, stranger := newCoreWith(t), newCoreWith(t)
	hub := newCoreWith(t, MaxInboundPeers(1), AllowedPublicKey(trusted.PublicKey()), AllowedPublicKey(stranger.PublicKey()))
	addr := "tcp://" + listenTCP(t, hub)
	requireNoError(t, stranger.CallPeer(mustParseURL(t, addr), ""))
	requireNoError(t, trusted.CallPeer(mustParseURL(t, addr), ""))
	ok := waitForCondition(5*time.Second, func() bool {
		_, a := hasPeer(hub, stranger.PublicKey())
		_, b := hasPeer(hub, trusted.PublicKey())
		return a && b
	})
	if !ok {
		t.Fatalf("keys listed in AllowedPublicKeys must bypass the limit; peers: %d (%s)", len(hub.GetPeers()), hex.EncodeToString(trusted.PublicKey())[:8])
	}
}
