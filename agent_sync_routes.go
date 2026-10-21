package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gigabytegrove/syncria/agentcontrol"
	"github.com/gigabytegrove/syncria/agentpeer"
	"github.com/gigabytegrove/syncria/syncpolicy"
)

// A device pairing credential is distinct from the one-time enrollment secret.
// Its raw value is stored only in a separate owner-only per-job file, never in
// state.json, browser route parameters, run-history or audit event payloads.
func agentPeerTokenPath(configFile, jobID string) (string, error) {
	if len(jobID) != 32 {
		return "", errors.New("invalid job ID")
	}
	for _, r := range jobID {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
			return "", errors.New("invalid job ID")
		}
	}
	return filepath.Join(filepath.Dir(configFile), "agent-peer-keys", jobID+".key"), nil
}
func saveAgentPeerCredential(configFile, jobID, token string) error {
	if _, err := agentpeer.NewHandler(nil, token); err != nil {
		return errors.New("invalid device peer credential")
	}
	path, err := agentPeerTokenPath(configFile, jobID)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err = os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	stat, err := os.Lstat(dir)
	if err != nil || !stat.IsDir() || stat.Mode()&os.ModeSymlink != 0 || runtime.GOOS != "windows" && stat.Mode().Perm()&0077 != 0 {
		return errors.New("unsafe device credential directory")
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	if _, err = file.Write([]byte(token + "\n")); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return err
	}
	if err = file.Sync(); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return err
	}
	return file.Close()
}
func readAgentPeerCredential(configFile, jobID string) (string, error) {
	path, err := agentPeerTokenPath(configFile, jobID)
	if err != nil {
		return "", err
	}
	return agentpeer.ReadKey(path)
}
func removeAgentPeerCredential(configFile, jobID string) error {
	path, err := agentPeerTokenPath(configFile, jobID)
	if err != nil {
		return err
	}
	if err = os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
func validateAgentPeerURL(value string) error {
	u, err := url.Parse(value)
	if err != nil || u == nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return errors.New("invalid device address")
	}
	if u.Scheme == "https" {
		return nil
	}
	if u.Scheme == "http" {
		host := u.Hostname()
		ip := net.ParseIP(host)
		if strings.EqualFold(host, "localhost") || ip != nil && ip.IsLoopback() {
			return nil
		}
	}
	return errors.New("device connection requires HTTPS except local loopback testing")
}

type agentSyncCreateRequest struct {
	Name       string `json:"name"`
	DeviceID   string `json:"device_id"`
	PeerURL    string `json:"peer_url"`
	GrantID    string `json:"grant_id"`
	PeerToken  string `json:"peer_token"`
	NASShareID string `json:"nas_share_id"`
	NASFolder  string `json:"nas_folder"`
	Interval   int    `json:"interval_minutes"`
	Policy *syncpolicy.Policy `json:"policy,omitempty"`
}

