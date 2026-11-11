package main

import (
 "strings"
 "testing"
 "time"
)

// Preview must never suggest an overwrite that live execution will reject.
func TestAgentFirstRunPreviewBlocksEvenSameSizeAndTimestamp(t *testing.T) {
 stamp:=time.Date(2026,10,20,12,0,0,0,time.UTC)
 j:=&Job{Provider:"agent_bridge",Direction:"one_way",PairStates:map[string]map[string]Item{}}
 primary:=map[string]bridgeEntry{"same.txt":{Size:6,Modified:stamp,Fingerprint:"agentmeta"}}
 secondary:=map[string]bridgeEntry{"same.txt":{Size:6,Modified:stamp,Fingerprint:"nasmeta"}}
 result:=syncPreviewResult{Counts:map[string]int{}}
 planSyncPair(&result,j,"device-pair",primary,secondary,true,conflictManual)
 if len(result.Rows)!=1||result.Rows[0].Action!="blocked"{t.Fatalf("unsafe preview plan: %+v",result.Rows)}
 if result.Counts["blocked"]!=1||!strings.Contains(result.Rows[0].Reason,"first-run"){t.Fatalf("missing collision warning: %+v",result.Rows)}
}

func TestAgentPreviewDoesNotProposeDeletionForAbsentDeviceFile(t *testing.T) {
 disabled:=false
 j:=&Job{Provider:"agent_bridge",Direction:"one_way",PropagateDeletions:&disabled,PairStates:map[string]map[string]Item{"pair":{"missing.txt":fileState("missing.txt","before","before")}}}
 p:=syncPreviewResult{Counts:map[string]int{}}
 planSyncPair(&p,j,"pair",map[string]bridgeEntry{},map[string]bridgeEntry{"missing.txt":{Size:6}},true,conflictManual)
 if len(p.Rows)!=1||p.Rows[0].Action!="skip"{t.Fatalf("preview proposed deletion: %+v",p.Rows)}
}

func TestAgentPreviewBlocksIndependentNASModification(t *testing.T) {
 original:=fileState("edited.txt","prior","prior")
 j:=&Job{Provider:"agent_bridge",Direction:"one_way",PairStates:map[string]map[string]Item{"pair":{"edited.txt":original}}}
 p:=syncPreviewResult{Counts:map[string]int{}}
 now:=time.Now()
 planSyncPair(&p,j,"pair",map[string]bridgeEntry{"edited.txt":{Fingerprint:original.MD5,Size:5,Modified:now}},map[string]bridgeEntry{"edited.txt":{Fingerprint:"modified externally",Size:22,Modified:now}},true,conflictManual)
 if len(p.Rows)!=1||p.Rows[0].Action!="blocked"{t.Fatalf("unsafe preview proposed replacement: %+v",p.Rows)}
}

func TestAgentPreviewCreatesOnlyNewApprovedFiles(t *testing.T) {
 j:=&Job{Provider:"agent_bridge",Direction:"one_way",PairStates:map[string]map[string]Item{}}
 p:=syncPreviewResult{Counts:map[string]int{}}
 planSyncPair(&p,j,"pair",map[string]bridgeEntry{"new.txt":{Size:17,Modified:time.Now()}},map[string]bridgeEntry{},true,conflictManual)
 if len(p.Rows)!=1||p.Rows[0].Action!="copy"{t.Fatalf("new file was not planned: %+v",p.Rows)}
}
