package main

import (
	"errors"
	"net/netip"
	"os/exec"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func TestParseNetworkTestOptions(t *testing.T) {
	opts, err := parseNetworkTestOptions([]string{"200:1234::1", "count=20", "idle=75s"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Target != netip.MustParseAddr("200:1234::1") || opts.Count != 20 || opts.Idle != 75*time.Second {
		t.Fatalf("unexpected options: %+v", opts)
	}
	defaults, err := parseNetworkTestOptions([]string{"200:1234::1"})
	if err != nil || defaults.Count != 5 || defaults.Idle != 0 {
		t.Fatalf("unexpected defaults: %+v, %v", defaults, err)
	}
}

func TestSystemEchoProbeLoopback(t *testing.T) {
	command, _, err := pingInvocation(runtime.GOOS, "::1")
	if err != nil {
		t.Skip(err)
	}
	if _, err := exec.LookPath(command); err != nil {
		t.Skipf("%s not installed: %v", command, err)
	}
	if elapsed, err := systemEchoProbe("::1"); err != nil {
		t.Fatalf("IPv6 loopback probe failed after %s: %v", elapsed, err)
	}
}

func TestParseNetworkTestOptionsRejectsUnsafeInputs(t *testing.T) {
	for _, args := range [][]string{
		nil,
		{"127.0.0.1"},
		{"::1"},
		{"example.com"},
		{"200:1234::1", "count=0"},
		{"200:1234::1", "count=21"},
		{"200:1234::1", "idle=-1s"},
		{"200:1234::1", "idle=6m"},
		{"200:1234::1", "count=2", "count=3"},
		{"200:1234::1", "unknown=1"},
	} {
		if _, err := parseNetworkTestOptions(args); err == nil {
			t.Fatalf("accepted invalid options %q", args)
		}
	}
}

func TestPingInvocation(t *testing.T) {
	for _, tc := range []struct {
		os      string
		command string
		args    []string
	}{
		{"linux", "ping", []string{"-6", "-n", "-c", "1", "-W", "2", "200:1234::1"}},
		{"windows", "ping", []string{"-6", "-n", "1", "-w", "2000", "200:1234::1"}},
		{"darwin", "ping6", []string{"-c", "1", "200:1234::1"}},
	} {
		command, args, err := pingInvocation(tc.os, "200:1234::1")
		if err != nil || command != tc.command || !reflect.DeepEqual(args, tc.args) {
			t.Fatalf("%s: %s %q %v", tc.os, command, args, err)
		}
	}
	if _, _, err := pingInvocation("unsupported", "200:1234::1"); err == nil {
		t.Fatal("accepted an unsupported platform")
	}
}

func TestCollectNetworkTestMeasuresLossAndFirstProbe(t *testing.T) {
	opts := networkTestOptions{Target: netip.MustParseAddr("200:1234::1"), Count: 3}
	attempt := 0
	report := collectNetworkTest(opts, func(target string) (time.Duration, error) {
		if target != opts.Target.String() {
			t.Fatalf("wrong target: %s", target)
		}
		attempt++
		switch attempt {
		case 1:
			return 0, errors.New("timeout")
		case 2:
			return 5 * time.Millisecond, nil
		default:
			return 10 * time.Millisecond, nil
		}
	})
	if report.Status != "fail" || report.Sent != 3 || report.Received != 2 ||
		report.LossPercent != 100.0/3.0 || report.FirstProbeReceived ||
		report.FirstProbeElapsedMS != nil || *report.MedianElapsedMS != 5 || *report.P95ElapsedMS != 10 {
		t.Fatalf("unexpected loss report: %+v", report)
	}
}

func TestCollectNetworkTestAllPass(t *testing.T) {
	opts := networkTestOptions{Target: netip.MustParseAddr("200:1234::1"), Count: 1}
	report := collectNetworkTest(opts, func(string) (time.Duration, error) { return 7 * time.Millisecond, nil })
	if report.Status != "pass" || !report.FirstProbeReceived ||
		*report.FirstProbeElapsedMS != 7 || report.LossPercent != 0 {
		t.Fatalf("unexpected pass report: %+v", report)
	}
}
