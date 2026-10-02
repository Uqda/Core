package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"suah.dev/protect"

	"github.com/Uqda/Core/internal/cli"
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

	cmdLineEnv := newCmdLineEnv()
	var flagOutput bytes.Buffer
	if err := cmdLineEnv.parseFlagsAndArgs(os.Args[1:], &flagOutput); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(os.Stdout, flagOutput.String())
			return 0
		}
		fmt.Fprint(os.Stderr, flagOutput.String())
		fmt.Fprintln(os.Stderr, "Uqda:", err)
		return 2
	}
	pledge := "stdio rpath inet unix dns"
	if len(cmdLineEnv.args) > 0 && strings.EqualFold(cmdLineEnv.args[0], "test") {
		pledge += " proc exec"
	}
	if err := protect.Pledge(pledge); err != nil {
		return fail(logger, logbuffer, "apply initial process restrictions: %v", err)
	}

	if cmdLineEnv.ver || (len(cmdLineEnv.args) == 1 && strings.EqualFold(cmdLineEnv.args[0], "version")) {
		fmt.Println(version.DisplayName())
		return 0
	}

	if len(cmdLineEnv.args) > 0 && strings.EqualFold(cmdLineEnv.args[0], "find") {
		return runFind(cmdLineEnv.args[1:], cmdLineEnv.injson)
	}
	if err := cmdLineEnv.setEndpoint(logger); err != nil {
		return fail(logger, logbuffer, "%v", err)
	}
	if len(cmdLineEnv.args) == 1 && isDoctorCommand(cmdLineEnv.args[0]) {
		return runDoctor(cmdLineEnv.endpoint, cmdLineEnv.injson)
	}
	if len(cmdLineEnv.args) > 0 && strings.EqualFold(cmdLineEnv.args[0], "test") {
		return runNetworkTest(cmdLineEnv.endpoint, cmdLineEnv.args[1:], cmdLineEnv.injson)
	}
	var available admin.ListResponse
	if err := doctorRequest(cmdLineEnv.endpoint, "list", &available); err != nil {
		if isAdminAccessError(err) {
			return fail(logger, logbuffer, "%v.\n  Next: %s", err, adminAccessHint(cmdLineEnv.args[0]))
		}
		return fail(logger, logbuffer, "cannot read available commands: %v; run 'uqda' to check node health", err)
	}
	if err := validateAdminArguments(cmdLineEnv.args, available); err != nil {
		fmt.Fprintln(os.Stderr, "Uqda:", err)
		return 2
	}

	conn, err := dialAdminEndpoint(cmdLineEnv.endpoint, logger)
	if err != nil {
		if isAdminAccessError(err) {
			return fail(logger, logbuffer, "%v.\n  Next: %s", err, adminAccessHint(cmdLineEnv.args[0]))
		}
		return fail(logger, logbuffer, "%v", err)
	}
	if err := conn.SetDeadline(time.Now().Add(30 * time.Second)); err != nil {
		_ = conn.Close()
		return fail(logger, logbuffer, "set admin request deadline: %v", err)
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
			logger.Printf("Sending request: %v\n", a)
			send.Name = a
			continue
		}
		tokens := strings.SplitN(a, "=", 2)
		args[tokens[0]] = tokens[1]
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
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{Symbols: tw.NewSymbols(tw.StyleASCII)})),
		tablewriter.WithRowAlignment(tw.AlignLeft),
		tablewriter.WithHeaderAlignment(tw.AlignCenter),
		tablewriter.WithHeaderAutoFormat(tw.Off),
		tablewriter.WithDebug(false),
	}
	if !cmdLineEnv.borders {
		opts = append(opts, tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Borders: tw.BorderNone,
			Symbols: tw.NewSymbols(tw.StyleASCII),
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
		renderCommands(os.Stdout, resp)

	case "getself":
		var resp admin.GetSelfResponse
		if err := json.Unmarshal(recv.Response, &resp); err != nil {
			return fail(logger, logbuffer, "decode getSelf response: %v", err)
		}
		cli.Header(os.Stdout, "NODE IDENTITY", resp.BuildName+" "+resp.BuildVersion)
		cli.Field(os.Stdout, "Address", resp.IPAddress)
		cli.Field(os.Stdout, "Subnet", resp.Subnet)
		cli.Field(os.Stdout, "Routes", fmt.Sprintf("%d", resp.RoutingEntries))
		fmt.Fprintln(os.Stdout, "  Public key")
		cli.Text(os.Stdout, "    ", resp.PublicKey)
		fmt.Println()

	case "getpeers":
		var resp admin.GetPeersResponse
		if err := json.Unmarshal(recv.Response, &resp); err != nil {
			return fail(logger, logbuffer, "decode getPeers response: %v", err)
		}
		renderPeers(os.Stdout, resp)

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
		keys := make([]string, 0, len(resp))
		for key := range resp {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Printf("%s: %s\n", key, resp[key])
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
		cli.Header(os.Stdout, "PEER UPDATED", version.DisplayName())
		if strings.EqualFold(send.Name, "addpeer") {
			fmt.Println("  [PASS] Peer added. Run 'uqda peers' to check its connection.")
		} else {
			fmt.Println("  [PASS] Peer removed.")
		}

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
	conn, err := net.DialTimeout(network, address, 5*time.Second)
	if err != nil {
		if network == "unix" && errors.Is(err, os.ErrPermission) {
			return nil, &adminAccessError{cause: err}
		}
		return nil, fmt.Errorf("connect to admin endpoint %q: %w", endpoint, err)
	}
	return conn, nil
}

func fail(logger *log.Logger, logbuffer *bytes.Buffer, format string, args ...interface{}) int {
	fmt.Fprintf(os.Stderr, "Uqda: "+format+"\n", args...)
	return 1
}
