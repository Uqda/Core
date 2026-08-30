//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package main

import (
	"errors"
	"os"
)

func requireAdministrator() error {
	if os.Geteuid() != 0 {
		return errors.New("administrative access requires root privileges; rerun with sudo (for example: sudo uqdactl getSelf)")
	}
	return nil
}
