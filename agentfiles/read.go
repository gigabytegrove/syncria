// Package agentfiles exposes read-only, locally approved folders to the
// Syncria peer transport. File opens are handle-relative and reject symlinks.
// No path on the device is remotely accessible without a local grant.
package agentfiles

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gigabytegrove/syncria/agentpolicy"
)

var ErrDenied = errors.New("file is outside an approved folder")
var ErrScanLimit = errors.New("approved folder inventory exceeds safety limit")

const MaxEntries = 100000

type Entry struct {
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	Modified  time.Time `json:"modified"`
	Directory bool      `json:"directory"`
}

type Reader struct{ Grants *agentpolicy.Store }

// Open checks current locally approved selections BEFORE opening the file.
// It returns a stable file handle; a concurrent filesystem rename cannot
// redirect the open to a symlink escape (see openUnderRoot).
func (r Reader) Open(grantID, relative string) (*os.File, os.FileInfo, error) {
	if r.Grants == nil {
		return nil, nil, ErrDenied
	}
	if err := r.Grants.Allows(grantID, relative, false); err != nil {
		return nil, nil, ErrDenied
	}
	grant, err := r.Grants.Get(grantID)
	if err != nil {
		return nil, nil, ErrDenied
	}
	f, err := openUnderRoot(grant.Root, relative)
	if err != nil {
		return nil, nil, fmt.Errorf("approved file unavailable: %w", err)
	}
	st, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, nil, err
	}
	if !st.Mode().IsRegular() {
		f.Close()
		return nil, nil, ErrDenied
	}
	return f, st, nil
}

// Inventory fails rather than returning partial results: partial scans are
// never evidence of file deletion. It exposes metadata, not raw local roots.
func (r Reader) Inventory(ctx context.Context, grantID string) ([]Entry, error) {
	if r.Grants == nil {
		return nil, ErrDenied
	}
	g, err := r.Grants.Get(grantID)
	if err != nil {
		return nil, ErrDenied
	}
	rootInfo, err := os.Lstat(g.Root)
	if err != nil || !rootInfo.IsDir() || rootInfo.Mode()&os.ModeSymlink != 0 {
		return nil, ErrDenied
	}
	var out []Entry
	count := 0
	err = filepath.WalkDir(g.Root, func(name string, de fs.DirEntry, walkErr error) error {
		if e := ctx.Err(); e != nil {
			return e
		}
		if walkErr != nil {
			return walkErr
		}
		if name == g.Root {
			return nil
		}
		count++
		if count > MaxEntries {
			return ErrScanLimit
		}
		rel, e := filepath.Rel(g.Root, name)
		if e != nil {
			return e
		}
		rel = filepath.ToSlash(rel)
		// Never silently follow or omit dangerous filesystem entries. The sync
		// engine must know the inventory was invalid rather than infer deletion.
		if de.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symbolic link in selected folder %s: %w", rel, ErrDenied)
		}
		if strings.HasPrefix(de.Name(), ".syncria-") {
			if de.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if de.IsDir() {
			return nil
		}
		if !de.Type().IsRegular() {
			return fmt.Errorf("unsupported entry %s: %w", rel, ErrDenied)
		}
		if r.Grants.Allows(grantID, rel, false) != nil {
			return nil
		}
		f, st, e := r.Open(grantID, rel)
		if e != nil {
			return e
		}
		_ = f.Close()
		out = append(out, Entry{Path: rel, Size: st.Size(), Modified: st.ModTime().UTC()})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

// Copy reads a handle-relative approved file into a bounded writer and checks
// its size and modification metadata after the copy. The caller must still
// verify the content digest before accepting a staged transfer.
func (r Reader) Copy(ctx context.Context, grantID, relative string, target io.Writer) (int64, error) {
	f, initial, err := r.Open(grantID, relative)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	n, err := io.Copy(target, &contextReader{ctx: ctx, r: f})
	if err != nil {
		return n, err
	}
	after, err := f.Stat()
	if err != nil {
		return n, err
	}
	if n != initial.Size() || initial.Size() != after.Size() || !initial.ModTime().Equal(after.ModTime()) {
		return n, errors.New("source file changed while it was being read")
	}
	return n, nil
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (r *contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}
