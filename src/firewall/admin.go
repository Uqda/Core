package firewall

import (
	"encoding/json"

	"github.com/Uqda/Core/src/admin"
)

// Source returns the firewall currently in effect, which may be nil.
type Source func() *Firewall

// SetupAdminHandlers registers getFirewall, which reports the firewall state.
func SetupAdminHandlers(a *admin.AdminSocket, current Source) {
	_ = a.AddHandler(
		"getFirewall", "Show the inbound firewall state, rules and counters", []string{},
		func(in json.RawMessage) (interface{}, error) {
			var req struct{}
			if err := json.Unmarshal(in, &req); err != nil {
				return nil, err
			}
			return current().Status(), nil
		},
	)
}
