package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gigabytegrove/syncria/syncpolicy"

	"github.com/gigabytegrove/syncria/agentpeer"
)

// agentReadEndpoint adapts a locally approved device folder into Syncria's
// existing indexed bridge engine. This endpoint is ALWAYS read-only. Device
// deletion, local mutation, and cloud-account operations cannot be called.
type agentReadEndpoint struct {
	peer      agentpeer.Client
	grant     string
	scanHook  func(string, int64)
	inventory map[string]bridgeEntry
	policy *syncpolicy.Policy
}

func (p *agentReadEndpoint) Close() {}
func (p *agentReadEndpoint) Scan(ctx context.Context) (map[string]bridgeEntry, error) {
	entries, err := p.peer.Inventory(ctx, p.grant)
	if err != nil {
		return nil, fmt.Errorf("device inventory unavailable: %w", err)
	}
	items := make(map[string]bridgeEntry, len(entries))
	folded := make(map[string]string, len(entries))
	for i, e := range entries {
		if e.Directory {
			return nil, errors.New("device inventory contains unsupported directory record")
		}
		if !agentSafeRelative(e.Path) || e.Size < 0 || e.Size > 1<<40 || e.Modified.IsZero() {
			return nil, fmt.Errorf("unsafe device file metadata at entry %d", i)
		}
		name := strings.ToLower(e.Path)
		if former, ok := folded[name]; ok {
			return nil, fmt.Errorf("colliding device paths %q and %q", former, e.Path)
		}
		folded[name] = e.Path
		items[e.Path] = bridgeEntry{Size: e.Size, Modified: e.Modified, Fingerprint: fmt.Sprintf("agentmeta|%020d|%020d", e.Size, e.Modified.UnixNano())}
		if p.scanHook != nil {
			p.scanHook(e.Path, int64(i+1))
		}
	}
	// A file cannot be its own directory ancestor on a NAS. Reject case-folded
	// conflicts before performing any remote writes.
	for path := range items {
		parts := strings.Split(path, "/")
		for n := 1; n < len(parts); n++ {
			prefix := strings.ToLower(strings.Join(parts[:n], "/"))
			if _, ok := folded[prefix]; ok {
				return nil, fmt.Errorf("file is an ancestor of another file: %q", path)
			}
		}
	}
	p.inventory = items
	return items, nil
}
func agentSafeRelative(rel string) bool {
	if rel == "" || len(rel) > 2048 || strings.HasPrefix(rel, "/") || strings.ContainsAny(rel, "\\\x00\r\n") {
		return false
	}
	for _, part := range strings.Split(rel, "/") {
		if part == "" || part == "." || part == ".." || len(part) > 255 || strings.HasPrefix(part, ".syncria-") {
			return false
		}
		for _, r := range part {
			if r < 32 || r == 127 {
				return false
			}
		}
	}
	return true
}
func (p *agentReadEndpoint) Open(ctx context.Context, rel string) (io.ReadCloser, int64, error) {
	recorded, ok := p.inventory[rel]
	if !ok {
		return nil, 0, fmt.Errorf("%q was not in approved device inventory", rel)
	}
	r, size, err := p.peer.OpenVerified(ctx, p.grant, rel)
	if err != nil {
		return nil, 0, err
	}
	if size != recorded.Size {
		_ = r.Close()
		return nil, 0, errors.New("device file changed after inventory")
	}
	if p.policy != nil {
        rate := &syncpolicy.RateReader{Source: r, Context: ctx, BytesPerSecond: int64(p.policy.KiBPerSecond) * 1024}
        if len(p.policy.Windows) > 0 {
            rate.Allowed = func(now time.Time) bool { return p.policy.Allowed(now) }
        }
        return rate, size, nil
    }
    return r, size, nil
}
func (p *agentReadEndpoint) Mkdir(context.Context, string) (bridgeEntry, error) {
	return bridgeEntry{}, errors.New("device endpoint is read-only")
}
func (p *agentReadEndpoint) Put(context.Context, string, io.Reader, int64, string, func(int64)) (bridgeEntry, error) {
	return bridgeEntry{}, errors.New("device endpoint is read-only")
}
func (p *agentReadEndpoint) StageDelete(context.Context, string, string) (string, error) {
	return "", errors.New("device deletion disabled")
}
func (p *agentReadEndpoint) Restore(context.Context, string, string) error {
	return errors.New("device mutation disabled")
}
func (p *agentReadEndpoint) Purge(context.Context, string) error {
	return errors.New("device mutation disabled")
}
func (p *agentReadEndpoint) Delete(context.Context, string) error {
	return errors.New("device deletion disabled")
}

