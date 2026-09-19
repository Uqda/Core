package core

import (
	"testing"
	"time"
)

// TestJitteredBackoffDurationBounds characterizes jitteredBackoffDuration's
// contract: the reconnect backoff used by links.add's backoffNow closure.
// Equal jitter should keep the result within [duration/2, duration], and
// duration itself must never exceed max regardless of what was asked for.
func TestJitteredBackoffDurationBounds(t *testing.T) {
	const max = 10 * time.Second

	for _, duration := range []time.Duration{
		0,
		1,
		time.Second,
		2 * time.Second,
		9 * time.Second,
		10 * time.Second,
		time.Hour, // deliberately larger than max
	} {
		for i := 0; i < 200; i++ {
			got := jitteredBackoffDuration(duration, max)
			if got > max {
				t.Fatalf("jitteredBackoffDuration(%s, %s) = %s, exceeds max", duration, max, got)
			}
			capped := duration
			if capped > max {
				capped = max
			}
			half := capped / 2
			if got < half {
				t.Fatalf("jitteredBackoffDuration(%s, %s) = %s, below the guaranteed half (%s)", duration, max, got, half)
			}
			if got > capped {
				t.Fatalf("jitteredBackoffDuration(%s, %s) = %s, exceeds the capped duration (%s)", duration, max, got, capped)
			}
		}
	}
}

// TestJitteredBackoffDurationVaries is a sanity check that jitter is
// actually happening - not a statistical test, just confirmation that
// repeated calls with the same input don't all return the exact same
// value, which would indicate the randomization was accidentally
// stripped out (e.g. a future edit collapsing this to duration/2+0).
func TestJitteredBackoffDurationVaries(t *testing.T) {
	const duration = 10 * time.Second
	const max = time.Minute

	first := jitteredBackoffDuration(duration, max)
	for i := 0; i < 100; i++ {
		if jitteredBackoffDuration(duration, max) != first {
			return
		}
	}
	t.Fatalf("jitteredBackoffDuration(%s, %s) returned %s on every one of 101 calls - jitter does not appear to be applied", duration, max, first)
}
