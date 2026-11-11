// Package agentrestore connects locally approved device folders to durable
// historical versions. Neither the Syncria controller nor a peer can invoke
// this package remotely, and restoration never overwrites the current file.
package agentrestore

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/gigabytegrove/syncria/agentfiles"
	"github.com/gigabytegrove/syncria/agentpolicy"
	"github.com/gigabytegrove/syncria/agentpublish"
	"github.com/gigabytegrove/syncria/agentversion"
)

var ErrSourceChanged = errors.New("source changed while its version was captured")

type Manager struct {
	Grants   *agentpolicy.Store
	Versions *agentversion.Store
}

// Save keeps an immutable copy on this device, outside Syncria's controller.
// A partially changing source is never accepted as a valid version.
func (m Manager) Save(ctx context.Context, jobID, grantID, relative string) (agentversion.Version, error) {
	if m.Grants == nil || m.Versions == nil {
		return agentversion.Version{}, agentpublish.ErrDenied
	}
	reader := agentfiles.Reader{Grants: m.Grants}
	f, original, err := reader.Open(grantID, relative)
	if err != nil {
		return agentversion.Version{}, err
	}
	defer f.Close()
	if err := ctx.Err(); err != nil {
		return agentversion.Version{}, err
	}
	version, err := m.Versions.Capture(ctx, jobID, relative, f, original.Size())
	if err != nil {
		return agentversion.Version{}, err
	}
	current, e := f.Stat()
	policyErr := m.Grants.Allows(grantID, relative, false)
	if e != nil || policyErr != nil || current.Size() != original.Size() || !current.ModTime().Equal(original.ModTime()) {
		// A validated content version is never retained under a misleading
		// source snapshot when the source changed or permission was revoked.
		if removeErr := m.Versions.Drop(version.ID); removeErr != nil {
			return agentversion.Version{}, fmt.Errorf("%w; abandoned archive cleanup: %v", ErrSourceChanged, removeErr)
		}
		return agentversion.Version{}, ErrSourceChanged
	}
	return version, nil
}

// RecoveredName avoids replacing any current data. It preserves extensions
// and embeds an unguessably generated version identifier to avoid collisions.
func RecoveredName(relative, versionID string) (string, error) {
	if len(versionID) != 32 {
		return "", errors.New("invalid version ID")
	}
	for _, r := range versionID {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
			return "", errors.New("invalid version ID")
		}
	}
	if relative == "" || relative == "." || path.Clean(relative) != relative || strings.HasPrefix(relative, "/") || strings.ContainsAny(relative, "\\\x00\r\n") {
		return "", agentpublish.ErrDenied
	}
	for _, part := range strings.Split(relative, "/") {
		if part == "" || part == "." || part == ".." || strings.HasPrefix(part, ".syncria-") {
			return "", agentpublish.ErrDenied
		}
		if len(part) > 255 {
			return "", agentpublish.ErrDenied
		}
	}
	base := path.Base(relative)
	ext := path.Ext(base)
	name := strings.TrimSuffix(base, ext) + " (restored-" + versionID[:12] + ")" + ext
	if len(name) > 255 {
		return "", agentpublish.ErrDenied
	}
	return path.Join(path.Dir(relative), name), nil
}

// Restore verifies the archived bytes twice: once before returning a stable
// open descriptor, and again during atomic no-replace publication. The
// destination's writable grant must still permit the recovered filename.
func (m Manager) Restore(ctx context.Context, versionID, grantID string) (agentpublish.Result, error) {
	if m.Grants == nil || m.Versions == nil {
		return agentpublish.Result{}, agentpublish.ErrDenied
	}
	f, v, err := m.Versions.OpenVerified(ctx, versionID)
	if err != nil {
		return agentpublish.Result{}, err
	}
	defer f.Close()
	name, err := RecoveredName(v.Relative, v.ID)
	if err != nil {
		return agentpublish.Result{}, err
	}
	return (agentpublish.Publisher{Grants: m.Grants}).PublishVerified(ctx, grantID, name, f, v.Bytes, v.SHA256)
}
