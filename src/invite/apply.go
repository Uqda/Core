package invite

import (
	"github.com/Uqda/Core/src/config"
)

// Apply merges a grant into cfg and returns a human-readable list of what
// changed. The node's identity is never touched.
func Apply(cfg *config.NodeConfig, g Grant) []string {
	var changes []string
	if g.GroupPassword != cfg.GroupPassword {
		switch {
		case g.GroupPassword == "":
			changes = append(changes, "joined a public network (no group password)")
		case cfg.GroupPassword == "":
			changes = append(changes, "joined the private group")
		default:
			changes = append(changes, "replaced the previous group password with the inviter's group")
		}
		cfg.GroupPassword = g.GroupPassword
	}
	known := make(map[string]bool, len(cfg.Peers))
	for _, p := range cfg.Peers {
		known[p] = true
	}
	added := 0
	for _, p := range g.Peers {
		if p == "" || known[p] {
			continue
		}
		known[p] = true
		cfg.Peers = append(cfg.Peers, p)
		added++
	}
	if added > 0 {
		changes = append(changes, plural(added, "peer")+" added")
	}
	return changes
}

func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	return itoa(n) + " " + word + "s"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
