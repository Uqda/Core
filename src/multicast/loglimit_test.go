package multicast

import (
	"errors"
	"testing"
	"time"
)

func TestWarnLimiterDemotesRepeats(t *testing.T) {
	w := newWarnLimiter()
	now := time.Unix(1_000, 0)
	w.now = func() time.Time { return now }
	bind := errors.New("bind: address already in use")
	other := errors.New("bind: cannot assign requested address")

	if !w.allow(warnKey("wlan0", "listen", bind)) {
		t.Fatal("first warning must be logged")
	}
	for i := 0; i < 60; i++ { // one minute of one-second scans
		now = now.Add(time.Second)
		if w.allow(warnKey("wlan0", "listen", bind)) {
			t.Fatalf("repeat %d was logged at warning level", i)
		}
	}
	if !w.allow(warnKey("wlan0", "listen", other)) {
		t.Fatal("a different error must be logged")
	}
	if !w.allow(warnKey("docker0", "listen", bind)) {
		t.Fatal("the same error on another interface must be logged")
	}
	now = now.Add(repeatWarnInterval)
	if !w.allow(warnKey("wlan0", "listen", bind)) {
		t.Fatal("the warning must be repeated after the quiet interval")
	}
	w.reset("wlan0")
	if !w.allow(warnKey("wlan0", "listen", bind)) {
		t.Fatal("after the interface recovers, a new failure must be logged at once")
	}
}
