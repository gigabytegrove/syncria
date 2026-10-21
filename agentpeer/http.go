// Package agentpeer provides an opt-in, read-only peer transport for locally
// approved Syncria folders. Peer credentials are separate from enrollment.
// No endpoint exposes writes, deletion, directory creation or remote grants.
package agentpeer

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gigabytegrove/syncria/agentfiles"
	"github.com/gigabytegrove/syncria/agentpolicy"
)

const MaxInventoryBytes = 32 << 20

// Handler requires a separately issued peer token and HTTPS for non-loopback
// connections. TLS must terminate directly on the agent, never on an
// arbitrary reverse proxy trusting an unverified forwarded header.
type Handler struct {
	Reader  agentfiles.Reader
	KeyHash [32]byte
	// Active must confirm a recent accepted controller heartbeat. Without it
	// the peer endpoint is disabled (fail closed).
	Active func() bool
}

func NewHandler(grants *agentpolicy.Store, token string) (Handler, error) {
	if len(token) != 43 {
		return Handler{}, errors.New("invalid peer token")
	}
	bytes, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(bytes) != 32 || base64.RawURLEncoding.EncodeToString(bytes) != token {
		return Handler{}, errors.New("invalid peer token")
	}
	return Handler{Reader: agentfiles.Reader{Grants: grants}, KeyHash: sha256.Sum256([]byte(token))}, nil
}
func isSecure(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return false
	}
	peer := net.ParseIP(host)
	if peer == nil || !peer.IsLoopback() {
		return false
	}
	destination := r.Host
	if h, _, e := net.SplitHostPort(destination); e == nil {
		destination = h
	}
	if strings.EqualFold(destination, "localhost") {
		return true
	}
	addr := net.ParseIP(destination)
	return addr != nil && addr.IsLoopback()
}
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if h.Active == nil || !h.Active() {
		http.Error(w, "agent authorization unavailable", http.StatusServiceUnavailable)
		return
	}
	if !isSecure(r) {
		http.Error(w, "TLS is required", http.StatusForbidden)
		return
	}
	raw := r.Header.Get("Authorization")
	if !strings.HasPrefix(raw, "Bearer ") || len(raw) != 50 {
		http.Error(w, "unauthorized", 401)
		return
	}
	sum := sha256.Sum256([]byte(strings.TrimPrefix(raw, "Bearer ")))
	if subtle.ConstantTimeCompare(h.KeyHash[:], sum[:]) != 1 {
		http.Error(w, "unauthorized", 401)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", 405)
		return
	}
	// Go's routing normalizes some URL paths before handler invocation. Reject
	// double encoding and control characters before resolving an approved path.
	if strings.ContainsAny(r.URL.Path, "\\\x00\r\n") {
		http.Error(w, "invalid path", 400)
		return
	}
	const prefix = "/v1/roots/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		http.NotFound(w, r)
		return
	}
	rest := strings.TrimPrefix(r.URL.Path, prefix)
	chunks := strings.SplitN(rest, "/", 3)
	if len(chunks) < 2 || len(chunks[0]) != 32 {
		http.NotFound(w, r)
		return
	}
	grant := chunks[0]
	if chunks[1] == "inventory" && len(chunks) == 2 {
		items, err := h.Reader.Inventory(r.Context(), grant)
		if err != nil {
			http.Error(w, "inventory unavailable", 409)
			return
		}
		body, err := json.Marshal(map[string]any{"entries": items})
		if err != nil || len(body) > MaxInventoryBytes {
			http.Error(w, "inventory too large", 413)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(200)
		if r.Method == http.MethodGet {
			_, _ = w.Write(body)
		}
		return
	}
	if chunks[1] != "files" || len(chunks) != 3 || chunks[2] == "" || path.Clean(chunks[2]) != chunks[2] {
		http.NotFound(w, r)
		return
	}
	relative := chunks[2]
	f, stat, err := h.Reader.Open(grant, relative)
	if err != nil {
		http.Error(w, "file unavailable", 404)
		return
	}
	defer f.Close()
	// Strong checksum is established on the opened file handle before sending
	// bytes. A concurrent source edit is detected at the receiver by digest.
	digest := sha256.New()
	if _, err = io.Copy(digest, f); err != nil {
		http.Error(w, "file read failed", 409)
		return
	}
	after, err := f.Stat()
	if err != nil || stat.Size() != after.Size() || !stat.ModTime().Equal(after.ModTime()) {
		http.Error(w, "source changed while hashing", 409)
		return
	}
	sumHex := hex.EncodeToString(digest.Sum(nil))
	etag := `"` + sumHex + `"`
	if value := r.Header.Get("If-Match"); value != "" && value != etag {
		http.Error(w, "source revision changed", 412)
		return
	}
	w.Header().Set("ETag", etag)
	w.Header().Set("X-Syncria-SHA256", sumHex)
	w.Header().Set("Content-Type", "application/octet-stream")
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return
	}
	http.ServeContent(w, r, relative, stat.ModTime(), f)
}

// GenerateKey is called locally by the device owner. The raw key is printed
// only at creation and kept separate from the controller enrollment secret.
func GenerateKey(file string) (string, error) {
	if !filepath.IsAbs(file) {
		return "", errors.New("peer key path must be absolute")
	}
	if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
		return "", err
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(b)
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return "", err
	}
	if _, err = f.Write([]byte(token + "\n")); err != nil {
		_ = f.Close()
		_ = os.Remove(file)
		return "", err
	}
	if err = f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(file)
		return "", err
	}
	if err = f.Close(); err != nil {
		return "", err
	}
	return token, nil
}
func ReadKey(file string) (string, error) {
	st, err := os.Lstat(file)
	if err != nil {
		return "", err
	}
	if !st.Mode().IsRegular() || st.Size() > 128 || runtime.GOOS != "windows" && st.Mode().Perm()&0077 != 0 {
		return "", errors.New("peer token file must be private and regular")
	}
	b, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	token := strings.TrimSpace(string(b))
	_, err = NewHandler(nil, token)
	if err != nil {
		return "", fmt.Errorf("invalid peer token file: %w", err)
	}
	return token, nil
}
