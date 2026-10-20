//go:build linux

package agentpublish

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

// openat directory handles, O_NOFOLLOW and an atomic linkat(no-replace)
// ensure that an existing file is never replaced, even when an attacker races
// to insert a file at the target name. There is no string-based absolute-path
// operation after the approved root directory has been opened.
func publishNoReplace(ctx context.Context, root, relative string, copyAndVerify func(io.Writer) error, recheck func() error) error {
	if err := validateRoot(root); err != nil {
		return err
	}
	parts, err := pathSegments(relative)
	if err != nil {
		return err
	}
	rootFD, err := openRootByHandle(root)
	if err != nil {
		return err
	}
	defer syscall.Close(rootFD)
	parentFD := rootFD
	for _, dir := range parts[:len(parts)-1] {
		if err = ctx.Err(); err != nil {
			return err
		}
		fd, e := syscall.Openat(parentFD, dir, syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
		if errors.Is(e, syscall.ENOENT) {
			if err = recheck(); err != nil {
				return err
			}
			e = syscall.Mkdirat(parentFD, dir, 0700)
			if e != nil && !errors.Is(e, syscall.EEXIST) {
				return e
			}
			if e == nil {
				if err = syscall.Fsync(parentFD); err != nil {
					return err
				}
			}
			fd, e = syscall.Openat(parentFD, dir, syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
		}
		if e != nil {
			return e
		}
		defer syscall.Close(fd)
		parentFD = fd
	}
	target := parts[len(parts)-1]
	// Try the exclusive create first only as a quick refusal; the final
	// atomic linkat remains the authority and cannot replace a racer.
	if fd, e := syscall.Openat(parentFD, target, syscall.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0); e == nil {
		syscall.Close(fd)
		return ErrExists
	} else if !errors.Is(e, syscall.ENOENT) {
		return e
	}
	nonce := make([]byte, 16)
	if _, err = rand.Read(nonce); err != nil {
		return err
	}
	temporary := ".syncria-stage-" + hex.EncodeToString(nonce)
	fd, err := syscall.Openat(parentFD, temporary, syscall.O_WRONLY|syscall.O_CREAT|syscall.O_EXCL|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0600)
	if err != nil {
		return err
	}
	defer syscall.Unlinkat(parentFD, temporary)
	f := os.NewFile(uintptr(fd), temporary)
	if err = copyAndVerify(f); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = recheck(); err != nil {
		return err
	}
	// Linux linkat is atomic, fails EEXIST, and preserves checked dirfds.
	// SYS_LINKAT is used because Go's historical syscall package omits
	// its convenience wrapper. No unsafe pointer escapes the syscall.
	oldptr, err := syscall.BytePtrFromString(temporary)
	if err != nil {
		return err
	}
	newptr, err := syscall.BytePtrFromString(target)
	if err != nil {
		return err
	}
	_, _, errno := syscall.Syscall6(syscall.SYS_LINKAT, uintptr(parentFD), uintptr(unsafe.Pointer(oldptr)), uintptr(parentFD), uintptr(unsafe.Pointer(newptr)), 0, 0)
	if errno != 0 {
		if errors.Is(errno, syscall.EEXIST) {
			return ErrExists
		}
		return errno
	}
	// File is durable before publication; after unlinking the temporary name
	// flush the parent directory so the final directory entry survives a crash.
	if err = syscall.Unlinkat(parentFD, temporary); err != nil {
		return err
	}
	return syscall.Fsync(parentFD)
}

// openRootByHandle walks every component of the authorized root from the
// filesystem root without following symlinks. This prevents an ancestor swap
// between validating the approved path and opening the destination root.
func openRootByHandle(root string) (int, error) {
	fd, err := syscall.Open("/", syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		return -1, err
	}
	for _, component := range strings.Split(strings.TrimPrefix(filepath.Clean(root), "/"), "/") {
		if component == "" {
			continue
		}
		next, e := syscall.Openat(fd, component, syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
		_ = syscall.Close(fd)
		if e != nil {
			return -1, e
		}
		fd = next
	}
	return fd, nil
}
