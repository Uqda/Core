package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"suah.dev/protect"

	"github.com/Uqda/Core/src/admin"
	"github.com/Uqda/Core/src/core"
	"github.com/Uqda/Core/src/multicast"
	"github.com/Uqda/Core/src/tun"
	"github.com/Uqda/Core/src/version"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

func main() {
	os.Exit(run())
}

func run() int {
	logbuffer := &bytes.Buffer{}
	logger := log.New(logbuffer, "", log.Flags())

	if err := protect.Pledge("stdio rpath inet unix dns"); err != nil {
		return fail(logger, logbuffer, "apply initial process restrictions: %v", err)
	}

	cmdLineEnv := newCmdLineEnv()
	cmdLineEnv.parseFlagsAndArgs()

	if cmdLineEnv.ver || (len(cmdLineEnv.args) == 1 && cmdLineEnv.args[0] == "version") {
		fmt.Println(version.DisplayName())
		return 0
	}

	if len(cmdLineEnv.args) == 0 {
		flag.Usage()
		return 0
	}

	if err := cmdLineEnv.setEndpoint(logger); err != nil {
		return fail(logger, logbuffer, "%v", err)
	}

	conn, err := dialAdminEndpoint(cmdLineEnv.endpoint, logger)
	if err != nil {
		return fail(logger, logbuffer, "%v", err)
	}

	if err := protect.Pledge("stdio"); err != nil {
		_ = conn.Close()
		return fail(logger, logbuffer, "apply process restrictions after connecting: %v", err)
	}

	logger.Println("Connected")
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Println("Close admin connection:", err)
		}
	}()

	decoder := json.NewDecoder(conn)
	send := &admin.AdminSocketRequest{}
	recv := &admin.AdminSocketResponse{}
	args := map[string]string{}
	for c, a := range cmdLineEnv.args {
		if c == 0 {
			if strings.HasPrefix(a, "-") {
				logger.Printf("Ignoring flag %s as it should be specified before other parameters\n", a)
				continue
			}
			logger.Printf("Sending request: %v\n", a)
			send.Name = a
			continue
		}
		tokens := strings.SplitN(a, "=", 2)
		switch {
		case len(tokens) == 1:
			logger.Println("Ignoring invalid argument:", a)
		default:
			args[tokens[0]] = tokens[1]
		}
	}
	if send.Arguments, err = json.Marshal(args); err != nil {
		return fail(logger, logbuffer, "encode command arguments: %v", err)
	}
	request, err := json.Marshal(send)
	if err != nil {
		return fail(logger, logbuffer, "encode admin request: %v", err)
	}
	if _, err := io.Copy(conn, bytes.NewReader(request)); err != nil {
		return fail(logger, logbuffer, "send admin request: %v", err)
	}
	logger.Printf("Request sent")
	if err := decoder.Decode(&recv); err != nil {
		return fail(logger, logbuffer, "read admin response: %v", err)
	}
	if recv.Status == "error" {
		if recv.Error != "" {
			fmt.Fprintln(os.Stderr, "Admin socket returned an error:", recv.Error)
		} else {
			fmt.Fprintln(os.Stderr, "Admin socket returned an error without details")
		}
		return 1
	}
	if cmdLineEnv.injson {
		output, err := json.MarshalIndent(recv.Response, "", "  ")
		if err != nil {
			return fail(logger, logbuffer, "format admin response as JSON: %v", err)
		}
		fmt.Println(string(output))
		return 0
	}

	opts := []tablewriter.Option{
		tablewriter.WithRowAlignment(tw.AlignLeft),
		tablewriter.WithHeaderAlignment(tw.AlignCenter),
		tablewriter.WithHeaderAutoFormat(tw.Off),
		tablewriter.WithDebug(false),
	}
	if !cmdLineEnv.borders {
		opts = append(opts, tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Borders: tw.BorderNone,
			Settings: tw.Settings{
				Lines:      tw.LinesNone,
				Separators: tw.SeparatorsNone,
			},
		})))
	}
	table := tablewriter.NewTable(os.Stdout, opts...)

	switch strings.ToLower(send.Name) {
	case "list":
		var resp admin.ListResponse
		if err := json.Unmarshal(recv.Response, &resp); err != nil {
			return fail(logger, logbuffer, "decode list response: %v", err)
		}
		table.Header([]string{"Command", "Arguments", "Description"})
		for _, entry := range resp.List {
			for i := range entry.Fields {
				entry.Fields[i] = entry.Fields[i] + "=..."
			}
			_ = table.Append([]string{entry.Command, strings.Join(entry.Fields, ", "), entry.Description})
		}
		_ = table.Render()

	case "getself":
		var resp admin.GetSelfResponse
		if err := json.Unmarshal(recv.Response, &resp); err != nil {
			return fail(logger, logbuffer, "decode getSelf response: %v", err)
		}
		_ = table.Append([]string{"Build name:", resp.BuildName})
		_ = table.Append([]string{"Build version:", resp.BuildVersion})
		_ = table.Append([]string{"IPv6 address:", resp.IPAddress})
		_ = table.Append([]string{"IPv6 subnet:", resp.Subnet})
		_ = table.Append([]string{"Routing table size:", fmt.Sprintf("%d", resp.RoutingEntries)})
		_ = table.Append([]string{"Public key:", resp.PublicKey})
		_ = table.Render()

	case "getpeers":
		var resp admin.GetPeersResponse
		if err := json.Unmarshal(recv.Response, &resp); err != nil {
			return fail(logger, logbuffer, "decode getPeers response: %v", err)
		}
		table.Header([]string{"URI", "State", "Dir", "IP Address", "Uptime", "RTT", "RX", "TX", "Down", "Up", "Pr", "Cost", "Last Error"})
		for _, peer := range resp.Peers {
			state, lasterr, dir, rtt, rxr, txr := "Up", "-", "Out", "-", "-", "-"
			if !peer.Up {
				if state = "Down"; peer.LastError != "" {
					lasterr = fmt.Sprintf("%s ago: %s", peer.LastErrorTime.Round(time.Second), peer.LastError)
				}
			} else if rttms := float64(peer.Latency.Microseconds()) / 1000; rttms > 0 {
				rtt = fmt.Sprintf("%.02fms", rttms)
			}
			if peer.Inbound {
				dir = "In"
			}
			uristring := peer.URI
			if uri, err := url.Parse(peer.URI); err == nil {
				uri.RawQuery = ""
				uristring = uri.String()
			}
			if peer.RXRate > 0 {
				rxr = peer.RXRate.String() + "/s"
			}
			if peer.TXRate > 0 {
				txr = peer.TXRate.String() + "/s"
			}
			_ = table.Append([]string{
				uristring,
				state,
				dir,
				peer.IPAddress,
				(time.Duration(peer.Uptime) * time.Second).String(),
				rtt,
				peer.RXBytes.String(),
				peer.TXBytes.String(),
				rxr,
				txr,
				fmt.Sprintf("%d", peer.Priority),
				fmt.Sprintf("%d", peer.Cost),
				lasterr,
			})
		}
		_ = table.Render()

	case "gettree":
		var resp admin.GetTreeResponse
		if err := json.Unmarshal(recv.Response, &resp); err != nil {
			return fail(logger, logbuffer, "decode getTree response: %v", err)
		}
		table.Header([]string{"Public Key", "IP Address", "Parent", "Sequence"})
		for _, tree := range resp.Tree {
			_ = table.Append([]string{
				tree.PublicKey,
				tree.IPAddress,
				tree.Parent,
				fmt.Sprintf("%d", tree.Sequence),
			})
		}
		_ = table.Render()

	case "getpaths":
		var resp admin.GetPathsResponse
		if err := json.Unmarshal(recv.Response, &resp); err != nil {
			return fail(logger, logbuffer, "decode getPaths response: %v", err)
		}
		table.Header([]string{"Public Key", "IP Address", "Path", "Seq"})
		for _, p := range resp.Paths {
			_ = table.Append([]string{
				p.PublicKey,
				p.IPAddress,
				fmt.Sprintf("%v", p.Path),
				fmt.Sprintf("%d", p.Sequence),
			})
		}
		_ = table.Render()

	case "getsessions":
		var resp admin.GetSessionsResponse
		if err := json.Unmarshal(recv.Response, &resp); err != nil {
			return fail(logger, logbuffer, "decode getSessions response: %v", err)
		}
		table.Header([]string{"Public Key", "IP Address", "Uptime", "RX", "TX"})
		for _, p := range resp.Sessions {
			_ = table.Append([]string{
				p.PublicKey,
				p.IPAddress,
				(time.Duration(p.Uptime) * time.Second).String(),
				p.RXBytes.String(),
				p.TXBytes.String(),
			})
		}
		_ = table.Render()

	case "getnodeinfo":
		var resp core.GetNodeInfoResponse
		if err := json.Unmarshal(recv.Response, &resp); err != nil {
			return fail(logger, logbuffer, "decode getNodeInfo response: %v", err)
		}
		for _, v := range resp {
			fmt.Println(string(v))
			break
		}

	case "getmulticastinterfaces":
		var resp multicast.GetMulticastInterfacesResponse
		if err := json.Unmarshal(recv.Response, &resp); err != nil {
			return fail(logger, logbuffer, "decode getMulticastInterfaces response: %v", err)
		}
		fmtBool := func(b bool) string {
			if b {
				return "Yes"
			}
			return "-"
		}
		table.Header([]string{"Name", "Listen Address", "Beacon", "Listen", "Password"})
		for _, p := range resp.Interfaces {
			_ = table.Append([]string{
				p.Name,
				p.Address,
				fmtBool(p.Beacon),
				fmtBool(p.Listen),
				fmtBool(p.Password),
			})
		}
		_ = table.Render()

	case "gettun":
		var resp tun.GetTUNResponse
		if err := json.Unmarshal(recv.Response, &resp); err != nil {
			return fail(logger, logbuffer, "decode getTUN response: %v", err)
		}
		_ = table.Append([]string{"TUN enabled:", fmt.Sprintf("%#v", resp.Enabled)})
		if resp.Enabled {
			_ = table.Append([]string{"Interface name:", resp.Name})
			_ = table.Append([]string{"Interface MTU:", fmt.Sprintf("%d", resp.MTU)})
		}
		_ = table.Render()

	case "addpeer", "removepeer":

	default:
		fmt.Println(string(recv.Response))
	}

	return 0
}

func dialAdminEndpoint(endpoint string, logger *log.Logger) (net.Conn, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse admin endpoint %q: %w", endpoint, err)
	}

	var network, address string
	switch strings.ToLower(u.Scheme) {
	case "unix":
		network, address = "unix", u.Path
	case "tcp":
		network, address = "tcp", u.Host
	case "":
		network, address = "tcp", endpoint
	default:
		return nil, fmt.Errorf("unsupported admin endpoint scheme %q", u.Scheme)
	}
	if address == "" {
		return nil, fmt.Errorf("admin endpoint %q has no address", endpoint)
	}
	logger.Printf("Connecting to %s endpoint %s", strings.ToUpper(network), address)
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, fmt.Errorf("connect to admin endpoint %q: %w", endpoint, err)
	}
	return conn, nil
}

func fail(logger *log.Logger, logbuffer *bytes.Buffer, format string, args ...interface{}) int {
	logger.Printf("Error: "+format, args...)
	_, _ = fmt.Fprint(os.Stderr, logbuffer.String())
	return 1
}
