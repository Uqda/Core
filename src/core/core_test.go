package core

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"net"
	"net/url"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/gologme/log"
	"github.com/yggdrasil-network/yggdrasil-go/src/config"
)

// GetLoggerWithPrefix creates a new logger instance with prefix.
// If verbose is set to true, three log levels are enabled: "info", "warn", "error".
func GetLoggerWithPrefix(prefix string, verbose bool) *log.Logger {
	l := log.New(os.Stderr, prefix, log.Flags())
	if !verbose {
		return l
	}
	l.EnableLevel("info")
	l.EnableLevel("warn")
	l.EnableLevel("error")
	return l
}

func require_NoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func require_Error(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error")
	}
}

func require_Equal[T comparable](t *testing.T, a, b T) {
	t.Helper()
	if a != b {
		t.Fatalf("%v != %v", a, b)
	}
}

func require_True(t *testing.T, a bool) {
	t.Helper()
	if !a {
		t.Fatal("expected true")
	}
}

// CreateAndConnectTwo creates two nodes. nodeB connects to nodeA.
// Verbosity flag is passed to logger.
func CreateAndConnectTwo(t testing.TB, verbose bool) (nodeA *Core, nodeB *Core) {
	var err error

	cfgA, cfgB := config.GenerateConfig(), config.GenerateConfig()
	if err = cfgA.GenerateSelfSignedCertificate(); err != nil {
		t.Fatal(err)
	}
	if err = cfgB.GenerateSelfSignedCertificate(); err != nil {
		t.Fatal(err)
	}

	logger := GetLoggerWithPrefix("", false)
	logger.EnableLevel("debug")

	if nodeA, err = New(cfgA.Certificate, logger); err != nil {
		t.Fatal(err)
	}
	if nodeB, err = New(cfgB.Certificate, logger); err != nil {
		t.Fatal(err)
	}

	nodeAListenURL, err := url.Parse("tcp://localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	nodeAListener, err := nodeA.Listen(nodeAListenURL, "")
	if err != nil {
		t.Fatal(err)
	}
	nodeAURL, err := url.Parse("tcp://" + nodeAListener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	if err = nodeB.CallPeer(nodeAURL, ""); err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)

	if l := len(nodeA.GetPeers()); l != 1 {
		t.Fatal("unexpected number of peers", l)
	}
	if l := len(nodeB.GetPeers()); l != 1 {
		t.Fatal("unexpected number of peers", l)
	}

	return nodeA, nodeB
}

// WaitConnected blocks until either nodes negotiated DHT or 5 seconds passed.
func WaitConnected(nodeA, nodeB *Core) bool {
	// It may take up to 3 seconds, but let's wait 5.
	for i := 0; i < 50; i++ {
		time.Sleep(100 * time.Millisecond)
		/*
			if len(nodeA.GetPeers()) > 0 && len(nodeB.GetPeers()) > 0 {
				return true
			}
		*/
		if len(nodeA.GetTree()) > 1 && len(nodeB.GetTree()) > 1 {
			time.Sleep(3 * time.Second) // FIXME hack, there's still stuff happening internally
			return true
		}
	}
	return false
}

// CreateEchoListener creates a routine listening on nodeA. It expects repeats messages of length bufLen.
// It returns a channel used to synchronize the routine with caller.
func CreateEchoListener(t testing.TB, nodeA *Core, bufLen int, repeats int) chan struct{} {
	// Start routine
	done := make(chan struct{})
	go func() {
		buf := make([]byte, bufLen)
		res := make([]byte, bufLen)
		for i := 0; i < repeats; i++ {
			n, from, err := nodeA.ReadFrom(buf)
			if err != nil {
				t.Error(err)
				return
			}
			if n != bufLen {
				t.Error("missing data")
				return
			}
			copy(res, buf)
			copy(res[8:24], buf[24:40])
			copy(res[24:40], buf[8:24])
			_, err = nodeA.WriteTo(res, from)
			if err != nil {
				t.Error(err)
			}
		}
		done <- struct{}{}
	}()

	return done
}

// TestCore_Start_Connect checks if two nodes can connect together.
func TestCore_Start_Connect(t *testing.T) {
	CreateAndConnectTwo(t, true)
}

// TestCore_Start_Transfer checks that messages can be passed between nodes (in both directions).
func TestCore_Start_Transfer(t *testing.T) {
	nodeA, nodeB := CreateAndConnectTwo(t, true)
	defer nodeA.Stop()
	defer nodeB.Stop()

	msgLen := 1500
	done := CreateEchoListener(t, nodeA, msgLen, 1)

	if !WaitConnected(nodeA, nodeB) {
		t.Fatal("nodes did not connect")
	}

	// Send
	msg := make([]byte, msgLen)
	_, _ = rand.Read(msg[40:])
	msg[0] = 0x60
	copy(msg[8:24], nodeB.Address())
	copy(msg[24:40], nodeA.Address())
	_, err := nodeB.WriteTo(msg, nodeA.LocalAddr())
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, msgLen)
	_, _, err = nodeB.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(msg[40:], buf[40:]) {
		t.Fatal("expected echo")
	}
	<-done
}

