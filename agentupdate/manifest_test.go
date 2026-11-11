package agentupdate

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func signed(t *testing.T) (ed25519.PublicKey, []byte, []byte) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	binary := bytes.Repeat([]byte("verified-agent-binary"), 400)
	sum := sha256.Sum256(binary)
	m := Manifest{Schema: 1, Version: "v1.0.0-alpha.2", NotBefore: time.Now().Add(-time.Hour).UTC(), Expires: time.Now().Add(time.Hour).UTC(), Artifacts: []Artifact{{OS: "linux", Arch: "amd64", Size: int64(len(binary)), SHA256: hex.EncodeToString(sum[:]), Filename: "syncria-agent-linux-amd64"}}}
	payload, _ := json.Marshal(m)
	sig := ed25519.Sign(priv, payload)
	doc, _ := json.Marshal(signedManifest{Payload: m, Signature: base64.RawStdEncoding.EncodeToString(sig)})
	return pub, doc, binary
}
func TestSignedManifestAndStaging(t *testing.T) {
	key, document, binary := signed(t)
	release, err := DecodeAndVerify(document, key, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if release.Version() != "v1.0.0-alpha.2" {
		t.Fatal("incorrect verified version")
	}
	dir := filepath.Join(t.TempDir(), "staging")
	target, err := release.Stage(context.Background(), "linux", "amd64", dir, bytes.NewReader(binary))
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(target)
	if err != nil || !bytes.Equal(got, binary) {
		t.Fatal("staged content mismatch", err)
	}
	if _, err = release.Stage(context.Background(), "linux", "amd64", dir, bytes.NewReader(binary)); err == nil {
		t.Fatal("update replaced existing staged file")
	}
	if _, err = release.Stage(context.Background(), "windows", "amd64", dir, bytes.NewReader(binary)); err == nil {
		t.Fatal("wrong platform accepted")
	}
}
func TestTamperingAndExpiry(t *testing.T) {
	key, document, _ := signed(t)
	for _, v := range [][]byte{bytes.Replace(document, []byte("alpha.2"), []byte("alpha.3"), 1), append(append([]byte(nil), document...), []byte("{}")...), []byte("{}")} {
		if _, err := DecodeAndVerify(v, key, time.Now()); err == nil {
			t.Fatal("tampered manifest accepted")
		}
	}
	another, _, _ := ed25519.GenerateKey(rand.Reader)
	if _, err := DecodeAndVerify(document, another, time.Now()); err == nil {
		t.Fatal("wrong signing key accepted")
	}
	if _, err := DecodeAndVerify(document, key, time.Now().Add(48*time.Hour)); err == nil {
		t.Fatal("expired release accepted")
	}
}
func TestTruncatedAndCorruptBinary(t *testing.T) {
	key, document, binary := signed(t)
	release, _ := DecodeAndVerify(document, key, time.Now())
	dir := t.TempDir()
	for _, v := range [][]byte{binary[:len(binary)-1], bytes.Repeat([]byte("x"), len(binary)+1), bytes.Repeat([]byte("x"), len(binary))} {
		if _, err := release.Stage(context.Background(), "linux", "amd64", dir, bytes.NewReader(v)); err == nil {
			t.Fatal("corrupt binary staged")
		}
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) > 0 {
		t.Fatalf("unverified files left behind: %v", files)
	}
}
func TestVersionPrecedence(t *testing.T) {
	for _, v := range []struct {
		a, b  string
		newer bool
	}{{"v1.0.0-alpha.1", "v1.0.0-alpha.2", true}, {"1.0.0-alpha.3", "1.0.0", true}, {"1.0.1", "1.0.0", false}, {"1.0.0", "1.0.0-beta", false}, {"1.0.0-rc.2", "1.0.0-rc.10", true}} {
		got, err := IsNewer(v.a, v.b)
		if err != nil || got != v.newer {
			t.Errorf("%s to %s: got %v,%v", v.a, v.b, got, err)
		}
	}
	if _, err := IsNewer("garbage", "v1.0.0"); err == nil {
		t.Fatal("invalid version accepted")
	}
}
func TestStageCancellation(t *testing.T) {
	key, document, _ := signed(t)
	release, _ := DecodeAndVerify(document, key, time.Now())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := release.Stage(ctx, "linux", "amd64", t.TempDir(), strings.NewReader("no")); err == nil {
		t.Fatal("canceled transfer allowed")
	}
}

var _ io.Reader = (*bytes.Reader)(nil)
