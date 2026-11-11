package agentfiles

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gigabytegrove/syncria/agentpolicy"
)

func granted(t *testing.T) (Reader, string, string) {
	t.Helper()
	folder := t.TempDir()
	if err := os.MkdirAll(filepath.Join(folder, "photos"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(folder, "private"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "photos", "a.txt"), []byte("hello sync"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "private", "secret"), []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}
	store, err := agentpolicy.Open(filepath.Join(t.TempDir(), "grants.json"))
	if err != nil {
		t.Fatal(err)
	}
	g, err := store.Add("Photos", folder, true, []string{"photos"}, []string{"photos/excluded"})
	if err != nil {
		t.Fatal(err)
	}
	return Reader{Grants: store}, g.ID, folder
}
func TestReadApprovedFolderAndInventory(t *testing.T) {
	reader, id, _ := granted(t)
	got, err := reader.Inventory(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Path != "photos/a.txt" || got[0].Size != 10 {
		t.Fatalf("wrong inventory: %#v", got)
	}
	var buf bytes.Buffer
	n, err := reader.Copy(context.Background(), id, "photos/a.txt", &buf)
	if err != nil || n != 10 || buf.String() != "hello sync" {
		t.Fatalf("bad read: %d %q %v", n, buf.String(), err)
	}
	for _, path := range []string{"private/secret", "photos/../private/secret", "/etc/passwd", "photos/excluded/x", "photos/../../etc/passwd", "photos/.syncria-internal"} {
		f, _, err := reader.Open(id, path)
		if err == nil {
			f.Close()
			t.Errorf("unexpected access to %q", path)
		}
	}
}
func TestRemovedGrantImmediatelyDeniesAccess(t *testing.T) {
	reader, id, _ := granted(t)
	if err := reader.Grants.Remove(id); err != nil {
		t.Fatal(err)
	}
	if _, _, err := reader.Open(id, "photos/a.txt"); !errors.Is(err, ErrDenied) {
		t.Fatal(err)
	}
	if _, err := reader.Inventory(context.Background(), id); !errors.Is(err, ErrDenied) {
		t.Fatal(err)
	}
}
func TestSymlinkAndPartialInventoryAreRejected(t *testing.T) {
	reader, id, root := granted(t)
	if runtime.GOOS != "linux" && runtime.Version() < "go1.24" {
		t.Skip("secure root API needs Go 1.24")
	}
	if err := os.Symlink(filepath.Join(root, "private", "secret"), filepath.Join(root, "photos", "linked")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, _, err := reader.Open(id, "photos/linked"); err == nil {
		t.Fatal("symlink allowed")
	}
	if items, err := reader.Inventory(context.Background(), id); err == nil || items != nil {
		t.Fatalf("partial inventory returned: %v %v", items, err)
	}
}
func TestCancelInventory(t *testing.T) {
	reader, id, _ := granted(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if out, err := reader.Inventory(ctx, id); err == nil || out != nil {
		t.Fatalf("cancel failed: %v %v", out, err)
	}
}
func TestNoSymlinkRootEscape(t *testing.T) {
	reader, id, root := granted(t)
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "photos", "outside")); err != nil {
		t.Skip(err)
	}
	_, _, err := reader.Open(id, "photos/outside/data")
	if err == nil || strings.TrimSpace(err.Error()) == "" {
		t.Fatalf("symlink escape allowed: %v", err)
	}
}
