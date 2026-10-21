package agenttransfer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func sum(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func newStore(t *testing.T) (*Store, string) {
	t.Helper()
	root := filepath.Join(t.TempDir(), "transfer")
	s, e := Open(root, 8<<20)
	if e != nil {
		t.Fatal(e)
	}
	return s, root
}
func request(b []byte) Request {
	return Request{SourceID: "device-123456", Revision: "revision-12345", RelativePath: "docs/file.bin", TotalBytes: int64(len(b)), SHA256: sum(b)}
}
func begin(t *testing.T, s *Store, b []byte) Transfer {
	t.Helper()
	tr, e := s.Begin(request(b))
	if e != nil {
		t.Fatal(e)
	}
	return tr
}
func TestStreamResumesAndVerifiesAcrossRestart(t *testing.T) {
	data := bytes.Repeat([]byte("file bytes "), 11234)
	s, root := newStore(t)
	tr := begin(t, s, data)
	if _, e := s.Complete(tr.ID); !errors.Is(e, ErrIncomplete) {
		t.Fatalf("should be incomplete: %v", e)
	}
	n := len(data) / 2
	first, e := s.Append(tr.ID, 0, bytes.NewReader(data[:n]), n, sum(data[:n]))
	if e != nil {
		t.Fatal(e)
	}
	if first.ReceivedBytes != int64(n) {
		t.Fatal(first)
	}
	restarted, e := Open(root, 8<<20)
	if e != nil {
		t.Fatal(e)
	}
	snapshot, e := restarted.Get(tr.ID)
	if e != nil || snapshot.ReceivedBytes != int64(n) {
		t.Fatalf("checkpoint lost: %v %v", snapshot, e)
	}
	repeat, e := restarted.Append(tr.ID, 0, bytes.NewReader(data[:n]), n, sum(data[:n]))
	if e != nil || repeat.ReceivedBytes != int64(n) {
		t.Fatalf("idempotency failed: %v", e)
	}
	final, e := restarted.Append(tr.ID, int64(n), bytes.NewReader(data[n:]), len(data[n:]), sum(data[n:]))
	if e != nil {
		t.Fatal(e)
	}
	if final.ReceivedBytes != int64(len(data)) {
		t.Fatal(final)
	}
	done, e := restarted.Complete(tr.ID)
	if e != nil || done.Status != StatusComplete {
		t.Fatalf("complete: %v %v", done, e)
	}
	f, _, e := restarted.OpenCompleted(tr.ID)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	output, e := io.ReadAll(f)
	if e != nil || !bytes.Equal(data, output) {
		t.Fatalf("bytes mismatch: %v", e)
	}
	done2, e := restarted.Complete(tr.ID)
	if e != nil || done2.ID != done.ID {
		t.Fatalf("completion retry failed: %v", e)
	}
}
func TestCrashAfterPayloadBeforeManifest(t *testing.T) {
	s, root := newStore(t)
	content := []byte("this is a staged transfer")
	tr := begin(t, s, content)
	if _, e := s.Append(tr.ID, 0, bytes.NewReader(content[:5]), 5, sum(content[:5])); e != nil {
		t.Fatal(e)
	}
	f, e := os.OpenFile(s.payloadPath(tr.ID), os.O_APPEND|os.O_WRONLY, 0)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = f.Write([]byte("uncertain bytes")); e != nil {
		t.Fatal(e)
	}
	if e = f.Close(); e != nil {
		t.Fatal(e)
	}
	again, e := Open(root, 8<<20)
	if e != nil {
		t.Fatal(e)
	}
	state, e := again.Get(tr.ID)
	if e != nil || state.ReceivedBytes != 5 {
		t.Fatalf("checkpoint mismatch: %v", e)
	}
	fi, e := os.Stat(again.payloadPath(tr.ID))
	if e != nil || fi.Size() != 5 {
		t.Fatalf("uncommitted data not discarded: %v %v", fi, e)
	}
	if _, e = again.Append(tr.ID, 5, bytes.NewReader(content[5:]), len(content)-5, sum(content[5:])); e != nil {
		t.Fatal(e)
	}
	if _, e = again.Complete(tr.ID); e != nil {
		t.Fatal(e)
	}
}
func TestQuotaIdempotencyAndIdentityConflict(t *testing.T) {
	s, _ := newStore(t)
	s.maxPending = 14
	a := []byte("12345678")
	first := begin(t, s, a)
	replay, e := s.Begin(request(a))
	if e != nil || replay.ID != first.ID {
		t.Fatalf("duplicate session: %v", e)
	}
	conflicting := request([]byte("abcdefgh"))
	if _, e = s.Begin(conflicting); !errors.Is(e, ErrConflict) {
		t.Fatalf("unexpected conflict result: %v", e)
	}
	next := request([]byte("1234567"))
	next.Revision = "different-revision"
	if _, e = s.Begin(next); !errors.Is(e, ErrQuota) {
		t.Fatalf("quota failed: %v", e)
	}
	if e = s.Abort(first.ID); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Begin(next); e != nil {
		t.Fatalf("quota should be released: %v", e)
	}
}
func TestInvalidOffsetAndBadChunkCannotAdvance(t *testing.T) {
	s, _ := newStore(t)
	content := []byte("abcdefghijklmnop")
	tr := begin(t, s, content)
	cases := []struct {
		name    string
		offset  int64
		payload []byte
		digest  string
		want    error
	}{
		{"skip", 1, content[:3], sum(content[:3]), ErrConflict},
		{"oversized", 14, content[:3], sum(content[:3]), ErrInvalid},
		{"wrong checksum", 0, content[:4], sum([]byte("wrong")), ErrCorrupt},
		{"short read", 0, content[:3], sum(content[:4]), io.ErrUnexpectedEOF},
		{"extra byte", 0, content[:5], sum(content[:4]), ErrInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			length := len(tc.payload)
			if tc.name == "short read" || tc.name == "extra byte" {
				length = 4
			}
			_, e := s.Append(tr.ID, tc.offset, bytes.NewReader(tc.payload), length, tc.digest)
			if !errors.Is(e, tc.want) {
				t.Fatalf("got %v want %v", e, tc.want)
			}
			x, e := s.Get(tr.ID)
			if e != nil || x.ReceivedBytes != 0 {
				t.Fatalf("bad append advanced offset: %+v %v", x, e)
			}
		})
	}
	_, e := s.Append(tr.ID, 0, bytes.NewReader(content[:4]), 4, sum(content[:4]))
	if e != nil {
		t.Fatal(e)
	}
	_, e = s.Append(tr.ID, 2, bytes.NewReader(content[2:6]), 4, sum(content[2:6]))
	if !errors.Is(e, ErrConflict) {
		t.Fatalf("overlap was allowed: %v", e)
	}
	_, e = s.Append(tr.ID, 0, bytes.NewReader([]byte("zzzz")), 4, sum([]byte("zzzz")))
	if !errors.Is(e, ErrConflict) {
		t.Fatalf("conflicting replay accepted: %v", e)
	}
}
func TestTamperedFinalPayloadNotCommitted(t *testing.T) {
	s, _ := newStore(t)
	data := []byte("verified transfer")
	tr := begin(t, s, data)
	_, e := s.Append(tr.ID, 0, bytes.NewReader(data), len(data), sum(data))
	if e != nil {
		t.Fatal(e)
	}
	f, e := os.OpenFile(s.payloadPath(tr.ID), os.O_WRONLY, 0)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = f.WriteAt([]byte("X"), 0); e != nil {
		t.Fatal(e)
	}
	_ = f.Close()
	if _, e = s.Complete(tr.ID); !errors.Is(e, ErrCorrupt) {
		t.Fatalf("corruption accepted: %v", e)
	}
	if _, _, e = s.OpenCompleted(tr.ID); !errors.Is(e, ErrIncomplete) {
		t.Fatalf("unverified file exposed: %v", e)
	}
}
func TestRejectInvalidRequestsAndPaths(t *testing.T) {
	data := []byte("test")
	for _, bad := range []string{"", "../outside", "/absolute", "./file", "a/../b", "a//b", "a\\b", "a/./b", "a/.syncria-private", "a/\x01b"} {
		q := request(data)
		q.RelativePath = bad
		if validRequest(q) {
			t.Errorf("accepted path %q", bad)
		}
	}
	for _, bad := range []string{"", "revision with whitespace", "x/filename"} {
		if validIdentifier(bad) {
			t.Errorf("accepted identifier %q", bad)
		}
	}
	if validDigest(strings.ToUpper(sum(data))) {
		t.Fatal("uppercase digests accepted")
	}
	s, _ := newStore(t)
	q := request(data)
	q.SHA256 = "not-sha"
	if _, e := s.Begin(q); !errors.Is(e, ErrInvalid) {
		t.Fatalf("bad digest accepted: %v", e)
	}
	if _, e := s.Get("../../etc/passwd"); !errors.Is(e, ErrInvalid) {
		t.Fatalf("unsafe id: %v", e)
	}
}
func TestCorruptManifestAndSymlinkRefused(t *testing.T) {
	s, root := newStore(t)
	tr := begin(t, s, []byte("test"))
	if e := os.WriteFile(s.manifestPath(tr.ID), []byte("garbage"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := Open(root, 8<<20); !errors.Is(e, ErrCorrupt) {
		t.Fatalf("corrupted state accepted: %v", e)
	}
	t.Run("symlink state", func(t *testing.T) {
		s, _ := newStore(t)
		tr := begin(t, s, []byte("hello"))
		real := s.payloadPath(tr.ID)
		if e := os.Remove(real); e != nil {
			t.Fatal(e)
		}
		if e := os.Symlink(filepath.Join(t.TempDir(), "outside"), real); e != nil {
			if runtime.GOOS == "windows" {
				t.Skip(e)
			}
			t.Fatal(e)
		}
		if _, e := s.Get(tr.ID); !errors.Is(e, ErrCorrupt) {
			t.Fatalf("symlink accepted: %v", e)
		}
	})
}
func TestConcurrentChunksSerialized(t *testing.T) {
	s, _ := newStore(t)
	data := bytes.Repeat([]byte("X"), 8192)
	tr := begin(t, s, data)
	var wg sync.WaitGroup
	ch := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.Append(tr.ID, 0, bytes.NewReader(data), len(data), sum(data))
			ch <- err
		}()
	}
	wg.Wait()
	close(ch)
	for e := range ch {
		if e != nil {
			t.Fatalf("concurrent idempotent append failed: %v", e)
		}
	}
	status, e := s.Get(tr.ID)
	if e != nil || status.ReceivedBytes != int64(len(data)) {
		t.Fatalf("bad offset after races: %+v %v", status, e)
	}
}
func TestSweepOnlyPending(t *testing.T) {
	s, _ := newStore(t)
	payload := []byte("hello world")
	done := begin(t, s, payload)
	_, e := s.Append(done.ID, 0, bytes.NewReader(payload), len(payload), sum(payload))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Complete(done.ID); e != nil {
		t.Fatal(e)
	}
	pendingQ := request([]byte("another"))
	pendingQ.Revision = "revision-pending-1"
	pending, e := s.Begin(pendingQ)
	if e != nil {
		t.Fatal(e)
	}
	s.now = func() time.Time { return time.Now().Add(48 * time.Hour) }
	removed, e := s.Sweep(time.Hour)
	if e != nil || removed != 1 {
		t.Fatalf("sweep count=%d err=%v", removed, e)
	}
	if _, e = s.Get(pending.ID); !errors.Is(e, ErrNotFound) {
		t.Fatalf("pending retained: %v", e)
	}
	if _, e = s.Get(done.ID); e != nil {
		t.Fatalf("completed removed: %v", e)
	}
}
func TestEmptyFileAndWrongFinalChecksum(t *testing.T) {
	s, _ := newStore(t)
	tr := begin(t, s, nil)
	done, e := s.Complete(tr.ID)
	if e != nil || done.Status != StatusComplete {
		t.Fatalf("zero-byte transfer: %v", e)
	}
	q := request([]byte("hello"))
	q.Revision = "revision-different"
	q.SHA256 = sum([]byte("other"))
	x, e := s.Begin(q)
	if e != nil {
		t.Fatal(e)
	}
	_, e = s.Append(x.ID, 0, bytes.NewReader([]byte("hello")), 5, sum([]byte("hello")))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Complete(x.ID); !errors.Is(e, ErrCorrupt) {
		t.Fatalf("wrong source expected digest accepted: %v", e)
	}
}
