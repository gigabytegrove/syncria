package main

import (
 "fmt"
 "net/http"
 "path/filepath"

 "github.com/gigabytegrove/syncria/agentcontrol"
)

func registerAgentAPIs(mux *http.ServeMux, a *App, dataDir string) error {
 store, err := agentcontrol.Open(filepath.Join(dataDir, "agents.json"))
 if err != nil { return fmt.Errorf("agent identity registry: %w", err) }
 mux.Handle("/api/v1/agents/", agentcontrol.Handler{
  Store: store,
  Admin: a.authenticated,
  SameOriginPost: localPost,
 })
 return nil
}
