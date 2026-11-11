//go:build linux

package agentpublish

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestSecureRootTraversalRejectsSymlinkedAncestor(t *testing.T) {
	base := t.TempDir()
	real := filepath.Join(base, "real")
	if err := os.MkdirAll(filepath.Join(real, "child"), 0700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(base, "alias")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	if fd, err := openRootByHandle(filepath.Join(link, "child")); err == nil {
		syscall.Close(fd)
		t.Fatal("approved root ancestor symlink followed")
	}
	fd, err := openRootByHandle(filepath.Join(real, "child"))
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Close(fd); err != nil {
		t.Fatal(err)
	}
}
