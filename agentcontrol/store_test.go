package agentcontrol

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	s, e := Open(filepath.Join(t.TempDir(), "agents.json"))
	if e != nil {
		t.Fatal(e)
	}
	return s
}
func TestEnrollmentPersistenceAndRevocation(t *testing.T) {
	s := testStore(t)
	invite, token, err := s.Issue("Office PC", 10*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if len(token) != 43 {
		t.Fatal("weak pairing token")
	}
	if _, _, err = s.Redeem(token, "windows", "amd64", "1.0.0"); !errors.Is(err, ErrPending) {
		t.Fatalf("unapproved invitation accepted: %v", err)
	}
	if err = s.Approve(invite.ID); err != nil {
		t.Fatal(err)
	}
	d, key, err := s.Redeem(token, "windows", "amd64", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if d.ID != invite.ID || len(key) != 43 {
		t.Fatal("bad issued device credential")
	}
	b, err := os.ReadFile(s.path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), token) || strings.Contains(string(b), key) {
		t.Fatal("plaintext credential persisted")
	}
	if _, _, err = s.Redeem(token, "windows", "amd64", "1.0.0"); !errors.Is(err, ErrUsed) {
		t.Fatalf("token reused: %v", err)
	}
	reopened, err := Open(s.path)
	if err != nil {
		t.Fatal(err)
	}
	if len(reopened.ListDevices()) != 1 {
		t.Fatal("device missing after restart")
	}
	if _, err = reopened.Heartbeat(d.ID, key, "windows", "amd64", "1.0.1"); err != nil {
		t.Fatal(err)
	}
	if _, err = reopened.Heartbeat(d.ID, "bogus", "windows", "amd64", "1.0.1"); !errors.Is(err, ErrInvalid) {
		t.Fatal("unauthorized heartbeat accepted")
	}
	if err = reopened.Revoke(d.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = reopened.Heartbeat(d.ID, key, "windows", "amd64", "1.0.1"); !errors.Is(err, ErrRevoked) {
		t.Fatalf("revoked device active: %v", err)
	}
	reloaded, err := Open(s.path)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.ListDevices()[0].RevokedAt == nil {
		t.Fatal("revocation not persisted")
	}
	if len(reloaded.Events()) != 4 {
		t.Fatalf("missing audit events: %d", len(reloaded.Events()))
	}
}
func TestConcurrentEnrollmentSingleWinner(t *testing.T) {
	s := testStore(t)
	i, code, err := s.Issue("server", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Approve(i.ID); err != nil {
		t.Fatal(err)
	}
	var won atomic.Int64
	var wg sync.WaitGroup
	for j := 0; j < 20; j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, e := s.Redeem(code, "linux", "amd64", "1.0.0")
			if e == nil {
				won.Add(1)
			}
		}()
	}
	wg.Wait()
	if won.Load() != 1 {
		t.Fatalf("redeemed %d times", won.Load())
	}
	if len(s.ListDevices()) != 1 {
		t.Fatal("incorrect device count")
	}
}
func TestExpiryInvalidInputAndCorruptState(t *testing.T) {
	s := testStore(t)
	for _, ttl := range []time.Duration{0, time.Second, 31 * time.Minute} {
		if _, _, err := s.Issue("Test", ttl); err == nil {
			t.Fatal("invalid TTL accepted")
		}
	}
	if _, _, err := s.Issue("\n", time.Minute); err == nil {
		t.Fatal("invalid device name accepted")
	}
	i, code, err := s.Issue("Test", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	s.clock = func() time.Time { return time.Now().Add(time.Hour) }
	if err = s.Approve(i.ID); !errors.Is(err, ErrExpired) {
		t.Fatalf("expired approved: %v", err)
	}
	if _, _, err = s.Redeem(code, "linux", "amd64", "1.0.0"); !errors.Is(err, ErrExpired) {
		t.Fatalf("expired redeemed: %v", err)
	}
	if err = os.WriteFile(s.path, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = Open(s.path); err == nil {
		t.Fatal("corrupt persisted state silently accepted")
	}
}
func TestSymlinkAndPermissions(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "agents.json")
	if err := os.WriteFile(file, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(file); err == nil {
		t.Fatal("world-readable credential file accepted")
	}
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "elsewhere"), file); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := Open(file); err == nil {
		t.Fatal("symlinked credential file accepted")
	}
}
