package main

import (
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/Uqda/Core/internal/cli"
	"github.com/Uqda/Core/src/admin"
	"github.com/Uqda/Core/src/version"
)

func renderDoctor(w io.Writer, report doctorReport) {
	cli.Header(w, "STATUS", version.DisplayName())
	for _, check := range report.Checks {
		cli.Text(w, fmt.Sprintf("  [%-4s] %-10s ", strings.ToUpper(check.Status), check.Name), check.Detail)
		if check.Hint != "" {
			cli.Text(w, "         Next: ", check.Hint)
		}
	}
	fmt.Fprint(w, "\n  Details: uqda peers   |   Network test: uqda test IP\n\n")
}

func renderCommands(w io.Writer, report admin.ListResponse) {
	cli.Header(w, "ADVANCED COMMANDS", version.DisplayName())
	for _, entry := range report.List {
		fmt.Fprintf(w, "  uqdactl %s\n", entry.Command)
		if len(entry.Fields) > 0 {
			fields := make([]string, len(entry.Fields))
			for i, field := range entry.Fields {
				fields[i] = field + "=..."
			}
			cli.Text(w, "    Arguments: ", strings.Join(fields, " "))
		}
		cli.Text(w, "    ", entry.Description)
		fmt.Fprintln(w)
	}
}

func renderNetworkTest(w io.Writer, report networkTestReport) {
	cli.Header(w, "NETWORK TEST", version.DisplayName())
	cli.Field(w, "Target", report.Target)
	cli.Field(w, "Result", "["+strings.ToUpper(report.Status)+"]")
	cli.Field(w, "Replies", fmt.Sprintf("%d/%d / %.0f%% loss", report.Received, report.Sent, report.LossPercent))
	first := "no reply"
	if report.FirstProbeReceived && report.FirstProbeElapsedMS != nil {
		first = fmt.Sprintf("%.1f ms", *report.FirstProbeElapsedMS)
	}
	cli.Field(w, "First probe", first)
	if report.MedianElapsedMS != nil && report.P95ElapsedMS != nil {
		cli.Field(w, "Median / p95", fmt.Sprintf("%.1f / %.1f ms", *report.MedianElapsedMS, *report.P95ElapsedMS))
	}
	cli.Text(w, "  Note: ", report.MeasurementNote)
	fmt.Fprintln(w)
}

func renderPeers(w io.Writer, report admin.GetPeersResponse) {
	cli.Header(w, "PEERS", version.DisplayName())
	if len(report.Peers) == 0 {
		fmt.Fprint(w, "  No peers configured or connected.\n\n")
		return
	}
	for i, peer := range report.Peers {
		state, direction := "DOWN", "outbound"
		if peer.Up {
			state = "UP"
		}
		if peer.Inbound {
			direction = "inbound"
		}
		fmt.Fprintf(w, "  [%s] Peer %d / %s\n", state, i+1, direction)
		if uri, err := url.Parse(peer.URI); err == nil {
			uri.RawQuery, uri.Fragment, uri.User = "", "", nil
			cli.Text(w, "  Connection     ", uri.String())
		} else {
			cli.Field(w, "Connection", "unavailable")
		}
		cli.Field(w, "Address", peer.IPAddress)
		cli.Field(w, "Uptime", (time.Duration(peer.Uptime) * time.Second).String())
		if peer.Up {
			cli.Field(w, "Latency", fmt.Sprintf("%.2f ms", float64(peer.Latency.Microseconds())/1000))
		}
		cli.Field(w, "Received", peer.RXBytes.String()+" / "+peer.RXRate.String()+"/s")
		cli.Field(w, "Sent", peer.TXBytes.String()+" / "+peer.TXRate.String()+"/s")
		cli.Field(w, "Routing", fmt.Sprintf("priority %d / cost %d", peer.Priority, peer.Cost))
		if !peer.Up && peer.LastError != "" {
			// The detailed admin JSON remains available explicitly. Human output
			// does not copy arbitrary transport errors that may contain secrets.
			cli.Field(w, "Last error", fmt.Sprintf("%s ago; check the local daemon log", peer.LastErrorTime.Round(time.Second)))
		}
		fmt.Fprintln(w)
	}
}
