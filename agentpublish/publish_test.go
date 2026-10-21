package agentpublish

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/gigabytegrove/syncria/agentpolicy"
	"github.com/gigabytegrove/syncria/agenttransfer"
)

func digest(b []byte) string { sum := sha256.Sum256(b); return hex.EncodeToString(sum[:]) }
func fixtures(t *testing.T, contents []byte) (Publisher, string, string, string) {
	t.Helper()
	parent := t.TempDir()
	root := filepath.Join(parent, "destination")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	g, err := agentpolicy.Open(filepath.Join(parent, "grants.json"))
	if err != nil {
		t.Fatal(err)
	}
	grant, err := g.Add("My destination", root, false, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	s, err := agenttransfer.Open(filepath.Join(parent, "stage"), 16<<20)
	if err != nil {
		t.Fatal(err)
	}
	tr, err := s.Begin(agenttransfer.Request{SourceID: "source-device-123", Revision: digest(contents), RelativePath: "photos/2026/image.jpg", TotalBytes: int64(len(contents)), SHA256: digest(contents)})
	if err != nil {
		t.Fatal(err)
	}
	if len(contents) > 0 {
		_, err = s.Append(tr.ID, 0, bytes.NewReader(contents), len(contents), digest(contents))
		if err != nil {
			t.Fatal(err)
		}
	}
	_, err = s.Complete(tr.ID)
	if err != nil {
		t.Fatal(err)
	}
	return Publisher{Grants: g, Transfers: s}, grant.ID, tr.ID, root
}
func TestPublishAtomicNewFile(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux atomic writer only")
	}
	data := bytes.Repeat([]byte("Syncria transfer verified."), 4096)
	p, id, tr, root := fixtures(t, data)
	result, err := p.Publish(context.Background(), id, tr)
	if err != nil {
		t.Fatal(err)
	}
	if result.Bytes != int64(len(data)) || result.SHA256 != digest(data) {
		t.Fatal(result)
	}
	target := filepath.Join(root, "photos/2026/image.jpg")
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, got) {
		t.Fatal("published bytes mismatch")
	}
	entries, err := os.ReadDir(filepath.Dir(target))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "image.jpg" {
		t.Fatalf("temporary file leaked: %v", entries)
	}
	if _, err = p.Publish(context.Background(), id, tr); !errors.Is(err, ErrExists) {
		t.Fatalf("overwrite allowed: %v", err)
	}
	if got, err = os.ReadFile(target); err != nil || !bytes.Equal(got, data) {
		t.Fatalf("overwrite modified content: %v", err)
	}
}
func TestPublishRejectsReadOnlyExclusionsAndRevocation(t *testing.T) {
	data := []byte("hello")
	p, id, tr, root := fixtures(t, data)
	if err := p.Grants.Remove(id); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Publish(context.Background(), id, tr); !errors.Is(err, ErrDenied) {
		t.Fatalf("revocation not honored: %v", err)
	}
	ro, err := p.Grants.Add("readonly", root, true, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Publish(context.Background(), ro.ID, tr); !errors.Is(err, ErrDenied) {
		t.Fatalf("readonly not honored: %v", err)
	}
	if err = p.Grants.Remove(ro.ID); err != nil {
		t.Fatal(err)
	}
	excluded, err := p.Grants.Add("selected", root, false, []string{"documents"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Publish(context.Background(), excluded.ID, tr); !errors.Is(err, ErrDenied) {
		t.Fatalf("selection not honored: %v", err)
	}
}
func TestPublishRejectsPreexistingAndSymlinkedParents(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux atomic writer only")
	}
	data := []byte("new content")
	p, id, tr, root := fixtures(t, data)
	if err := os.MkdirAll(filepath.Join(root, "photos/2026"), 0700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "photos/2026/image.jpg")
	if err := os.WriteFile(target, []byte("important old version"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Publish(context.Background(), id, tr); !errors.Is(err, ErrExists) {
		t.Fatalf("existing file replaced: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil || string(got) != "important old version" {
		t.Fatal("existing data was modified", err)
	}
	if err := os.RemoveAll(filepath.Join(root, "photos")); err != nil {
		t.Fatal(err)
	}
	other := t.TempDir()
	if err := os.Symlink(other, filepath.Join(root, "photos")); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Publish(context.Background(), id, tr); err == nil {
		t.Fatal("symlinked destination accepted")
	}
	if _, err := os.Stat(filepath.Join(other, "2026/image.jpg")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("wrote outside approval: %v", err)
	}
}
func TestPublishNoReplaceUnderConcurrency(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux atomic writer only")
	}
	p, id, tr, root := fixtures(t, []byte("winner"))
	var wg sync.WaitGroup
	results := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, e := p.Publish(context.Background(), id, tr); results <- e }()
	}
	wg.Wait()
	close(results)
	successes, duplicates := 0, 0
	for e := range results {
		if e == nil {
			successes++
		} else if errors.Is(e, ErrExists) {
			duplicates++
		} else {
			t.Errorf("unexpected error %v", e)
		}
	}
	if successes != 1 || duplicates != 9 {
		t.Fatalf("successes=%d duplicates=%d", successes, duplicates)
	}
	b, err := os.ReadFile(filepath.Join(root, "photos/2026/image.jpg"))
	if err != nil || string(b) != "winner" {
		t.Fatal(err)
	}
}
func TestPublishCanceledAndCorruptStage(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux atomic writer only")
	}
	p, id, tr, root := fixtures(t, []byte("payload"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := p.Publish(ctx, id, tr); !errors.Is(err, context.Canceled) {
		t.Fatalf("context cancellation ignored: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "photos/2026/image.jpg")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("canceled transfer published: %v", err)
	}
	// The full SHA-256 is verified every time the completed stage is opened.
	items, err := p.Transfers.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatal(items)
	}
	stageRoot := filepath.Join(filepath.Dir(root), "stage", tr, "payload.part")
	if err := os.WriteFile(stageRoot, []byte(strings.Repeat("x", 7)), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = p.Publish(context.Background(), id, tr); !errors.Is(err, agenttransfer.ErrCorrupt) {
		t.Fatalf("corrupted stage accepted: %v", err)
	}
}
