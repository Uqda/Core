// Package cli provides plain, portable terminal output for Uqda commands.
package cli

import (
	"fmt"
	"io"
	"strings"
)

const rule = "------------------------------------------------------------"

func Header(w io.Writer, title, subtitle string) {
	fmt.Fprintf(w, "\n  UQDA / %s\n", title)
	if subtitle != "" {
		fmt.Fprintf(w, "  %s\n", subtitle)
	}
	fmt.Fprintf(w, "  %s\n\n", rule)
}

func Field(w io.Writer, label, value string) {
	fmt.Fprintf(w, "  %-14s %s\n", label, value)
}

// Text wraps prose without terminal escape sequences or Unicode decorations.
func Text(w io.Writer, prefix, value string) {
	line := prefix
	for _, word := range strings.Fields(value) {
		if len(line)+len(word)+1 > 78 && line != prefix {
			fmt.Fprintln(w, line)
			line = strings.Repeat(" ", len(prefix))
		}
		if line != prefix && strings.TrimSpace(line) != "" {
			line += " "
		}
		line += word
	}
	fmt.Fprintln(w, line)
}

func Help(w io.Writer, program, release string) {
	Header(w, "COMMAND CENTER", release)
	fmt.Fprintf(w, "  Start here:  %s\n\n", program)
	for _, entry := range [][2]string{
		{program, "Node health and connection status"},
		{program + " peers", "Connected peers and traffic"},
		{program + " test IP", "Test another Uqda IPv6 address"},
		{program + " info", "Node address and public identity"},
		{program + " version", "Installed release"},
	} {
		fmt.Fprintf(w, "  %-20s %s\n", entry[0], entry[1])
	}
	fmt.Fprintf(w, "\n  Automation:  %s status --json\n", program)
	fmt.Fprintf(w, "  Advanced:    %s commands\n\n", program)
	Text(w, "  ", "Use help COMMAND for command details without a running daemon. IP means a numeric Uqda IPv6 address. URI means a reachable transport endpoint, not an overlay address. KEY means a node's hexadecimal public key, never its private key.")
	for _, entry := range commandGuide {
		fmt.Fprintf(w, "\n  %s\n", entry.syntax)
		Text(w, "    ", entry.description)
	}
	fmt.Fprintln(w, "\n  INSTALLATION LIFECYCLE (HELP TOPICS, NOT UQDA OPERATIONS)")
	for _, entry := range lifecycleGuide {
		fmt.Fprintf(w, "\n  %s\n", entry.syntax)
		Text(w, "    ", entry.description)
		guideExamples(w, entry.name)
	}
	fmt.Fprintln(w, "\n  CONTROLLER OPTIONS")
	Text(w, "    ", "--json: raw machine-readable output; it may contain sensitive peer data. --endpoint ADDRESS: select the local admin socket. --borders=false: omit advanced table borders. --version: installed controller version.")
	Text(w, "    ", "With uqda, put the command before its options. With uqdactl, options may appear before or after the command. Advanced operations use uqdactl, not uqda.")
	fmt.Fprintln(w, "\n  DAEMON AND CONFIGURATION")
	for _, entry := range daemonGuide {
		fmt.Fprintf(w, "  %s\n", entry[0])
		Text(w, "    ", entry[1])
	}
	fmt.Fprintln(w, "\n  SAFETY AND RESULTS")
	Text(w, "    ", "Bare uqda inspects the existing service; it does not start a daemon. Admin operations require a running daemon and permission to access its socket. Never expose the unauthenticated admin API publicly.")
	Text(w, "    ", "Private groups require the same strong GroupPassword and compatible versions on every member. Incoming peer allowlists must also permit the new node's public key. Peering alone does not prove encrypted session reachability.")
	Text(w, "    ", "Exit codes: 0 success, 1 operational failure, 2 invalid usage. A health report may contain warnings even with exit code 0. Read the findings. Use commands to discover the operations actually advertised by your daemon.")
}

type guideEntry struct {
	name, syntax, description string
}

