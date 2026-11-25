//go:build !unix

package main

// clearUmask is a no-op on platforms that have no process umask, such as
// Windows, so the tool can still be built and run there.
func clearUmask() (restore func()) {
	return func() {}
}
