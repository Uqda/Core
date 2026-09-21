package multicast

import "regexp"

func (m *Multicast) _applyOption(opt SetupOption) {
	switch v := opt.(type) {
	case MulticastInterface:
		m.config._interfaces[v] = struct{}{}
	case GroupAddress:
		m.config._groupAddr = v
	}
}

// SetupOption configures multicast discovery during construction.
type SetupOption interface {
	isSetupOption()
}

// MulticastInterface controls discovery on interfaces matching Regex.
type MulticastInterface struct {
	Regex    *regexp.Regexp
	Beacon   bool
	Listen   bool
	Port     uint16
	Priority uint8
	Password string
}

// GroupAddress sets the IPv6 multicast discovery group.
type GroupAddress string

func (a MulticastInterface) isSetupOption() {}
func (a GroupAddress) isSetupOption()       {}
