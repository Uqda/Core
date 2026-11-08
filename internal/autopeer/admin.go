package autopeer

import (
	"encoding/json"

	"github.com/Uqda/Core/src/admin"
)

// SetupAdminHandlers registers getAutoPeers. m may be nil when the feature
// is disabled, which is reported as such.
func SetupAdminHandlers(a *admin.AdminSocket, m *Manager) {
	_ = a.AddHandler(
		"getAutoPeers", "Show automatic public peer selection and the peers it manages", []string{},
		func(in json.RawMessage) (interface{}, error) {
			return m.Status(), nil
		},
	)
}
