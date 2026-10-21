// Package synctopology creates a deterministic, non-synthetic graph from
// existing Syncria endpoint records and persisted synchronization jobs.
// It has no scanner, scheduler, network client or independent state store.
package synctopology

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Node struct {
	ID       string     `json:"id"`
	Kind     string     `json:"kind"`
	Name     string     `json:"name"`
	Status   string     `json:"status"`
	LastSeen *time.Time `json:"last_seen,omitempty"`
}
type Relationship struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	SourceID  string   `json:"source_id"`
	TargetIDs []string `json:"target_ids"`
	Mode      string   `json:"mode"`
	Enabled   bool     `json:"enabled"`
	Error     string   `json:"error,omitempty"`
}
type Edge struct {
	ID             string `json:"id"`
	RelationshipID string `json:"relationship_id"`
	Name           string `json:"name"`
	SourceID       string `json:"source_id"`
	TargetID       string `json:"target_id"`
	Mode           string `json:"mode"`
	Status         string `json:"status"`
}
type Graph struct {
	Nodes    []Node   `json:"nodes"`
	Edges    []Edge   `json:"edges"`
	Warnings []string `json:"warnings"`
}

func DeviceStatus(now time.Time, lastSeen, revoked *time.Time) string {
	if revoked != nil {
		return "revoked"
	}
	if lastSeen == nil || lastSeen.After(now.Add(2*time.Minute)) {
		return "not_recently_seen"
	}
	if now.Sub(*lastSeen) <= 2*time.Minute {
		return "recently_seen"
	}
	return "not_recently_seen"
}
func Build(nodes []Node, relationships []Relationship) (Graph, error) {
	if len(nodes) > 2000 || len(relationships) > 10000 {
		return Graph{}, errors.New("topology too large")
	}
	g := Graph{Nodes: []Node{}, Edges: []Edge{}, Warnings: []string{}}
	keys := make(map[string]struct{}, len(nodes))
	for _, n := range nodes {
		if n.ID == "" || n.Name == "" || !validKind(n.Kind) || len(n.Name) > 256 || len(n.ID) > 320 {
			return Graph{}, errors.New("invalid topology endpoint")
		}
		if _, ok := keys[n.ID]; ok {
			return Graph{}, fmt.Errorf("duplicate endpoint ID %q", n.ID)
		}
		keys[n.ID] = struct{}{}
		g.Nodes = append(g.Nodes, n)
	}
	sort.Slice(g.Nodes, func(i, j int) bool {
		if g.Nodes[i].Kind != g.Nodes[j].Kind {
			return g.Nodes[i].Kind < g.Nodes[j].Kind
		}
		return g.Nodes[i].ID < g.Nodes[j].ID
	})
	seenJobs := map[string]bool{}
	for _, r := range relationships {
		if r.ID == "" || r.Name == "" || seenJobs[r.ID] || len(r.TargetIDs) > 256 {
			return Graph{}, errors.New("invalid or duplicate synchronization relationship")
		}
		seenJobs[r.ID] = true
		if _, ok := keys[r.SourceID]; !ok {
			g.Warnings = append(g.Warnings, fmt.Sprintf("Relationship %q has an unavailable source", r.Name))
			continue
		}
		targets := map[string]bool{}
		for _, target := range r.TargetIDs {
			if target == r.SourceID || targets[target] {
				g.Warnings = append(g.Warnings, fmt.Sprintf("Relationship %q has an invalid destination", r.Name))
				continue
			}
			targets[target] = true
			if _, ok := keys[target]; !ok {
				g.Warnings = append(g.Warnings, fmt.Sprintf("Relationship %q has an unavailable destination", r.Name))
				continue
			}
			status := "paused"
			if r.Enabled {
				status = "configured"
			}
			if r.Error != "" {
				status = "attention"
			}
			g.Edges = append(g.Edges, Edge{ID: r.ID + ":" + target, RelationshipID: r.ID, Name: r.Name, SourceID: r.SourceID, TargetID: target, Mode: mode(r.Mode), Status: status})
		}
	}
	sort.Slice(g.Edges, func(i, j int) bool { return strings.Compare(g.Edges[i].ID, g.Edges[j].ID) < 0 })
	return g, nil
}
func validKind(s string) bool { return s == "agent" || s == "nas" || s == "cloud" }
func mode(s string) string {
	switch s {
	case "one_way", "bidirectional":
		return s
	default:
		return "unknown"
	}
}
