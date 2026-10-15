package main

import (
 "context"
 "net/http"
 "path/filepath"
 "sync"
 "sync/atomic"
 "testing"
 "time"
)

func TestNASWatcherHealthyOnlyAfterVerifiedFullScan(t *testing.T){
 a:=&App{}
 a.setNASWatchExpected("one",1)
 a.watchConnected("one")
 if a.canTrustNASWatch("one"){t.Fatal("A new watcher does not establish an authoritative NAS baseline")}
 a.validatedNAS("one")
 if !a.canTrustNASWatch("one"){t.Fatal("Healthy watcher following full reconciliation must permit fast checks")}
 a.watchChanged("one")
 if a.canTrustNASWatch("one"){t.Fatal("External NAS event must force fresh scan")}
 a.validatedNAS("one")
 a.watchDisconnected("one")
 if a.canTrustNASWatch("one"){t.Fatal("Disconnected SMB notify invalidates trust")}
 a.watchConnected("one")
 if a.canTrustNASWatch("one"){t.Fatal("Reconnect without fresh scan is not enough to restore trust")}
 a.validatedNAS("one")
 a.watchStateMu.Lock();a.watchState["one"].fullAt=time.Now().Add(-nasFullVerificationInterval-time.Minute);a.watchStateMu.Unlock()
 if a.canTrustNASWatch("one"){t.Fatal("Expired NAS reconciliation must require full scan")}
}
func TestMultipleNASMonitorsMustAllBeConnected(t *testing.T){
 a:=&App{}
 a.setNASWatchExpected("many",2)
 a.watchConnected("many")
 a.validatedNAS("many")
 if a.canTrustNASWatch("many"){t.Fatal("One connected NAS watcher must not cover two destinations")}
 a.watchConnected("many");a.validatedNAS("many")
 if !a.canTrustNASWatch("many"){t.Fatal("All healthy watchers should be eligible")}
}
func TestDriveNoChangesFastCheckAvoidsNASListing(t *testing.T){
 idx,err:=loadSyncIndex(filepath.Join(t.TempDir(),"index.db"));if err!=nil{t.Fatal(err)}
 defer idx.Close()
 snap:=exampleDriveSnapshot()
 snap.FullScanAt=time.Now()
 if err=idx.saveDriveTree(snap);err!=nil{t.Fatal(err)}
 a:=&App{index:idx,cfg:Config{Accounts:[]Account{{ID:"account",AccessToken:"fake",Expires:time.Now().Add(time.Hour)}}}}
 a.client=&http.Client{Transport:driveTestTransport(func(r *http.Request)(*http.Response,error){
  if r.URL.Path!="/drive/v3/changes"{t.Errorf("Unexpected full Drive inventory request: %s",r.URL.Path)}
  return driveTestResponse(200,`{"newStartPageToken":"newcursor","changes":[]}`),nil
 })}
 a.setNASWatchExpected("relationship",1)
 a.watchConnected("relationship")
 a.validatedNAS("relationship")
 job:=Job{ID:"relationship",Provider:"bridge",CloudProvider:"google_drive",AccountID:"account",RemoteID:"root",Direction:"bidirectional",ScheduleMode:"changes"}
 empty,err:=a.noRemoteChanges(context.Background(),job)
 if err!=nil||!empty{t.Fatalf("Unchanged relationship must avoid recursive SMB scan: empty=%v err=%v",empty,err)}
 a.watchChanged(job.ID)
 empty,err=a.noRemoteChanges(context.Background(),job)
 if err!=nil||empty{t.Fatal("NAS change notification must disable fast skip")}
}
func TestSingleFlightLaunchReservationAcrossManyTriggers(t *testing.T){
 a:=&App{}
 var wg sync.WaitGroup
 var accepted int32
 for i:=0;i<200;i++{
  wg.Add(1)
  go func(){defer wg.Done();if a.reserveScheduledRun("same-sync"){atomic.AddInt32(&accepted,1)}}()
 }
 wg.Wait()
 if accepted!=1{t.Fatalf("Expected exactly one reserved launch; got %d",accepted)}
 a.queuedRuns.Delete("same-sync")
 if a.reserveScheduledRun("same-sync"){t.Fatal("Self-triggered rescan within 45s cooldown must not start")}
 a.lastAttempt.Store("same-sync",time.Now().Add(-time.Hour))
 if !a.reserveScheduledRun("same-sync"){t.Fatal("Later legitimate run must be accepted")}
 a.queuedRuns.Delete("same-sync")
 a.lastAttempt.Store("another",time.Now().Add(-time.Hour))
 a.runs.Store("another",true)
 if a.reserveScheduledRun("another"){t.Fatal("Already executing run must not be queued again")}
}
