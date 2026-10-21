// Package agentupdate verifies signed, reproducible native agent artifacts.
// Applying a staged update and service rollback are intentionally separate
// platform operations: no unverified network data is ever executed here.
package agentupdate

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var ErrUntrusted = errors.New("signed agent release could not be verified")
var semver = regexp.MustCompile(`^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z.-]+))?$`)

type Artifact struct {
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Size     int64  `json:"size"`
	SHA256   string `json:"sha256"`
	Filename string `json:"filename"`
}
type Manifest struct {
	Schema    int        `json:"schema"`
	Version   string     `json:"version"`
	NotBefore time.Time  `json:"not_before"`
	Expires   time.Time  `json:"expires"`
	Artifacts []Artifact `json:"artifacts"`
}
type signedManifest struct {
	Payload   Manifest `json:"payload"`
	Signature string   `json:"signature"`
}

// Release can be constructed only after successful signature validation.
type Release struct{ payload Manifest }

func (r *Release) Version() string { return r.payload.Version }
func (r *Release) Select(goos, goarch string) (Artifact, error) {
	if r == nil {
		return Artifact{}, ErrUntrusted
	}
	for _, a := range r.payload.Artifacts {
		if a.OS == goos && a.Arch == goarch {
			return a, nil
		}
	}
	return Artifact{}, errors.New("platform not available in verified release")
}
func validate(m Manifest) error {
	if m.Schema != 1 || !semver.MatchString(m.Version) || m.NotBefore.IsZero() || m.Expires.IsZero() || !m.NotBefore.Before(m.Expires) {
		return ErrUntrusted
	}
	if len(m.Artifacts) == 0 || len(m.Artifacts) > 32 {
		return ErrUntrusted
	}
	found := map[string]bool{}
	for _, a := range m.Artifacts {
		if !validPlatform(a.OS, a.Arch) || a.Size < 1 || a.Size > 2<<30 || len(a.SHA256) != 64 || len(a.Filename) < 4 || len(a.Filename) > 128 || filepath.Base(a.Filename) != a.Filename || strings.ContainsAny(a.Filename, "\\/\x00\r\n") {
			return ErrUntrusted
		}
		if _, e := hex.DecodeString(a.SHA256); e != nil {
			return ErrUntrusted
		}
		k := a.OS + "/" + a.Arch
		if found[k] {
			return ErrUntrusted
		}
		found[k] = true
	}
	return nil
}
func validPlatform(osName, arch string) bool {
	supported := map[string]map[string]bool{"windows": {"amd64": true, "arm64": true}, "linux": {"amd64": true, "arm64": true, "arm": true}, "darwin": {"amd64": true, "arm64": true}}
	return supported[osName][arch]
}
func DecodeAndVerify(document []byte, publicKey ed25519.PublicKey, now time.Time) (*Release, error) {
	if len(publicKey) != ed25519.PublicKeySize || len(document) == 0 || len(document) > 64<<10 {
		return nil, ErrUntrusted
	}
	dec := json.NewDecoder(strings.NewReader(string(document)))
	dec.DisallowUnknownFields()
	var s signedManifest
	if err := dec.Decode(&s); err != nil {
		return nil, ErrUntrusted
	}
	var extra any
	if !errors.Is(dec.Decode(&extra), io.EOF) {
		return nil, ErrUntrusted
	}
	if err := validate(s.Payload); err != nil {
		return nil, err
	}
	sig, err := base64.RawStdEncoding.DecodeString(s.Signature)
	if err != nil || len(sig) != ed25519.SignatureSize {
		return nil, ErrUntrusted
	}
	payload, err := json.Marshal(s.Payload)
	if err != nil {
		return nil, ErrUntrusted
	}
	if !ed25519.Verify(publicKey, payload, sig) {
		return nil, ErrUntrusted
	}
	if now.Before(s.Payload.NotBefore) || !now.Before(s.Payload.Expires) {
		return nil, errors.New("agent update signature valid but manifest outside permitted time window")
	}
	return &Release{payload: s.Payload}, nil
}

