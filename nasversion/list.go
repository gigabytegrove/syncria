package nasversion

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

// Lister extends Remote with metadata-only listing; no archived payload is
// returned or staged on the controller.
type Lister interface {
	ReadDir(string) ([]os.FileInfo, error)
}

// Version exposes the read-only location of an old file within its SMB share.
type Version struct {
	Path      string    `json:"path"`
	Relative  string    `json:"relative"`
	VersionID string    `json:"version_id"`
	Bytes     int64     `json:"bytes"`
	Modified  time.Time `json:"modified"`
}

// List is a strict, bounded, all-or-error listing of retained NAS versions.
// A missing archive directory is an empty inventory; other errors fail closed.
func List(ctx context.Context, remote Lister, root, jobID string) ([]Version, error) {
	if remote == nil || !safeRoot(root) || !validID(jobID) {
		return nil, ErrUnsafePath
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	parent := path.Join(root, ".syncria-versions", jobID)
	versions, e := remote.ReadDir(parent)
	if os.IsNotExist(e) {
		return []Version{}, nil
	}
	if e != nil {
		return nil, fmt.Errorf("list NAS versions: %w", e)
	}
	if len(versions) > 10000 {
		return nil, errors.New("too many retained NAS versions")
	}
	found := make([]Version, 0)
	for _, v := range versions {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if !v.IsDir() || !validID(v.Name()) {
			return nil, errors.New("invalid NAS version directory")
		}
		var walk func(dir, relative string, depth int) error
		walk = func(dir, relative string, depth int) error {
			if depth > 64 {
				return errors.New("NAS version directory depth exceeded")
			}
			entries, err := remote.ReadDir(dir)
			if err != nil {
				return err
			}
			for _, info := range entries {
				if err := ctx.Err(); err != nil {
					return err
				}
				name := info.Name()
				if name == "" || name == "." || name == ".." || strings.ContainsAny(name, "/\\\x00\r\n") || strings.HasPrefix(name, ".syncria-") || len(name) > 255 || info.Mode()&os.ModeSymlink != 0 {
					return ErrUnsafePath
				}
				child := path.Join(relative, name)
				full := path.Join(dir, name)
				if !safeRel(child) {
					return ErrUnsafePath
				}
				if info.IsDir() {
					if err := walk(full, child, depth+1); err != nil {
						return err
					}
					continue
				}
				if !info.Mode().IsRegular() {
					return ErrUnsafePath
				}
				found = append(found, Version{Path: full, Relative: child, VersionID: v.Name(), Bytes: info.Size(), Modified: info.ModTime()})
				if len(found) > 10000 {
					return errors.New("NAS version file count exceeds limit")
				}
			}
			return nil
		}
		if err := walk(path.Join(parent, v.Name()), "", 0); err != nil {
			return nil, fmt.Errorf("inspect NAS version %s: %w", v.Name(), err)
		}
	}
	sort.Slice(found, func(i, j int) bool {
		if found[i].Modified.Equal(found[j].Modified) {
			return found[i].Path < found[j].Path
		}
		return found[i].Modified.After(found[j].Modified)
	})
	return found, nil
}
