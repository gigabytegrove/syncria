package main

import (
 "strings"
 "testing"
 "time"
)

func TestFirstRunCloudNewerTimestampPlanAndExecutionAgree(t *testing.T) {
 older:=time.Date(2026,1,5,8,0,0,0,time.UTC)
 newer:=older.Add(90*time.Hour)
 job:=&Job{Provider:"bridge",PairStates:map[string]map[string]Item{"drive|nas":{}}}
 primary:=map[string]bridgeEntry{
  "drive-newer.docx":{Size:120,Modified:newer,Fingerprint:"cloud-hash"},
  "nas-newer.png":{Size:100,Modified:older,Fingerprint:"cloud-hash2"},
 }
 secondary:=map[string]bridgeEntry{
  "drive-newer.docx":{Size:110,Modified:older,Fingerprint:"120:mtime"},
  "nas-newer.png":{Size:140,Modified:newer,Fingerprint:"140:mtime"},
 }
 plan:=syncPreviewResult{Rows:[]syncPreviewRow{},Counts:map[string]int{}}
 planSyncPair(&plan,job,"drive|nas",primary,secondary,false,conflictManual)
 got:=map[string]syncPreviewRow{}
 for _,row:=range plan.Rows {got[row.Path]=row}
 if a:=got["drive-newer.docx"];a.Action!="replace"||a.From!="Primary"||a.To!="Secondary"||!strings.Contains(a.Reason,"2026-01-"){t.Errorf("Drive newer not shown with evidence: %+v",a)}
 if a:=got["nas-newer.png"];a.Action!="replace"||a.From!="Secondary"||a.To!="Primary"{t.Errorf("NAS newer not chosen: %+v",a)}
 for name,row:=range got {
  dir:=decideFirstRun(job,"drive|nas",name,primary[name],secondary[name],false,conflictManual)
  expected:="Primary";if dir.Direction=="secondary"{expected="Secondary"}
  if row.From!=expected{t.Errorf("%s: preview %q differs from engine %q",name,row.From,dir.Direction)}
 }
 if len(job.PairStates["drive|nas"])!=0{t.Fatal("Planning may not create synchronization checkpoints")}
}
func TestFirstRunAmbiguousTimestampsRemainProtected(t *testing.T) {
 base:=time.Date(2026,1,1,0,0,0,0,time.UTC)
 job:=&Job{Provider:"bridge"}
 cases:=[]struct{name string;primary,secondary time.Time}{
  {"missing NAS",base,time.Time{}},
  {"identical",base,base},
  {"SMB precision",base,base.Add(time.Second)},
 }
 for _,tc:=range cases{
  d:=decideFirstRun(job,"pair",tc.name,bridgeEntry{Modified:tc.primary},bridgeEntry{Modified:tc.secondary},false,conflictManual)
  if d.Direction!=""{t.Errorf("%s: unsafe automatic winner %q",tc.name,d.Direction)}
 }
}
func TestFirstRunHonorsExplicitPolicyOneWayAndOverride(t *testing.T) {
 old:=time.Date(2026,1,1,0,0,0,0,time.UTC)
 newTime:=old.Add(time.Hour)
 l:=bridgeEntry{Modified:old};r:=bridgeEntry{Modified:newTime}
 job:=&Job{Provider:"bridge",ConflictPolicy:conflictManual,InitialConflictPolicy:conflictManual}
 d:=decideFirstRun(job,"p","f",l,r,false,conflictManual)
 if d.Direction!=""{t.Fatal("explicit initial manual selection must remain manual")}
 job.InitialConflictPolicy=""
 job.InitialConflictPolicy=""
 job.ConflictPolicy=conflictManual
 d=decideFirstRun(job,"p","f",l,r,false,conflictManual)
 if d.Direction!="secondary"{t.Fatal("legacy cloud manual preference was implicit; initial run must choose newer")}
 job.ConflictOverrides=map[string]string{"p|f":"primary"}
 d=decideFirstRun(job,"p","f",l,r,false,conflictManual)
 if d.Direction!="primary"{t.Fatal("explicit file override must beat newest timestamp")}
 d=decideFirstRun(job,"p","f",l,r,true,conflictManual)
 if d.Direction!=""{t.Fatal("one-way destination must not be automatically overwritten")}
 nas:=&Job{Provider:"nas_bridge",ConflictPolicy:conflictManual}
 d=decideFirstRun(nas,"p","f",l,r,false,conflictManual)
 if d.Direction!=""{t.Fatal("manual NAS policy must not be overridden")}
 nas.ConflictPolicy=conflictNewer
 d=decideFirstRun(nas,"p","f",l,r,false,conflictNewer)
 if d.Direction!="secondary"{t.Fatal("newer NAS policy should choose newer side")}
}
