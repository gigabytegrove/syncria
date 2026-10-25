//go:build !linux && go1.24

package agentfiles

import (
	"os"
	"path/filepath"
	"strings"
)

// Go 1.24+ os.Root uses platform-specific traversal resistance, including
// Windows reparse points, to anchor access within the approved directory.
func openUnderRoot(root, relative string) (*os.File, error) {
	if !filepath.IsAbs(root) || relative == "" || filepath.IsAbs(relative) || strings.ContainsAny(relative, "\\\x00") {
		return nil, ErrDenied
	}
	for _, p := range strings.Split(relative, "/") {
		if p == "" || p == "." || p == ".." || strings.HasPrefix(p, ".syncria-") {
			return nil, ErrDenied
		}
	}
	r, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return r.Open(filepath.FromSlash(relative))
}
