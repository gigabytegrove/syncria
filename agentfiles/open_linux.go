//go:build linux

package agentfiles

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// openUnderRoot uses openat on every component, O_NOFOLLOW on directories
// and the file, and rejects pipes/devices. Check+open races cannot redirect
// reads outside the originally opened directory handles.
func openUnderRoot(root, relative string) (*os.File, error) {
	if !filepath.IsAbs(root) || relative == "" || filepath.IsAbs(relative) || strings.ContainsAny(relative, "\\\x00") {
		return nil, ErrDenied
	}
	parts := strings.Split(relative, "/")
	for _, p := range parts {
		if p == "" || p == "." || p == ".." || strings.HasPrefix(p, ".syncria-") {
			return nil, ErrDenied
		}
	}
	dir, err := syscall.Open(root, syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	defer func() { _ = syscall.Close(dir) }()
	for _, part := range parts[:len(parts)-1] {
		child, e := syscall.Openat(dir, part, syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
		if e != nil {
			return nil, e
		}
		syscall.Close(dir)
		dir = child
	}
	fd, err := syscall.Openat(dir, parts[len(parts)-1], syscall.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(fd), "approved-file")
	st, e := f.Stat()
	if e != nil {
		f.Close()
		return nil, e
	}
	if !st.Mode().IsRegular() {
		f.Close()
		return nil, errors.New("file is not regular")
	}
	return f, nil
}
