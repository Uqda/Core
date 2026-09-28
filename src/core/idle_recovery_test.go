package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"testing"
	"time"
)

// TestIdleSessionRecovery exercises the real encrypted session timeout in
// Ironwood. It is opt-in because each cycle must be idle for over a minute.
// A two-node localhost test cannot reproduce every LAN/multicast failure mode.
func TestIdleSessionRecovery(t *testing.T) {
	if os.Getenv("UQDA_TEST_IDLE_RECOVERY") != "1" {
		t.Skip("set UQDA_TEST_IDLE_RECOVERY=1 to run the long idle-session test")
	}

	nodeA, nodeB := CreateAndConnectTwo(t, false)
	defer nodeA.Stop()
	defer nodeB.Stop()
	reads := map[*Core]chan []byte{
		nodeA: make(chan []byte, 2),
		nodeB: make(chan []byte, 2),
	}
	for node, out := range reads {
		go func() {
			buf := make([]byte, 1500)
			for {
				n, _, err := node.ReadFrom(buf)
				if err != nil {
					return
				}
				out <- bytes.Clone(buf[:n])
			}
		}()
	}
	if !WaitConnected(nodeA, nodeB) {
		t.Fatal("nodes did not connect")
	}

	transfer := func(from, to *Core, label string) {
		t.Helper()
		payload := make([]byte, 40+len(label))
		payload[0] = 0x60 // IPv6 header, as required by the packet path.
		binary.BigEndian.PutUint16(payload[4:6], uint16(len(label)))
		payload[6] = 253 // Experimental next-header value.
		payload[7] = 64
		copy(payload[8:24], from.Address())
		copy(payload[24:40], to.Address())
		copy(payload[40:], label)
		start := time.Now()
		if _, err := from.WriteTo(payload, to.LocalAddr()); err != nil {
			t.Fatalf("%s write: %v", label, err)
		}
		select {
		case got := <-reads[to]:
			if !bytes.Equal(got, payload) {
				t.Fatalf("%s received %q", label, got)
			}
		case <-time.After(10 * time.Second):
			t.Fatalf("%s delivery timed out after %s", label, time.Since(start))
		}
		t.Logf("%s delivered in %s", label, time.Since(start))
	}

	transfer(nodeA, nodeB, "initial A to B")
	transfer(nodeB, nodeA, "initial B to A")
	for cycle := 1; cycle <= 2; cycle++ {
		t.Logf("idle cycle %d: waiting beyond Ironwood's one-minute session timeout", cycle)
		time.Sleep(70 * time.Second)
		if sessionsA, sessionsB := len(nodeA.GetSessions()), len(nodeB.GetSessions()); sessionsA != 0 || sessionsB != 0 {
			t.Fatalf("idle cycle %d did not expire both sessions: A=%d B=%d", cycle, sessionsA, sessionsB)
		}
		transfer(nodeA, nodeB, fmt.Sprintf("after idle %d A to B", cycle))
		transfer(nodeB, nodeA, fmt.Sprintf("after idle %d B to A", cycle))
	}
}
