package agentversion

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDropCommitsManifestBeforeBlobRemoval(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "versions")
	s, err := Open(dir, 12)
	if err != nil {
		t.Fatal(err)
	}
	id := "aabbccddeeff00112233445566778899"
	one, err := s.Capture(context.Background(), id, "test.txt", strings.NewReader("abcd"), 4)
	if err != nil {
		t.Fatal(err)
	}
	two, err := s.Capture(context.Background(), id, "test.txt", strings.NewReader("efgh"), 4)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Drop(one.ID); err != nil {
		t.Fatal(err)
	}
	if len(s.List(id, "test.txt")) != 1 {
		t.Fatal("drop removed wrong version")
	}
	if _, err := os.Stat(s.blobPath(one.ID)); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("blob not removed")
	}
	if err := s.Drop(one.ID); !errors.Is(err, ErrNotFound) {
		t.Fatal("drop was not idempotently rejected")
	}
	s, err = Open(dir, 12)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.List(id, "test.txt")) != 1 || s.List(id, "test.txt")[0].ID != two.ID {
		t.Fatal("archive changed across restart")
	}
	if _, err = s.Capture(context.Background(), id, "new.txt", strings.NewReader("01234567"), 8); err != nil {
		t.Fatalf("quota not reclaimed: %v", err)
	}
}
