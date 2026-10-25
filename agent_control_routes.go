package main

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gigabytegrove/syncria/agentcontrol"
)

// registerAgentAPIs adds identity-only management routes to the existing
// server without changing cloud/NAS jobs or stored synchronization data.
func registerAgentAPIs(mux *http.ServeMux, a *App, dataDir string) error {
	store, err := agentcontrol.Open(filepath.Join(dataDir, "agents.json"))
	if err != nil {
		return fmt.Errorf("agent identity registry: %w", err)
	}
	mux.Handle("/api/v1/agents/", agentcontrol.Handler{
		Store:          store,
		Admin:          a.authenticated,
		SameOriginPost: localPost,
	})
	a.registerAgentSyncRoutes(mux, store)
	mux.HandleFunc("/api/v1/agent-sync/versions", a.agentVersionAPI)
	mux.HandleFunc("/api/v1/agent-sync/versions/restore", a.agentVersionRestoreAPI)
	mux.HandleFunc("/api/v1/agent-sync/versions/retention", a.agentVersionRetentionAPI)
	return nil
}
