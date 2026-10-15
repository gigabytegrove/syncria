package main

import (
 "context"
 "errors"
 "fmt"
 "io"
 "strings"
 "testing"
 "time"
)

type noOpenRecoveryEndpoint struct {
 *recoveryFake
 opens int
}
func (e *noOpenRecoveryEndpoint) Open(ctx context.Context,rel string)(io.ReadCloser,int64,error){
 e.opens++
 return nil,0,errors.New("Unexpected file-body read for unchanged "+rel)
}

func TestTrackedUnchangedFilesNeverOpenFileBodies(t *testing.T){
 name:="Personal Data/Kaleb/Kaleb Tablet Backup/Bedtime Music/Holiday/Kaleb_Xmas1.mp4"
 key:="google_drive:account:root|nas|."
 data:="same-existing-video-metadata"
 job:=Job{ID:"unchanged-no-open",Provider:"bridge",Direction:"one_way",PairStates:map[string]map[string]Item{key:{name:fileState(name,data,data)}}}
 app:=policyTestApp(t,job)
 primary:=&noOpenRecoveryEndpoint{recoveryFake:newRecoveryFake(map[string]string{name:data})}
 secondary:=&noOpenRecoveryEndpoint{recoveryFake:newRecoveryFake(map[string]string{name:data})}
 if err:=app.executePolicyPair(context.Background(),&job,primary,secondary,key,true,conflictManual);err!=nil{t.Fatal(err)}
 if primary.opens!=0||secondary.opens!=0{t.Fatalf("Unchanged file was opened: source=%d destination=%d",primary.opens,secondary.opens)}
 if job.LastUploads!=0||job.LastDownloads!=0{t.Fatal("Unchanged file was retransferred")}
 if a:=app.getSyncProgress(job.ID);len(a.Plan)!=0 || len(a.PlanCounts)!=0{
  t.Fatalf("Live sync must not rebuild a speculative transfer plan: %+v",a)
 }
}

func TestPreviewListsAll89SkipsAnd64ReplacementsAfter5800Unchanged(t *testing.T) {
 stamp:=time.Date(2026,10,14,19,45,0,0,time.UTC)
 job:=&Job{Provider:"bridge",Direction:"one_way",PairStates:map[string]map[string]Item{}}
 primary:=map[string]bridgeEntry{}
 secondary:=map[string]bridgeEntry{}
 for i:=0;i<5800;i++{
  name:=fmt.Sprintf("0000-unchanged/%05d.txt",i)
  primary[name]=bridgeEntry{Size:100,Modified:stamp}
  secondary[name]=bridgeEntry{Size:100,Modified:stamp}
 }
 for i:=0;i<89;i++{
  name:=fmt.Sprintf("9000-NAS-only/skip-%03d.txt",i)
  secondary[name]=bridgeEntry{Size:200,Modified:stamp}
 }
 for i:=0;i<64;i++{
  name:=fmt.Sprintf("9100-replacement/replace-%03d.mov",i)
  primary[name]=bridgeEntry{Size:300,Modified:stamp.Add(time.Hour)}
  secondary[name]=bridgeEntry{Size:300,Modified:stamp}
 }
 for _,changesOnly:=range []bool{false,true}{
  p:=syncPreviewResult{Rows:[]syncPreviewRow{},Counts:map[string]int{},ChangesOnly:changesOnly}
  planSyncPair(&p,job,"relationship",primary,secondary,true,conflictManual)
  if p.Counts["unchanged"]!=5800||p.Counts["skip"]!=89||p.Counts["replace"]!=64{
   t.Fatalf("counts incorrect: %+v",p.Counts)
  }
  if p.Truncated ||len(p.Rows)!=153{t.Fatalf("expected all actionable rows, got %d, truncated=%v",len(p.Rows),p.Truncated)}
  for _,row:=range p.Rows{
   if row.Action=="unchanged"{t.Fatalf("unchanged appeared in actionable plan: %s",row.Path)}
   if row.Reason==""||row.Path==""{t.Fatalf("missing path/reason: %+v",row)}
  }
 }
}
func TestFirstRunIgnoresSubsecondTimestampRoundingWithoutReading(t *testing.T){
 ctx:=context.Background()
 stamp:=time.Date(2026,10,14,19,45,0,0,time.UTC)
 a:=newRecoveryFake(nil);b:=newRecoveryFake(nil)
 same,err:=bridgeSameFirstRun(ctx,a,b,"30g-video.mov",bridgeEntry{Size:30*1024*1024*1024,Modified:stamp},bridgeEntry{Size:30*1024*1024*1024,Modified:stamp.Add(800*time.Millisecond)})
 if err!=nil||!same{t.Fatalf("Same size, subsecond timestamp rounding must skip: same=%v err=%v",same,err)}
}
func TestRetryOnlyHandlesTransientReadErrors(t *testing.T){
 if !retryableBridgeRead(fmt.Errorf("network: %w",context.DeadlineExceeded)){t.Fatal("deadline should be retryable if job context remains active")}
 if retryableBridgeRead(context.Canceled){t.Fatal("user cancellation must never retry")}
 if retryableBridgeRead(errors.New("file permission denied")){t.Fatal("permanent errors must not retry")}
}
func TestPreviewContainsFilterControlsAndDoesNotAdvertiseHashReads(t *testing.T){
 for _,want:=range []string{"Skipped and why","Overwrites and why","action!=='unchanged'","row.reason||'No reason provided'"}{
  if !strings.Contains(newUI,want){t.Fatalf("UI missing %q",want)}
 }
 if strings.Contains(newUI,"The real sync verifies content before replacing a file"){t.Fatal("Outdated preview copy still claims full-file comparison")}
}
