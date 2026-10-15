package main

import (
 "context"
 "net/http"
 "net/http/httptest"
 "path/filepath"
 "strings"
 "testing"
 "time"
)

func TestRecoveryEligibilityStopsExactlyAtTenMinuteDeadline(t *testing.T){
 now:=time.Date(2026,10,15,5,30,0,0,time.UTC)
 entries:=[]PendingDeletion{
  {Path:"Last night.mov",Expires:now.Add(-8*time.Hour)},
  {Path:"Exactly expired.txt",Expires:now},
  {Path:"Unknown expiry.dat"},
  {Path:"Still recoverable.mp4",Expires:now.Add(20*time.Second)},
 }
 got:=eligiblePendingDeletions(entries,now)
 if len(got)!=1||got[0].Path!="Still recoverable.mp4"{
  t.Fatalf("Only a strictly future recovery should remain: %+v",got)
 }
 if len(entries)!=4||entries[0].Path!="Last night.mov"{
  t.Fatal("The persistent cleanup queue must never be modified by presentation filtering")
 }
 if extra:=eligiblePendingDeletions(entries,now.Add(20*time.Second));len(extra)!=0{
  t.Fatalf("Restoration remained available at exact deadline: %+v",extra)
 }
}

func TestHomeNeverOffersToRestoreExpiredFilesEvenIfCleanupFailed(t *testing.T){
 dir:=t.TempDir()
 now:=time.Now()
 app:=&App{file:filepath.Join(dir,"state.json"),cfg:Config{Secret:b64rand()}}
 setup:=httptest.NewRequest(http.MethodPost,"http://localhost:9764/action",
  strings.NewReader("action=setup&password=strongpassphrase1234"))
 setup.Header.Set("Content-Type","application/x-www-form-urlencoded")
 session:=httptest.NewRecorder()
 app.action(session,setup)
 if session.Code!=http.StatusSeeOther{t.Fatalf("test login failed: %d %s",session.Code,session.Body.String())}
 cookies:=session.Result().Cookies()
 if len(cookies)==0{t.Fatal("No authenticated session")}
 expired:=PendingDeletion{Pair:"pair1",Path:"Family/Last Night/old.mov",MissingSide:"primary",
  RecoveryToken:".syncria-recovery/old-token",Expires:now.Add(-12*time.Hour)}
 active:=PendingDeletion{Pair:"pair2",Path:"Family/Recent/new.jpg",MissingSide:"primary",
  RecoveryToken:".syncria-recovery/active-token",Expires:now.Add(8*time.Minute)}
 app.cfg.Jobs=[]Job{
  {ID:"only-expired",Name:"Old cleanup pending",PendingDeletions:[]PendingDeletion{expired}},
  {ID:"mixed",Name:"Mixed recovery",PendingDeletions:[]PendingDeletion{expired,active}},
 }
 req:=httptest.NewRequest(http.MethodGet,"http://localhost:9764/",nil)
 req.AddCookie(cookies[0])
 page:=httptest.NewRecorder()
 app.home(page,req)
 if page.Code!=http.StatusOK{t.Fatalf("home returned %d: %s",page.Code,page.Body.String())}
 html:=page.Body.String()
 if strings.Contains(html,"Family/Last Night/old.mov")||
  strings.Contains(html,`data-recovery-summary="only-expired"`){
  t.Fatal("Expired record remained visibly restorable after the 10-minute window")
 }
 if !strings.Contains(html,"Family/Recent/new.jpg")||
  !strings.Contains(html,`data-recovery-summary="mixed"`)||
  !strings.Contains(html,`class="recovery-count">1</span>`){
  t.Fatal("Active recovery was removed or expired entries were counted")
 }
 if len(app.cfg.Jobs[0].PendingDeletions)!=1||len(app.cfg.Jobs[1].PendingDeletions)!=2{
  t.Fatal("Filtering for the page must not remove records needed for purge retry")
 }
 if err:=app.restorePendingDeletion(context.Background(),"only-expired","pair1",expired.Path);err==nil||
   !strings.Contains(err.Error(),"expired"){
  t.Fatalf("Backend must reject expired requests even if stale UI sends one: %v",err)
 }
}

func TestBrowserExpiryRemovesEntireRestoreFormAtDeadline(t *testing.T){
 for _,needle:=range []string{
  `data-recovery-expires-ms="{{.Expires.UnixMilli}}"`,
  `if(!Number.isFinite(deadline)||deadline<=now)`,
  `item.remove();`,
  `summary.hidden=total===0;`,
  `refreshRecoveryExpirations`,
  `visibilitychange`,
  `window.addEventListener('focus',refreshRecoveryExpirations)`,
 }{
  if !strings.Contains(newUI,needle){t.Errorf("Missing browser expiration safeguard: %s",needle)}
 }
 if !strings.Contains(syncriaDesignCSS,".recovery-dialog-item[hidden]{display:none!important}"){
  t.Fatal("CSS may resurrect filename-filtered Restore buttons")
 }
}
