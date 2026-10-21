package nasversion

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func (d *disk) ReadDir(p string) ([]os.FileInfo, error) {
	entries, e := os.ReadDir(d.name(p))
	if e != nil {
		return nil, e
	}
	infos := make([]os.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, e := os.Lstat(d.name(filepath.ToSlash(filepath.Join(p, entry.Name()))))
		if e != nil {
			return nil, e
		}
		infos = append(infos, info)
	}
	return infos, nil
}
func TestListRetainedCopies(t *testing.T) {
	d, dest, temp := setup(t)
	os.WriteFile(d.name(dest), []byte("original"), 0600)
	r, e := Promote(context.Background(), d, "folder", "x.txt", temp, job)
	if e != nil {
		t.Fatal(e)
	}
	found, e := List(context.Background(), d, "folder", job)
	if e != nil || len(found) != 1 || found[0].Path != r.VersionPath || found[0].Relative != "x.txt" || found[0].Bytes != 8 {
		t.Fatalf("found=%+v err=%v", found, e)
	}
	verify(t, d, found[0].Path, "original")
}
func TestEmptyListOnFreshNAS(t *testing.T) {
	d, _, _ := setup(t)
	v, e := List(context.Background(), d, "folder", job)
	if e != nil || len(v) != 0 {
		t.Fatal(v, e)
	}
}
func TestListFailsClosedOnInvalidFiles(t *testing.T) {
	d, _, _ := setup(t)
	os.MkdirAll(d.name("folder/.syncria-versions/"+job), 0700)
	os.WriteFile(d.name("folder/.syncria-versions/"+job+"/junk"), []byte("x"), 0600)
	if _, e := List(context.Background(), d, "folder", job); e == nil {
		t.Fatal("invalid archive inventory accepted")
	}
}
func TestListRejectsUnsafeRootAndID(t *testing.T) {
	d, _, _ := setup(t)
	if _, e := List(context.Background(), d, "../root", job); !errors.Is(e, ErrUnsafePath) {
		t.Fatal(e)
	}
	if _, e := List(context.Background(), d, "folder", strings.Repeat("z", 32)); !errors.Is(e, ErrUnsafePath) {
		t.Fatal(e)
	}
}
func TestListCancel(t *testing.T) {
	d, _, _ := setup(t)
	os.MkdirAll(d.name("folder/.syncria-versions/"+job), 0700)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := List(ctx, d, "folder", job); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
}
