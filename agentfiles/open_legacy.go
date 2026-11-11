//go:build !linux && !go1.24

package agentfiles

import "os"

// Fail closed on older, non-Linux toolchains without safe root-relative opens.
func openUnderRoot(_, _ string) (*os.File, error) { return nil, ErrDenied }
