package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gigabytegrove/syncria/nasversion"
)

// agentVersionAPI lists versions retained on the destination NAS for a single
// approved agent-sync relationship. It streams no file data to the controller.
func (a *App) agentVersionAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if !a.authenticated(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	jobID := strings.TrimSpace(r.URL.Query().Get("job_id"))
	a.mu.Lock()
	found := a.job(jobID)
	var job Job
	if found != nil {
		job = *found
	}
	a.mu.Unlock()
	if found == nil || job.Provider != "agent_bridge" || len(job.NASDestinations) != 1 {
		http.Error(w, "agent synchronization job not found", http.StatusNotFound)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	endpoint, err := a.openNAS(ctx, job.NASDestinations[0])
	if err != nil {
		http.Error(w, "NAS version listing unavailable", http.StatusServiceUnavailable)
		return
	}
	defer endpoint.Close()
	smb, ok := endpoint.(*bridgeSMBEndpoint)
	if !ok {
		http.Error(w, "NAS version listing unsupported", http.StatusUnprocessableEntity)
		return
	}
	entries, err := nasversion.List(ctx, smb.client.share, smb.root, job.ID)
	if err != nil {
		http.Error(w, "NAS version listing failed; check server logs", http.StatusConflict)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{"job_id": job.ID, "versions": entries})
}
