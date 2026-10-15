package main

import (
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "os"
 "path/filepath"
 "strings"
 "testing"
)

func TestDismissOldFailurePreservesRunHistory(t *testing.T){
 dir:=t.TempDir()
 a:=&App{file:filepath.Join(dir,"state.json"),cfg:Config{Secret:b64rand()}}
 setup:=httptest.NewRequest(http.MethodPost,"http://localhost:9764/action",strings.NewReader("action=setup&password=strongpassphrase1234"))
 setup.Header.Set("Content-Type","application/x-www-form-urlencoded")
 sw:=httptest.NewRecorder()
 a.action(sw,setup)
 if sw.Code!=http.StatusSeeOther{t.Fatalf("setup failed: %d: %s",sw.Code,sw.Body.String())}
 cookies:=sw.Result().Cookies()
 if len(cookies)==0{t.Fatal("missing login session")}

 a.cfg.Jobs=[]Job{{ID:"job123",Name:"Brad's GD Sync",LastError:"read Kaleb_Xmas1.mp4: context deadline exceeded",FailedRuns:3,SuccessfulRuns:2,TotalUploads:58,LastSummary:"Previous sync failed"}}
 if err:=a.save();err!=nil{t.Fatal(err)}
 req:=httptest.NewRequest(http.MethodPost,"http://localhost:9764/action",strings.NewReader("action=dismiss_job_error&id=job123"))
 req.Header.Set("Content-Type","application/x-www-form-urlencoded")
 req.AddCookie(cookies[0])
 w:=httptest.NewRecorder()
 a.action(w,req)
 if w.Code!=http.StatusSeeOther{t.Fatalf("dismiss failed: %d %s",w.Code,w.Body.String())}
 if a.cfg.Jobs[0].LastError!=""{t.Fatal("dismiss did not clear visible failure")}
 if a.cfg.Jobs[0].FailedRuns!=3||a.cfg.Jobs[0].SuccessfulRuns!=2||a.cfg.Jobs[0].TotalUploads!=58{t.Fatal("dismiss reset historical run statistics")}
 if a.cfg.Jobs[0].LastSummary!="Previous sync failed"{t.Fatal("dismiss unexpectedly altered historical summary")}
 var stored Config
 b,err:=os.ReadFile(a.file);if err!=nil{t.Fatal(err)}
 if err=json.Unmarshal(b,&stored);err!=nil{t.Fatal(err)}
 if stored.Jobs[0].LastError!=""||stored.Jobs[0].FailedRuns!=3{t.Fatal("dismiss was not persisted safely")}
}
func TestDismissFailureRejectsAnonymousRequest(t *testing.T){
 a:=&App{cfg:Config{Secret:b64rand(),PasswordHash:"configured"},file:filepath.Join(t.TempDir(),"state.json")}
 req:=httptest.NewRequest(http.MethodPost,"http://localhost:9764/action",strings.NewReader("action=dismiss_job_error&id=job123"))
 req.Header.Set("Content-Type","application/x-www-form-urlencoded")
 w:=httptest.NewRecorder()
 a.action(w,req)
 if w.Code!=http.StatusUnauthorized{t.Fatalf("unauthorized failure dismissal: %d",w.Code)}
}
