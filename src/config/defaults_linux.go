//go:build linux
// +build linux

package config

// Sane defaults for the Linux platform. The "default" options may be
// may be replaced by the running configuration.
func getDefaults() platformDefaultParameters {
	return platformDefaultParameters{
		// Admin
		DefaultAdminListen: "unix:///var/run/uqda.sock",

		// Configuration (used for uqdactl)
		DefaultConfigFile: "/etc/uqda.conf",

		// Multicast interfaces
		DefaultMulticastInterfaces: []MulticastInterfaceConfig{
			{Regex: ".*", Beacon: true, Listen: true},
		},

		// TUN
		MaximumIfMTU:  65535,
		DefaultIfMTU:  65535,
		DefaultIfName: "auto",
	}
}
