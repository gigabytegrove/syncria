package main

import (
 "context"
 "strings"
 "testing"
 "time"
)

func TestStopRunningJobCancelsOnlyActiveRun(t *testing.T) {
 app:=&App{syncProgress:map[string]SyncProgress{}}
 ctx,cancel:=context.WithCancel(context.Background())
 defer cancel()
 app.runs.Store("active",true)
 app.runCancels.Store("active",cancel)
 if !app.stopRunningJob("active"){t.Fatal("active job should accept cancellation")}
 if ctx.Err()!=context.Canceled{t.Fatalf("active context not canceled: %v",ctx.Err())}
 if _,ok:=app.stopRequests.Load("active");!ok{t.Fatal("explicit stop flag missing")}
 if got:=app.getSyncProgress("active").Phase;got!="stopping"{t.Fatalf("phase=%q",got)}
 if app.stopRunningJob("not-running"){t.Fatal("non-running job cannot be stopped")}
}
func TestActiveRunPlanContainsOnlyPendingChanges(t *testing.T) {
 stamp:=time.Date(2026,10,14,12,14,0,0,time.UTC)
 j:=&Job{Provider:"bridge",Direction:"one_way",PairStates:map[string]map[string]Item{}}
 l:=map[string]bridgeEntry{
  "unchanged.mov":{Size:30*1024*1024*1024,Modified:stamp},
  "new-photo.jpg":{Size:25,Modified:stamp},
  "changed.txt":{Size:120,Modified:stamp.Add(time.Hour)},
 }
 r:=map[string]bridgeEntry{
  "unchanged.mov":{Size:30*1024*1024*1024,Modified:stamp},
  "changed.txt":{Size:100,Modified:stamp},
 }
 p:=syncPreviewResult{Counts:map[string]int{},Rows:[]syncPreviewRow{},ChangesOnly:true,MaxRows:120}
 planSyncPair(&p,j,"drive|nas",l,r,true,conflictManual)
 if p.Counts["unchanged"]!=1 || len(p.Rows)!=2 {t.Fatalf("plan must count unchanged but only list work: %+v",p)}
 for _,row:=range p.Rows{if row.Action=="unchanged"||row.Action=="skip"{t.Fatal("unchanged entry leaked into actionable queue")}}
}
func TestActivityHasActiveQueueAndStopControls(t *testing.T) {
 for _,part:=range []string{
  `id="activeSyncList"`,`data-job-cancel="{{.ID}}"`,`requestStopSync(`,
  `renderActiveSyncs(jobs)`,`items discovered`,
 }{
  if !strings.Contains(newUI,part){t.Errorf("Missing live activity feature %s",part)}
 }
}
