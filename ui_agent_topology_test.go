package main

import (
	"strings"
	"testing"
)

// The feature must remain part of the existing addressable/hash-routed
// workspace, not introduce a second dashboard or placeholder controls.
func TestIntegratedSyncTopologyDashboard(t *testing.T) {
	for _, needle := range []string{`data-view="topology"`, `id="view-topology"`, `id="syncTopologyEdges"`, `id="syncTopologyNodes"`, `function refreshSyncTopology()`, `/api/v1/sync/topology`, `if(view==='topology')refreshSyncTopology();`} {
		if strings.Count(newUI, needle) != 1 {
			t.Errorf("topology UI marker missing or duplicated %q", needle)
		}
	}
	if strings.Contains(newUI, `innerHTML=node.name`) {
		t.Fatal("unsafe rendering of user supplied device names")
	}
}
