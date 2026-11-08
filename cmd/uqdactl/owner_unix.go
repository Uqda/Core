//go:build !windows

package main

import (
	"os"
	"syscall"
)

// preserveOwner gives f the owner and group of an existing file.
func preserveOwner(f *os.File, existing os.FileInfo) error {
	st, ok := existing.Sys().(*syscall.Stat_t)
	if !ok {
		return nil
	}
	if int(st.Uid) == os.Getuid() && int(st.Gid) == os.Getgid() {
		return nil
	}
	return f.Chown(int(st.Uid), int(st.Gid))
}
