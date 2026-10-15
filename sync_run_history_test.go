package main

import (
 "strings"
 "fmt"
 "net/http"
 "net/http/httptest"
 "path/filepath"
 "testing"
 "time"
)

func TestDurableRunHistoryHasEveryFileAfterRestart(t *testing.T){
 db:=filepath.Join(t.TempDir(),"syncria-index.db")
 idx,err:=loadSyncIndex(db);if err!=nil{t.Fatal(err)}
 id,err:=idx.beginRun("job");if err!=nil{t.Fatal(err)}
 for i:=0;i<350;i++{
  event:=SyncActionEvent{At:time.Now(),Action:"copy",Path:fmt.Sprintf("Family/Video%03d.mov",i),From:"Google Drive",To:"NAS"}
  if err=idx.appendRunEvent("job",id,event);err!=nil{t.Fatal(err)}
 }
 j:=Job{LastUploads:350,LastUploadBytes:123456}
 if err=idx.finishRun("job",id,"success","350 files copied","",j);err!=nil{t.Fatal(err)}
 if err=idx.Close();err!=nil{t.Fatal(err)}
 idx,err=loadSyncIndex(db);if err!=nil{t.Fatal(err)}
 defer idx.Close()
 list,err:=idx.listRuns("job",0,25)
 if err!=nil||len(list.Runs)!=1||list.Runs[0].EventCount!=350||list.Runs[0].Uploads!=350{t.Fatalf("Run metadata missing: %+v, %v",list,err)}
 first,err:=idx.readRun("job",id,0,100)
 if err!=nil||len(first.Events)!=100||!first.More{t.Fatalf("First page incorrect: %+v, %v",first,err)}
 final,err:=idx.readRun("job",id,300,100)
 if err!=nil||len(final.Events)!=50||final.More{t.Fatalf("Final page incorrect: %+v, %v",final,err)}
 if final.Events[49].Path!="Family/Video349.mov"{t.Fatal("Journal lost last transferred file")}
}
func TestRunHistorySupportsSeveralRunsAndStatuses(t *testing.T){
 idx,err:=loadSyncIndex(filepath.Join(t.TempDir(),"index.db"));if err!=nil{t.Fatal(err)}
 defer idx.Close()
 for i:=0;i<4;i++{
  id,err:=idx.beginRun("sync");if err!=nil{t.Fatal(err)}
  status:="success";if i==2{status="failed"}
  errText:="";if status=="failed"{errText="NAS read timed out"}
  if err=idx.finishRun("sync",id,status,"test",errText,Job{});err!=nil{t.Fatal(err)}
 }
 page,err:=idx.listRuns("sync",1,2)
 if err!=nil||len(page.Runs)!=2||!page.More{t.Fatalf("History pagination wrong: %+v %v",page,err)}
 for _,run:=range page.Runs{if run.Status==""||run.Ended.IsZero(){t.Fatal("Run status/time not stored")}}
}
func TestRunHistoryIsAuthorizedAndScopedToJob(t *testing.T){
 idx,err:=loadSyncIndex(filepath.Join(t.TempDir(),"index.db"));if err!=nil{t.Fatal(err)}
 defer idx.Close()
 a:=&App{index:idx,cfg:Config{Secret:b64rand(),PasswordHash:"protected",Jobs:[]Job{{ID:"known"}}}}
 req:=httptest.NewRequest(http.MethodGet,"/api/sync/history?id=known",nil)
 w:=httptest.NewRecorder()
 a.syncHistoryAPI(w,req)
 if w.Code!=http.StatusUnauthorized{t.Fatalf("Unauthenticated access to history: %d",w.Code)}
}

func TestHistorySearchFiltersEntirePersistedRunNotOnlyCurrentPage(t *testing.T){
 idx,err:=loadSyncIndex(filepath.Join(t.TempDir(),"history-search.db"));if err!=nil{t.Fatal(err)}
 defer idx.Close()
 run,err:=idx.beginRun("sync-search");if err!=nil{t.Fatal(err)}
 for i:=0;i<420;i++{
  name:=fmt.Sprintf("Bulk/%03d-unchanged-name.txt",i)
  action:="copy"
  if i==318{name="Family/Holiday/Kaleb_Xmas1.mp4";action="transfer failed"}
  if i==400{name="Family/Holiday/Another.mp4";action="replace"}
  if err=idx.appendRunEvent("sync-search",run,SyncActionEvent{Path:name,Action:action,At:time.Now()});err!=nil{t.Fatal(err)}
 }
 if err=idx.finishRun("sync-search",run,"failed","one failed file","timeout",Job{});err!=nil{t.Fatal(err)}
 results,err:=idx.readRunFiltered("sync-search",run,0,100,"kaleb","transfer failed")
 if err!=nil||len(results.Events)!=1||results.Events[0].Path!="Family/Holiday/Kaleb_Xmas1.mp4"{
  t.Fatalf("Search must find files after the first 100 rows: result=%+v err=%v",results,err)
 }
 matches,err:=idx.readRunFiltered("sync-search",run,0,100,"Holiday","")
 if err!=nil||len(matches.Events)!=2||matches.More{t.Fatalf("Folder search should find both entries across full history: %+v %v",matches,err)}
 no,err:=idx.readRunFiltered("sync-search",run,0,100,"nonexistent","")
 if err!=nil||len(no.Events)!=0{t.Fatalf("No-match search incorrect: %+v %v",no,err)}
 if _,err=idx.readRunFiltered("sync-search",run,0,100,strings.Repeat("a",257),"");err==nil{t.Fatal("Search length guard missing")}
}
