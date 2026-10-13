package main

import (
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "os"
 "path/filepath"
 "strings"
 "testing"
 "time"
)

func TestV051UpdateTimelineIsDurableAndSpecific(t *testing.T){
 a:=&App{file:filepath.Join(t.TempDir(),"state.json")}
 if err:=a.beginUpdate("install");err!=nil{t.Fatal(err)}
 a.setUpdateProgress("manifest","Receiving SHA256SUMS","v0.5.1",-1,-1,-1,"")
 a.updateAsset("syncria-linux-amd64")
 a.setUpdateProgress("downloading","Downloading syncria-linux-amd64","v0.5.1",-1,2048,8192,"")
 a.setUpdateProgress("verifying","Calculating SHA-256 of syncria-linux-amd64","v0.5.1",-1,1024,8192,"")
 a.setUpdateProgress("staging","Saving rollback binary","v0.5.1",-1,-1,-1,"")
 p:=a.updateProgressSnapshot()
 if p.From!=appVersion||p.Target!="v0.5.1"||p.Asset!="syncria-linux-amd64"{
  t.Fatalf("release identifiers lost: %+v",p)
 }
 if p.Downloaded!=2048||p.Total!=8192||p.VerifiedBytes!=1024||p.VerifiedTotal!=8192{
  t.Fatalf("actual byte counters were conflated: %+v",p)
 }
 if len(p.Events)<4{t.Fatalf("expected durable stage events: %+v",p.Events)}
 b,err:=os.ReadFile(a.updateProgressPath());if err!=nil{t.Fatal(err)}
 var persisted UpdateProgress
 if err=json.Unmarshal(b,&persisted);err!=nil{t.Fatal(err)}
 if len(persisted.Events)!=len(p.Events)||persisted.Downloaded!=2048||persisted.VerifiedBytes!=1024{
  t.Fatal("Update log or progress did not survive file persistence")
 }
 if persisted.Percent!=-1 {t.Fatal("Unmeasured stages must not claim an arbitrary overall completion percentage")}
}

func TestV051RestartMustConfirmExactRelease(t *testing.T){
 file:=filepath.Join(t.TempDir(),"state.json")
 a:=&App{file:file}
 a.updateMu.Lock()
 err:=a.writeUpdateProgressLocked(UpdateProgress{Phase:"restarting",Operation:"install",Target:"v99.1.2",From:"v0.5.0",Active:true})
 a.updateMu.Unlock()
 if err!=nil{t.Fatal(err)}
 next:=&App{file:file}
 next.reconcileUpdateOnStartup()
 p:=next.updateProgressSnapshot()
 if p.Phase!="failed"||p.Active||p.Error==""{t.Fatalf("Mismatched running binary was reported as success: %+v",p)}
}

func TestV051ManualRestartConfirmedFromNewProcess(t *testing.T){
 file:=filepath.Join(t.TempDir(),"state.json")
 a:=&App{file:file}
 a.updateMu.Lock()
 err:=a.writeUpdateProgressLocked(UpdateProgress{Phase:"restarting",Operation:"manual",From:"v0.5.0",Active:true,StartedAt:time.Now()})
 a.updateMu.Unlock()
 if err!=nil{t.Fatal(err)}
 after:=&App{file:file}
 after.reconcileUpdateOnStartup()
 p:=after.updateProgressSnapshot()
 if p.Phase!="complete"||p.Active||p.Percent!=100{
  t.Fatalf("Manual in-app restart was not confirmed: %+v",p)
 }
}

func TestV051UpdateUIExposesRealEvidenceAndNeverInventsOverallProgress(t *testing.T){
 for _,required:=range []string{
  "updateEventTimeline","updateFromVersion","updateToVersion","updateProgressAsset",
  "updateVerifiedBytes","updateElapsed","updateConnectionStatus",
  "data-update-phase=\"manifest\"","renderUpdateEvents(","elapsed_seconds",
  "setUpdateProgressBar(","update-live-banner","/api/update/progress",
 } {
  if !strings.Contains(newUI,required){t.Errorf("Missing measurable updater detail: %s",required)}
 }
 for _,obsolete:=range []string{"Update staged. Syncria is restarting; refresh shortly.","percent+=int(downloaded*60/total)"}{
  if strings.Contains(newUI,obsolete){t.Errorf("Obsolete ambiguous status or fake percent remains in UI: %s",obsolete)}
 }
}

func TestV051StatusAPINotAccessibleWithoutAuthentication(t *testing.T){
 a:=&App{file:filepath.Join(t.TempDir(),"state.json")}
 rec:=httptest.NewRecorder()
 a.updateProgressAPI(rec,httptest.NewRequest(http.MethodGet,"/api/update/progress",nil))
 if rec.Code!=http.StatusUnauthorized{t.Fatalf("Status API leaked without session: %d",rec.Code)}
}
