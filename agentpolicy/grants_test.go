package agentpolicy

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

func fresh(t *testing.T) (*Store, string, string) {
	t.Helper()
	base := t.TempDir()
	file := filepath.Join(base, "private", "grants.json")
	root := filepath.Join(base, "docs")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	s, err := Open(file)
	if err != nil {
		t.Fatal(err)
	}
	return s, file, root
}
func TestGrantLifeCycleAndRestart(t *testing.T) {
	s, file, root := fresh(t)
	g, err := s.Add("Work documents", root, true, []string{"reports", "photos/family"}, []string{"reports/private"})
	if err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"reports", "reports/2026/file.txt", "photos/family/new.jpg"} {
		if err = s.Allows(g.ID, rel, false); err != nil {
			t.Errorf("unexpected denied %q: %v", rel, err)
		}
	}
	for _, rel := range []string{"reports/private", "reports/private/secrets", "photos/other", "../escape", "/etc/passwd", "reports/../../etc"} {
		if err = s.Allows(g.ID, rel, false); !errors.Is(err, ErrDenied) {
			t.Errorf("unsafe path allowed %q", rel)
		}
	}
	if err = s.Allows(g.ID, "reports/file.txt", true); !errors.Is(err, ErrDenied) {
		t.Fatalf("read-only grant allowed write: %v", err)
	}
	s2, err := Open(file)
	if err != nil {
		t.Fatal(err)
	}
	if list := s2.List(); len(list) != 1 || list[0].ID != g.ID {
		t.Fatalf("not persisted %+v", list)
	}
	if err = s2.Remove(g.ID); err != nil {
		t.Fatal(err)
	}
	if e := s2.Allows(g.ID, "reports/file.txt", false); !errors.Is(e, ErrDenied) {
		t.Fatalf("removed grant still allows access: %v", e)
	}
	s3, err := Open(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(s3.List()) != 0 {
		t.Fatal("removal not durable")
	}
}
func TestRejectOverlappingRootsAndSymlinks(t *testing.T) {
	s, _, root := fresh(t)
	if _, e := s.Add("first", root, false, nil, nil); e != nil {
		t.Fatal(e)
	}
	if _, e := s.Add("same", root, false, nil, nil); !errors.Is(e, ErrDuplicate) {
		t.Fatalf("duplicate root accepted: %v", e)
	}
	child := filepath.Join(root, "child")
	if e := os.Mkdir(child, 0700); e != nil {
		t.Fatal(e)
	}
	if _, e := s.Add("nested", child, false, nil, nil); !errors.Is(e, ErrDuplicate) {
		t.Fatalf("nested root accepted: %v", e)
	}
	other := filepath.Join(t.TempDir(), "alias")
	if e := os.Symlink(root, other); e != nil {
		if runtime.GOOS == "windows" {
			t.Skip(e)
		}
		t.Fatal(e)
	}
	if _, e := s.Add("alias", other, false, nil, nil); e == nil {
		t.Fatal("symlink root accepted")
	}
}
func TestBadSelectorsAndTamperedFile(t *testing.T) {
	s, file, root := fresh(t)
	for _, bad := range []string{"/abs", "../up", "dir/../up", "a//b", "a/./b", "a\\b", ".syncria-staging", "some\nfile"} {
		if _, e := s.Add("bad", root, false, []string{bad}, nil); !errors.Is(e, ErrInvalid) {
			t.Errorf("bad selector %q: %v", bad, e)
		}
	}
	if _, e := s.Add("dup", root, false, []string{"a", "a"}, nil); !errors.Is(e, ErrInvalid) {
		t.Fatal("duplicate selection accepted")
	}
	if _, e := s.Add("valid", root, false, nil, nil); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(file, []byte("invalid"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := Open(file); e == nil {
		t.Fatal("corrupt grants silently reset")
	}
}
func TestPrivatePermissionsAndSymlinkedState(t *testing.T) {
	s, file, root := fresh(t)
	if _, e := s.Add("docs", root, false, nil, nil); e != nil {
		t.Fatal(e)
	}
	if runtime.GOOS != "windows" {
		if fi, e := os.Stat(file); e != nil || fi.Mode().Perm() != 0600 {
			t.Fatalf("grant permissions %v %v", fi, e)
		}
	}
	if e := os.Remove(file); e != nil {
		t.Fatal(e)
	}
	if e := os.Symlink(root, file); e != nil {
		if runtime.GOOS == "windows" {
			t.Skip(e)
		}
		t.Fatal(e)
	}
	if _, e := Open(file); e == nil {
		t.Fatal("symlink state accepted")
	}
}
func TestConcurrencyAndImmutableViews(t *testing.T) {
	s, _, root := fresh(t)
	g, e := s.Add("docs", root, false, []string{"docs"}, nil)
	if e != nil {
		t.Fatal(e)
	}
	view := s.List()
	view[0].Selections[0] = "changed"
	if s.List()[0].Selections[0] != "docs" {
		t.Fatal("caller mutated grant")
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 30; j++ {
				if err := s.Allows(g.ID, "docs/file.txt", true); err != nil {
					t.Error(err)
				}
			}
		}()
	}
	wg.Wait()
}
func TestSelectorValidationAndReadOnly(t *testing.T) {
	s, _, root := fresh(t)
	g, e := s.Add("Personal", root, false, nil, []string{"private"})
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Allows(g.ID, "family/hello.txt", true); e != nil {
		t.Fatal(e)
	}
	if e = s.Allows(g.ID, "private/accounts.txt", true); !errors.Is(e, ErrDenied) {
		t.Fatal("exclusion failed")
	}
	if e = s.Allows(strings.Repeat("0", 32), "file", false); !errors.Is(e, ErrDenied) {
		t.Fatal("unknown grant accepted")
	}
	if e = s.Allows(g.ID, "", false); !errors.Is(e, ErrDenied) {
		t.Fatal("empty path accepted")
	}
}
