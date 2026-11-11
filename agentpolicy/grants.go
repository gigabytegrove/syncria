// Package agentpolicy holds locally approved device folders and selective sync
// rules. A controller is not allowed to create or modify these grants.
// The grant resolver is a policy gate, NOT a race-safe filesystem opener.
package agentpolicy

import (
	"crypto/rand"
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
	ErrInvalid   = errors.New("invalid folder grant or selection")
	ErrMissing   = errors.New("folder grant not found")
	ErrDenied    = errors.New("path not approved for synchronization")
	ErrDuplicate = errors.New("overlapping approved folder")
)

const maxFileSize = 1 << 20

type Grant struct {
	ID         string    `json:"id"`
	Label      string    `json:"label"`
	Root       string    `json:"root"`
	ReadOnly   bool      `json:"read_only"`
	Selections []string  `json:"selections,omitempty"`
	Exclusions []string  `json:"exclusions,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
type state struct {
	Schema int     `json:"schema"`
	Grants []Grant `json:"grants"`
}
type Store struct {
	mu    sync.Mutex
	file  string
	state state
}

func validName(name string) bool {
	return name != "" && len(name) <= 128 && strings.TrimSpace(name) == name && !strings.ContainsAny(name, "\x00\r\n")
}
func validSegment(segment string) bool {
	if segment == "" || segment == "." || segment == ".." || len(segment) > 255 || strings.HasPrefix(segment, ".syncria-") {
		return false
	}
	for _, r := range segment {
		if r < 32 || r == 127 || r == '\\' {
			return false
		}
	}
	return true
}
func validSelection(relative string) bool {
	if relative == "" || len(relative) > 2048 || strings.HasPrefix(relative, "/") || strings.ContainsRune(relative, '\\') || path.Clean(relative) != relative {
		return false
	}
	for _, part := range strings.Split(relative, "/") {
		if !validSegment(part) {
			return false
		}
	}
	return true
}
func cleanSelectionList(values []string) ([]string, error) {
	if len(values) > 256 {
		return nil, ErrInvalid
	}
	unique := make(map[string]struct{}, len(values))
	for _, value := range values {
		if !validSelection(value) {
			return nil, ErrInvalid
		}
		if _, exists := unique[value]; exists {
			return nil, ErrInvalid
		}
		unique[value] = struct{}{}
	}
	out := make([]string, 0, len(unique))
	for v := range unique {
		out = append(out, v)
	}
	sort.Strings(out)
	return out, nil
}
func validID(id string) bool {
	if len(id) != 32 {
		return false
	}
	for _, r := range id {
		if r < '0' || r > '9' {
			if r < 'a' || r > 'f' {
				return false
			}
		}
	}
	return true
}
func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
func canonicalRoot(root string) (string, error) {
	if !filepath.IsAbs(root) {
		return "", ErrInvalid
	}
	clean := filepath.Clean(root)
	real, err := filepath.EvalSymlinks(clean)
	if err != nil {
		return "", err
	}
	if real != clean {
		return "", errors.New("symlinked folder roots are not permitted")
	}
	info, err := os.Lstat(clean)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", errors.New("approved root must be an existing directory")
	}
	return clean, nil
}
func overlap(a, b string) bool {
	a = filepath.Clean(a)
	b = filepath.Clean(b)
	return a == b || strings.HasPrefix(a, b+string(filepath.Separator)) || strings.HasPrefix(b, a+string(filepath.Separator))
}
func cloneGrant(g Grant) Grant {
	g.Selections = append([]string(nil), g.Selections...)
	g.Exclusions = append([]string(nil), g.Exclusions...)
	return g
}
func (s *Store) persist(next state) error {
	body, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return err
	}
	if len(body) > maxFileSize {
		return ErrInvalid
	}
	dir := filepath.Dir(s.file)
	if err = os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	info, err := os.Lstat(s.file)
	if err == nil {
		if !info.Mode().IsRegular() {
			return errors.New("grant file cannot be a symlink or special file")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	f, err := os.CreateTemp(dir, ".agent-grants-*.tmp")
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
	if err = os.Rename(f.Name(), s.file); err != nil {
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
func Open(file string) (*Store, error) {
	if file == "" || !filepath.IsAbs(file) {
		return nil, ErrInvalid
	}
	s := &Store{file: file, state: state{Schema: 1, Grants: []Grant{}}}
	info, err := os.Lstat(file)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maxFileSize {
		return nil, ErrInvalid
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0077 != 0 {
		return nil, errors.New("grant file must be private")
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	decoder := json.NewDecoder(io.LimitReader(f, maxFileSize+1))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&s.state); err != nil {
		return nil, err
	}
	var extra any
	if err = decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, ErrInvalid
	}
	if s.state.Schema != 1 || s.state.Grants == nil || len(s.state.Grants) > 1000 {
		return nil, ErrInvalid
	}
	ids := make(map[string]bool)
	for i, g := range s.state.Grants {
		if !validID(g.ID) || !validName(g.Label) || !filepath.IsAbs(g.Root) || filepath.Clean(g.Root) != g.Root || ids[g.ID] || g.CreatedAt.IsZero() {
			return nil, fmt.Errorf("grant %d invalid: %w", i, ErrInvalid)
		}
		if _, err = cleanSelectionList(g.Selections); err != nil {
			return nil, err
		}
		if _, err = cleanSelectionList(g.Exclusions); err != nil {
			return nil, err
		}
		ids[g.ID] = true
		for j := 0; j < i; j++ {
			if overlap(g.Root, s.state.Grants[j].Root) {
				return nil, ErrDuplicate
			}
		}
	}
	return s, nil
}

// Add must be invoked locally with permission from the device owner. The
// controller cannot remotely grant itself new folders.
func (s *Store) Add(label, root string, readOnly bool, selections, exclusions []string) (Grant, error) {
	if !validName(label) {
		return Grant{}, ErrInvalid
	}
	resolved, err := canonicalRoot(root)
	if err != nil {
		return Grant{}, err
	}
	selected, err := cleanSelectionList(selections)
	if err != nil {
		return Grant{}, err
	}
	excluded, err := cleanSelectionList(exclusions)
	if err != nil {
		return Grant{}, err
	}
	id, err := newID()
	if err != nil {
		return Grant{}, err
	}
	g := Grant{ID: id, Label: label, Root: resolved, ReadOnly: readOnly, Selections: selected, Exclusions: excluded, CreatedAt: time.Now().UTC()}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.state.Grants) >= 1000 {
		return Grant{}, ErrInvalid
	}
	for _, prev := range s.state.Grants {
		if overlap(prev.Root, g.Root) {
			return Grant{}, ErrDuplicate
		}
	}
	next := state{Schema: 1, Grants: append(append([]Grant(nil), s.state.Grants...), g)}
	if err = s.persist(next); err != nil {
		return Grant{}, err
	}
	s.state = next
	return cloneGrant(g), nil
}
func (s *Store) List() []Grant {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Grant, 0, len(s.state.Grants))
	for _, g := range s.state.Grants {
		out = append(out, cloneGrant(g))
	}
	return out
}
func (s *Store) Remove(id string) error {
	if !validID(id) {
		return ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	next := state{Schema: 1, Grants: make([]Grant, 0, len(s.state.Grants))}
	found := false
	for _, g := range s.state.Grants {
		if g.ID == id {
			found = true
		} else {
			next.Grants = append(next.Grants, g)
		}
	}
	if !found {
		return ErrMissing
	}
	if err := s.persist(next); err != nil {
		return err
	}
	s.state = next
	return nil
}
func selected(relative string, prefixes []string) bool {
	for _, root := range prefixes {
		if relative == root || strings.HasPrefix(relative, root+"/") {
			return true
		}
	}
	return false
}

// Allows checks preapproved relative paths and selection rules. It does not
// open files: callers must use platform-specific handle-relative traversal
// that rejects symlinks and checks every filesystem component at operation time.
func (s *Store) Allows(grantID, relative string, write bool) error {
	if !validID(grantID) || !validSelection(relative) {
		return ErrDenied
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, g := range s.state.Grants {
		if g.ID == grantID {
			if write && g.ReadOnly {
				return ErrDenied
			}
			if selected(relative, g.Exclusions) {
				return ErrDenied
			}
			if len(g.Selections) > 0 && !selected(relative, g.Selections) {
				return ErrDenied
			}
			return nil
		}
	}
	return ErrDenied
}

// Get returns an independent copy of a locally approved grant. File access
// callers must still check Allows for each requested relative path and use
// handle-relative filesystem operations; Get does not grant additional rights.
func (s *Store) Get(id string) (Grant, error) {
	if !validID(id) {
		return Grant{}, ErrMissing
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, grant := range s.state.Grants {
		if grant.ID == id {
			return cloneGrant(grant), nil
		}
	}
	return Grant{}, ErrMissing
}