var commandGuide = []guideEntry{
	{"status", "uqda [status|doctor]", "Read-only health checks: daemon, identity, peers and TUN interface. status and doctor are aliases. Example: uqda status --json."},
	{"peers", "uqda peers", "Show direct connections, traffic and latency. Example: uqda peers --json. The raw admin equivalent is uqdactl getPeers; sort=uptime or sort=cost changes ordering."},
	{"info", "uqda info", "Show your IPv6 address, subnet and public key. The admin equivalent is uqdactl getSelf."},
	{"test", "uqda test IP [count=N] [idle=DURATION]", "Probe overlay reachability. count is 1..20 (default 5); idle is 0s..5m. Example: uqda test 200:1234::1 count=10 idle=75s. Use count=10, not --count=10. Requires TUN and an OS ping utility. Timings include process startup, not just network RTT."},
	{"version", "uqda version", "Show the installed release without contacting the daemon. uqdactl version shows its own installed release."},
	{"commands", "uqda commands / uqdactl list", "Discover the running daemon's supported operations and parameter names. Custom or disabled components may change this list."},
	{"addPeer", "uqdactl addPeer uri=URI [interface=NAME]", "Add an outbound peer and retry it if its link drops. Example: uqdactl addPeer uri=tcp://203.0.113.10:33443. This changes runtime state, not the configuration file. For persistence, also add the URI to Peers in your existing uqda.conf. Do not replace your identity."},
	{"removePeer", "uqdactl removePeer uri=URI [interface=NAME]", "Remove a runtime peer using the same URI and interface. Example: uqdactl removePeer uri=tcp://203.0.113.10:33443. Also remove it from Peers in your configuration to keep it removed after restart. This is not a firewall or an immediate disconnection guarantee."},
	{"getTree", "uqdactl getTree", "Read known routing-tree entries."},
	{"getPaths", "uqdactl getPaths", "Read established paths through this node."},
	{"getSessions", "uqdactl getSessions", "Read established encrypted traffic sessions with remote nodes."},
	{"getTun", "uqdactl getTun", "Read virtual network interface information; this does not create or enable TUN."},
	{"getMulticastInterfaces", "uqdactl getMulticastInterfaces", "Read local discovery interface information; this does not enable multicast."},
	{"getNodeInfo", "uqdactl getNodeInfo key=KEY", "Request published node information from a reachable remote node. Replace KEY with its public key. Remote privacy settings or reachability may prevent a response."},
	{"debug_remoteGetSelf", "uqdactl debug_remoteGetSelf key=KEY", "Advanced remote-node identity diagnostics. Requires a reachable, responding node."},
	{"debug_remoteGetPeers", "uqdactl debug_remoteGetPeers key=KEY", "Advanced remote peer diagnostics. Requires a reachable, responding node."},
	{"debug_remoteGetTree", "uqdactl debug_remoteGetTree key=KEY", "Advanced remote routing-tree diagnostics. Requires a reachable, responding node."},
	{"help", "uqda help [COMMAND] / uqda help advanced", "Show this complete guide, command details, or all daemon flag defaults. No daemon connection is needed."},
}

var daemonGuide = [][2]string{
	{"uqda -useconffile PATH / uqda -useconf", "Start a daemon from an existing file or stdin. Do not start a second daemon over an installed service."},
	{"uqda -genconf [-json]", "Generate NEW configuration and identity. Never redirect it over an existing configuration. Output contains a private key."},
	{"uqda -useconffile PATH -checkconf", "Validate configuration without starting a daemon."},
	{"uqda -useconffile PATH -normaliseconf [-json]", "Print normalized configuration. Output may contain secrets; this does not save changes."},
	{"uqda -useconffile PATH -address / -subnet / -publickey", "Print one identity field from the existing configuration."},
	{"uqda -useconffile PATH -exportkey", "Export the SECRET private key in PEM format. Do not share it or post it in logs."},
	{"uqda -autoconf", "Start with an ephemeral identity and automatic local peering; not a persistent configured service."},
	{"uqda -version", "Show the installed daemon release; equivalent to uqda version."},
	{"-loglevel LEVEL / -logto DESTINATION", "Daemon logging level and destination. See uqda help advanced for defaults."},
	{"-user USER[:GROUP] / -notifyfd NUMBER", "Unix privilege switching and service-manager readiness integration; not ordinary peer-management commands."},
}

var lifecycleGuide = []guideEntry{
	{"install", "uqda help install", "macOS: brew install --cask Uqda/tap/uqda. Linux/systemd: download contrib/install/linux.sh from Uqda/Core on GitHub, then sudo sh uqda-install.sh install. Windows: install the correct official MSI for your architecture. Verify release SHA256SUMS. Installers are unsigned. Never copy one node's private identity to a second live node."},
	{"update", "uqda help update", "macOS: brew update && brew upgrade --cask Uqda/tap/uqda. Linux quick installation: sudo sh uqda-install.sh update using the current official script. Windows: install the newer MSI. Existing identity is retained. Private groups migrating from 26.0.1 or earlier must update every member together."},
	{"uninstall", "uqda help uninstall", "macOS: brew uninstall --cask Uqda/tap/uqda. Linux quick installation: sudo sh uqda-install.sh uninstall. Windows: remove Uqda through Installed Apps. Normal removal retains configuration and identity. Homebrew --zap or Linux --purge --yes permanently removes identity too; back up protected keys first. Do not run the quick installer over a package-managed installation."},
}

func HelpTopic(w io.Writer, program, release, topic string) bool {
	aliases := map[string]string{"doctor": "status", "getself": "info", "getpeers": "peers", "list": "commands"}
	if name, ok := aliases[strings.ToLower(topic)]; ok {
		topic = name
	}
	for _, entry := range append(append([]guideEntry{}, commandGuide...), lifecycleGuide...) {
		if strings.EqualFold(entry.name, topic) {
			Header(w, "COMMAND HELP", release)
			fmt.Fprintf(w, "  %s\n", entry.syntax)
			Text(w, "    ", entry.description)
			guideExamples(w, entry.name)
			fmt.Fprintf(w, "\n  Complete guide: %s help\n\n", program)
			return true
		}
	}
	return false
}

func guideExamples(w io.Writer, topic string) {
	if topic == "install" {
		fmt.Fprintln(w, "\n    Linux script download (sh, not CMD):")
		fmt.Fprintln(w, "    curl -fsSLo uqda-install.sh \\")
		fmt.Fprintln(w, "      https://raw.githubusercontent.com/Uqda/Core/main/contrib/install/linux.sh")
		fmt.Fprintln(w, "    sudo sh uqda-install.sh install")
		Text(w, "    ", "Read the script before running it as root. It selects the newest published release, including prereleases, and checks its archive against that release's SHA256SUMS.")
	}
}
