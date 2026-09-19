//go:build windows

package config

import "os"

// permissionCheckMeaningful is false on Windows: os.FileMode there is
// synthesized from the read-only attribute, not from real per-user ACLs,
// so applying the Unix "group/other bits" check would produce misleading
// warnings (e.g. a file could be ACL-restricted to the owner alone while
// still reporting permissive-looking Unix mode bits, or vice versa).
// Meaningful permission auditing on Windows would require inspecting the
// file's actual ACL, which is out of scope here - so callers must treat
// every check on this platform as "not checked" rather than "safe".
const permissionCheckMeaningful = false

func filePermissionsAreUnsafe(mode os.FileMode) bool {
	return false
}
