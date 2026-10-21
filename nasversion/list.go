package nasversion

import (
	"context"
	"errors"
	"fmt"
 "io"
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
	Captured time.Time `json:"captured"`
 CaptureSHA256 string `json:"capture_sha256,omitempty"`
 CaptureProofPresent bool `json:"capture_proof_present"`
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
    if depth==0 && name==integrityName{
      if !info.Mode().IsRegular()||info.Size()<1||info.Size()>integrityLimit{return errors.New("unsafe NAS archive proof file")}
      continue
    }
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
				found = append(found, Version{Path: full, Relative: child, VersionID: v.Name(), Bytes: info.Size(), Modified: info.ModTime(), Captured: v.ModTime()})
				if len(found) > 10000 {
					return errors.New("NAS version file count exceeds limit")
				}
			}
			return nil
		}
		var proof CaptureProof
  var hasProof bool
  if reader,ok:=remote.(interface{Open(string)(io.ReadCloser,error)});ok{
    record,exists,e:=readProof(reader,root,jobID,v.Name())
    if e!=nil{return nil,fmt.Errorf("NAS archive proof invalid: %w",e)}
    proof,hasProof=record,exists
  }
  start:=len(found)
  if err := walk(path.Join(parent, v.Name()), "", 0); err != nil {
			return nil, fmt.Errorf("inspect NAS version %s: %w", v.Name(), err)
		}
  if hasProof{
    if len(found)-start==0{continue}
    if len(found)-start!=1{return nil,errors.New("capture proof must identify one archived file")}
    entry:=&found[start]
    if entry.Relative!=proof.Relative||entry.Bytes!=proof.Bytes{return nil,errors.New("capture proof does not match archive")}
    entry.CaptureProofPresent=true
    entry.CaptureSHA256=proof.SHA256
  }
	}
	sort.Slice(found, func(i, j int) bool {
		if found[i].Captured.Equal(found[j].Captured) {
			return found[i].Path < found[j].Path
		}
		return found[i].Captured.After(found[j].Captured)
	})
	return found, nil
}
