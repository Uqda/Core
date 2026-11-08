//go:build windows

package main

import "os"

// preserveOwner is a no-op on Windows, where the installer's directory ACLs
// apply to files created inside it.
func preserveOwner(*os.File, os.FileInfo) error { return nil }