// IsNewer enforces monotonically increasing SemVer 2.0 precedence for common
// major/minor/patch + prerelease versions. Build metadata is unsupported.
func IsNewer(current, candidate string) (bool, error) {
	a, b := semver.FindStringSubmatch(current), semver.FindStringSubmatch(candidate)
	if a == nil || b == nil {
		return false, errors.New("invalid semantic version")
	}
	for n := 1; n <= 3; n++ {
		av, _ := strconv.ParseUint(a[n], 10, 64)
		bv, _ := strconv.ParseUint(b[n], 10, 64)
		if bv > av {
			return true, nil
		}
		if bv < av {
			return false, nil
		}
	}
	if a[4] == b[4] {
		return false, nil
	}
	if a[4] == "" {
		return false, nil
	}
	if b[4] == "" {
		return true, nil
	}
	x, y := strings.Split(a[4], "."), strings.Split(b[4], ".")
	for i := 0; i < len(x) && i < len(y); i++ {
		if x[i] == y[i] {
			continue
		}
		u, eu := strconv.ParseUint(x[i], 10, 64)
		v, ev := strconv.ParseUint(y[i], 10, 64)
		if eu == nil && ev == nil {
			return v > u, nil
		}
		if eu == nil {
			return true, nil
		}
		if ev == nil {
			return false, nil
		}
		return y[i] > x[i], nil
	}
	return len(y) > len(x), nil
}

// Stage streams a known signed-release artifact into an owner-only staging
// directory, validates exact length and SHA256, syncs and atomically renames.
// It refuses existing target files. Executing and swapping binaries requires
// a separate platform-specific supervised update/rollback transaction.
func (r *Release) Stage(ctx context.Context, osName, arch, stagingDir string, src io.Reader) (string, error) {
	a, err := r.Select(osName, arch)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(stagingDir) {
		return "", errors.New("staging directory must be absolute")
	}
	if err := os.MkdirAll(stagingDir, 0700); err != nil {
		return "", err
	}
	info, err := os.Lstat(stagingDir)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm()&0077 != 0 {
		return "", errors.New("staging directory permissions unsafe")
	}
	f, err := os.CreateTemp(stagingDir, ".syncria-verify-*.tmp")
	if err != nil {
		return "", err
	}
	defer os.Remove(f.Name())
	if err := f.Chmod(0700); err != nil {
		f.Close()
		return "", err
	}
	hash := sha256.New()
	// io.LimitReader catches oversized and short downloads without allocating.
	buf := make([]byte, 128<<10)
	var total int64
	for {
		if err := ctx.Err(); err != nil {
			f.Close()
			return "", err
		}
		n, e := io.LimitReader(src, a.Size+1-total).Read(buf)
		if n > 0 {
			total += int64(n)
			if total > a.Size {
				f.Close()
				return "", errors.New("agent update exceeds manifest size")
			}
			if _, err := f.Write(buf[:n]); err != nil {
				f.Close()
				return "", err
			}
			_, _ = hash.Write(buf[:n])
		}
		if e == io.EOF {
			break
		}
		if e != nil {
			f.Close()
			return "", e
		}
		if n == 0 {
			f.Close()
			return "", io.ErrNoProgress
		}
	}
	if total != a.Size || !strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), a.SHA256) {
		f.Close()
		return "", errors.New("agent update checksum or length mismatch")
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return "", err
	}
	if err = f.Close(); err != nil {
		return "", err
	}
	target := filepath.Join(stagingDir, a.Filename)
	// Hard-link only if nonexistent: rename could silently replace a known-good file.
	if err = os.Link(f.Name(), target); err != nil {
		return "", fmt.Errorf("verified update cannot replace staged target: %w", err)
	}
	if dir, e := os.Open(stagingDir); e == nil {
		_ = dir.Sync()
		_ = dir.Close()
	}
	return target, nil
}
