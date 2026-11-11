package nasversion

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type disk struct {
	root         string
	failRename   int
	counter      int
	failRollback bool
}

func (d *disk) name(v string) string                { return filepath.Join(d.root, filepath.FromSlash(v)) }
func (d *disk) Stat(p string) (os.FileInfo, error)  { return os.Stat(d.name(p)) }
func (d *disk) Mkdir(p string, m os.FileMode) error { return os.Mkdir(d.name(p), m) }
func (d *disk) Rename(old, new string) error {
	d.counter++
	if d.failRename == d.counter || (d.failRollback && strings.Contains(old, ".syncria-versions") && strings.HasSuffix(new, "x.txt")) {
		return errors.New("simulated NAS rename error")
	}
	return os.Rename(d.name(old), d.name(new))
}
func setup(t *testing.T) (*disk, string, string) {
	t.Helper()
	d := &disk{root: t.TempDir()}
	if err := os.MkdirAll(d.name("folder"), 0700); err != nil {
		t.Fatal(err)
	}
	dest := "folder/x.txt"
	temp := dest + ".syncria-pending-" + strings.Repeat("f", 32)
	if err := os.WriteFile(d.name(temp), []byte("new"), 0600); err != nil {
		t.Fatal(err)
	}
	return d, dest, temp
}

const job = "aabbccddeeff00112233445566778899"

func verify(t *testing.T, d *disk, name, expected string) {
	t.Helper()
	v, e := os.ReadFile(d.name(name))
	if e != nil || string(v) != expected {
		t.Fatalf("file %q contents = %q, err %v, want %q", name, v, e, expected)
	}
}
func TestFirstPublicationNoArchive(t *testing.T) {
	d, dest, temp := setup(t)
	r, e := Promote(context.Background(), d, "folder", "x.txt", temp, job)
	if e != nil || r.Replaced || r.VersionPath != "" {
		t.Fatalf("result=%+v err=%v", r, e)
	}
	verify(t, d, dest, "new")
}
func TestReplacementPreservesOriginalOnNAS(t *testing.T) {
	d, dest, temp := setup(t)
	os.WriteFile(d.name(dest), []byte("old"), 0600)
	r, e := Promote(context.Background(), d, "folder", "x.txt", temp, job)
	if e != nil || !r.Replaced || !strings.Contains(r.VersionPath, ".syncria-versions/") {
		t.Fatalf("result=%+v err=%v", r, e)
	}
	verify(t, d, dest, "new")
	verify(t, d, r.VersionPath, "old")
	if _, e := os.Stat(d.name(temp)); !os.IsNotExist(e) {
		t.Fatal("pending file not removed")
	}
}
func TestFailureRestoresPrevious(t *testing.T) {
	d, dest, temp := setup(t)
	os.WriteFile(d.name(dest), []byte("old"), 0600)
	d.failRename = 2
	_, e := Promote(context.Background(), d, "folder", "x.txt", temp, job)
	if e == nil {
		t.Fatal("expected publish failure")
	}
	verify(t, d, dest, "old")
	verify(t, d, temp, "new")
}
func TestDoubleFailureLeavesRecoverablePrevious(t *testing.T) {
	d, dest, temp := setup(t)
	os.WriteFile(d.name(dest), []byte("old"), 0600)
	d.failRename = 2
	d.failRollback = true
	r, e := Promote(context.Background(), d, "folder", "x.txt", temp, job)
	if e == nil || r.VersionPath == "" {
		t.Fatalf("expected recovery instructions: %+v %v", r, e)
	}
	verify(t, d, r.VersionPath, "old")
	verify(t, d, temp, "new")
}
func TestUnsafeInputs(t *testing.T) {
	cases := []struct{ root, rel, temp, job string }{{"../home", "x.txt", "../home/x.txt.syncria-pending-x", job}, {"folder", "../secret", "folder/../secret.syncria-pending-x", job}, {"folder", "x.txt", "folder/z.txt", job}, {"folder", "x.txt", "folder/x.txt.syncria-pending-a", "../../bad"}, {"folder", ".syncria-versions/private", "folder/.syncria-versions/private.syncria-pending-a", job}}
	for _, c := range cases {
		d, _, _ := setup(t)
		if _, e := Promote(context.Background(), d, c.root, c.rel, c.temp, c.job); !errors.Is(e, ErrUnsafePath) {
			t.Errorf("accepted unsafe: %+v: %v", c, e)
		}
	}
}
func TestCanceledBeforePublication(t *testing.T) {
	d, dest, temp := setup(t)
	os.WriteFile(d.name(dest), []byte("old"), 0600)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, e := Promote(ctx, d, "folder", "x.txt", temp, job)
	if !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	verify(t, d, dest, "old")
}
func TestRejectsDestinationDirectory(t *testing.T) {
	d, _, temp := setup(t)
	os.Mkdir(d.name("folder/x.txt"), 0700)
	_, e := Promote(context.Background(), d, "folder", "x.txt", temp, job)
	if e == nil {
		t.Fatal("directory replaced")
	}
}
func TestCannotReenterInternalArchive(t *testing.T) {
	d, _, temp := setup(t)
	_, e := Promote(context.Background(), d, "folder", ".syncria-recovery/x", temp, job)
	if !errors.Is(e, ErrUnsafePath) {
		t.Fatal(e)
	}
}
func TestNoStagingFileFailsWithoutTouchingDestination(t *testing.T) {
	d, dest, temp := setup(t)
	os.WriteFile(d.name(dest), []byte("old"), 0600)
	os.Remove(d.name(temp))
	_, e := Promote(context.Background(), d, "folder", "x.txt", temp, job)
	if e == nil {
		t.Fatal("expected missing staging")
	}
	verify(t, d, dest, "old")
}
