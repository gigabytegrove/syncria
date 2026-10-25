// Package agentversion retains verified immutable historical file versions on
// the destination side. It deliberately does not modify existing live files:
// callers must approve both the source and any eventual restore destination.
package agentversion

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
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

var (
	ErrNotFound  = errors.New("version not found")
	ErrCorrupted = errors.New("saved version failed SHA256 integrity validation")
	ErrQuota     = errors.New("version storage quota exceeded")
)

const manifestLimit = 8 << 20

type Version struct {
	ID       string    `json:"id"`
	JobID    string    `json:"job_id"`
	Relative string    `json:"relative"`
	Bytes    int64     `json:"bytes"`
	SHA256   string    `json:"sha256"`
	Created  time.Time `json:"created"`
}
type manifest struct {
	Schema   int       `json:"schema"`
	Versions []Version `json:"versions"`
}
type Store struct {
	mu        sync.Mutex
	directory string
	data      manifest
	MaxBytes  int64
}

func validKey(x string) bool {
	if len(x) != 32 {
		return false
	}
	for _, c := range x {
		if c < '0' || c > '9' {
			if c < 'a' || c > 'f' {
				return false
			}
		}
	}
	return true
}
func validRelative(s string) bool {
	if len(s) == 0 || len(s) > 2048 || strings.HasPrefix(s, "/") || strings.ContainsAny(s, "\\\x00\r\n") || path.Clean(s) != s {
		return false
	}
	for _, c := range strings.Split(s, "/") {
		if c == "" || c == "." || c == ".." || strings.HasPrefix(c, ".syncria-") {
			return false
		}
	}
	return true
}
func privateDir(dir string) error {
	if !filepath.IsAbs(dir) {
		return errors.New("absolute version directory required")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || (runtime.GOOS != "windows" && info.Mode().Perm()&0077 != 0) {
		return errors.New("version directory must be private and not a symlink")
	}
	return nil
}
func Open(dir string, limitBytes int64) (*Store, error) {
	if limitBytes <= 0 || limitBytes > 1<<40 {
		return nil, errors.New("invalid archive size quota")
	}
	if err := privateDir(dir); err != nil {
		return nil, err
	}
	s := &Store{directory: dir, MaxBytes: limitBytes, data: manifest{Schema: 1, Versions: []Version{}}}
	file := filepath.Join(dir, "manifest.json")
	stat, err := os.Lstat(file)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if !stat.Mode().IsRegular() || (runtime.GOOS != "windows" && stat.Mode().Perm()&0077 != 0) || stat.Size() > manifestLimit {
		return nil, errors.New("archive manifest not a private regular file")
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := json.NewDecoder(io.LimitReader(f, manifestLimit+1))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&s.data); err != nil {
		return nil, err
	}
	var extra any
	if !errors.Is(dec.Decode(&extra), io.EOF) {
		return nil, errors.New("invalid archive manifest trailing data")
	}
	if s.data.Schema != 1 || s.data.Versions == nil {
		return nil, errors.New("unsupported archive manifest")
	}
	seen := map[string]bool{}
	total := int64(0)
	for _, v := range s.data.Versions {
		if !validKey(v.ID) || !validKey(v.JobID) || !validRelative(v.Relative) || v.Created.IsZero() || v.Bytes < 0 || v.Bytes > limitBytes || seen[v.ID] {
			return nil, errors.New("invalid archive manifest entry")
		}
		decoded, e := hex.DecodeString(v.SHA256)
		if e != nil || len(decoded) != sha256.Size {
			return nil, errors.New("invalid archive digest")
		}
		seen[v.ID] = true
		if v.Bytes > limitBytes-total {
			return nil, ErrQuota
		}
		total += v.Bytes
		info, e := os.Lstat(s.blobPath(v.ID))
		if e != nil || !info.Mode().IsRegular() || info.Size() != v.Bytes {
			return nil, errors.New("version blob missing, changed or unsafe")
		}
	}
	return s, nil
}
func newID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
func (s *Store) blobPath(id string) string { return filepath.Join(s.directory, id+".blob") }
func (s *Store) persist(entries []Version) error {
	b, err := json.MarshalIndent(manifest{Schema: 1, Versions: entries}, "", "  ")
	if err != nil {
		return err
	}
	if len(b) > manifestLimit {
		return errors.New("version manifest too large")
	}
	tmp, err := os.CreateTemp(s.directory, ".manifest-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp.Name(), filepath.Join(s.directory, "manifest.json")); err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		if d, e := os.Open(s.directory); e == nil {
			_ = d.Sync()
			_ = d.Close()
		}
	}
	return nil
}
func (s *Store) used() int64 {
	var n int64
	for _, v := range s.data.Versions {
		n += v.Bytes
	}
	return n
}

