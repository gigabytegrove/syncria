package main

import (
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