func (a *App) registerAgentSyncRoutes(mux *http.ServeMux, control *agentcontrol.Store) {
	mux.HandleFunc("/api/v1/agent-sync/jobs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !a.authenticated(r) {
			http.Error(w, "unauthorized", 401)
			return
		}
		if r.Method == "GET" {
			a.mu.Lock()
			var jobs []map[string]any
			for _, j := range a.cfg.Jobs {
				if j.Provider == "agent_bridge" {
					jobs = append(jobs, map[string]any{"id": j.ID, "name": j.Name, "device_id": j.AgentDeviceID, "grant_id": j.AgentGrantID, "nas_share_id": j.NASShareID, "nas_folder": j.NASFolder, "enabled": j.Enabled, "last_error": j.LastError, "last_summary": j.LastSummary})
				}
			}
			a.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"jobs": jobs})
			return
		}
		if r.Method != "POST" {
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", 405)
			return
		}
		if !localPost(r) {
			http.Error(w, "same-origin POST required", 403)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 8192)
		var req agentSyncCreateRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			http.Error(w, "invalid JSON", 400)
			return
		}
		var extra any
		if !errors.Is(dec.Decode(&extra), io.EOF) {
			http.Error(w, "invalid trailing JSON", 400)
			return
		}
		if req.Name == "" || len(req.Name) > 128 || strings.TrimSpace(req.Name) != req.Name || req.Interval < 1 || req.Interval > 1440 || len(req.GrantID) != 32 || len(req.DeviceID) == 0 {
			http.Error(w, "invalid relationship settings", 400)
			return
		}
		if req.Policy != nil { if err:=req.Policy.Validate(); err != nil { http.Error(w,"invalid bandwidth or schedule policy",400);return } }
		if err := validateAgentPeerURL(req.PeerURL); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if _, err := agentpeer.NewHandler(nil, req.PeerToken); err != nil {
			http.Error(w, "invalid peer token", 400)
			return
		}
		known := false
		for _, d := range control.ListDevices() {
			if d.ID == req.DeviceID && d.RevokedAt == nil {
				known = true
				break
			}
		}
		if !known {
			http.Error(w, "enroll and approve this device first", 409)
			return
		}
		destination := NASDestination{ShareID: req.NASShareID, Folder: req.NASFolder}
		share, err := a.directShare(destination.ShareID)
		if err != nil {
			http.Error(w, "NAS connection not found", 400)
			return
		}
		_ = share
		normalized, err := safeRemotePath(destination.Folder)
		if err != nil {
			http.Error(w, "invalid NAS destination path", 400)
			return
		}
		destination.Folder = normalized
		// Source and target must be reachable at creation, before recording the
		// new relationship. Offline source never means 'empty' when scheduled.
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		if _, err = (agentpeer.Client{Server: req.PeerURL, Token: req.PeerToken}).Inventory(ctx, req.GrantID); err != nil {
			http.Error(w, "approved device folder is unavailable: "+err.Error(), 409)
			return
		}
		if err = a.verifyNASFolder(ctx, destination); err != nil {
			http.Error(w, "NAS destination unavailable: "+err.Error(), 409)
			return
		}
		id := newID()
		job := Job{Provider: "agent_bridge", RelationshipType: "agent", ID: id, Name: req.Name, Direction: "one_way", ConflictPolicy: conflictManual, AgentDeviceID: req.DeviceID, AgentGrantID: req.GrantID, AgentPeerURL: req.PeerURL, NASShareID: destination.ShareID, NASFolder: destination.Folder, NASDestinations: []NASDestination{destination}, ScheduleMode: "interval", Interval: req.Interval, Enabled: false, AgentSyncPolicy: req.Policy, State: map[string]Item{}, PropagateDeletions: boolPtr(false), RecoveryEnabled: boolPtr(true)}
		if err = saveAgentPeerCredential(a.file, id, req.PeerToken); err != nil {
			http.Error(w, "could not save device credential", 500)
			return
		}
		a.mu.Lock()
		// No alpha destination may overlap any existing enabled/disabled NAS sync:
		// protect existing data even when the relationship is paused.
		conflict := false
		for _, other := range a.cfg.Jobs {
			for _, d := range other.NASDestinations {
				if relationshipOverlap(destination, d) {
					conflict = true
					break
				}
			}
			if other.NASShareID != "" && relationshipOverlap(destination, NASDestination{ShareID: other.NASShareID, Folder: other.NASFolder}) {
				conflict = true
			}
			if other.NASSourceID != "" && relationshipOverlap(destination, NASDestination{ShareID: other.NASSourceID, Folder: other.NASSourceFolder}) {
				conflict = true
			}
		}
		if conflict {
			a.mu.Unlock()
			_ = removeAgentPeerCredential(a.file, id)
			http.Error(w, "NAS folder overlaps an existing relationship", 409)
			return
		}
		a.cfg.Jobs = append(a.cfg.Jobs, job)
		err = a.save()
		if err != nil {
			a.cfg.Jobs = a.cfg.Jobs[:len(a.cfg.Jobs)-1]
		}
		a.mu.Unlock()
		if err != nil {
			_ = removeAgentPeerCredential(a.file, id)
			http.Error(w, "failed to save relationship", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": id, "enabled": false, "message": "Relationship created paused. Preview and enable after checking destination."})
	})
}
