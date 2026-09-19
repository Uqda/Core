//go:build !windows

package config

import "os"

// permissionCheckMeaningful is true on platforms where os.FileMode's
// permission bits reflect real Unix-style file permissions.
const permissionCheckMeaningful = true

// filePermissionsAreUnsafe reports whether mode grants any access
// (read, write, or execute) to group or other - a sensitive file such as
// a private key or a config file embedding one should be readable only by
// its owner.
func filePermissionsAreUnsafe(mode os.FileMode) bool {
	return mode.Perm()&0o077 != 0
}
