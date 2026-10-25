package synctopology

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestGraphBuildUsesConfiguredJobsOnly(t *testing.T) {
	nodes := []Node{{ID: "agent:a", Kind: "agent", Name: "Desktop", Status: "recently_seen"}, {ID: "nas:b", Kind: "nas", Name: "NAS", Status: "configured"}, {ID: "cloud:c", Kind: "cloud", Name: "Drive", Status: "configured"}}
	relationships := []Relationship{{ID: "j2", Name: "NAS Backup", SourceID: "nas:b", TargetIDs: []string{"agent:a"}, Mode: "bidirectional", Enabled: false}, {ID: "j1", Name: "Photos", SourceID: "agent:a", TargetIDs: []string{"nas:b"}, Mode: "one_way", Enabled: true}}
	graph, err := Build(nodes, relationships)
	if err != nil {
		t.Fatal(err)
	}
	if len(graph.Nodes) != 3 || len(graph.Edges) != 2 {
		t.Fatalf("unexpected graph: %+v", graph)
	}
	if graph.Edges[0].ID != "j1:nas:b" || graph.Edges[0].Status != "configured" || graph.Edges[1].Status != "paused" {
		t.Fatal("graph order/status incorrect")
	}
	if len(graph.Warnings) != 0 {
		t.Fatalf("spurious topology warning: %v", graph.Warnings)
	}
	sorted, err := Build([]Node{nodes[2], nodes[1], nodes[0]}, []Relationship{relationships[1], relationships[0]})
	if err != nil || !reflect.DeepEqual(graph, sorted) {
		t.Fatal("graph should be deterministic", err)
	}
}
func TestMissingEndpointsNeverSynthesized(t *testing.T) {
	g, err := Build([]Node{{ID: "nas:b", Kind: "nas", Name: "NAS"}}, []Relationship{{ID: "job", Name: "Offsite", SourceID: "agent:offline", TargetIDs: []string{"nas:b"}}, {ID: "job2", Name: "Copy", SourceID: "nas:b", TargetIDs: []string{"nas:missing"}}})
	if err != nil || len(g.Edges) != 0 || len(g.Warnings) != 2 {
		t.Fatalf("fabricated relationship: %+v error %v", g, err)
	}
}
func TestRejectDuplicatesAndInvalidInputs(t *testing.T) {
	_, err := Build([]Node{{ID: "a", Kind: "nas", Name: "NAS"}, {ID: "a", Kind: "nas", Name: "Duplicate"}}, nil)
	if err == nil {
		t.Fatal("duplicate nodes allowed")
	}
	_, err = Build([]Node{{ID: "a", Kind: "remote_desktop", Name: "PC"}}, nil)
	if err == nil {
		t.Fatal("unapproved endpoint type accepted")
	}
	n := []Node{{ID: "a", Kind: "nas", Name: "NAS"}}
	_, err = Build(n, []Relationship{{ID: "x", Name: "A", SourceID: "a"}, {ID: "x", Name: "B", SourceID: "a"}})
	if err == nil {
		t.Fatal("duplicate jobs accepted")
	}
}
func TestHeartbeatStatusesAreNotOpenNowClaims(t *testing.T) {
	now := time.Now().UTC()
	fresh := now.Add(-30 * time.Second)
	stale := now.Add(-10 * time.Minute)
	revoked := now
	if v := DeviceStatus(now, &fresh, nil); v != "recently_seen" {
		t.Fatal(v)
	}
	if v := DeviceStatus(now, &stale, nil); v != "not_recently_seen" {
		t.Fatal(v)
	}
	if v := DeviceStatus(now, nil, nil); v != "not_recently_seen" {
		t.Fatal(v)
	}
	if v := DeviceStatus(now, &fresh, &revoked); v != "revoked" {
		t.Fatal(v)
	}
}
func TestDanglingTargetCannotEscapeKnownGraph(t *testing.T) {
	nodes := []Node{{ID: "a", Kind: "agent", Name: "A"}, {ID: "b", Kind: "nas", Name: "B"}}
	graph, err := Build(nodes, []Relationship{{ID: "j", Name: "J", SourceID: "a", TargetIDs: []string{"b", "../../hidden", "b"}, Enabled: true}})
	if err != nil || len(graph.Edges) != 1 || len(graph.Warnings) != 2 {
		t.Fatalf("unexpected graph %+v %v", graph, err)
	}
	if strings.Contains(graph.Edges[0].TargetID, "hidden") {
		t.Fatal("invented untrusted edge")
	}
}
