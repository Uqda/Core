package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/netip"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Uqda/Core/src/admin"
	"github.com/Uqda/Core/src/tun"
)

const probeTimeout = 3 * time.Second

var overlayPrefix = netip.MustParsePrefix("200::/7")

type networkTestOptions struct {
	Target netip.Addr
	Count  int
	Idle   time.Duration
}

type networkTestReport struct {
	Target              string   `json:"target"`
	Status              string   `json:"status"`
	Sent                int      `json:"sent"`
	Received            int      `json:"received"`
	LossPercent         float64  `json:"loss_percent"`
	FirstProbeReceived  bool     `json:"first_probe_received"`
	FirstProbeElapsedMS *float64 `json:"first_probe_elapsed_ms,omitempty"`
	MedianElapsedMS     *float64 `json:"median_elapsed_ms,omitempty"`
	P95ElapsedMS        *float64 `json:"p95_elapsed_ms,omitempty"`
	IdleBeforeProbe     string   `json:"idle_before_probe"`
	MeasurementNote     string   `json:"measurement_note"`
}

func parseNetworkTestOptions(args []string) (networkTestOptions, error) {
	opts := networkTestOptions{Count: 5}
	if len(args) == 0 {
		return opts, fmt.Errorf("usage: uqdactl test IPv6 [count=5] [idle=75s]")
	}
	addr, err := netip.ParseAddr(args[0])
	if err != nil || !addr.Is6() || !overlayPrefix.Contains(addr) {
		return opts, fmt.Errorf("target must be a numeric Uqda IPv6 address in 200::/7")
	}
	opts.Target = addr
	seen := map[string]bool{}
	for _, arg := range args[1:] {
		key, value, ok := strings.Cut(arg, "=")
		if !ok || value == "" || seen[key] {
			return opts, fmt.Errorf("invalid or duplicate option %q; use count=N or idle=DURATION", arg)
		}
		seen[key] = true
		switch key {
		case "count":
			count, err := strconv.Atoi(value)
			if err != nil || count < 1 || count > 20 {
				return opts, fmt.Errorf("count must be between 1 and 20")
			}
			opts.Count = count
		case "idle":
			idle, err := time.ParseDuration(value)
			if err != nil || idle < 0 || idle > 5*time.Minute {
				return opts, fmt.Errorf("idle must be a duration from 0s to 5m")
			}
			opts.Idle = idle
		default:
			return opts, fmt.Errorf("unknown option %q; use count=N or idle=DURATION", key)
		}
	}
	return opts, nil
}

func pingInvocation(goos, target string) (string, []string, error) {
	switch goos {
	case "linux":
		return "ping", []string{"-6", "-n", "-c", "1", "-W", "2", target}, nil
	case "windows":
		return "ping", []string{"-6", "-n", "1", "-w", "2000", target}, nil
	case "darwin", "freebsd", "openbsd", "netbsd":
		return "ping6", []string{"-c", "1", target}, nil
	default:
		return "", nil, fmt.Errorf("network test is not supported on %s", goos)
	}
}

type echoProbe func(string) (time.Duration, error)

func systemEchoProbe(target string) (time.Duration, error) {
	command, args, err := pingInvocation(runtime.GOOS, target)
	if err != nil {
		return 0, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	start := time.Now()
	err = cmd.Run()
	return time.Since(start), err
}

func collectNetworkTest(opts networkTestOptions, probe echoProbe) networkTestReport {
	report := networkTestReport{
		Target: opts.Target.String(), Status: "fail", Sent: opts.Count,
		IdleBeforeProbe: opts.Idle.String(),
		MeasurementNote: "Elapsed times include ping process startup; they are not raw network RTT.",
	}
	durations := make([]time.Duration, 0, opts.Count)
	for attempt := 0; attempt < opts.Count; attempt++ {
		if attempt > 0 {
			time.Sleep(200 * time.Millisecond)
		}
		elapsed, err := probe(report.Target)
		if err != nil {
			continue
		}
		report.Received++
		durations = append(durations, elapsed)
		if attempt == 0 {
			report.FirstProbeReceived = true
			value := float64(elapsed.Microseconds()) / 1000
			report.FirstProbeElapsedMS = &value
		}
	}
	report.LossPercent = 100 * float64(report.Sent-report.Received) / float64(report.Sent)
	if report.Received == report.Sent {
		report.Status = "pass"
	}
	if len(durations) > 0 {
		slices.Sort(durations)
		median := float64(durations[nearestRank(len(durations), .5)].Microseconds()) / 1000
		p95 := float64(durations[nearestRank(len(durations), .95)].Microseconds()) / 1000
		report.MedianElapsedMS = &median
		report.P95ElapsedMS = &p95
	}
	return report
}

func nearestRank(count int, percentile float64) int {
	return int(math.Ceil(float64(count)*percentile)) - 1
}

func runNetworkTest(endpoint string, args []string, inJSON bool) int {
	opts, err := parseNetworkTestOptions(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	var self admin.GetSelfResponse
	if err := doctorSelfWithRetry(endpoint, &self); err != nil {
		fmt.Fprintln(os.Stderr, "Cannot reach the Uqda admin socket; run uqdactl doctor first.")
		return 1
	}
	if selfAddress, err := netip.ParseAddr(self.IPAddress); err == nil && selfAddress == opts.Target {
		fmt.Fprintln(os.Stderr, "Target is this node; enter the other node's Uqda IPv6 address.")
		return 2
	}
	var interfaceState tun.GetTUNResponse
	if err := doctorRequest(endpoint, "getTun", &interfaceState); err != nil || !interfaceState.Enabled {
		fmt.Fprintln(os.Stderr, "The local Uqda TUN interface is unavailable; run uqdactl doctor.")
		return 1
	}
	command, _, err := pingInvocation(runtime.GOOS, opts.Target.String())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if _, err := exec.LookPath(command); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			fmt.Fprintln(os.Stderr, command, "is not installed; install your OS ping utility and retry.")
		} else {
			fmt.Fprintln(os.Stderr, "Cannot run", command, ":", err)
		}
		return 1
	}
	if opts.Idle > 0 && !inJSON {
		fmt.Printf("Waiting %s without probe traffic...\n", opts.Idle)
	}
	time.Sleep(opts.Idle)
	report := collectNetworkTest(opts, systemEchoProbe)
	if inJSON {
		encoded, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not format network test:", err)
			return 1
		}
		fmt.Println(string(encoded))
	} else {
		renderNetworkTest(os.Stdout, report)
	}
	if report.Status != "pass" {
		return 1
	}
	return 0
}
