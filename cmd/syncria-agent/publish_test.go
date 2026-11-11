package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gigabytegrove/syncria/agentclient"
	"github.com/gigabytegrove/syncria/agentpolicy"
	"github.com/gigabytegrove/syncria/agenttransfer"
)

func TestAgentPeerPublishCLI(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-only atomic publication")
	}
	tmp := t.TempDir()
	id := filepath.Join(tmp, "identity.json")
	if err := agentclient.Save(id, agentclient.Credentials{ID: "device-local", Name: "Receiver", Server: "https://example.com", Secret: strings.Repeat("a", 43)}); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(tmp, "approved")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	grantDB := filepath.Join(tmp, "grants.json")
	grants, err := agentpolicy.Open(grantDB)
	if err != nil {
		t.Fatal(err)
	}
	grant, err := grants.Add("Receiver", root, false, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	stagePath := filepath.Join(tmp, "incoming")
	transfer, err := agenttransfer.Open(stagePath, agenttransfer.DefaultMaxPendingBytes)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("published by the Syncria peer command")
	sum := sha256.Sum256(data)
	digest := hex.EncodeToString(sum[:])
	session, err := transfer.Begin(agenttransfer.Request{SourceID: "remote-12345678", Revision: digest, RelativePath: "folder/arrived.txt", TotalBytes: int64(len(data)), SHA256: digest})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = transfer.Append(session.ID, 0, bytes.NewReader(data), len(data), digest); err != nil {
		t.Fatal(err)
	}
	if _, err = transfer.Complete(session.ID); err != nil {
		t.Fatal(err)
	}
	args := []string{"peer", "publish", "-identity", id, "-grants", grantDB, "-staging", stagePath, "-destination-grant", grant.ID, "-transfer", session.ID}
	var output bytes.Buffer
	if err := run(args, strings.NewReader(""), &output); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(root, "folder/arrived.txt"))
	if err != nil || !bytes.Equal(got, data) {
		t.Fatalf("published incorrect data: %v", err)
	}
	if !strings.Contains(output.String(), digest) {
		t.Fatalf("receipt missing checksum: %s", output.String())
	}
	if err = run(args, strings.NewReader(""), &output); err == nil {
		t.Fatal("CLI replaced existing file")
	}
	if err := run([]string{"peer", "publish", "-identity", id}, strings.NewReader(""), &output); err == nil {
		t.Fatal("accepted incomplete publish arguments")
	}
}
