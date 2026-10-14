package main

import (
 "encoding/json"
 "os"
 "path/filepath"
 "strings"
 "testing"
 "time"
)

func TestCompletedUpdateDoesNotStayInPersistentState(t *testing.T){
 a:=&App{file:filepath.Join(t.TempDir(),"state.json")}
 if err:=a.beginUpdate("install");err!=nil{t.Fatal(err)}
 a.setUpdateProgress("verifying","Checked bytes","v0.5.3",-1,100,100,"")
 a.finishUpdate("complete","The latest version is already installed","")
 if _,err:=os.Stat(a.updateProgressPath());!os.IsNotExist(err){
  t.Fatalf("a successful update must not remain on disk: %v",err)
 }
 p:=a.updateProgressSnapshot()
 if p.Phase!="complete"||p.Active{t.Fatalf("brief success confirmation was lost: %+v",p)}
 a.updateMu.Lock()
 a.updateProgress.UpdatedAt=time.Now().Add(-completedUpdateGrace-time.Second)
 a.updateMu.Unlock()
 if p=a.updateProgressSnapshot();p.Phase!="idle"{t.Fatalf("completed update remained on screen: %+v",p)}
}

func TestConfirmedRestartClearsStatusFile(t *testing.T){
 file:=filepath.Join(t.TempDir(),"state.json")
 a:=&App{file:file}
 a.updateMu.Lock()
 err:=a.writeUpdateProgressLocked(UpdateProgress{Phase:"restarting",Operation:"install",From:"v0.5.2",Target:appVersion,Active:true,StartedAt:time.Now()})
 a.updateMu.Unlock()
 if err!=nil{t.Fatal(err)}
 next:=&App{file:file}
 next.reconcileUpdateOnStartup()
 if p:=next.updateProgressSnapshot();p.Phase!="complete"||p.Active{t.Fatalf("restart was not verified: %+v",p)}
 if _,err:=os.Stat(next.updateProgressPath());!os.IsNotExist(err){
  t.Fatalf("successful restart left permanent status: %v",err)
 }
}

func TestLegacyCompletedStatusRemovedOnRestart(t *testing.T){
 file:=filepath.Join(t.TempDir(),"state.json")
 a:=&App{file:file}
 a.updateMu.Lock()
 err:=a.writeUpdateProgressLocked(UpdateProgress{Phase:"complete",Active:false,Target:"v0.5.2"})
 a.updateMu.Unlock()
 if err!=nil{t.Fatal(err)}
 next:=&App{file:file}
 next.reconcileUpdateOnStartup()
 if p:=next.updateProgressSnapshot();p.Phase!="idle"{t.Fatalf("old completed state returned: %+v",p)}
 if _,err:=os.Stat(next.updateProgressPath());!os.IsNotExist(err){t.Fatalf("old completed status remains: %v",err)}
}

func TestFailedUpdatesRemainVisible(t *testing.T){
 file:=filepath.Join(t.TempDir(),"state.json")
 a:=&App{file:file}
 if err:=a.beginUpdate("install");err!=nil{t.Fatal(err)}
 a.finishUpdate("failed","Could not check checksum","File hash mismatch")
 b,err:=os.ReadFile(a.updateProgressPath());if err!=nil{t.Fatal(err)}
 var state UpdateProgress
 if err=json.Unmarshal(b,&state);err!=nil{t.Fatal(err)}
 if state.Phase!="failed"||state.Error==""{t.Fatalf("failure details not persisted: %+v",state)}
 next:=&App{file:file}
 next.reconcileUpdateOnStartup()
 if p:=next.updateProgressSnapshot();p.Phase!="failed"||p.Error==""{t.Fatalf("error disappeared after restart: %+v",p)}
}

func TestPolishedUIKeepsAllFunctionality(t *testing.T){
 elements:=[]string{
  `id="view-overview"`,`id="view-sync"`,`id="view-storage"`,`id="view-accounts"`,
  `id="view-activity"`,`id="view-updates"`,`id="view-settings"`,
  `id="syncWizard"`,`id="nasTargets"`,`id="conflictPolicy"`,`id="scheduleMode"`,
  `id="propagateDeletions"`,`id="recoveryEnabled"`,
  `id="updateFlow"`,`id="updateHistoryPanel"`,`id="updateReadyPanel"`,
  `id="updateEventTimeline"`,`id="updateVerifiedBytes"`,`id="updateProgressBar"`,
  `/api/recovery/undo`,`/api/conflicts/resolve`,
 }
 for _,term:=range elements{if !strings.Contains(newUI,term){t.Errorf("Missing UI element: %s",term)}}
 for _,term:=range []string{"completedUpdateTimer","resetUpdateView()","hiddenCompletedUpdate","phase==='idle'","progressBox.classList.toggle('hidden',!running)"}{
  if !strings.Contains(newUI,term){t.Errorf("Missing clean completed/idle handling: %s",term)}
 }
 for _,term:=range []string{".sync-library",".onboarding-disclosure",".update-ready",".update-explainer",".relationship-card",".quick-start-grid"}{
  if !strings.Contains(v051Theme,term){t.Errorf("Incomplete unified UI styling: %s",term)}
 }
}
