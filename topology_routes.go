package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gigabytegrove/syncria/agentcontrol"
	"github.com/gigabytegrove/syncria/synctopology"
)

// The topology view is derived solely from existing relationships and
// enrolled devices. It creates no independent graph database or scheduler.
func (a *App) syncTopologyHandler(devices *agentcontrol.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, "method not allowed", 405)
			return
		}
		if !a.authenticated(r) {
			http.Error(w, "unauthorized", 401)
			return
		}
		now := time.Now().UTC()
		nodes := []synctopology.Node{}
		for _, d := range devices.ListDevices() {
			nodes = append(nodes, synctopology.Node{ID: "agent:" + d.ID, Kind: "agent", Name: d.Name, Status: synctopology.DeviceStatus(now, d.LastSeen, d.RevokedAt), LastSeen: d.LastSeen})
		}
		a.mu.Lock()
		for _, s := range a.cfg.Shares {
			label := s.Label
			if label == "" {
				label = s.ID
			}
			nodes = append(nodes, synctopology.Node{ID: "nas:" + s.ID, Kind: "nas", Name: label, Status: "configured"})
		}
		for _, c := range a.cfg.Accounts {
			label := c.Label
			if label == "" {
				label = c.Email
			}
			if label == "" {
				label = c.ID
			}
			nodes = append(nodes, synctopology.Node{ID: "cloud:google:" + c.ID, Kind: "cloud", Name: label, Status: "configured"})
		}
		for _, c := range a.cfg.DropboxAccounts {
			label := c.Email
			if label == "" {
				label = c.ID
			}
			nodes = append(nodes, synctopology.Node{ID: "cloud:dropbox:" + c.ID, Kind: "cloud", Name: label, Status: "configured"})
		}
		relationships := []synctopology.Relationship{}
		unsupported := 0
		for _, j := range a.cfg.Jobs {
			relation := synctopology.Relationship{ID: j.ID, Name: j.Name, Mode: j.Direction, Enabled: j.Enabled, Error: j.LastError}
			switch j.Provider {
			case "agent_bridge":
				relation.SourceID = "agent:" + j.AgentDeviceID
			case "nas_bridge":
				relation.SourceID = "nas:" + j.NASSourceID
			case "bridge":
				if j.CloudProvider == "dropbox" {
					relation.SourceID = "cloud:dropbox:" + j.DropboxAccountID
				} else {
					relation.SourceID = "cloud:google:" + j.AccountID
				}
			default:
				unsupported++
				continue
			}
			for _, dest := range j.NASDestinations {
				relation.TargetIDs = append(relation.TargetIDs, "nas:"+dest.ShareID)
			}
			if len(relation.TargetIDs) == 0 && j.NASShareID != "" {
				relation.TargetIDs = append(relation.TargetIDs, "nas:"+j.NASShareID)
			}
			relationships = append(relationships, relation)
		}
		a.mu.Unlock()
		graph, err := synctopology.Build(nodes, relationships)
		if err != nil {
			http.Error(w, "topology unavailable", 503)
			return
		}
		if unsupported > 0 {
			graph.Warnings = append(graph.Warnings, "Some legacy local-storage relationships do not have endpoint graph metadata")
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(graph)
	}
}
