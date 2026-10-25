package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gigabytegrove/syncria/syncpolicy"
)

func TestAgentBridgeRespectsConfiguredClosedWindow(t *testing.T) {
	today := time.Now().UTC().Weekday()
	tomorrow := (today + 1) % 7
	j := Job{Provider: "agent_bridge", Direction: "one_way", ID: "aabbccddeeff00112233445566778899", AgentDeviceID: "approved-device", AgentGrantID: "approved-grant", AgentPeerURL: "https://agent.example:9774", NASDestinations: []NASDestination{{ShareID: "nas", Folder: "test"}}, AgentSyncPolicy: &syncpolicy.Policy{Windows: []syncpolicy.Window{{Timezone: "UTC", Days: []time.Weekday{tomorrow}, StartMinute: 60, EndMinute: 120}}}}
	a := &App{}
	err := a.syncAgentBridge(context.Background(), &j)
	if !errors.Is(err, syncpolicy.ErrOutsideWindow) {
		t.Fatalf("out-of-window device sync should be denied before connecting to source/NAS: %v", err)
	}
}
func TestAgentBridgeRejectsInvalidPersistedRatePolicy(t *testing.T) {
	j := Job{Provider: "agent_bridge", Direction: "one_way", AgentDeviceID: "approved-device", AgentGrantID: "approved-grant", AgentPeerURL: "https://agent.example:9774", NASDestinations: []NASDestination{{ShareID: "nas", Folder: "test"}}, AgentSyncPolicy: &syncpolicy.Policy{KiBPerSecond: -10}}
	if err := (&App{}).syncAgentBridge(context.Background(), &j); err == nil {
		t.Fatal("invalid policy should fail before transfer")
	}
}

func TestScheduledAgentJobsWaitForApprovedWindow(t *testing.T) {
	now := time.Now().UTC()
	today := now.Weekday()
	tomorrow := (today + 1) % 7
	job := Job{Provider: "agent_bridge", Enabled: true, Interval: 1, AgentSyncPolicy: &syncpolicy.Policy{Windows: []syncpolicy.Window{{Timezone: "UTC", Days: []time.Weekday{tomorrow}, StartMinute: 60, EndMinute: 120}}}}
	if syncCheckDue(job, now) {
		t.Fatal("job launched outside approved sync window")
	}
	job.AgentSyncPolicy = nil
	if !syncCheckDue(job, now) {
		t.Fatal("unrestricted job should remain due")
	}
}
