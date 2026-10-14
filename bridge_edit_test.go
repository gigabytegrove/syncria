package main

import (
 "strings"
 "testing"
)

func editorExample()Job{
 return Job{
  Provider:"bridge",CloudProvider:"google_drive",ID:"job-1",Name:"Original",
  AccountID:"account",RemoteID:"root",Direction:"bidirectional",
  NASDestinations:[]NASDestination{{ShareID:"share-1",Folder:"docs"}},
  ScheduleMode:"interval",Interval:5,Enabled:false,
  State:map[string]Item{"original.txt":{MD5:"local-old"}},
  PairStates:map[string]map[string]Item{"drive|share-1": {"original.txt":{MD5:"cloud-old"}}},
  ConflictOverrides:map[string]string{"drive|share-1|a.txt":"primary"},
  CloudOriginHashes:map[string]string{"a.txt":"cloud-old"},
  LastError:"previous error",SuccessfulRuns:8,FailedRuns:1,
 }
}
func TestEditSameEndpointsPreservesAllCheckpoints(t *testing.T){
 old:=editorExample()
 changed:=old
 changed.Name="Updated"
 changed.Direction="one_way"
 changed.ScheduleMode="changes"
 changed.Interval=10
 changed.PairStates=nil
 changed.State=nil
 changed.CloudOriginHashes=nil
 changed.ConflictOverrides=nil
 changed.LastError=""
 if err:=carryEditedSync(&changed,old,false);err!=nil{t.Fatalf("Same endpoints should edit safely: %v",err)}
 if changed.ID!=old.ID||changed.SuccessfulRuns!=8||changed.LastError!="previous error"{t.Fatal("Edit must preserve job identity and history")}
 if len(changed.PairStates)!=1||len(changed.State)!=1||len(changed.ConflictOverrides)!=1||len(changed.CloudOriginHashes)!=1{t.Fatal("Edit lost comparison history")}
 if changed.Direction!="one_way"||changed.Name!="Updated"{t.Fatal("Edits weren't saved")}
 if changed.Enabled{t.Fatal("Editor must keep sync paused")}
}
func TestEditEndpointChangeRequiresExplicitBaselineReset(t *testing.T){
 old:=editorExample()
 changed:=old
 changed.RemoteID="new-folder"
 changed.PairStates=map[string]map[string]Item{}
 changed.State=map[string]Item{}
 if err:=carryEditedSync(&changed,old,false);err!=errSyncResetRequired{t.Fatalf("Expected reset acknowledgment, got %v",err)}
 if err:=carryEditedSync(&changed,old,true);err!=nil{t.Fatalf("acknowledged remap should be supported: %v",err)}
 if len(changed.PairStates)!=0||len(changed.State)!=0{t.Fatal("Stale checkpoint leaked into a different source folder")}
 if changed.LastError!=""{t.Fatal("Stale scan error should clear after remapping")}
}
func TestEditRejectsActiveAndPendingRecovery(t *testing.T){
 old:=editorExample();old.Enabled=true
 changed:=old
 if err:=carryEditedSync(&changed,old,false);err!=errSyncMustBePaused{t.Fatalf("Expected paused requirement, got %v",err)}
 old.Enabled=false;old.PendingDeletions=[]PendingDeletion{{Path:"a.txt"}}
 changed=old;changed.RemoteID="new"
 if err:=carryEditedSync(&changed,old,true);err!=errSyncPendingRecovery{t.Fatalf("Expected protected deletion, got %v",err)}
}
func TestLegacyCloudEndpointCanEditWithoutDroppingState(t *testing.T){
 old:=editorExample()
 old.CloudProvider=""
 changed:=old;changed.CloudProvider="google_drive"
 if !sameSyncEndpoints(old,changed){t.Fatal("Legacy cloud source misidentified as an endpoint change")}
}
func TestCloudDirectionProposedAndExecutedForFirstRun(t *testing.T){
 job:=editorExample();job.Direction="one_way"
 l:=bridgeEntry{Size:10,Fingerprint:"cloud"};r:=bridgeEntry{Size:11,Fingerprint:"nas"}
 d:=decideFirstRun(&job,"pair","file",l,r,true,conflictManual)
 if d.Direction!="primary"||d.Method!=conflictPrimary{t.Fatal("Source must win explicit one-way cloud relationship")}
 plan:=syncPreviewResult{Counts:map[string]int{},Rows:[]syncPreviewRow{}}
 planSyncPair(&plan,&job,"new-pair",map[string]bridgeEntry{"file":l},map[string]bridgeEntry{"file":r},true,conflictManual)
 if len(plan.Rows)!=1||plan.Rows[0].From!="Primary"||plan.Rows[0].To!="Secondary"{t.Fatalf("Preview does not match one-way source direction: %+v",plan.Rows)}
 nas:=job;nas.Provider="nas_bridge"
 d=decideFirstRun(&nas,"pair","file",l,r,true,conflictManual)
 if d.Direction!=""{t.Fatal("Legacy one-way NAS protections must remain unchanged")}
}
func TestSyncEditorUIControlsPresent(t *testing.T){
 for _,expect:=range []string{
  `onclick="editSync('`,`id="editingSyncId"`,`id="cloudDirection"`,
  `name="cloud_direction"`,`id="nasDirection"`,`id="editResetMapping"`,
  `/api/sync/config?id=`,`function editSync(id)`,`name="edit_id"`,
 }{
  if !strings.Contains(newUI,expect){t.Errorf("Missing editor control %s",expect)}
 }
}
