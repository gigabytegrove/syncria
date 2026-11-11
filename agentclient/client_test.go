package agentclient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gigabytegrove/syncria/agentcontrol"
)

func TestServerValidation(t *testing.T) {
	for _, ok := range []string{"https://syncria.example", "https://syncria.example:8443", "http://127.0.0.1:9764", "http://localhost:9764"} {
		if _, e := ValidServer(ok); e != nil {
			t.Errorf("rejected %q: %v", ok, e)
		}
	}
	for _, bad := range []string{"http://example.com", "ftp://localhost", "https://user:pass@example.com", "https://syncria.example/a", "https://syncria.example?token=abc", "https://syncria.example/#fragment", ""} {
		if _, e := ValidServer(bad); e == nil {
			t.Errorf("accepted %q", bad)
		}
	}
}
func TestSaveAndReloadPrivateIdentity(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent", "identity.json")
	x := Credentials{Server: "https://example.com", ID: "123", Name: "host", Secret: strings.Repeat("x", 43), EnrolledAt: time.Now().UTC()}
	if err := Save(path, x); err != nil {
		t.Fatal(err)
	}
	read, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if read.ID != x.ID || read.Secret != x.Secret {
		t.Fatal("stored identity mismatch")
	}
	if _, err := os.Lstat(path); err != nil {
		t.Fatal(err)
	}
	if err := Save(path, x); err == nil {
		t.Fatal("existing credential overwritten")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(t.TempDir(), "fake"), path); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	if _, err := Read(path); err == nil {
		t.Fatal("symlinked credential accepted")
	}
	if err := Save(path, x); err == nil {
		t.Fatal("overwrote symlink")
	}
}
func TestAgentClientEndToEnd(t *testing.T) {
	store, err := agentcontrol.Open(filepath.Join(t.TempDir(), "agents.json"))
	if err != nil {
		t.Fatal(err)
	}
	h := agentcontrol.Handler{Store: store, Admin: func(r *http.Request) bool { return true }, SameOriginPost: func(r *http.Request) bool { return true }}
	server := httptest.NewTLSServer(h)
	defer server.Close()
	client := Client{HTTP: server.Client()}
	invite, token, err := store.Issue("Test Device", 10*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Enroll(context.Background(), server.URL, token, "linux", "amd64", "1.0.0"); err == nil {
		t.Fatal("enrolled before approval")
	}
	if err := store.Approve(invite.ID); err != nil {
		t.Fatal(err)
	}
	creds, err := client.Enroll(context.Background(), server.URL, token, "linux", "amd64", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if creds.ID != invite.ID || creds.Name != "Test Device" {
		t.Fatalf("unexpected credentials: %+v", creds)
	}
	if err := client.Heartbeat(context.Background(), creds, "linux", "amd64", "1.0.1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Enroll(context.Background(), server.URL, token, "linux", "amd64", "1.0.0"); err == nil {
		t.Fatal("redeemed twice")
	}
	if err := store.Revoke(creds.ID); err != nil {
		t.Fatal(err)
	}
	if err := client.Heartbeat(context.Background(), creds, "linux", "amd64", "1.0.1"); !errors.Is(err, ErrRevoked) {
		t.Fatalf("revocation not detected: %v", err)
	}
}
func TestRedirectDoesNotLeakCredential(t *testing.T) {
	var seen bool
	attacker := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { seen = true }))
	defer attacker.Close()
	redirect := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, attacker.URL, http.StatusTemporaryRedirect)
	}))
	defer redirect.Close()
	identity := Credentials{Server: redirect.URL, ID: "id", Secret: strings.Repeat("x", 43)}
	c := Client{HTTP: redirect.Client()}
	if err := c.Heartbeat(context.Background(), identity, "linux", "amd64", "1.0"); err == nil {
		t.Fatal("redirect accepted")
	}
	if seen {
		t.Fatal("credential forwarded to unrelated origin")
	}
}
func TestNoSecretsInStatusJSON(t *testing.T) {
	x := Identity{ID: "id", Name: "host"}
	b, e := json.Marshal(x)
	if e != nil {
		t.Fatal(e)
	}
	if strings.Contains(string(b), "secret") {
		t.Fatal("public identity unexpectedly contains credential")
	}
}
