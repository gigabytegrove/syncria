// Package agenttransfer provides crash-resilient verified staging of agent
// transfers. Staged payloads are private and are never promoted into a sync
// destination by this package. A future endpoint adapter must perform scoped,
// handle-relative writes and enforce the active sync relationship policy.
package agenttransfer

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	MaxChunkBytes          = 4 << 20
	MaxManifestBytes       = 16 << 10
	DefaultMaxPendingBytes = 32 << 30
	StatusPending          = "pending"
	StatusComplete         = "complete"
)

var (
	ErrInvalid    = errors.New("invalid transfer request")
	ErrConflict   = errors.New("transfer identity or offset conflict")
	ErrQuota      = errors.New("pending transfer quota exceeded")
	ErrIncomplete = errors.New("transfer is incomplete")
	ErrCorrupt    = errors.New("staged transfer is corrupt")
	ErrNotFound   = errors.New("transfer does not exist")
)

type Request struct {
	SourceID     string `json:"source_id"`
	Revision     string `json:"revision"`
	RelativePath string `json:"relative_path"`
	TotalBytes   int64  `json:"total_bytes"`
	SHA256       string `json:"sha256"`
}

type Transfer struct {
	ID string `json:"id"`
	Request
	ReceivedBytes int64     `json:"received_bytes"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Store struct {
	mu         sync.Mutex
	directory  string
	maxPending int64
	now        func() time.Time
}

func randomID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
func validID(s string) bool {
	if len(s) != 32 {
		return false
	}
	for _, r := range s {
		if r != '0' && r != '1' && r != '2' && r != '3' && r != '4' && r != '5' && r != '6' && r != '7' && r != '8' && r != '9' && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}
func validDigest(s string) bool {
	if len(s) != sha256.Size*2 {
		return false
	}
	b, e := hex.DecodeString(s)
	return e == nil && len(b) == sha256.Size && strings.ToLower(s) == s
}
func validRelative(s string) bool {
	if s == "" || len(s) > 2048 || strings.ContainsAny(s, "\\\x00\r\n") || strings.HasPrefix(s, "/") || path.Clean(s) != s {
		return false
	}
	for _, component := range strings.Split(s, "/") {
		if component == "" || component == "." || component == ".." || len(component) > 255 || strings.HasPrefix(component, ".syncria-") {
			return false
		}
		for _, r := range component {
			if r < 32 || r == 127 {
				return false
			}
		}
	}
	return true
}
func validRequest(q Request) bool {
	return len(q.SourceID) >= 8 && len(q.SourceID) <= 128 && validIdentifier(q.SourceID) &&
		len(q.Revision) >= 8 && len(q.Revision) <= 256 && validIdentifier(q.Revision) &&
		validRelative(q.RelativePath) && q.TotalBytes >= 0 && q.TotalBytes <= 1<<40 && validDigest(q.SHA256)
}
func validIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '!' || r > '~' || r == '/' || r == '\\' || r == '?' || r == '#' {
			return false
		}
	}
	return true
}

// Open verifies the storage root and all session entries before making any
// staged content available. This directory must be private to the agent user.
func Open(directory string, maxPending int64) (*Store, error) {
	if directory == "" || !filepath.IsAbs(directory) || maxPending <= 0 {
		return nil, ErrInvalid
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, err
	}
	fi, err := os.Lstat(directory)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() || runtime.GOOS != "windows" && fi.Mode().Perm()&0077 != 0 {
		return nil, errors.New("transfer directory must be a private real directory")
	}
	s := &Store{directory: directory, maxPending: maxPending, now: time.Now}
	if _, err = s.listLocked(); err != nil {
		return nil, err
	}
	return s, nil
}
func (s *Store) subdir(id string) string       { return filepath.Join(s.directory, id) }
func (s *Store) manifestPath(id string) string { return filepath.Join(s.subdir(id), "manifest.json") }
func (s *Store) payloadPath(id string) string  { return filepath.Join(s.subdir(id), "payload.part") }

// loadLocked refuses symlinks, malformed manifests, unexpected file sizes and
// unexpected sizes. Uncommitted trailing bytes are removed only
// after validation so the last durable offset remains authoritative.
func (s *Store) loadLocked(id string) (Transfer, error) {
	if !validID(id) {
		return Transfer{}, ErrInvalid
	}
	base := s.subdir(id)
	fi, err := os.Lstat(base)
	if errors.Is(err, os.ErrNotExist) {
		return Transfer{}, ErrNotFound
	}
	if err != nil {
		return Transfer{}, err
	}
	if !fi.IsDir() || runtime.GOOS != "windows" && fi.Mode().Perm()&0077 != 0 {
		return Transfer{}, ErrCorrupt
	}
	mp := s.manifestPath(id)
	fi, err = os.Lstat(mp)
	if err != nil {
		return Transfer{}, err
	}
	if !fi.Mode().IsRegular() || fi.Size() > MaxManifestBytes || runtime.GOOS != "windows" && fi.Mode().Perm()&0077 != 0 {
		return Transfer{}, ErrCorrupt
	}
	body, err := os.ReadFile(mp)
	if err != nil {
		return Transfer{}, err
	}
	var tr Transfer
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&tr); err != nil {
		return Transfer{}, ErrCorrupt
	}
	var extra any
	if err = decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return Transfer{}, ErrCorrupt
	}
	if tr.ID != id || !validRequest(tr.Request) || tr.ReceivedBytes < 0 || tr.ReceivedBytes > tr.TotalBytes || (tr.Status != StatusPending && tr.Status != StatusComplete) || tr.CreatedAt.IsZero() || tr.UpdatedAt.IsZero() || tr.UpdatedAt.Before(tr.CreatedAt) || tr.Status == StatusComplete && tr.ReceivedBytes != tr.TotalBytes {
		return Transfer{}, ErrCorrupt
	}
	pf := s.payloadPath(id)
	fi, err = os.Lstat(pf)
	if err != nil {
		return Transfer{}, err
	}
	if !fi.Mode().IsRegular() || runtime.GOOS != "windows" && fi.Mode().Perm()&0077 != 0 {
		return Transfer{}, ErrCorrupt
	}
	if fi.Size() < tr.ReceivedBytes || fi.Size() > tr.TotalBytes {
		return Transfer{}, ErrCorrupt
	}
	// If a crash occurred after fsyncing the payload but before committing the
	// manifest, discard the uncommitted tail; re-send from the last checkpoint.
	if fi.Size() > tr.ReceivedBytes {
		if tr.Status == StatusComplete {
			return Transfer{}, ErrCorrupt
		}
		if err := os.Truncate(pf, tr.ReceivedBytes); err != nil {
			return Transfer{}, err
		}
	}
	return tr, nil
}
func (s *Store) listLocked() ([]Transfer, error) {
	entries, err := os.ReadDir(s.directory)
	if err != nil {
		return nil, err
	}
	out := make([]Transfer, 0, len(entries))
	for _, e := range entries {
		if !validID(e.Name()) || !e.IsDir() {
			return nil, fmt.Errorf("unrecognized transfer entry %q: %w", e.Name(), ErrCorrupt)
		}
		tr, err := s.loadLocked(e.Name())
		if err != nil {
			return nil, fmt.Errorf("transfer %s: %w", e.Name(), err)
		}
		out = append(out, tr)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}
func (s *Store) List() ([]Transfer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.listLocked()
}
func (s *Store) Get(id string) (Transfer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked(id)
}
func atomicManifest(location string, tr Transfer) error {
	body, err := json.MarshalIndent(tr, "", "  ")
	if err != nil {
		return err
	}
	if len(body) > MaxManifestBytes {
		return ErrInvalid
	}
	dir := filepath.Dir(location)
	f, err := os.CreateTemp(dir, ".manifest-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = f.Chmod(0600); err != nil {
		_ = f.Close()
		return err
	}
	if _, err = f.Write(body); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(f.Name(), location); err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		if d, e := os.Open(dir); e == nil {
			_ = d.Sync()
			_ = d.Close()
		}
	}
	return nil
}

// Begin reserves a private staged transfer and enforces a global capacity
// budget across unfinished sessions. The caller must authenticate and ensure
// the selected job and endpoint allow the intended path BEFORE calling Begin.
func (s *Store) Begin(q Request) (Transfer, error) {
	if !validRequest(q) {
		return Transfer{}, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := s.listLocked()
	if err != nil {
		return Transfer{}, err
	}
	pending := int64(0)
	for _, e := range entries {
		if e.Status != StatusPending {
			continue
		}
		if e.SourceID == q.SourceID && e.Revision == q.Revision && e.RelativePath == q.RelativePath {
			if e.Request == q {
				return e, nil
			}
			return Transfer{}, ErrConflict
		}
		if e.TotalBytes > s.maxPending-pending {
			return Transfer{}, ErrQuota
		}
		pending += e.TotalBytes
	}
	if q.TotalBytes > s.maxPending-pending {
		return Transfer{}, ErrQuota
	}
	id, err := randomID()
	if err != nil {
		return Transfer{}, err
	}
	now := s.now().UTC()
	tr := Transfer{ID: id, Request: q, Status: StatusPending, CreatedAt: now, UpdatedAt: now}
	dir := s.subdir(id)
	if err := os.Mkdir(dir, 0700); err != nil {
		return Transfer{}, err
	}
	success := false
	defer func() {
		if !success {
			_ = os.RemoveAll(dir)
		}
	}()
	f, err := os.OpenFile(s.payloadPath(id), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return Transfer{}, err
	}
	if err = f.Sync(); err != nil {
		_ = f.Close()
		return Transfer{}, err
	}
	if err = f.Close(); err != nil {
		return Transfer{}, err
	}
	if err = atomicManifest(s.manifestPath(id), tr); err != nil {
		return Transfer{}, err
	}
	success = true
	return tr, nil
}

// Append receives exactly length bytes at a verified checkpoint. An identical
// previously-committed chunk is idempotent; overlapping or stale writes fail.
// Buffering is bounded by MaxChunkBytes; no payload goes into Syncria /data.
func (s *Store) Append(id string, offset int64, source io.Reader, length int, expectedSHA256 string) (Transfer, error) {
	if !validID(id) || offset < 0 || length <= 0 || length > MaxChunkBytes || !validDigest(expectedSHA256) || source == nil {
		return Transfer{}, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tr, err := s.loadLocked(id)
	if err != nil {
		return Transfer{}, err
	}
	if offset > tr.TotalBytes || int64(length) > tr.TotalBytes-offset {
		return Transfer{}, ErrInvalid
	}
	if tr.Status == StatusComplete && offset == tr.ReceivedBytes {
		return Transfer{}, ErrConflict
	}
	if offset > tr.ReceivedBytes || offset < tr.ReceivedBytes && offset+int64(length) > tr.ReceivedBytes {
		return Transfer{}, ErrConflict
	}
	data := make([]byte, length)
	if _, err = io.ReadFull(source, data); err != nil {
		return Transfer{}, err
	}
	digest := sha256.Sum256(data)
	want, _ := hex.DecodeString(expectedSHA256)
	if subtle.ConstantTimeCompare(digest[:], want) != 1 {
		return Transfer{}, ErrCorrupt
	}
	// Read exactly the declared chunk; a malicious extra byte must not be
	// silently consumed as part of this protocol frame.
	var excess [1]byte
	n, e := source.Read(excess[:])
	if n > 0 || e == nil {
		return Transfer{}, ErrInvalid
	}
	if e != nil && !errors.Is(e, io.EOF) {
		return Transfer{}, e
	}
	pf, err := os.OpenFile(s.payloadPath(id), os.O_RDWR, 0600)
	if err != nil {
		return Transfer{}, err
	}
	defer pf.Close()
	if offset < tr.ReceivedBytes {
		old := make([]byte, length)
		if _, err = pf.ReadAt(old, offset); err != nil {
			return Transfer{}, ErrCorrupt
		}
		oldsum := sha256.Sum256(old)
		if subtle.ConstantTimeCompare(oldsum[:], want) != 1 {
			return Transfer{}, ErrConflict
		}
		return tr, nil
	}
	if _, err = pf.WriteAt(data, offset); err != nil {
		return Transfer{}, err
	}
	if err = pf.Sync(); err != nil {
		return Transfer{}, err
	}
	tr.ReceivedBytes += int64(length)
	tr.UpdatedAt = s.now().UTC()
	if err = atomicManifest(s.manifestPath(id), tr); err != nil {
		return Transfer{}, err
	}
	return tr, nil
}

// Complete verifies the entire staged payload, including size and SHA-256,
// and records a durable completion state. It does NOT overwrite a live file.
func (s *Store) Complete(id string) (Transfer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tr, err := s.loadLocked(id)
	if err != nil {
		return Transfer{}, err
	}
	if tr.Status == StatusComplete {
		return tr, nil
	}
	if tr.ReceivedBytes != tr.TotalBytes {
		return Transfer{}, ErrIncomplete
	}
	f, err := os.Open(s.payloadPath(id))
	if err != nil {
		return Transfer{}, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return Transfer{}, err
	}
	if n != tr.TotalBytes {
		return Transfer{}, ErrCorrupt
	}
	want, _ := hex.DecodeString(tr.SHA256)
	if subtle.ConstantTimeCompare(h.Sum(nil), want) != 1 {
		return Transfer{}, ErrCorrupt
	}
	tr.Status = StatusComplete
	tr.UpdatedAt = s.now().UTC()
	if err = atomicManifest(s.manifestPath(id), tr); err != nil {
		return Transfer{}, err
	}
	return tr, nil
}

// OpenCompleted opens only a completed, checksum-validated staged transfer.
// The returned handle must be closed by the caller. Promotion into a remote
// endpoint is exclusively the responsibility of the future sync adapter.
func (s *Store) OpenCompleted(id string) (*os.File, Transfer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tr, err := s.loadLocked(id)
	if err != nil {
		return nil, Transfer{}, err
	}
	if tr.Status != StatusComplete {
		return nil, Transfer{}, ErrIncomplete
	}
	f, err := os.Open(s.payloadPath(id))
	if err != nil {
		return nil, Transfer{}, err
	}
	h := sha256.New()
	size, e := io.Copy(h, f)
	want, _ := hex.DecodeString(tr.SHA256)
	if e != nil || size != tr.TotalBytes || subtle.ConstantTimeCompare(h.Sum(nil), want) != 1 {
		_ = f.Close()
		return nil, Transfer{}, ErrCorrupt
	}
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		_ = f.Close()
		return nil, Transfer{}, err
	}
	return f, tr, nil
}
func (s *Store) Abort(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.loadLocked(id); err != nil {
		return err
	}
	return os.RemoveAll(s.subdir(id))
}

// Sweep discards stale uncompleted sessions only. A completed session stays
// available for explicit promotion or manual removal; it is not expired here.
func (s *Store) Sweep(olderThan time.Duration) (int, error) {
	if olderThan <= 0 {
		return 0, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	items, err := s.listLocked()
	if err != nil {
		return 0, err
	}
	cutoff := s.now().Add(-olderThan)
	removed := 0
	for _, tr := range items {
		if tr.Status == StatusPending && tr.UpdatedAt.Before(cutoff) {
			if err := os.RemoveAll(s.subdir(tr.ID)); err != nil {
				return removed, err
			}
			removed++
		}
	}
	return removed, nil
}
