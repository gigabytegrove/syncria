package agentpeer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenVerifiedStreaming(t *testing.T) {
	body := bytes.Repeat([]byte("hello agent"), 20000)
	sum := sha256.Sum256(body)
	digest := hex.EncodeToString(sum[:])
	token := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{'x'}, 32))
	handle := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+token {
			http.Error(w, "no", 401)
			return
		}
		w.Header().Set("X-Syncria-SHA256", digest)
		w.Header().Set("ETag", `"`+digest+`"`)
		w.Header().Set("Content-Length", strings.TrimSpace(stringInt(len(body))))
		if r.Method == "GET" {
			if r.Header.Get("If-Match") != `"`+digest+`"` {
				http.Error(w, "changed", 412)
				return
			}
			_, _ = w.Write(body)
		}
	}
	server := httptest.NewServer(http.HandlerFunc(handle))
	defer server.Close()
	c := Client{Server: server.URL, Token: token}
	rd, n, err := c.OpenVerified(context.Background(), strings.Repeat("a", 32), "sub/file.txt")
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(rd)
	_ = rd.Close()
	if err != nil {
		t.Fatal(err)
	}
	if n != int64(len(body)) || !bytes.Equal(got, body) {
		t.Fatal("data mismatch")
	}
}
func TestOpenVerifiedDetectsCorruptionAndBadPath(t *testing.T) {
	body := []byte("original")
	sum := sha256.Sum256(body)
	digest := hex.EncodeToString(sum[:])
	token := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{'y'}, 32))
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", `"`+digest+`"`)
		w.Header().Set("X-Syncria-SHA256", digest)
		w.Header().Set("Content-Length", "8")
		if r.Method == "GET" {
			_, _ = w.Write([]byte("tampered"))
		}
	}))
	defer s.Close()
	c := Client{Server: s.URL, Token: token}
	if _, _, err := c.OpenVerified(context.Background(), strings.Repeat("a", 32), "../secret"); err == nil {
		t.Fatal("traversal accepted")
	}
	stream, _, err := c.OpenVerified(context.Background(), strings.Repeat("a", 32), "clean.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.Copy(io.Discard, stream)
	_ = stream.Close()
	if err == nil {
		t.Fatal("corruption accepted")
	}
}
func stringInt(i int) string {
	if i == 0 {
		return "0"
	}
	var b []byte
	for i > 0 {
		b = append([]byte{byte('0' + i%10)}, b...)
		i /= 10
	}
	return string(b)
}
