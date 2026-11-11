package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gigabytegrove/syncria/agentclient"
	"github.com/gigabytegrove/syncria/agentpeer"
	"github.com/gigabytegrove/syncria/agentpolicy"
)

func TestPeerReceiveEndToEndCLI(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-only publication")
	}
	from := t.TempDir()
	to := t.TempDir()
	local := t.TempDir()
	payload := bytes.Repeat([]byte("streamed peer data "), 64000)
	if err := os.Mkdir(filepath.Join(from, "export"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(from, "export/copy.bin"), payload, 0600); err != nil {
		t.Fatal(err)
	}
	sourceDB, err := agentpolicy.Open(filepath.Join(local, "source.json"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := sourceDB.Add("Shared", from, true, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	token := base64.RawURLEncoding.EncodeToString(key)
	handler, err := agentpeer.NewHandler(sourceDB, token)
	if err != nil {
		t.Fatal(err)
	}
	handler.Active = func() bool { return true }
	server := httptest.NewServer(handler)
	defer server.Close()
	targetPath := filepath.Join(local, "target-grants.json")
	destinationDB, err := agentpolicy.Open(targetPath)
	if err != nil {
		t.Fatal(err)
	}
	destination, err := destinationDB.Add("Receive here", to, false, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	identity := filepath.Join(local, "identity.json")
	if err := agentclient.Save(identity, agentclient.Credentials{ID: "local-abc", Name: "My Device", Server: "https://controller.example", Secret: strings.Repeat("x", 43)}); err != nil {
		t.Fatal(err)
	}
	args := []string{"peer", "receive", "-identity", identity, "-grants", targetPath, "-staging", filepath.Join(local, "incoming"), "-peer", server.URL, "-grant", source.ID, "-relative", "export/copy.bin", "-source-id", "device-abc123456", "-destination-grant", destination.ID}
	var output bytes.Buffer
	if err := run(args, strings.NewReader(token+"\n"), &output); err != nil {
		t.Fatal(err)
	}
	copied, err := os.ReadFile(filepath.Join(to, "export/copy.bin"))
	if err != nil || !bytes.Equal(copied, payload) {
		t.Fatalf("cli peer receive mismatch: %v", err)
	}
	if !strings.Contains(output.String(), "Received export/copy.bin") {
		t.Fatal(output.String())
	}
	if err := run(args, strings.NewReader(token+"\n"), &output); err == nil {
		t.Fatal("receive overwrote destination on retry")
	}
	copyAgain, err := os.ReadFile(filepath.Join(to, "export/copy.bin"))
	if err != nil || !bytes.Equal(copyAgain, payload) {
		t.Fatalf("destination changed on second request: %v", err)
	}
}
