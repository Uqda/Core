package tun

func (m *TunAdapter) _applyOption(opt SetupOption) {
	switch v := opt.(type) {
	case InterfaceName:
		m.config.name = v
	case InterfaceMTU:
		m.config.mtu = v
	case FileDescriptor:
		m.config.fd = int32(v)
	}
}

// SetupOption configures a TunAdapter during construction.
type SetupOption interface {
	isSetupOption()
}

// InterfaceName selects the operating-system TUN interface.
type InterfaceName string

// InterfaceMTU sets the requested TUN MTU.
type InterfaceMTU uint64

// FileDescriptor supplies an existing TUN descriptor where supported.
type FileDescriptor int32

func (a InterfaceName) isSetupOption()  {}
func (a InterfaceMTU) isSetupOption()   {}
func (a FileDescriptor) isSetupOption() {}
