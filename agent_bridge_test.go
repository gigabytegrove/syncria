package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gigabytegrove/syncria/agentpeer"
)

func TestAgentBridgeScanAndReadVerified(t *testing.T) {
	data := bytes.Repeat([]byte("alpha transfer "), 8000)
	digest := sha256.Sum256(data)
	sum := hex.EncodeToString(digest[:])
	token := "YXh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHg" // Valid 32-byte URL-safe token.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/inventory") {
			_ = json.NewEncoder(w).Encode(map[string]any{"entries": []map[string]any{{"path": "notes/daily.txt", "size": len(data), "modified": "2026-10-20T00:00:00Z"}}})
			return
		}
		w.Header().Set("ETag", `"`+sum+`"`)
		w.Header().Set("X-Syncria-SHA256", sum)
		w.Header().Set("Content-Length", fmt.Sprint(len(data)))
		if r.Method == http.MethodGet {
			_, _ = w.Write(data)
		}
	}))
	defer server.Close()
	e := &agentReadEndpoint{peer: agentpeer.Client{Server: server.URL, Token: token}, grant: strings.Repeat("a", 32)}
	entries, err := e.Scan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("inventory %d", len(entries))
	}
	if len(entries["notes/daily.txt"].Fingerprint) == 32 || len(entries["notes/daily.txt"].Fingerprint) == 64 {
		t.Fatal("metadata could be mistaken for a content digest")
	}
	rd, size, err := e.Open(context.Background(), "notes/daily.txt")
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(rd)
	_ = rd.Close()
	if err != nil || size != int64(len(data)) || !bytes.Equal(got, data) {
		t.Fatalf("file stream %v", err)
	}
	if _, _, err = e.Open(context.Background(), "not-listed.txt"); err == nil {
		t.Fatal("unlisted file accepted")
	}
	if _, err = e.Mkdir(context.Background(), "folder"); err == nil {
		t.Fatal("source mutation permitted")
	}
}
func TestAgentBridgeRejectsPartialAndUnsafeInventory(t *testing.T) {
	for _, paths := range [][]string{{"../secret"}, {"same.txt", "SAME.TXT"}, {"file", "file/child"}, {"x/.syncria-hidden"}} {
		t.Run(strings.Join(paths, "_"), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				records := []map[string]any{}
				for _, path := range paths {
					records = append(records, map[string]any{"path": path, "size": 1, "modified": time.Now().UTC()})
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"entries": records})
			}))
			defer server.Close()
			e := &agentReadEndpoint{peer: agentpeer.Client{Server: server.URL, Token: "YXh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHh4eHg"}, grant: strings.Repeat("a", 32)}
			if _, err := e.Scan(context.Background()); err == nil {
				t.Fatal("invalid inventory accepted")
			}
		})
	}
}