// syncAgentBridge is executed by the exact same scheduler, per-job run lock,
// progress history, failure tracking and bbolt checkpointing as NAS↔NAS jobs.
// It is an opt-in one-way policy only; absence never propagates deletions.
func (a *App) syncAgentBridge(ctx context.Context, j *Job) error {
	if j.Provider != "agent_bridge" || j.Direction != "one_way" {
		return errors.New("device relationship must be one-way")
	}
	if j.AgentDeviceID == "" || j.AgentGrantID == "" || j.AgentPeerURL == "" || len(j.NASDestinations) != 1 {
		return errors.New("incomplete device/NAS relationship")
	}
	if j.AgentSyncPolicy != nil {
        if err := j.AgentSyncPolicy.Validate(); err != nil { return fmt.Errorf("invalid agent sync policy: %w", err) }
        if !j.AgentSyncPolicy.Allowed(time.Now()) { return syncpolicy.ErrOutsideWindow }
    }
    secret, err := readAgentPeerCredential(a.file, j.ID)
	if err != nil {
		return fmt.Errorf("device peer credential unavailable: %w", err)
	}
	device := &agentReadEndpoint{peer: agentpeer.Client{Server: j.AgentPeerURL, Token: secret}, grant: j.AgentGrantID, policy: j.AgentSyncPolicy}
	device.scanHook = func(rel string, count int64) {
		if j.scanHook != nil {
			j.scanHook("Device", rel, count)
		}
	}
	nas, err := a.openNAS(ctx, j.NASDestinations[0])
	if err != nil {
		return fmt.Errorf("destination NAS: %w", err)
	}
    smb,ok:=nas.(*bridgeSMBEndpoint)
    if !ok {nas.Close();return errors.New("NAS version retention requires a direct SMB destination")}
    smb.archiveJobID=j.ID
	// Reuse the established one-way NAS conflict rule: no replacement when the
	// destination changed independently, no automatic first-run overwrite.
	// The job on disk remains provider=agent_bridge and is never modified here.
	safe := *j
	// Keep the actual device provider in the shared policy engine; never masquerade
	// as a NAS-only job to bypass or reinterpret device-specific safety rules.
	safe.PropagateDeletions = boolPtr(false)
	safe.RecoveryEnabled = boolPtr(true)
	key := relationPairKey("device:"+j.AgentDeviceID+":"+j.AgentGrantID, j.NASDestinations[0])
	err = a.executePolicyPair(ctx, &safe, device, nas, key, true, conflictManual)
	j.PairStates = safe.PairStates
	j.PendingDeletions = safe.PendingDeletions
	j.Conflicts = safe.Conflicts
	j.LastUploads = safe.LastUploads
	j.LastDownloads = safe.LastDownloads
	j.LastUploadBytes = safe.LastUploadBytes
	j.LastDownloadBytes = safe.LastDownloadBytes
	j.TotalUploads = safe.TotalUploads
	j.TotalDownloads = safe.TotalDownloads
	j.TotalUploadBytes = safe.TotalUploadBytes
	j.TotalDownloadBytes = safe.TotalDownloadBytes
	j.LastSummary = safe.LastSummary
	return err
}

var _ recoverableBridgeEndpoint = (*agentReadEndpoint)(nil)
