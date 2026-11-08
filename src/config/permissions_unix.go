//go:build !windows

package config

import "os"

// permissionCheckMeaningful is true on platforms where os.FileMode's
// permission bits reflect real Unix-style file permissions.
const permissionCheckMeaningful = true

// filePermissionsAreUnsafe reports whether mode grants access to other
// users, or lets the group change or execute the file. Group read alone is
// allowed: packages store uqda.conf as root:uqda 0640 so that the service's
// dedicated group can read it, the same pattern as /etc/shadow.
func filePermissionsAreUnsafe(mode os.FileMode) bool {
	return mode.Perm()&0o037 != 0
}
