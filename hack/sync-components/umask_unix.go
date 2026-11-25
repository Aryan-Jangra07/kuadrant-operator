//go:build unix

package main

import "syscall"

// clearUmask sets the process umask to 0 and returns a function that
// restores the previous value.
func clearUmask() (restore func()) {
	old := syscall.Umask(0)
	return func() { syscall.Umask(old) }
}
