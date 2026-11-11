// Package agentcontrol manages durable device enrollment for Syncria.
// It grants device identity, never permission to access any filesystem path.
package agentcontrol

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalid = errors.New("invalid enrollment or credential")
	ErrExpired = errors.New("enrollment invitation expired")
	ErrPending = errors.New("administrator approval required")
	ErrUsed    = errors.New("enrollment invitation already redeemed")
	ErrRevoked = errors.New("device revoked")
	ErrMissing = errors.New("device or invitation not found")
)

const maxEvents = 2000

type Invitation struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	TokenHash  string     `json:"token_hash"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  time.Time  `json:"expires_at"`
	ApprovedAt *time.Time `json:"approved_at,omitempty"`
	RedeemedAt *time.Time `json:"redeemed_at,omitempty"`
}
type InvitationView struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	ExpiresAt time.Time `json:"expires_at"`
	Approved  bool      `json:"approved"`
	Redeemed  bool      `json:"redeemed"`
}
type Device struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Platform       string     `json:"platform"`
	Arch           string     `json:"arch"`
	Version        string     `json:"version"`
	CredentialHash string     `json:"credential_hash"`
	EnrolledAt     time.Time  `json:"enrolled_at"`
	LastSeen       *time.Time `json:"last_seen,omitempty"`
	RevokedAt      *time.Time `json:"revoked_at,omitempty"`
}
type DeviceView struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Platform   string     `json:"platform"`
	Arch       string     `json:"arch"`
	Version    string     `json:"version"`
	EnrolledAt time.Time  `json:"enrolled_at"`
	LastSeen   *time.Time `json:"last_seen,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}
type Event struct {
	At      time.Time `json:"at"`
	Kind    string    `json:"kind"`
	Subject string    `json:"subject"`
}
type state struct {
	Schema      int                   `json:"schema"`
	Invitations map[string]Invitation `json:"invitations"`
	Devices     map[string]Device     `json:"devices"`
	Events      []Event               `json:"events"`
}
type Store struct {
	mu    sync.Mutex
	path  string
	state state
	clock func() time.Time
}

