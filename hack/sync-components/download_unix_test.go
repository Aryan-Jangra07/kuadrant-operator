//go:build unix

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// TestExtractChart_PermissionsUnderRestrictiveUmask verifies extracted files
// and directories get their intended permissions even when the process umask
// would otherwise mask them out. os.OpenFile/os.MkdirAll's mode argument is
// ANDed with ~umask by the OS, so extractChart zeroes the umask for the
// duration of extraction rather than relying on the creation mode alone.
// Includes a deeply nested file with no explicit tar directory entries for
// any of its ancestors, since those are created purely as a side effect of
// os.MkdirAll and are just as exposed to the umask as anything named directly
// in the tarball.
func TestExtractChart_PermissionsUnderRestrictiveUmask(t *testing.T) {
	old := syscall.Umask(0o077)
	defer syscall.Umask(old)

	files := map[string]string{
		"Kuadrant-dns-operator-abc1234/charts/dns-operator/Chart.yaml":                       "apiVersion: v2\nname: dns-operator",
		"Kuadrant-dns-operator-abc1234/charts/dns-operator/templates/manifests.yaml":         "kind: Deployment",
		"Kuadrant-dns-operator-abc1234/charts/dns-operator/templates/tests/nested/deep.yaml": "kind: Test",
	}
	tarball := buildTestTarball(t, files)
	outputDir := t.TempDir()

	if err := extractChart(bytes.NewReader(tarball), "charts/dns-operator", outputDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, relPath := range []string{"Chart.yaml", "templates/manifests.yaml", "templates/tests/nested/deep.yaml"} {
		info, err := os.Stat(filepath.Join(outputDir, relPath))
		if err != nil {
			t.Fatalf("stat %s: %v", relPath, err)
		}
		if got := info.Mode().Perm(); got != 0o644 {
			t.Errorf("file %s permission = %v, want 0644 (umask should not affect this)", relPath, got)
		}
	}

	// outputDir itself (".") is created by the caller before extractChart
	// runs (os.MkdirTemp in the real cmd_sync.go flow, always 0700 regardless
	// of umask) -- fixing it is the caller's responsibility, not
	// extractChart's, so it's deliberately not asserted on here.
	for _, relPath := range []string{"templates", "templates/tests", "templates/tests/nested"} {
		info, err := os.Stat(filepath.Join(outputDir, relPath))
		if err != nil {
			t.Fatalf("stat %s: %v", relPath, err)
		}
		if got := info.Mode().Perm(); got != 0o755 {
			t.Errorf("dir %s permission = %v, want 0755 (umask should not affect this, including ancestors with no tar entry of their own)", relPath, got)
		}
	}
}
