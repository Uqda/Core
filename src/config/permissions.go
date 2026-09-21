//go:build !mobile

package config

import "os"

// FilePermissionsAreUnsafe reports whether the file at path appears to
// grant read or write access to users other than its owner - the config
// file and any PrivateKeyPath file are exactly the kind of sensitive file
// this matters for, since both can contain (or point at) a node's private
// key.
//
// ok is false when this platform's permission model can't be meaningfully
// checked this way (see permissions_windows.go); callers must not emit a
// warning when ok is false, since doing so would be a false positive
// based on a Unix assumption that doesn't hold on that platform.
func FilePermissionsAreUnsafe(path string) (unsafe bool, ok bool, err error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, false, err
	}
	return filePermissionsAreUnsafe(info.Mode()), permissionCheckMeaningful, nil
}
