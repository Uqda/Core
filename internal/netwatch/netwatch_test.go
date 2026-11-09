package netwatch

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

func TestOneCallPerSettledChange(t *testing.T) {
	states := []string{"wlan=192.168.1.5/24", "wlan=192.168.1.5/24", "", "", "wlan=10.0.0.7/24", "eth0=10.0.0.8/24,wlan=10.0.0.7/24", "eth0=10.0.0.8/24,wlan=10.0.0.7/24", "eth0=10.0.0.8/24,wlan=10.0.0.7/24"}
	var mu sync.Mutex
	i, calls := 0, 0
	done := make(chan struct{})
	w := &Watcher{
		Every: time.Millisecond,
		List: func() (string, error) {
			mu.Lock()
			defer mu.Unlock()
			if i >= len(states) {
				select {
				case <-done:
				default:
					close(done)
				}
				return states[len(states)-1], nil
			}
			s := states[i]
			i++
			return s, nil
		},
		OnChange: func() {
			mu.Lock()
			calls++
			mu.Unlock()
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	go w.Run(ctx)
	<-done
	time.Sleep(20 * time.Millisecond)
	cancel()
	mu.Lock()
	defer mu.Unlock()
	// Wi-Fi lost (settles once), then back with a new address and a cable
	// added in quick succession (settles once more).
	if calls != 2 {
		t.Fatalf("expected 2 settled changes, got %d", calls)
	}
}

func TestRoutable(t *testing.T) {
	for addr, want := range map[string]bool{
		"192.168.1.5": true, "2001:db8::1": true, "fd00::1": true,
		"127.0.0.1": false, "::1": false, "fe80::1": false, "169.254.1.1": false,
		"200:1a2b:3c4d::7": false, "300:1a2b:3c4d::1": false, "ff02::1": false,
	} {
		if got := Routable(net.ParseIP(addr)); got != want {
			t.Errorf("Routable(%s) = %v, want %v", addr, got, want)
		}
	}
}

func TestFingerprintRuns(t *testing.T) {
	if _, err := Fingerprint(); err != nil {
		t.Fatal(err)
	}
}
