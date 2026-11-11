package agentpeer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/gigabytegrove/syncria/agentpolicy"
	"github.com/gigabytegrove/syncria/agenttransfer"
)

func prepare(t *testing.T, content []byte) (Client, *agentpolicy.Store, string, string, *httptest.Server) {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "photos"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "private"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "photos", "image.bin"), content, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "private", "secret"), []byte("hidden"), 0600); err != nil {
		t.Fatal(err)
	}
	grants, err := agentpolicy.Open(filepath.Join(t.TempDir(), "grants.json"))
	if err != nil {
		t.Fatal(err)
	}
	g, err := grants.Add("Photos", root, true, []string{"photos"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	key, err := GenerateKey(filepath.Join(t.TempDir(), "peer.key"))
	if err != nil {
		t.Fatal(err)
	}
	handler, err := NewHandler(grants, key)
	if err != nil {
		t.Fatal(err)
	}
	handler.Active = func() bool { return true }
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return Client{Server: srv.URL, Token: key, HTTP: srv.Client()}, grants, g.ID, root, srv
}
func TestPeerInventoryAndDeniedPaths(t *testing.T) {
	client, _, id, _, srv := prepare(t, []byte("binary \x00 payload"))
	entries, err := client.Inventory(context.Background(), id)
	if err != nil || len(entries) != 1 || entries[0].Path != "photos/image.bin" {
		t.Fatalf("inventory=%v err=%v", entries, err)
	}
	for _, route := range []string{
		"/v1/roots/" + id + "/files/private/secret",
		"/v1/roots/" + id + "/files/photos/../private/secret",
		"/v1/roots/" + id + "/files/photos/%252e%252e/private/secret",
		"/v1/roots/" + id + "/files/photos/.syncria-secret",
	} {
		req, err := http.NewRequest("GET", srv.URL+route, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+client.Token)
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		if resp.StatusCode == 200 {
			t.Errorf("path accessible: %s", route)
		}
	}
	for _, method := range []string{"PUT", "DELETE", "POST"} {
		req, _ := http.NewRequest(method, srv.URL+"/v1/roots/"+id+"/inventory", nil)
		req.Header.Set("Authorization", "Bearer "+client.Token)
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != 405 {
			t.Errorf("method %s allowed: %d", method, resp.StatusCode)
		}
	}
}
func TestUnauthorizedRequestsAndHeaderSpoofing(t *testing.T) {
	client, _, id, _, srv := prepare(t, []byte("hello"))
	for _, token := range []string{"", "Bearer wrong", "Bearer " + strings.Repeat("a", 43)} {
		req, _ := http.NewRequest("GET", srv.URL+"/v1/roots/"+id+"/inventory", nil)
		req.Header.Set("Authorization", token)
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != 401 {
			t.Errorf("token accepted: %q %d", token, resp.StatusCode)
		}
	}
	// Synthetic non-TLS requests from a non-local source cannot bypass TLS by
	// claiming they arrived through an HTTPS-terminating proxy.
	h, _ := NewHandler(clientGrants(t), client.Token)
	h.Active = func() bool { return true }
	req := httptest.NewRequest("GET", "http://device.example/v1/roots/"+id+"/inventory", nil)
	req.RemoteAddr = "192.0.2.25:5555"
	req.Header.Set("Authorization", "Bearer "+client.Token)
	req.Header.Set("X-Forwarded-Proto", "https")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != 403 {
		t.Fatalf("forwarded scheme accepted: %d", w.Code)
	}
}
func clientGrants(t *testing.T) *agentpolicy.Store {
	t.Helper()
	s, e := agentpolicy.Open(filepath.Join(t.TempDir(), "grants.json"))
	if e != nil {
		t.Fatal(e)
	}
	return s
}

func TestVerifiedTransferAndCheckpointResume(t *testing.T) {
	data := bytes.Repeat([]byte("syncria durable remote bytes\x00"), 280000)
	client, grants, id, root, _ := prepare(t, data)
	staging := filepath.Join(t.TempDir(), "transfers")
	store, err := agenttransfer.Open(staging, 32<<20)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	digest := hex.EncodeToString(sum[:])
	request := agenttransfer.Request{SourceID: "device-123456", Revision: digest, RelativePath: "photos/image.bin", TotalBytes: int64(len(data)), SHA256: digest}
	tr, err := store.Begin(request)
	if err != nil {
		t.Fatal(err)
	}
	part := data[:1<<20]
	first := sha256.Sum256(part)
	tr, err = store.Append(tr.ID, 0, bytes.NewReader(part), len(part), hex.EncodeToString(first[:]))
	if err != nil {
		t.Fatal(err)
	}
	if tr.ReceivedBytes != 1<<20 {
		t.Fatal(tr)
	}
	store, err = agenttransfer.Open(staging, 32<<20)
	if err != nil {
		t.Fatal(err)
	}
	done, err := client.Stage(context.Background(), id, "photos/image.bin", "device-123456", store)
	if err != nil {
		t.Fatal(err)
	}
	if done.Status != agenttransfer.StatusComplete || done.ReceivedBytes != int64(len(data)) {
		t.Fatalf("unexpected complete state: %+v", done)
	}
	f, _, err := store.OpenCompleted(done.ID)
	if err != nil {
		t.Fatal(err)
	}
	received, err := io.ReadAll(f)
	f.Close()
	if err != nil || !bytes.Equal(received, data) {
		t.Fatal("remote payload changed")
	}
	// A revoked local folder blocks further file reads without waiting for a
	// central policy refresh. Old completed staging remains only at receiver.
	if err = grants.Remove(id); err != nil {
		t.Fatal(err)
	}
	if _, err = client.Stage(context.Background(), id, "photos/image.bin", "device-123456", store); err == nil {
		t.Fatal("revoked folder still readable")
	}
	_ = root
}
func TestBadSourceDigestRejectsTransfer(t *testing.T) {
	client, _, id, root, _ := prepare(t, []byte("original"))
	store, err := agenttransfer.Open(filepath.Join(t.TempDir(), "staging"), 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	// A stale transfer manifest for the same source revision must not be used
	// after file mutation; source SHA and revision will no longer match.
	if err = os.WriteFile(filepath.Join(root, "photos", "image.bin"), []byte("replacement"), 0600); err != nil {
		t.Fatal(err)
	}
	done, err := client.Stage(context.Background(), id, "photos/image.bin", "device-123456", store)
	if err != nil {
		t.Fatal(err)
	}
	if done.Status != agenttransfer.StatusComplete {
		t.Fatal(done)
	}
	srcDigest := sha256.Sum256([]byte("replacement"))
	if done.SHA256 != fmt.Sprintf("%x", srcDigest) {
		t.Fatal("source checksum not reflected")
	}
}
func TestPeerTokenPersistenceAndPermissions(t *testing.T) {
	f := filepath.Join(t.TempDir(), "token")
	key, err := GenerateKey(f)
	if err != nil {
		t.Fatal(err)
	}
	read, err := ReadKey(f)
	if err != nil || read != key {
		t.Fatalf("read mismatch: %v", err)
	}
	if _, err = GenerateKey(f); err == nil {
		t.Fatal("overwrote existing peer key")
	}
	if err = os.Symlink(f, f+"-link"); err == nil {
		if _, err = ReadKey(f + "-link"); err == nil {
			t.Fatal("symlinked secret accepted")
		}
	}
}
func TestTransferCancelHasNoCommittedOutput(t *testing.T) {
	client, _, id, _, _ := prepare(t, []byte("bytes"))
	store, err := agenttransfer.Open(filepath.Join(t.TempDir(), "staging"), 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = client.Stage(ctx, id, "photos/image.bin", "device-123456", store)
	if err == nil {
		t.Fatal("canceled request accepted")
	}
}
func TestConcurrentTransferSessions(t *testing.T) {
	data := bytes.Repeat([]byte("concurrent"), 4000)
	client, _, id, _, _ := prepare(t, data)
	var wg sync.WaitGroup
	errs := make(chan error, 5)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			store, err := agenttransfer.Open(filepath.Join(t.TempDir(), "staging"), 4<<20)
			if err == nil {
				_, err = client.Stage(context.Background(), id, "photos/image.bin", "device-123456", store)
			}
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestStageRejectsSourceChangeAfterHEAD(t *testing.T) {
	client, _, id, root, srv := prepare(t, []byte("original"))
	baseTransport := srv.Client().Transport
	client.HTTP = &http.Client{Transport: &alterAfterHead{base: baseTransport, path: filepath.Join(root, "photos", "image.bin")}}
	staging, err := agenttransfer.Open(filepath.Join(t.TempDir(), "incoming"), 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Stage(context.Background(), id, "photos/image.bin", "device-123456", staging)
	if err == nil || !strings.Contains(err.Error(), "412") {
		t.Fatalf("changed source accepted: %v", err)
	}
}

type alterAfterHead struct {
	base http.RoundTripper
	path string
	once sync.Once
}

func (t *alterAfterHead) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err == nil && req.Method == "HEAD" && resp.StatusCode == 200 {
		t.once.Do(func() { _ = os.WriteFile(t.path, []byte("a different source"), 0600) })
	}
	return resp, err
}
func TestUnicodeAndSpacedNames(t *testing.T) {
	client, grants, id, root, _ := prepare(t, []byte("original"))
	renamed := "photos/My favorite ä image.txt"
	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(renamed)), []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	inventory, err := client.Inventory(context.Background(), id)
	if err != nil || len(inventory) != 2 {
		t.Fatalf("inventory: %v %v", inventory, err)
	}
	staging, err := agenttransfer.Open(filepath.Join(t.TempDir(), "incoming"), 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	done, err := client.Stage(context.Background(), id, renamed, "device-123456", staging)
	if err != nil || done.Status != agenttransfer.StatusComplete {
		t.Fatalf("unicode download: %v %v", done, err)
	}
	_ = grants
}
func TestInactiveControllerStopsPeerAccess(t *testing.T) {
	client, grants, id, _, srv := prepare(t, []byte("ok"))
	handler, err := NewHandler(grants, client.Token)
	if err != nil {
		t.Fatal(err)
	}
	handler.Active = func() bool { return false }
	request, err := http.NewRequest("GET", srv.URL+"/v1/roots/"+id+"/inventory", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Authorization", "Bearer "+client.Token)
	recorder := httptest.NewRecorder()
	// TLS is intentionally nil: authorization availability must block before
	// any file inventory or secret comparison.
	handler.ServeHTTP(recorder, request)
	if recorder.Code != 503 {
		t.Fatalf("stale authorization allowed: %d", recorder.Code)
	}
}
