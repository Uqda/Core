//go:build windows

package main

import (
	"errors"

	"golang.org/x/sys/windows"
)

func requireAdministrator() error {
	if !windows.GetCurrentProcessToken().IsElevated() {
		return errors.New("administrative access requires an elevated terminal; rerun Command Prompt or PowerShell as Administrator")
	}
	return nil
}
