//go:build !linux

package agentpublish

import (
	"context"
	"io"
)

// Fail closed until platform-native handle-relative, atomic no-replace writes
// are implemented and tested. Do not approximate with os.Rename: that can
// replace the destination and expose unverified content.
func publishNoReplace(context.Context, string, string, func(io.Writer) error, func() error) error {
	return ErrUnsupported
}
