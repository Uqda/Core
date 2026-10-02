package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Uqda/Core/src/admin"
	"github.com/Uqda/Core/src/tun"
)

// A doctor report contains only local health findings, never configuration,
// peer URIs, error strings from peers, or private key material.
type doctorCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail"`
	Hint   string `json:"hint,omitempty"`
}

type doctorReport struct {
	Checks []doctorCheck `json:"checks"`
}

func isDoctorCommand(command string) bool {
	switch strings.ToLower(command) {
	case "doctor", "status":
		return true
	default:
		return false
	}
}

func runDoctor(endpoint string, inJSON bool) int {
	report := diagnoseNode(endpoint)
	if inJSON {
		encoded, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not format doctor report:", err)
			return 1
		}
		fmt.Println(string(encoded))
	} else {
		renderDoctor(os.Stdout, report)
	}
	for _, check := range report.Checks {
		if check.Status == "fail" {
			return 1
		}
	}
	return 0
}

func diagnoseNode(endpoint string) doctorReport {
	report := doctorReport{}
	if u, err := url.Parse(endpoint); err == nil && strings.EqualFold(u.Scheme, "tcp") {
		host := u.Hostname()
		if ip := net.ParseIP(host); host != "localhost" && (ip == nil || !ip.IsLoopback()) {
			report.Checks = append(report.Checks, doctorCheck{
				Name: "Admin", Status: "warn", Detail: "The admin endpoint is not a loopback address.",
				Hint: "The admin API has no authentication; bind it to a private local socket or protect it externally.",
			})
		}
	}
	var self admin.GetSelfResponse
	if err := doctorSelfWithRetry(endpoint, &self); err != nil {
		report.Checks = append(report.Checks, doctorCheck{
			Name: "Daemon", Status: "fail", Detail: "Cannot reach the admin API.",
			Hint: "Check that Uqda is running and that the local admin socket is accessible; use -endpoint=... for a custom socket.",
		})
		return report
	}
	report.Checks = append(report.Checks, doctorCheck{Name: "Daemon", Status: "pass", Detail: "Admin API is responding."})
	key, keyErr := hex.DecodeString(self.PublicKey)
	ip := net.ParseIP(self.IPAddress)
	if keyErr != nil || len(key) != 32 || ip == nil || ip.To4() != nil {
		report.Checks = append(report.Checks, doctorCheck{
			Name: "Identity", Status: "fail", Detail: "The node did not report a valid identity and IPv6 address.",
			Hint: "Check the daemon log and validate the configuration with uqda -useconffile PATH -checkconf.",
		})
	} else {
		report.Checks = append(report.Checks, doctorCheck{Name: "Identity", Status: "pass", Detail: "Node identity and IP address are available."})
	}

	var peers admin.GetPeersResponse
	if err := doctorRequest(endpoint, "getPeers", &peers); err != nil {
		report.Checks = append(report.Checks, doctorCheck{
			Name: "Peers", Status: "fail", Detail: "Could not read peer status.",
			Hint: "Run uqdactl getPeers and inspect the daemon log.",
		})
	} else {
		up := 0
		for _, peer := range peers.Peers {
			if peer.Up {
				up++
			}
		}
		if up == 0 {
			report.Checks = append(report.Checks, doctorCheck{
				Name: "Peers", Status: "warn", Detail: "No connected peers.",
				Hint: "Check your peer URLs, network access and local multicast settings; run uqdactl getPeers for details.",
			})
		} else {
			report.Checks = append(report.Checks, doctorCheck{Name: "Peers", Status: "pass", Detail: fmt.Sprintf("%d connected peer(s).", up)})
		}
	}

	var interfaceState tun.GetTUNResponse
	if err := doctorRequest(endpoint, "getTun", &interfaceState); err != nil {
		report.Checks = append(report.Checks, doctorCheck{
			Name: "Interface", Status: "warn", Detail: "Could not read the TUN interface state.",
			Hint: "Run uqdactl getTun and inspect the daemon log.",
		})
	} else if !interfaceState.Enabled {
		report.Checks = append(report.Checks, doctorCheck{
			Name: "Interface", Status: "warn", Detail: "TUN is disabled; this may be intentional for a router-only node.",
			Hint: "If this host needs an Uqda IP interface, check IfName in uqda.conf.",
		})
	} else {
		report.Checks = append(report.Checks, doctorCheck{Name: "Interface", Status: "pass", Detail: "TUN is enabled."})
	}
	return report
}

// A service manager can report the daemon as active before its admin socket is
// ready. Wait briefly for that startup window, but do not wait indefinitely.
func doctorSelfWithRetry(endpoint string, self *admin.GetSelfResponse) error {
	deadline := time.Now().Add(5 * time.Second)
	for {
		err := doctorRequest(endpoint, "getSelf", self)
		if err == nil || !time.Now().Before(deadline) {
			return err
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func doctorRequest(endpoint, name string, result interface{}) error {
	conn, err := dialAdminEndpoint(endpoint, log.New(io.Discard, "", 0))
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return err
	}
	request := admin.AdminSocketRequest{Name: name, Arguments: json.RawMessage("{}")}
	if err := json.NewEncoder(conn).Encode(request); err != nil {
		return err
	}
	var response admin.AdminSocketResponse
	if err := json.NewDecoder(conn).Decode(&response); err != nil {
		return err
	}
	if response.Status != "success" {
		return fmt.Errorf("admin request failed")
	}
	return json.Unmarshal(response.Response, result)
}
