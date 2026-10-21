package agentversion

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

const jobID = "aabbccddeeff00112233445566778899"

func TestArchiveRestoreAndRestart(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "versions")
	store, err := Open(dir, 5<<20)
	if err != nil {
		t.Fatal(err)
	}
	body := bytes.Repeat([]byte("source content"), 4000)
	entry, err := store.Capture(context.Background(), jobID, "photos/photo.txt", bytes.NewReader(body), int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}
	if entry.Bytes != int64(len(body)) || len(entry.SHA256) != 64 {
		t.Fatal("bad archive")
	}
	store, err = Open(dir, 5<<20)
	if err != nil {
		t.Fatal("reopen", err)
	}
	f, v, err := store.OpenVerified(context.Background(), entry.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil || !bytes.Equal(data, body) || v.ID != entry.ID {
		t.Fatal("restored wrong content", err)
	}
	got := store.List(jobID, "photos/photo.txt")
	if len(got) != 1 {
		t.Fatal("version missing")
	}
}
func TestArchiveErrorsAndQuota(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "versions"), 128)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Capture(context.Background(), jobID, "../secret", strings.NewReader("abc"), 3); err == nil {
		t.Fatal("unsafe path accepted")
	}
	if _, err := store.Capture(context.Background(), jobID, "ok", strings.NewReader("abc"), 4); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatal("truncated input accepted", err)
	}
	if _, err := store.Capture(context.Background(), jobID, "ok", strings.NewReader("abcd"), 3); err == nil {
		t.Fatal("extra bytes accepted")
	}
	if _, err := store.Capture(context.Background(), jobID, "ok", strings.NewReader(strings.Repeat("z", 90)), 90); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Capture(context.Background(), jobID, "ok", strings.NewReader(strings.Repeat("z", 50)), 50); !errors.Is(err, ErrQuota) {
		t.Fatal("quota not enforced", err)
	}
}
func TestTamperingFailClosed(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "versions")
	store, err := Open(dir, 1024)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := store.Capture(context.Background(), jobID, "item", strings.NewReader("original"), 8)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(store.blobPath(entry.ID), []byte("replaced"), 0600); err != nil {
		t.Fatal(err)
	}
	f, _, err := store.OpenVerified(context.Background(), entry.ID)
	if f != nil {
		f.Close()
	}
	if !errors.Is(err, ErrCorrupted) {
		t.Fatal("modified bytes accepted", err)
	}
}
func TestRetentionAndConcurrentCapture(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "versions")
	store, err := Open(dir, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, e := store.Capture(context.Background(), jobID, "a/b", strings.NewReader("test"), 4)
			errs <- e
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	if len(store.List(jobID, "a/b")) != 12 {
		t.Fatal("lost writes")
	}
	removed, err := store.Prune(3)
	if err != nil || removed != 9 {
		t.Fatalf("prune: %v count %d", err, removed)
	}
	restart, err := Open(dir, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	if len(restart.List(jobID, "a/b")) != 3 {
		t.Fatal("prune not durable")
	}
}
func TestRejectUntrustedStore(t *testing.T) {
	dir := t.TempDir()
	child := filepath.Join(dir, "alias")
	if err := os.Symlink(dir, child); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := Open(child, 1024); err == nil {
		t.Fatal("symlink archive directory accepted")
	}
}
func TestCancelCaptureNoManifest(t *testing.T) {
	store, _ := Open(filepath.Join(t.TempDir(), "versions"), 4096)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := store.Capture(ctx, jobID, "x", strings.NewReader("x"), 1); !errors.Is(err, context.Canceled) {
		t.Fatal("cancel ignored", err)
	}
	if len(store.List(jobID, "")) != 0 {
		t.Fatal("canceled capture committed")
	}
}
