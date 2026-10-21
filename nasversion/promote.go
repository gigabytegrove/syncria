// Package nasversion provides destination-owned, non-destructive NAS file
// replacement. Previous bytes remain on the NAS, never the controller.
package nasversion

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
)

// Remote is the small shared contract implemented by an SMB share.
// Implementations are expected to perform atomic same-share renames.
type Remote interface {
	Stat(string) (os.FileInfo, error)
	Mkdir(string, os.FileMode) error
	Rename(string, string) error
}

var ErrUnsafePath = errors.New("unsafe NAS version path")

// Result describes the previous file's retained location, if any.
// VersionPath is relative to the SMB share, not to Syncria's /data.
type Result struct {
	VersionPath string
	Replaced    bool
}

func safeRel(s string) bool {
	if s == "" || len(s) > 2048 || strings.HasPrefix(s, "/") || strings.ContainsAny(s, "\\\x00\r\n") || path.Clean(s) != s {
		return false
	}
	for _, c := range strings.Split(s, "/") {
		if c == "" || c == "." || c == ".." || len(c) > 255 || strings.HasPrefix(c, ".syncria-") {
			return false
		}
		for _, r := range c {
			if r < 32 || r == 127 {
				return false
			}
		}
	}
	return true
}
func safeRoot(s string) bool {
	if s == "." {
		return true
	}
	return safeRel(s)
}
func validID(s string) bool {
	if len(s) != 32 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			if r < 'a' || r > 'f' {
				return false
			}
		}
	}
	return true
}
func mkdirAll(remote Remote, dir string) error {
	current := ""
	for _, name := range strings.Split(dir, "/") {
		if name == "" || name == "." || name == ".." {
			return ErrUnsafePath
		}
		current = path.Join(current, name)
		if err := remote.Mkdir(current, 0700); err != nil {
			info, statErr := remote.Stat(current)
			if statErr != nil || !info.IsDir() {
				return fmt.Errorf("create NAS version folder %q: %w", current, err)
			}
		}
	}
	return nil
}

// Promote publishes a previously fully transferred and validated NAS staging
// file. When the live destination exists, its exact bytes are moved to a
// unique version path BEFORE publishing the new file. This API cannot infer
// whether the caller validated the source: the caller MUST do so first.
//
// If new publication fails, rollback restores the old live file. If the
// rollback also fails, it retains the old bytes at the reported version path
// and returns an error describing where an operator can recover them.
// Promote remains the legacy compatibility path.
func Promote(ctx context.Context, remote Remote, root, relative, temp, jobID string) (Result, error) {
 return promote(ctx,remote,nil,root,relative,temp,jobID)
}
func PromoteVerified(ctx context.Context,remote IntegrityRemote,root,relative,temp,jobID string)(Result,error){
 return promote(ctx,remote,remote,root,relative,temp,jobID)
}
func promote(ctx context.Context, remote Remote, integrity IntegrityRemote, root, relative, temp, jobID string) (Result, error) {
	if remote == nil || !safeRoot(root) || !safeRel(relative) || !validID(jobID) ||
		temp == "" || len(temp) > 4096 || strings.ContainsAny(temp, "\\\x00\r\n") || path.Clean(temp) != temp || strings.HasPrefix(temp, "/") {
		return Result{}, ErrUnsafePath
	}
	dest := path.Join(root, relative)
	if temp == dest || !strings.HasPrefix(temp, dest+".syncria-pending-") {
		return Result{}, ErrUnsafePath
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	staged, e := remote.Stat(temp)
	if e != nil {
		return Result{}, fmt.Errorf("verified staging file unavailable: %w", e)
	}
	if !staged.Mode().IsRegular() {
		return Result{}, errors.New("staging file is not regular")
	}
	old, e := remote.Stat(dest)
	if e != nil && !os.IsNotExist(e) {
		return Result{}, fmt.Errorf("cannot inspect NAS destination: %w", e)
	}
	if os.IsNotExist(e) {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		if err := remote.Rename(temp, dest); err != nil {
			return Result{}, fmt.Errorf("publish new NAS file: %w", err)
		}
		return Result{}, nil
	}
	if !old.Mode().IsRegular() {
		return Result{}, errors.New("cannot replace NAS folder or special file")
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return Result{}, err
	}
	version := path.Join(root, ".syncria-versions", jobID, hex.EncodeToString(nonce[:]), relative)
	if err := mkdirAll(remote, path.Dir(version)); err != nil {
		return Result{}, err
	}
	// A collision or hostile replacement of the unique backup path must fail
	// closed; do not replace an existing historical version.
	if _, err := remote.Stat(version); err == nil {
		return Result{}, errors.New("NAS version path already exists")
	} else if !os.IsNotExist(err) {
		return Result{}, err
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if err := remote.Rename(dest, version); err != nil {
		return Result{}, fmt.Errorf("preserve existing NAS version: %w", err)
	}
 if integrity!=nil{
   versionID:=path.Base(path.Dir(version))
   if err:=writeProof(ctx,integrity,root,jobID,versionID,relative,version,old);err!=nil{
     if undo:=remote.Rename(version,dest);undo!=nil{
       return Result{VersionPath:version,Replaced:true},fmt.Errorf("archive proof failed and rollback failed; old bytes at %s: %v: %w",version,undo,err)
     }
     return Result{},fmt.Errorf("archive proof failed; original restored: %w",err)
   }
 }
	if err := ctx.Err(); err != nil {
		if undo := remote.Rename(version, dest); undo != nil {
			return Result{VersionPath: version, Replaced: true}, fmt.Errorf("transfer canceled and rollback failed; original at %s: %v: %w", version, undo, err)
		}
		return Result{}, err
	}
	if err := remote.Rename(temp, dest); err != nil {
		if undo := remote.Rename(version, dest); undo != nil {
			return Result{VersionPath: version, Replaced: true}, fmt.Errorf("publication failed and rollback failed; original at %s: %v: %w", version, undo, err)
		}
		return Result{}, fmt.Errorf("publication failed; original restored: %w", err)
	}
	return Result{VersionPath: version, Replaced: true}, nil
}
