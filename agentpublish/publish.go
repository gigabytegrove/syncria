// Package agentpublish delivers verified staged Syncria files to a locally
// approved destination. The operation is initiated on the receiving computer;
// it is not a network write API and cannot silently broaden folder grants.
package agentpublish

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gigabytegrove/syncria/agentpolicy"
	"github.com/gigabytegrove/syncria/agenttransfer"
)

var (
	ErrDenied      = errors.New("destination is not locally approved for writes")
	ErrExists      = errors.New("destination file exists; Syncria will not overwrite it")
	ErrUnsupported = errors.New("safe atomic publication is not supported on this platform")
	ErrChanged     = errors.New("staged file changed during verification")
)

// Result is reported only once a verified staged file is atomically visible.
type Result struct {
	GrantID string `json:"grant_id"`
	Path    string `json:"path"`
	SHA256  string `json:"sha256"`
	Bytes   int64  `json:"bytes"`
}

// Publisher never changes an existing destination, removes a file, or deletes
// a stage. No new permissions are added by the controller or the peer.
type Publisher struct {
	Grants    *agentpolicy.Store
	Transfers *agenttransfer.Store
}

func (p Publisher) Publish(ctx context.Context, grantID, transferID string) (Result, error) {
	if p.Grants == nil || p.Transfers == nil {
		return Result{}, ErrDenied
	}
	input, item, err := p.Transfers.OpenCompleted(transferID)
	if err != nil {
		return Result{}, err
	}
	defer input.Close()
	relative := item.RelativePath
	if err = p.Grants.Allows(grantID, relative, true); err != nil {
		return Result{}, ErrDenied
	}
	grant, err := p.Grants.Get(grantID)
	if err != nil {
		return Result{}, ErrDenied
	}
	if grant.ReadOnly {
		return Result{}, ErrDenied
	}
	// Reject a disappearing or replaced approval before publication as well.
	recheck := func() error {
		if err := p.Grants.Allows(grantID, relative, true); err != nil {
			return ErrDenied
		}
		current, err := p.Grants.Get(grantID)
		if err != nil || current.Root != grant.Root || current.ReadOnly {
			return ErrDenied
		}
		return nil
	}
	want, err := hex.DecodeString(item.SHA256)
	if err != nil || len(want) != sha256.Size {
		return Result{}, ErrChanged
	}
	copyAndVerify := func(writer io.Writer) error {
		digest := sha256.New()
		n, e := io.Copy(io.MultiWriter(writer, digest), &readerWithContext{ctx: ctx, r: input})
		if e != nil {
			return e
		}
		if n != item.TotalBytes || subtle.ConstantTimeCompare(digest.Sum(nil), want) != 1 {
			return ErrChanged
		}
		return recheck()
	}
	if err = publishNoReplace(ctx, grant.Root, relative, copyAndVerify, recheck); err != nil {
		return Result{}, err
	}
	return Result{GrantID: grantID, Path: relative, SHA256: item.SHA256, Bytes: item.TotalBytes}, nil
}

type readerWithContext struct {
	ctx context.Context
	r   io.Reader
}

func (r *readerWithContext) Read(out []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(out)
}

// ValidateRoot should be applied to a local filesystem root before opening it.
// A symlinked final root, absent folder, or non-directory is never accepted.
func validateRoot(root string) error {
	if !filepath.IsAbs(root) {
		return ErrDenied
	}
	real, err := filepath.EvalSymlinks(root)
	if err != nil || real != filepath.Clean(root) {
		return fmt.Errorf("approved destination root unavailable or symlinked: %w", ErrDenied)
	}
	st, err := os.Lstat(root)
	if err != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return ErrDenied
	}
	return nil
}

func pathSegments(relative string) ([]string, error) {
	if relative == "" || len(relative) > 2048 || strings.HasPrefix(relative, "/") || strings.ContainsAny(relative, "\\\x00\r\n") || filepath.ToSlash(filepath.Clean(relative)) != relative {
		return nil, ErrDenied
	}
	parts := strings.Split(relative, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." || len(part) > 255 || strings.HasPrefix(part, ".syncria-") {
			return nil, ErrDenied
		}
		for _, r := range part {
			if r < 32 || r == 127 {
				return nil, ErrDenied
			}
		}
	}
	return parts, nil
}
