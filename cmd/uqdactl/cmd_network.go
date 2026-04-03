package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/Uqda/Core/src/admin"
)

type multiString []string

func (m *multiString) String() string { return strings.Join(*m, ",") }

func (m *multiString) Set(v string) error {
	*m = append(*m, v)
	return nil
}

// runNetworkCLI handles "uqdactl network ..." after the admin connection is open.
func runNetworkCLI(conn net.Conn, env *CmdLineEnv) int {
	args := env.args[1:] // after "network"
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: network <create|join|list|leave> [options]")
		return 1
	}
	sub := strings.ToLower(strings.TrimSpace(args[0]))
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	send := func(name string, arguments interface{}) (*admin.AdminSocketResponse, error) {
		raw, err := json.Marshal(arguments)
		if err != nil {
			return nil, err
		}
		req := admin.AdminSocketRequest{Name: name, Arguments: raw, Auth: env.adminAuth}
		if err := encoder.Encode(&req); err != nil {
			return nil, err
		}
		var recv admin.AdminSocketResponse
		if err := decoder.Decode(&recv); err != nil {
			return nil, err
		}
		return &recv, nil
	}

	switch sub {
	case "create":
		fs := flag.NewFlagSet("network create", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		name := fs.String("name", "", "network name")
		var peers multiString
		fs.Var(&peers, "peer", "bootstrap peer URI (repeatable)")
		expires := fs.Int("expires", 24, "invite validity (hours)")
		adm := fs.String("admin", "", "optional tcp://host:port included in token for inviteRegister")
		if err := fs.Parse(args[1:]); err != nil {
			return 1
		}
		if strings.TrimSpace(*name) == "" {
			fmt.Fprintln(os.Stderr, "network create: -name is required")
			return 1
		}
		if len(peers) == 0 {
			fmt.Fprintln(os.Stderr, "network create: at least one -peer URI is required")
			return 1
		}
		recv, err := send("createNetwork", admin.CreateNetworkRequest{
			Name:         *name,
			Peers:        []string(peers),
			ExpiresHours: *expires,
			Admin:        strings.TrimSpace(*adm),
		})
		if err != nil {
			panic(err)
		}
		if recv.Status == "error" {
			fmt.Fprintln(os.Stderr, recv.Error)
			return 1
		}
		var out admin.CreateNetworkResponse
		if err := json.Unmarshal(recv.Response, &out); err != nil {
			panic(err)
		}
		fmt.Printf("Network %q created.\n\nInvite token:\n\n  %s\n\nShare this with devices you want to join.\n", out.Network, out.Token)
		if out.Warning != "" {
			fmt.Printf("\nNote: %s\n", out.Warning)
		}
		return 0

	case "join":
		fs := flag.NewFlagSet("network join", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		ownerAuth := fs.String("register-auth", "", "owner AdminAuth if inviteRegister is used")
		if err := fs.Parse(args[1:]); err != nil {
			return 1
		}
		tok := strings.TrimSpace(strings.Join(fs.Args(), " "))
		if tok == "" {
			fmt.Fprintln(os.Stderr, "usage: network join [flags] <uqda-invite-v1-...>")
			return 1
		}
		recv, err := send("joinNetwork", admin.JoinNetworkRequest{
			Token:          tok,
			OwnerAdminAuth: strings.TrimSpace(*ownerAuth),
		})
		if err != nil {
			panic(err)
		}
		if recv.Status == "error" {
			fmt.Fprintln(os.Stderr, recv.Error)
			return 1
		}
		var out admin.JoinNetworkResponse
		if err := json.Unmarshal(recv.Response, &out); err != nil {
			panic(err)
		}
		fmt.Printf("Joined private network %q.\n", out.Network)
		if out.Registered {
			fmt.Println("Owner was notified (inviteRegister).")
		}
		if out.Message != "" {
			fmt.Println(out.Message)
		}
		return 0

	case "list":
		recv, err := send("listNetworks", map[string]string{})
		if err != nil {
			panic(err)
		}
		if recv.Status == "error" {
			fmt.Fprintln(os.Stderr, recv.Error)
			return 1
		}
		var out admin.ListNetworksResponse
		if err := json.Unmarshal(recv.Response, &out); err != nil {
			panic(err)
		}
		if env.injson {
			b, _ := json.MarshalIndent(out.Networks, "", "  ")
			fmt.Println(string(b))
			return 0
		}
		if len(out.Networks) == 0 {
			fmt.Println("No private networks configured.")
			return 0
		}
		for _, n := range out.Networks {
			fmt.Printf("- %s (owner=%v) peers=%d keys=%d\n", n.Name, n.IsOwner, len(n.Peers), len(n.AllowedKeys))
		}
		return 0

	case "leave":
		fs := flag.NewFlagSet("network leave", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		name := fs.String("name", "", "network name")
		if err := fs.Parse(args[1:]); err != nil {
			return 1
		}
		if strings.TrimSpace(*name) == "" {
			fmt.Fprintln(os.Stderr, "network leave: -name is required")
			return 1
		}
		recv, err := send("leaveNetwork", admin.LeaveNetworkRequest{Name: strings.TrimSpace(*name)})
		if err != nil {
			panic(err)
		}
		if recv.Status == "error" {
			fmt.Fprintln(os.Stderr, recv.Error)
			return 1
		}
		fmt.Printf("Left private network %q.\n", *name)
		return 0

	default:
		fmt.Fprintf(os.Stderr, "unknown network subcommand %q\n", sub)
		return 1
	}
}
