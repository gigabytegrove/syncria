package agentpublish

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/gigabytegrove/syncria/agentpolicy"
)

func TestPublishVerifiedRejectsWrongDigestAndLength(t *testing.T) {
	root := filepath.Join(t.TempDir(), "approved")
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	grants, err := agentpolicy.Open(filepath.Join(t.TempDir(), "grants.json"))
	if err != nil {
		t.Fatal(err)
	}
	g, err := grants.Add("Approved", root, false, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("hello verification")
	sha := sha256.Sum256(data)
	pub := Publisher{Grants: grants}
	for _, tc := range []struct {
		name    string
		content []byte
		size    int64
		digest  string
	}{
		{"wrong digest", data, int64(len(data)), hex.EncodeToString(bytes.Repeat([]byte{0xFF}, 32))},
		{"truncated input", data, int64(len(data) + 1), hex.EncodeToString(sha[:])},
		{"extra bytes", append(append([]byte{}, data...), '!'), int64(len(data)), hex.EncodeToString(sha[:])},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := pub.PublishVerified(context.Background(), g.ID, "recovered.txt", bytes.NewReader(tc.content), tc.size, tc.digest)
			if !errors.Is(err, ErrChanged) {
				t.Fatalf("invalid content accepted: %v", err)
			}
			if _, err := os.Stat(filepath.Join(root, "recovered.txt")); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("invalid file was published")
			}
		})
	}
}