// BenchmarkCore_Start_Transfer estimates the possible transfer between nodes (in MB/s).
func BenchmarkCore_Start_Transfer(b *testing.B) {
	nodeA, nodeB := CreateAndConnectTwo(b, false)

	msgLen := 1500 // typical MTU
	done := CreateEchoListener(b, nodeA, msgLen, b.N)

	if !WaitConnected(nodeA, nodeB) {
		b.Fatal("nodes did not connect")
	}

	// Send
	msg := make([]byte, msgLen)
	_, _ = rand.Read(msg[40:])
	msg[0] = 0x60
	copy(msg[8:24], nodeB.Address())
	copy(msg[24:40], nodeA.Address())

	buf := make([]byte, msgLen)

	b.SetBytes(int64(msgLen))
	b.ResetTimer()

	addr := nodeA.LocalAddr()
	for i := 0; i < b.N; i++ {
		_, err := nodeB.WriteTo(msg, addr)
		if err != nil {
			b.Fatal(err)
		}
		_, _, err = nodeB.ReadFrom(buf)
		if err != nil {
			b.Fatal(err)
		}
	}
	<-done
}

// waitForGoroutineCountAtMost polls runtime.NumGoroutine until it settles
// at or below max, or fails the test after timeout. Goroutine counts are
// inherently noisy (GC workers, finalizers, the test runner itself), so
// this is a coarse regression guard against gross leaks (an entire link's
// worth of goroutines failing to exit), not a precise leak detector for a
// single stray goroutine.
func waitForGoroutineCountAtMost(t testing.TB, max int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var last int
	for {
		runtime.GC()
		last = runtime.NumGoroutine()
		if last <= max {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("goroutine count did not settle within %s: have %d, want <= %d", timeout, last, max)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// TestCoreStopReleasesGoroutinesAndIsIdempotent characterizes two lifecycle
// properties of Core.Stop that any future refactor (see
// ../../RESTRUCTURING.md's peer/transport split) must preserve:
//
//  1. Stopping a node actually tears down the goroutines it started for
//     its links and listeners, rather than leaking them.
//  2. Stop is safe to call more than once - Core.Stop's own
//     implementation relies on context.CancelFunc and net.Conn.Close
//     both tolerating repeated calls, and callers (e.g. a future signal
//     handler plus an explicit admin "shutdown" command racing each
//     other) should be able to rely on that rather than tracking whether
//     they're the first caller.
func TestCoreStopReleasesGoroutinesAndIsIdempotent(t *testing.T) {
	runtime.GC()
	baseline := runtime.NumGoroutine()

	nodeA, nodeB := CreateAndConnectTwo(t, false)
	if !WaitConnected(nodeA, nodeB) {
		t.Fatal("nodes did not connect")
	}

	nodeA.Stop()
	nodeB.Stop()
	// Calling Stop again must not panic.
	nodeA.Stop()
	nodeB.Stop()

	// Small slack above baseline: this is deliberately loose to avoid
	// flakiness from goroutines outside this package's control, while
	// still catching a gross leak (e.g. an entire link handler or
	// listener accept loop failing to exit on shutdown).
	waitForGoroutineCountAtMost(t, baseline+4, 5*time.Second)
}

// TestRepeatedConnectDisconnectCycles characterizes current behavior of
// AddPeer/RemovePeer under repeated cycling against the same peer, ahead
// of the Phase 5 peer-lifecycle refactor (RESTRUCTURING.md). It checks
// three things a rewritten peer-lifecycle manager must preserve:
//
//  1. Add -> wait connected -> Remove -> Add again works repeatedly
//     without error - RemovePeer must actually clear the link-info entry
//     so a later AddPeer for the same URI doesn't hit
//     ErrLinkAlreadyConfigured forever.
//  2. Removing a peer that was never added, or removing it twice, returns
//     ErrLinkNotConfigured rather than panicking.
//  3. Cycling connect/disconnect several times doesn't leak goroutines -
//     the count after several cycles plus a final Stop should be no worse
//     than after a single cycle, using the same loose-tolerance approach
//     as TestCoreStopReleasesGoroutinesAndIsIdempotent.
func TestRepeatedConnectDisconnectCycles(t *testing.T) {
	nodeA, nodeB := CreateAndConnectTwoUnconnected(t)
	defer nodeA.Stop()
	defer nodeB.Stop()

	listener, err := nodeA.Listen(mustParseURL(t, "tcp://localhost:0"), "")
	if err != nil {
		t.Fatal(err)
	}
	peerURL := mustParseURL(t, "tcp://"+listener.Addr().String())

	// Removing a peer that was never configured must error cleanly.
	if err := nodeB.RemovePeer(peerURL, ""); err != ErrLinkNotConfigured {
		t.Fatalf("expected ErrLinkNotConfigured for an unconfigured peer, got %v", err)
	}

	runtime.GC()
	baseline := runtime.NumGoroutine()

	const cycles = 5
	for i := 0; i < cycles; i++ {
		if err := nodeB.AddPeer(peerURL, ""); err != nil {
			t.Fatalf("cycle %d: AddPeer failed: %s", i, err)
		}
		if !waitForCondition(2*time.Second, func() bool {
			return len(nodeB.GetPeers()) == 1 && nodeB.GetPeers()[0].Up
		}) {
			t.Fatalf("cycle %d: peer did not come up", i)
		}
		if err := nodeB.RemovePeer(peerURL, ""); err != nil {
			t.Fatalf("cycle %d: RemovePeer failed: %s", i, err)
		}
		// Removing twice in a row must not panic and must report the
		// same "not configured" error as removing something that was
		// never added.
		if err := nodeB.RemovePeer(peerURL, ""); err != ErrLinkNotConfigured {
			t.Fatalf("cycle %d: expected ErrLinkNotConfigured on double-remove, got %v", i, err)
		}
		if !waitForCondition(2*time.Second, func() bool {
			return len(nodeB.GetPeers()) == 0
		}) {
			t.Fatalf("cycle %d: peer did not disconnect", i)
		}
	}

	waitForGoroutineCountAtMost(t, baseline+4, 5*time.Second)
}

// TestImmediateReAddAfterRemove is a narrower regression test for the race
// TestRepeatedConnectDisconnectCycles's fix addresses: calling AddPeer for
// the same URI immediately after RemovePeer - with no time for the dial
// goroutine's own asynchronous cleanup to run in between - must not
// spuriously fail with ErrLinkAlreadyConfigured.
func TestImmediateReAddAfterRemove(t *testing.T) {
	nodeA, nodeB := CreateAndConnectTwoUnconnected(t)
	defer nodeA.Stop()
	defer nodeB.Stop()

	listener, err := nodeA.Listen(mustParseURL(t, "tcp://localhost:0"), "")
	if err != nil {
		t.Fatal(err)
	}
	peerURL := mustParseURL(t, "tcp://"+listener.Addr().String())

	if err := nodeB.AddPeer(peerURL, ""); err != nil {
		t.Fatalf("initial AddPeer failed: %s", err)
	}
	if err := nodeB.RemovePeer(peerURL, ""); err != nil {
		t.Fatalf("RemovePeer failed: %s", err)
	}
	// No sleep here - this is exactly the race window under test.
	if err := nodeB.AddPeer(peerURL, ""); err != nil {
		t.Fatalf("AddPeer immediately after RemovePeer should succeed, got: %s", err)
	}
}

// waitForCondition polls fn until it returns true or the timeout elapses,
// then evaluates it one final time so the caller's failure reflects the
// true final state rather than a stale poll.
func waitForCondition(timeout time.Duration, fn func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fn()
}

func mustParseURL(t testing.TB, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

// CreateAndConnectTwoUnconnected creates two nodes with fresh identities
// but does not peer them together, unlike CreateAndConnectTwo - useful for
// tests (like connect/disconnect cycling) that want to drive peering
// explicitly rather than start from an already-connected pair.
func CreateAndConnectTwoUnconnected(t testing.TB) (nodeA *Core, nodeB *Core) {
	t.Helper()
	cfgA, cfgB := config.GenerateConfig(), config.GenerateConfig()
	if err := cfgA.GenerateSelfSignedCertificate(); err != nil {
		t.Fatal(err)
	}
	if err := cfgB.GenerateSelfSignedCertificate(); err != nil {
		t.Fatal(err)
	}
	logger := GetLoggerWithPrefix("", false)
	var err error
	if nodeA, err = New(cfgA.Certificate, logger); err != nil {
		t.Fatal(err)
	}
	if nodeB, err = New(cfgB.Certificate, logger); err != nil {
		t.Fatal(err)
	}
	return nodeA, nodeB
}

func TestAllowedPublicKeys(t *testing.T) {
	logger := GetLoggerWithPrefix("", false)
	cfgA, cfgB := config.GenerateConfig(), config.GenerateConfig()
	require_NoError(t, cfgA.GenerateSelfSignedCertificate())
	require_NoError(t, cfgB.GenerateSelfSignedCertificate())

	nodeA, err := New(cfgA.Certificate, logger, AllowedPublicKey("abcdef"))
	require_NoError(t, err)
	defer nodeA.Stop()

	nodeB, err := New(cfgB.Certificate, logger)
	require_NoError(t, err)
	defer nodeB.Stop()

	u, err := url.Parse("tcp://localhost:0")
	require_NoError(t, err)

	l, err := nodeA.Listen(u, "")
	require_NoError(t, err)

	u, err = url.Parse("tcp://" + l.Addr().String())
	require_NoError(t, err)

	require_NoError(t, nodeB.AddPeer(u, ""))

	time.Sleep(time.Second)

	peers := nodeB.GetPeers()
	require_Equal(t, len(peers), 1)
	require_True(t, !peers[0].Up)
	require_True(t, peers[0].LastError != nil)
}

func TestAllowedPublicKeysLocal(t *testing.T) {
	logger := GetLoggerWithPrefix("", false)
	cfgA, cfgB := config.GenerateConfig(), config.GenerateConfig()
	require_NoError(t, cfgA.GenerateSelfSignedCertificate())
	require_NoError(t, cfgB.GenerateSelfSignedCertificate())

	nodeA, err := New(cfgA.Certificate, logger, AllowedPublicKey("abcdef"))
	require_NoError(t, err)
	defer nodeA.Stop()

	nodeB, err := New(cfgB.Certificate, logger)
	require_NoError(t, err)
	defer nodeB.Stop()

	u, err := url.Parse("tcp://localhost:0")
	require_NoError(t, err)

	l, err := nodeA.ListenLocal(u, "")
	require_NoError(t, err)

	u, err = url.Parse("tcp://" + l.Addr().String())
	require_NoError(t, err)

	require_NoError(t, nodeB.AddPeer(u, ""))

	time.Sleep(time.Second)

	peers := nodeB.GetPeers()
	require_Equal(t, len(peers), 1)
	require_True(t, peers[0].Up)
	require_True(t, peers[0].LastError == nil)
}

func TestGroupPassword(t *testing.T) {
	logger := GetLoggerWithPrefix("", false)
	cfgA, cfgB, cfgC := config.GenerateConfig(), config.GenerateConfig(), config.GenerateConfig()
	require_NoError(t, cfgA.GenerateSelfSignedCertificate())
	require_NoError(t, cfgB.GenerateSelfSignedCertificate())
	require_NoError(t, cfgC.GenerateSelfSignedCertificate())

	nodeA, err := New(cfgA.Certificate, logger, GroupPassword("test-group-password"))
	require_NoError(t, err)
	defer nodeA.Stop()

	nodeB, err := New(cfgB.Certificate, logger, GroupPassword("test-group-password"))
	require_NoError(t, err)
	defer nodeB.Stop()

	nodeC, err := New(cfgC.Certificate, logger, GroupPassword("different-test-group-password"))
	require_NoError(t, err)
	defer nodeC.Stop()

	pathFound := map[string]chan struct{}{
		nodeB.LocalAddr().String(): make(chan struct{}, 1),
		nodeC.LocalAddr().String(): make(chan struct{}, 1),
	}
	nodeA.SetPathNotify(func(key ed25519.PublicKey) {
		pathFound[hex.EncodeToString(key)] <- struct{}{}
	})

	u, err := url.Parse("tcp://localhost:0")
	require_NoError(t, err)

	l, err := nodeA.Listen(u, "")
	require_NoError(t, err)

	u, err = url.Parse("tcp://" + l.Addr().String())
	require_NoError(t, err)

	require_NoError(t, nodeB.AddPeer(u, ""))
	require_NoError(t, nodeC.AddPeer(u, ""))

	require_True(t, WaitConnected(nodeA, nodeB))
	require_True(t, WaitConnected(nodeA, nodeC))

	var connA net.PacketConn = nodeA.PacketConn
	var connB net.PacketConn = nodeB.PacketConn
	var connC net.PacketConn = nodeC.PacketConn

	go func() {
		var buf [1024]byte
		for t.Context().Err() == nil {
			// Needed as encrypted package relies on ReadFrom to process session acks.
			_, _, _ = connA.ReadFrom(buf[:])
		}
	}()

	matchingPasswordMessage := []byte("matching group password")
	differentPasswordMessage := []byte("different group password")

	_, err = connA.WriteTo(matchingPasswordMessage, connB.LocalAddr())
	require_NoError(t, err)

	_, err = connA.WriteTo(differentPasswordMessage, connC.LocalAddr())
	require_NoError(t, err)

	<-pathFound[nodeB.LocalAddr().String()]
	<-pathFound[nodeC.LocalAddr().String()]

	var buf [1024]byte
	deadline := time.Now().Add(3 * time.Second)
	require_NoError(t, connB.SetReadDeadline(deadline))
	require_NoError(t, connC.SetReadDeadline(deadline))

	n, from, err := connB.ReadFrom(buf[:])
	require_NoError(t, err)
	require_Equal(t, from.String(), connA.LocalAddr().String())
	require_True(t, bytes.Equal(buf[:n], matchingPasswordMessage))

	_, _, err = connC.ReadFrom(buf[:])
	require_Error(t, err)
}
