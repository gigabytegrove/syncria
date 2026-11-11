package agentrestore

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gigabytegrove/syncria/agentpolicy"
	"github.com/gigabytegrove/syncria/agentpublish"
	"github.com/gigabytegrove/syncria/agentversion"
)

const testJob = "aabbccddeeff00112233445566778899"

type fixture struct {
	root       string
	grant      string
	grants     *agentpolicy.Store
	versions   *agentversion.Store
	archiveDir string
}

func setup(t *testing.T, readOnly bool) fixture {
	t.Helper()
	root := filepath.Join(t.TempDir(), "files")
	if err := os.MkdirAll(filepath.Join(root, "photos"), 0700); err != nil {
		t.Fatal(err)
	}
	grants, err := agentpolicy.Open(filepath.Join(t.TempDir(), "grants.json"))
	if err != nil {
		t.Fatal(err)
	}
	g, err := grants.Add("Photos", root, readOnly, []string{"photos"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	archiveDir := filepath.Join(t.TempDir(), "archive")
	versionStore, err := agentversion.Open(archiveDir, 8<<20)
	if err != nil {
		t.Fatal(err)
	}
	return fixture{root: root, grant: g.ID, grants: grants, versions: versionStore, archiveDir: archiveDir}
}
func (f fixture) manager() Manager { return Manager{Grants: f.grants, Versions: f.versions} }

func TestLocalVersionSaveRestoreNoOverwrite(t *testing.T) {
	if testing.Short() {
		t.Skip("filesystem test")
	}
	f := setup(t, false)
	src := filepath.Join(f.root, "photos", "holiday.txt")
	old := bytes.Repeat([]byte("old photo list"), 10000)
	if err := os.WriteFile(src, old, 0600); err != nil {
		t.Fatal(err)
	}
	v, err := f.manager().Save(context.Background(), testJob, f.grant, "photos/holiday.txt")
	if err != nil {
		t.Fatal(err)
	}
	if len(v.SHA256) != 64 || v.Bytes != int64(len(old)) {
		t.Fatalf("inaccurate snapshot: %+v", v)
	}
	if err := os.WriteFile(src, []byte("new important changes"), 0600); err != nil {
		t.Fatal(err)
	}
	// Files are always restored to a NEW name, never overwriting the live version.
	result, err := f.manager().Restore(context.Background(), v.ID, f.grant)
	if err != nil {
		t.Fatal(err)
	}
	wantName, _ := RecoveredName("photos/holiday.txt", v.ID)
	if result.Path != wantName {
		t.Fatalf("wrong restore destination %q", result.Path)
	}
	got, err := os.ReadFile(filepath.Join(f.root, filepath.FromSlash(result.Path)))
	if err != nil || !bytes.Equal(got, old) {
		t.Fatalf("restored content mismatch %v", err)
	}
	live, err := os.ReadFile(src)
	if err != nil || string(live) != "new important changes" {
		t.Fatal("live file was modified")
	}
	if _, err := f.manager().Restore(context.Background(), v.ID, f.grant); !errors.Is(err, agentpublish.ErrExists) {
		t.Fatalf("existing recovered copy replaced: %v", err)
	}
	// Restart: the manifest and exact original content still exist.
	reopened, err := agentversion.Open(f.versionsDirectory(), 8<<20)
	if err != nil {
		t.Fatal(err)
	}
	if len(reopened.List(testJob, "photos/holiday.txt")) != 1 {
		t.Fatal("history lost across restart")
	}
}
func (f fixture) versionsDirectory() string { return f.archiveDir }

func TestReadOnlyAndRevokedGrants(t *testing.T) {
	f := setup(t, true)
	if err := os.WriteFile(filepath.Join(f.root, "photos", "a.txt"), []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	v, err := f.manager().Save(context.Background(), testJob, f.grant, "photos/a.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.manager().Restore(context.Background(), v.ID, f.grant); !errors.Is(err, agentpublish.ErrDenied) {
		t.Fatalf("read-only restore allowed %v", err)
	}
	if err := f.grants.Remove(f.grant); err != nil {
		t.Fatal(err)
	}
	if _, err := f.manager().Restore(context.Background(), v.ID, f.grant); !errors.Is(err, agentpublish.ErrDenied) {
		t.Fatalf("revoked grant used %v", err)
	}
}
func TestCorruptedArchivedVersionNotPublished(t *testing.T) {
	f := setup(t, false)
	if err := os.WriteFile(filepath.Join(f.root, "photos", "a.txt"), []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	v, err := f.manager().Save(context.Background(), testJob, f.grant, "photos/a.txt")
	if err != nil {
		t.Fatal(err)
	}
	archiveLocation := filepath.Join(f.versionsDirectory(), v.ID+".blob")
	if err := os.WriteFile(archiveLocation, []byte("tampered"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := f.manager().Restore(context.Background(), v.ID, f.grant); !errors.Is(err, agentversion.ErrCorrupted) {
		t.Fatalf("corruption accepted %v", err)
	}
	recovery, _ := RecoveredName("photos/a.txt", v.ID)
	if _, err := os.Stat(filepath.Join(f.root, recovery)); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("corrupted version produced destination")
	}
}
func TestConcurrentRestoresDoNotReplace(t *testing.T) {
	f := setup(t, false)
	if err := os.WriteFile(filepath.Join(f.root, "photos", "a.txt"), []byte(strings.Repeat("a", 40000)), 0600); err != nil {
		t.Fatal(err)
	}
	v, err := f.manager().Save(context.Background(), testJob, f.grant, "photos/a.txt")
	if err != nil {
		t.Fatal(err)
	}
	var ok atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, e := f.manager().Restore(context.Background(), v.ID, f.grant)
			if e == nil {
				ok.Add(1)
			} else if !errors.Is(e, agentpublish.ErrExists) {
				t.Errorf("unexpected concurrent failure: %v", e)
			}
		}()
	}
	wg.Wait()
	if ok.Load() != 1 {
		t.Fatalf("%d versions published, expected exactly one", ok.Load())
	}
}
func TestInvalidRecoveryName(t *testing.T) {
	id := strings.Repeat("a", 32)
	for _, name := range []string{"../x", "a/../b", "/root/file", ".syncria-internal/x", "x//y", "x/./y", "\\server\\share", ""} {
		if got, err := RecoveredName(name, id); err == nil {
			t.Errorf("unsafe name accepted %q -> %q", name, got)
		}
	}
	if _, err := RecoveredName("photos/a.jpg", strings.Repeat("x", 32)); err == nil {
		t.Fatal("invalid version ID")
	}
}
func TestCancelledRestoreLeavesNoFile(t *testing.T) {
	f := setup(t, false)
	if err := os.WriteFile(filepath.Join(f.root, "photos", "a.txt"), []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	v, err := f.manager().Save(context.Background(), testJob, f.grant, "photos/a.txt")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := f.manager().Restore(ctx, v.ID, f.grant); err == nil {
		t.Fatal("cancelled restore succeeded")
	}
	recovered, _ := RecoveredName("photos/a.txt", v.ID)
	if _, err := os.Stat(filepath.Join(f.root, recovered)); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("cancelled restore published file")
	}
}
