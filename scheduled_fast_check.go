package main

import (
 "context"
 "log"
 "strings"
 "time"
)

const nasFullVerificationInterval=45*time.Minute
const changeRunCooldown=45*time.Second

// A NAS change feed is useful only while all expected SMB notifications have
// stayed connected since a completed full inventory. On startup/disconnection
// we fail closed: no delta-only shortcut, so external NAS edits are discoverable.
type nasWatchCoverage struct {
 expected int
 connected int
 verified bool
 dirty bool
 fullAt time.Time
}
func (a *App) setNASWatchExpected(id string,n int){
 a.watchStateMu.Lock();defer a.watchStateMu.Unlock()
 if a.watchState==nil{a.watchState=map[string]*nasWatchCoverage{}}
 a.watchState[id]=&nasWatchCoverage{expected:n}
}
func (a *App) watchConnected(id string){
 a.watchStateMu.Lock();defer a.watchStateMu.Unlock()
 w:=a.watchState[id];if w!=nil&&w.connected<w.expected{w.connected++}
}
func (a *App) watchDisconnected(id string){
 a.watchStateMu.Lock();defer a.watchStateMu.Unlock()
 w:=a.watchState[id]
 if w!=nil{if w.connected>0{w.connected--};w.verified=false;w.dirty=true}
}
func (a *App) watchChanged(id string){
 a.watchStateMu.Lock();defer a.watchStateMu.Unlock()
 if w:=a.watchState[id];w!=nil{w.dirty=true}
}
func (a *App) validatedNAS(id string){
 a.watchStateMu.Lock();defer a.watchStateMu.Unlock()
 w:=a.watchState[id]
 if w!=nil&&w.expected>0&&w.connected==w.expected {
  w.verified=true;w.dirty=false;w.fullAt=time.Now()
 }
}
func (a *App) canTrustNASWatch(id string)bool{
 a.watchStateMu.Lock();defer a.watchStateMu.Unlock()
 w:=a.watchState[id]
 return w!=nil&&w.verified&&!w.dirty&&w.expected>0&&w.connected==w.expected&&time.Since(w.fullAt)<nasFullVerificationInterval
}

// For a scheduled (not user-requested) relationship run, check Drive's
// persisted change cursor without advancing it. If both cloud and continuously
// monitored NAS are unchanged, skip expensive recursive SMB listing altogether.
// Every 45 minutes do a full reconciliation regardless, to detect missed events.
func (a *App) noRemoteChanges(ctx context.Context,j Job)(bool,error){
 if j.ScheduleMode!="changes"||j.Provider!="bridge"||j.CloudProvider=="dropbox"||j.Direction==""{
  return false,nil
 }
 if !a.canTrustNASWatch(j.ID)||a.index==nil{return false,nil}
 snap,ok,err:=a.index.getDriveTree(j.AccountID,j.RemoteID)
 if err!=nil{return false,err}
 if !ok||time.Since(snap.FullScanAt)>=24*time.Hour{return false,nil}
 changes,_,err:=a.driveChanges(ctx,j.AccountID,snap.Cursor)
 if err!=nil{return false,err} // never interpret inability to check as no changes
 // Even unrelated Drive changes are conservatively resolved through a run.
 return len(changes)==0,nil
}

// Prevent scheduler and NAS watchers from creating multiple pending goroutines
// for the same relationship. syncJob provides a second, atomic running guard
// for manual starts and other call sites.
func (a *App) reserveScheduledRun(id string)bool{
 if _,busy:=a.runs.Load(id);busy{return false}
 if _,queued:=a.queuedRuns.LoadOrStore(id,true);queued{return false}
 now:=time.Now()
 if last,ok:=a.lastAttempt.Load(id);ok&&now.Sub(last.(time.Time))<changeRunCooldown{
  a.queuedRuns.Delete(id)
  return false
 }
 a.lastAttempt.Store(id,now)
 return true
}
func (a *App) launchScheduledRun(ctx context.Context,id string,fromWatcher bool){
 if ctx.Err()!=nil||!a.reserveScheduledRun(id){return}
 go func(){
  defer a.queuedRuns.Delete(id)
  a.mu.Lock()
  job:=a.job(id)
  if job==nil||!job.Enabled {a.mu.Unlock();return}
  candidate:=*job
  a.mu.Unlock()
  if _,busy:=a.runs.Load(id);busy{return}
  // Changes caused by Syncria's own just-completed transfer can arrive
  // after the run. Do not start a fresh complete scan immediately.
  if fromWatcher&&!candidate.LastRun.IsZero()&&time.Since(candidate.LastRun)<changeRunCooldown{return}
  if !fromWatcher {
   unchanged,err:=a.noRemoteChanges(ctx,candidate)
   if err!=nil{
    // A failed cloud probe must NOT be interpreted as an unchanged sync.
    // Fall back to the normal full sync, which records a real error if needed.
    log.Printf("Syncria quick change check for %s unavailable: %v; using normal reconciliation",id,err)
   }else if unchanged{
    a.mu.Lock()
    if actual:=a.job(id);actual!=nil&&actual.Enabled {
     actual.LastChecked=time.Now().UTC()
     if saveErr:=a.save();saveErr!=nil{log.Printf("Syncria could not record fast check for %s: %v",id,saveErr)}
    }
    a.mu.Unlock()
    return
   }
  }
  if err:=a.syncJob(ctx,id);err!=nil&&!strings.Contains(err.Error(),"job already running"){
   log.Printf("Syncria scheduled relationship %s: %v",id,err)
  }
 }()
}
