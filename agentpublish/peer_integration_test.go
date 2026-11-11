package agentpublish

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gigabytegrove/syncria/agentpeer"
	"github.com/gigabytegrove/syncria/agentpolicy"
	"github.com/gigabytegrove/syncria/agenttransfer"
)

// This exercises the real peer HTTP handler, peer download client, resumable
// transfer engine, owner-approved destination and atomic final publication.
func TestPeerToPeerVerifiedPublication(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("native Linux destination publisher")
	}
	sourceRoot := t.TempDir()
	data := bytes.Repeat([]byte("This is a large multi-chunk Syncria peer test. "), 125000)
	if err := os.MkdirAll(filepath.Join(sourceRoot, "photos"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, "photos/summer.bin"), data, 0600); err != nil {
		t.Fatal(err)
	}
	srcGrants, err := agentpolicy.Open(filepath.Join(t.TempDir(), "source-grants.json"))
	if err != nil {
		t.Fatal(err)
	}
	srcGrant, err := srcGrants.Add("Photos", sourceRoot, true, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	key := make([]byte, 32)
	if _, err = rand.Read(key); err != nil {
		t.Fatal(err)
	}
	secret := base64.RawURLEncoding.EncodeToString(key)
	handler, err := agentpeer.NewHandler(srcGrants, secret)
	if err != nil {
		t.Fatal(err)
	}
	handler.Active = func() bool { return true }
	server := httptest.NewServer(handler)
	defer server.Close()
	client := agentpeer.Client{Server: server.URL, Token: secret}
	entries, err := client.Inventory(context.Background(), srcGrant.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Path != "photos/summer.bin" {
		t.Fatalf("unexpected inventory: %v", entries)
	}
	destinationRoot := t.TempDir()
	dstGrants, err := agentpolicy.Open(filepath.Join(t.TempDir(), "destination-grants.json"))
	if err != nil {
		t.Fatal(err)
	}
	destGrant, err := dstGrants.Add("Destination", destinationRoot, false, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	staging, err := agenttransfer.Open(filepath.Join(t.TempDir(), "staging"), 16<<20)
	if err != nil {
		t.Fatal(err)
	}
	staged, err := client.Stage(context.Background(), srcGrant.ID, entries[0].Path, "source-abc123456", staging)
	if err != nil {
		t.Fatal(err)
	}
	if staged.Status != agenttransfer.StatusComplete {
		t.Fatal(staged)
	}
	receipt, err := (Publisher{Grants: dstGrants, Transfers: staging}).Publish(context.Background(), destGrant.ID, staged.ID)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Bytes != int64(len(data)) || receipt.SHA256 != digest(data) {
		t.Fatal(receipt)
	}
	final, err := os.ReadFile(filepath.Join(destinationRoot, entries[0].Path))
	if err != nil || !bytes.Equal(final, data) {
		t.Fatalf("end-to-end content mismatch: %v", err)
	}
	// Removed destination grants prevent replay even with a complete transfer.
	if err = dstGrants.Remove(destGrant.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = (Publisher{Grants: dstGrants, Transfers: staging}).Publish(context.Background(), destGrant.ID, staged.ID); err == nil {
		t.Fatal("revoked destination accepted")
	}
}
