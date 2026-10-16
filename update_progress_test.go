package main

import (
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "os"
 "path/filepath"
 "runtime"
 "strings"
 "testing"
)

func TestUpdateProgressRequiresAuthentication(t *testing.T){
 a:=&App{file:filepath.Join(t.TempDir(),"config.json")}
 for _,method:=range []string{http.MethodGet,http.MethodPost}{
  req:=httptest.NewRequest(method,"http://localhost/api/update/progress",strings.NewReader("operation=install"))
  if method==http.MethodPost{req.Header.Set("Content-Type","application/x-www-form-urlencoded")}
  recorder:=httptest.NewRecorder()
  a.updateProgressAPI(recorder,req)
  if recorder.Code!=http.StatusUnauthorized{t.Errorf("%s returned HTTP %d, want 401",method,recorder.Code)}
 }
}

func TestUpdateProgressPersistedAndReconciled(t *testing.T){
 if runtime.GOOS=="windows"{t.Skip("Windows updater is manual")}
 path:=filepath.Join(t.TempDir(),"config.json")
 a:=&App{file:path}
 if err:=a.beginUpdate("install");err!=nil{t.Fatal(err)}
 a.setUpdateProgress("downloading","Downloading a release","v0.4.1",35,1024,4096,"")
 p:=a.updateProgressSnapshot()
 if p.Phase!="downloading"||p.Downloaded!=1024||p.Total!=4096||!p.Active{t.Fatalf("bad progress: %+v",p)}
 if _,err:=os.Stat(a.updateProgressPath());err!=nil{t.Fatalf("missing durable state: %v",err)}
 if err:=a.beginUpdate("install");err==nil{t.Fatal("parallel update accepted")}
 a.setUpdateProgress("restarting","Restarting",appVersion,97,-1,-1,"")
 fresh:=&App{file:path}
 fresh.reconcileUpdateOnStartup()
 p=fresh.updateProgressSnapshot()
 if p.Phase!="complete"||p.Percent!=100||p.Active{t.Fatalf("new process did not confirm the upgrade: %+v",p)}
}

func TestUpdateProgressIncorrectVersionNeverClaimsSuccess(t *testing.T){
 a:=&App{file:filepath.Join(t.TempDir(),"config.json")}
 a.updateMu.Lock()
 err:=a.writeUpdateProgressLocked(UpdateProgress{Phase:"restarting",Operation:"install",From:"v0.0.1",Target:"v99.99.99",Active:true,Percent:97})
 a.updateMu.Unlock()
 if err!=nil{t.Fatal(err)}
 fresh:=&App{file:a.file}
 fresh.reconcileUpdateOnStartup()
 p:=fresh.updateProgressSnapshot()
 if p.Phase!="failed"||p.Active{t.Fatalf("unexpected success: %+v",p)}
}

func TestSchedulerModeJsonBackwardCompatibility(t *testing.T){
 legacy:=[]byte(`{"id":"legacy","interval":10,"enabled":true}`)
 var job Job
 if err:=json.Unmarshal(legacy,&job);err!=nil{t.Fatal(err)}
 if job.ScheduleMode!=""||job.Interval!=10{t.Fatalf("legacy schedule changed: %+v",job)}
}

func TestSyncUIShowsProgressAndFourWizardSteps(t *testing.T){
 for _,part:=range []string{"id=\"wizard1\"","id=\"wizard2\"","id=\"wizard3\"","id=\"wizard4\"","id=\"scheduleMode\"","data-job-progress=","/api/update/progress","id=\"updateProgressBar\""}{
  if !strings.Contains(newUI,part){t.Errorf("missing %s",part)}
 }
}

func TestChannelPrefixedReleaseRestartConfirmsInstalledBinary(t *testing.T){
 if runtime.GOOS=="windows"{t.Skip("in-app updater runs inside Linux")}
 tag:=computerAgentVersionTag()
 if got:=releaseTagAppVersion(tag);got!=appVersion{
  t.Fatalf("release channel prefix was not normalized: %q vs %q",got,appVersion)
 }
 if got:=releaseTagAppVersion("beta-v0.8.6-beta.2");got!="v0.8.6-beta.2"{
  t.Fatalf("beta channel release tag was not normalized: %q",got)
 }
 if got:=releaseTagAppVersion("v0.8.5");got!="v0.8.5"{
  t.Fatalf("stable tag changed: %q",got)
 }
 if got:=releaseTagAppVersion("alpha-v0.8.6");got!=""{
  t.Fatalf("malformed tag accepted: %q",got)
 }
 app:=&App{file:filepath.Join(t.TempDir(),"state.json")}
 p:=UpdateProgress{Phase:"restarting",Operation:"install",From:"v0.8.5",Target:tag,Active:true,Percent:97}
 app.updateMu.Lock()
 err:=app.writeUpdateProgressLocked(p)
 app.updateMu.Unlock()
 if err!=nil{t.Fatal(err)}
 app.reconcileUpdateOnStartup()
 state:=app.updateProgressSnapshot()
 if state.Phase!="complete"||state.Active||state.Percent!=100{
  t.Fatalf("valid prerelease restart was reported failed instead of completed: %+v",state)
 }
 if _,err:=os.Stat(app.updateProgressPath());!os.IsNotExist(err){
  t.Fatalf("completed prerelease update left stale status file: %v",err)
 }
 if !strings.Contains(newUI,"runtimeVersionOfRelease(expectedVersion)")||
  !strings.Contains(newUI,"runtimeVersionOfRelease(p.target)") {
  t.Fatal("browser version confirmation still compares channel-prefixed tags to executable versions")
 }
}
