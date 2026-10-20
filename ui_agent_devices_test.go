package main

import (
	"strings"
	"testing"
)

func TestAgentDashboardIntegrated(t *testing.T) {
	for _, marker := range []string{
		`data-view="devices"`,
		`id="view-devices"`,
		`id="agentInvitationForm"`,
		`id="agentDeviceList"`,
		`id="agentInvitationList"`,
		`id="agentAuditList"`,
		`async function refreshAgentDevices()`,
		`if(view==='devices')refreshAgentDevices();`,
		`devices:['Connected devices','DEVICES']`,
	} {
		if strings.Count(newUI, marker) != 1 {
			t.Errorf("device UI marker should appear once: %s", marker)
		}
	}
	if strings.Contains(newUI, `localStorage.setItem('agent`) || strings.Contains(newUI, `sessionStorage.setItem('agent`) {
		t.Fatal("agent bearer credentials must not be persisted in browser storage")
	}
}