// Capture preserves a source as a new immutable version, verifies exact length,
// and commits the metadata only AFTER the content is flushed to stable storage.
// It does not replace, delete or mutate the live destination file.
func (s *Store) Capture(ctx context.Context, jobID, relative string, source io.Reader, size int64) (Version, error) {
	if !validKey(jobID) || !validRelative(relative) || size < 0 || size > s.MaxBytes {
		return Version{}, errors.New("invalid version input")
	}
	if source == nil {
		return Version{}, errors.New("missing source")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if size > s.MaxBytes-s.used() {
		return Version{}, ErrQuota
	}
	id, err := newID()
	if err != nil {
		return Version{}, err
	}
	f, err := os.CreateTemp(s.directory, ".version-*.tmp")
	if err != nil {
		return Version{}, err
	}
	defer os.Remove(f.Name())
	if err := f.Chmod(0600); err != nil {
		f.Close()
		return Version{}, err
	}
	h := sha256.New()
	writer := io.MultiWriter(f, h)
	b := make([]byte, 128<<10)
	var transferred int64
	for {
		if err := ctx.Err(); err != nil {
			f.Close()
			return Version{}, err
		}
		n, e := io.LimitReader(source, size+1-transferred).Read(b)
		if n > 0 {
			transferred += int64(n)
			if transferred > size {
				f.Close()
				return Version{}, errors.New("archive source too large")
			}
			if _, err := writer.Write(b[:n]); err != nil {
				f.Close()
				return Version{}, err
			}
		}
		if e == io.EOF {
			break
		}
		if e != nil {
			f.Close()
			return Version{}, e
		}
		if n == 0 {
			f.Close()
			return Version{}, io.ErrNoProgress
		}
	}
	if transferred != size {
		f.Close()
		return Version{}, io.ErrUnexpectedEOF
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return Version{}, err
	}
	if err := f.Close(); err != nil {
		return Version{}, err
	}
	destination := s.blobPath(id)
	if err := os.Link(f.Name(), destination); err != nil {
		return Version{}, err
	}
	v := Version{ID: id, JobID: jobID, Relative: relative, Bytes: size, SHA256: hex.EncodeToString(h.Sum(nil)), Created: time.Now().UTC()}
	next := append(append([]Version(nil), s.data.Versions...), v)
	if err := s.persist(next); err != nil {
		_ = os.Remove(destination)
		return Version{}, err
	}
	s.data.Versions = next
	return v, nil
}
func (s *Store) List(jobID, relative string) []Version {
	s.mu.Lock()
	defer s.mu.Unlock()
	versions := []Version{}
	for _, v := range s.data.Versions {
		if (jobID == "" || v.JobID == jobID) && (relative == "" || v.Relative == relative) {
			versions = append(versions, v)
		}
	}
	sort.Slice(versions, func(i, j int) bool { return versions[i].Created.After(versions[j].Created) })
	return versions
}

// OpenVerified re-hashes the same open file handle before returning it. The
// consumer receives a stable descriptor positioned at offset 0, not a pathname.
func (s *Store) OpenVerified(ctx context.Context, id string) (io.ReadCloser, Version, error) {
	if !validKey(id) {
		return nil, Version{}, ErrNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var found Version
	ok := false
	for _, v := range s.data.Versions {
		if v.ID == id {
			found = v
			ok = true
			break
		}
	}
	if !ok {
		return nil, Version{}, ErrNotFound
	}
	f, err := os.Open(s.blobPath(id))
	if err != nil {
		return nil, Version{}, err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, Version{}, err
	}
	if !info.Mode().IsRegular() || info.Size() != found.Bytes {
		f.Close()
		return nil, Version{}, ErrCorrupted
	}
	h := sha256.New()
	buf := make([]byte, 128<<10)
	var copied int64
	for {
		if err := ctx.Err(); err != nil {
			f.Close()
			return nil, Version{}, err
		}
		n, e := f.Read(buf)
		if n > 0 {
			copied += int64(n)
			_, _ = h.Write(buf[:n])
		}
		if e == io.EOF {
			break
		}
		if e != nil {
			f.Close()
			return nil, Version{}, e
		}
	}
	if copied != found.Bytes || !strings.EqualFold(hex.EncodeToString(h.Sum(nil)), found.SHA256) {
		f.Close()
		return nil, Version{}, ErrCorrupted
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		f.Close()
		return nil, Version{}, err
	}
	return f, found, nil
}

// Prune removes versions beyond the newest keepPerFile entries for each file.
// Commit the new manifest first; a crash may leave only orphaned private blobs,
// never a manifest referencing deleted version data.
func (s *Store) Prune(keepPerFile int) (int, error) {
	if keepPerFile < 1 || keepPerFile > 10000 {
		return 0, errors.New("invalid retention count")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	sorted := append([]Version(nil), s.data.Versions...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Created.After(sorted[j].Created) })
	kept := []Version{}
	deleted := []Version{}
	count := map[string]int{}
	for _, v := range sorted {
		k := v.JobID + "|" + v.Relative
		count[k]++
		if count[k] <= keepPerFile {
			kept = append(kept, v)
		} else {
			deleted = append(deleted, v)
		}
	}
	if len(deleted) == 0 {
		return 0, nil
	}
	if err := s.persist(kept); err != nil {
		return 0, err
	}
	s.data.Versions = kept
	for _, v := range deleted {
		if err := os.Remove(s.blobPath(v.ID)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return len(deleted), fmt.Errorf("manifest committed but blob removal failed: %w", err)
		}
	}
	return len(deleted), nil
}