func newState() state {
	return state{Schema: 1, Invitations: map[string]Invitation{}, Devices: map[string]Device{}}
}
func validLabel(v string, max int) bool {
	return len(v) > 0 && len(v) <= max && strings.TrimSpace(v) == v && !strings.ContainsAny(v, "\x00\r\n")
}
func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func tokenHash(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
func validHex(s string) bool {
	if len(s) != 64 {
		return false
	}
	_, e := hex.DecodeString(s)
	return e == nil
}
func (s *Store) now() time.Time { return s.clock().UTC() }

// Open refuses malformed, incompatible, symlinked or loosely permissioned state.
// Call with a path inside Syncria's existing private data directory.
func Open(path string) (*Store, error) {
	if path == "" || !filepath.IsAbs(path) {
		return nil, errors.New("absolute device state path required")
	}
	s := &Store{path: path, state: newState(), clock: time.Now}
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("agent state must be a regular file")
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0077 != 0 {
		return nil, errors.New("agent state permissions must be owner-only")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(b) > 16<<20 {
		return nil, errors.New("agent state too large")
	}
	if err := json.Unmarshal(b, &s.state); err != nil {
		return nil, fmt.Errorf("decode agent state: %w", err)
	}
	if s.state.Schema != 1 || s.state.Invitations == nil || s.state.Devices == nil {
		return nil, errors.New("unsupported or malformed agent state")
	}
	for k, v := range s.state.Invitations {
		if k != v.ID || !validHex(v.TokenHash) {
			return nil, errors.New("invalid invitation record")
		}
	}
	for k, v := range s.state.Devices {
		if k != v.ID || !validHex(v.CredentialHash) {
			return nil, errors.New("invalid device record")
		}
	}
	return s, nil
}
func cloneState(v state) state {
	n := state{Schema: v.Schema, Invitations: make(map[string]Invitation, len(v.Invitations)), Devices: make(map[string]Device, len(v.Devices)), Events: append([]Event(nil), v.Events...)}
	for k, x := range v.Invitations {
		n.Invitations[k] = x
	}
	for k, x := range v.Devices {
		n.Devices[k] = x
	}
	return n
}
func (s *Store) persist(v state) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if len(b) > 16<<20 {
		return errors.New("agent state exceeds maximum size")
	}
	if err = os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return err
	}
	// Never follow a replacement symlink during normal updates.
	if info, e := os.Lstat(s.path); e == nil && !info.Mode().IsRegular() {
		return errors.New("agent state path is not regular")
	} else if e != nil && !errors.Is(e, os.ErrNotExist) {
		return e
	}
	f, err := os.CreateTemp(filepath.Dir(s.path), ".agents-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = f.Chmod(0600); err != nil {
		f.Close()
		return err
	}
	if _, err = f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(f.Name(), s.path); err != nil {
		return err
	}
	// Directory sync is not supported on every OS (notably Windows).
	if runtime.GOOS != "windows" {
		if d, e := os.Open(filepath.Dir(s.path)); e == nil {
			_ = d.Sync()
			_ = d.Close()
		}
	}
	return nil
}
func record(v *state, when time.Time, kind, subject string) {
	v.Events = append(v.Events, Event{At: when, Kind: kind, Subject: subject})
	if len(v.Events) > maxEvents {
		v.Events = append([]Event(nil), v.Events[len(v.Events)-maxEvents:]...)
	}
}
func (s *Store) change(fn func(*state) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := cloneState(s.state)
	if err := fn(&n); err != nil {
		return err
	}
	if err := s.persist(n); err != nil {
		return err
	}
	s.state = n
	return nil
}
func (s *Store) Issue(name string, ttl time.Duration) (InvitationView, string, error) {
	if !validLabel(name, 128) || ttl < time.Minute || ttl > 30*time.Minute {
		return InvitationView{}, "", errors.New("invalid name or TTL")
	}
	id, err := randomToken(16)
	if err != nil {
		return InvitationView{}, "", err
	}
	token, err := randomToken(32)
	if err != nil {
		return InvitationView{}, "", err
	}
	now := s.now()
	inv := Invitation{ID: id, Name: name, TokenHash: tokenHash(token), CreatedAt: now, ExpiresAt: now.Add(ttl)}
	err = s.change(func(v *state) error { v.Invitations[id] = inv; record(v, now, "invitation_created", id); return nil })
	return InvitationView{ID: id, Name: name, ExpiresAt: inv.ExpiresAt}, token, err
}
func (s *Store) Approve(id string) error {
	return s.change(func(v *state) error {
		x, ok := v.Invitations[id]
		if !ok {
			return ErrMissing
		}
		now := s.now()
		if !now.Before(x.ExpiresAt) {
			return ErrExpired
		}
		if x.RedeemedAt != nil {
			return ErrUsed
		}
		if x.ApprovedAt != nil {
			return nil
		}
		x.ApprovedAt = &now
		v.Invitations[id] = x
		record(v, now, "invitation_approved", id)
		return nil
	})
}
func (s *Store) Redeem(token, platform, arch, version string) (DeviceView, string, error) {
	if len(token) != 43 || !validLabel(platform, 32) || !validLabel(arch, 32) || !validLabel(version, 64) {
		return DeviceView{}, "", ErrInvalid
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) != 32 || base64.RawURLEncoding.EncodeToString(raw) != token {
		return DeviceView{}, "", ErrInvalid
	}
	secret, err := randomToken(32)
	if err != nil {
		return DeviceView{}, "", err
	}
	credential := tokenHash(secret)
	now := s.now()
	var device Device
	err = s.change(func(v *state) error {
		match := ""
		want := sha256.Sum256([]byte(token))
		for id, inv := range v.Invitations {
			digest, e := hex.DecodeString(inv.TokenHash)
			if e == nil && subtle.ConstantTimeCompare(digest, want[:]) == 1 {
				match = id
			}
		}
		if match == "" {
			return ErrInvalid
		}
		inv := v.Invitations[match]
		if !now.Before(inv.ExpiresAt) {
			return ErrExpired
		}
		if inv.RedeemedAt != nil {
			return ErrUsed
		}
		if inv.ApprovedAt == nil {
			return ErrPending
		}
		device = Device{ID: inv.ID, Name: inv.Name, Platform: platform, Arch: arch, Version: version, CredentialHash: credential, EnrolledAt: now}
		if _, exists := v.Devices[device.ID]; exists {
			return ErrUsed
		}
		inv.RedeemedAt = &now
		v.Invitations[match] = inv
		v.Devices[device.ID] = device
		record(v, now, "device_enrolled", device.ID)
		return nil
	})
	if err != nil {
		return DeviceView{}, "", err
	}
	return publicDevice(device), secret, nil
}
func publicDevice(d Device) DeviceView {
	return DeviceView{ID: d.ID, Name: d.Name, Platform: d.Platform, Arch: d.Arch, Version: d.Version, EnrolledAt: d.EnrolledAt, LastSeen: d.LastSeen, RevokedAt: d.RevokedAt}
}
func (s *Store) authenticate(v state, id, secret string) error {
	d, ok := v.Devices[id]
	if !ok {
		return ErrInvalid
	}
	if d.RevokedAt != nil {
		return ErrRevoked
	}
	if len(secret) != 43 {
		return ErrInvalid
	}
	supplied := sha256.Sum256([]byte(secret))
	expected, e := hex.DecodeString(d.CredentialHash)
	if e != nil || subtle.ConstantTimeCompare(supplied[:], expected) != 1 {
		return ErrInvalid
	}
	return nil
}
func (s *Store) Heartbeat(id, secret, platform, arch, version string) (DeviceView, error) {
	if !validLabel(platform, 32) || !validLabel(arch, 32) || !validLabel(version, 64) {
		return DeviceView{}, ErrInvalid
	}
	var result DeviceView
	err := s.change(func(v *state) error {
		if err := s.authenticate(*v, id, secret); err != nil {
			return err
		}
		d := v.Devices[id]
		now := s.now()
		d.LastSeen = &now
		d.Platform = platform
		d.Arch = arch
		d.Version = version
		v.Devices[id] = d
		result = publicDevice(d)
		return nil
	})
	return result, err
}
func (s *Store) Revoke(id string) error {
	return s.change(func(v *state) error {
		d, ok := v.Devices[id]
		if !ok {
			return ErrMissing
		}
		if d.RevokedAt != nil {
			return nil
		}
		now := s.now()
		d.RevokedAt = &now
		v.Devices[id] = d
		record(v, now, "device_revoked", id)
		return nil
	})
}
func (s *Store) ListDevices() []DeviceView {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := make([]DeviceView, 0, len(s.state.Devices))
	for _, d := range s.state.Devices {
		items = append(items, publicDevice(d))
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name || items[i].Name == items[j].Name && items[i].ID < items[j].ID
	})
	return items
}
func (s *Store) ListInvitations() []InvitationView {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := make([]InvitationView, 0, len(s.state.Invitations))
	for _, x := range s.state.Invitations {
		items = append(items, InvitationView{ID: x.ID, Name: x.Name, ExpiresAt: x.ExpiresAt, Approved: x.ApprovedAt != nil, Redeemed: x.RedeemedAt != nil})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ExpiresAt.After(items[j].ExpiresAt) })
	return items
}
func (s *Store) Events() []Event {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Event(nil), s.state.Events...)
}
