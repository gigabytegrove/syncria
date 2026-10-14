package main

import (
 "context"
 "encoding/json"
 "errors"
 "fmt"
 "log"
 "net/http"
 "os"
 "path/filepath"
 "runtime"
 "strings"
 "time"
)

// UpdateProgress is stored in the existing Syncria /data volume. It is
// deliberately independent of the web browser and the container lifecycle.
type UpdateEvent struct {
 At time.Time `json:"at"`
 Phase string `json:"phase"`
 Detail string `json:"detail"`
}
// All update evidence is persisted alongside the binary, not in the browser.
type UpdateProgress struct {
 Asset string `json:"asset,omitempty"`
 VerifiedBytes int64 `json:"verified_bytes,omitempty"`
 VerifiedTotal int64 `json:"verified_total,omitempty"`
 Events []UpdateEvent `json:"events,omitempty"`
 Phase string `json:"phase"`
 Operation string `json:"operation,omitempty"`
 From string `json:"from,omitempty"`
 Target string `json:"target,omitempty"`
 Message string `json:"message,omitempty"`
 Error string `json:"error,omitempty"`
 Percent int `json:"percent"`
 Downloaded int64 `json:"downloaded"`
 Total int64 `json:"total"`
 Active bool `json:"active"`
 StartedAt time.Time `json:"started_at,omitempty"`
 UpdatedAt time.Time `json:"updated_at,omitempty"`
}

func (a *App) updateProgressPath() string {
 return filepath.Join(filepath.Dir(a.file),"runtime","update-status.json")
}
func (a *App) writeUpdateProgressLocked(p UpdateProgress) error {
 path:=a.updateProgressPath()
 if err:=os.MkdirAll(filepath.Dir(path),0700);err!=nil{return err}
 p.UpdatedAt=time.Now().UTC()
 b,err:=json.MarshalIndent(p,"","  ");if err!=nil{return err}
 f,err:=os.CreateTemp(filepath.Dir(path),".update-status-");if err!=nil{return err}
 temp:=f.Name();defer os.Remove(temp)
 if err=f.Chmod(0600);err!=nil{f.Close();return err}
 if _,err=f.Write(b);err!=nil{f.Close();return err}
 if err=f.Sync();err!=nil{f.Close();return err}
 if err=f.Close();err!=nil{return err}
 if err=os.Rename(temp,path);err!=nil{return err}
 a.updateProgress=p
 return nil
}
func (a *App) setUpdateProgress(phase,message,target string,percent int,downloaded,total int64,detail string) {
 a.updateMu.Lock()
 defer a.updateMu.Unlock()
 p:=a.updateProgress
 oldPhase,oldMessage:=p.Phase,p.Message
 p.Phase=phase;p.Message=message
 if oldPhase!=phase || (oldMessage!=message && phase!="downloading" && phase!="verifying") {
  p.Events=append(p.Events,UpdateEvent{At:time.Now().UTC(),Phase:phase,Detail:message})
  if len(p.Events)>45 {p.Events=p.Events[len(p.Events)-45:]}
 }
 if target!=""{p.Target=target}
 if percent>=0{p.Percent=percent}
 if phase=="verifying"{
  if downloaded>=0 {p.VerifiedBytes=downloaded}
  if total>=0 {p.VerifiedTotal=total}
 }else if phase=="downloading"{
  if downloaded>=0{p.Downloaded=downloaded}
  if total>=0{p.Total=total}
 }
 p.Error=detail
 p.Active=phase!="complete"&&phase!="failed"&&phase!="idle"
 if err:=a.writeUpdateProgressLocked(p);err!=nil{log.Printf("cannot save Syncria update progress: %v",err)}
}
func (a *App) beginUpdate(operation string) error {
 a.updateMu.Lock()
 defer a.updateMu.Unlock()
 if a.updateBusy{return errors.New("an update is already in progress")}
 if runtime.GOOS=="windows"{return errors.New("in-app executable replacement is unavailable on Windows")}
 busy:=false
 a.runs.Range(func(_, _ any)bool{busy=true;return false})
 if busy{return errors.New("an active sync must finish before updating")}
 now:=time.Now().UTC()
 p:=UpdateProgress{Phase:"checking",Operation:operation,From:appVersion,Message:"Checking release metadata",Percent:-1,Active:true,StartedAt:now,
 Events:[]UpdateEvent{{At:now,Phase:"checking",Detail:"Checking release metadata from GitHub Releases"}}}
 if operation=="rollback"{
  p.Phase="staging";p.Message="Restoring the saved previous executable";p.Percent=-1
  p.Events[0].Phase=p.Phase;p.Events[0].Detail=p.Message
 }
 if operation=="manual"{
  p.Phase="receiving";p.Message="Receiving the selected executable and SHA256SUMS";p.Percent=-1
  p.Events[0].Phase=p.Phase;p.Events[0].Detail=p.Message
 }
 if err:=a.writeUpdateProgressLocked(p);err!=nil{return fmt.Errorf("cannot persist update progress: %w",err)}
 a.updateBusy=true
 return nil
}
// An update must survive a process restart while it is running.
// Successful updates are NOT history entries: after a brief, in-memory
// confirmation the Updates page returns to its normal ready state.
// Failures remain on disk so people can see what went wrong.
const completedUpdateGrace = 20 * time.Second

func (a *App) clearSuccessfulUpdateLocked() {
 if a.updateProgress.Phase!="complete" {return}
 if err:=os.Remove(a.updateProgressPath());err!=nil&&!os.IsNotExist(err){
  log.Printf("cannot clear completed Syncria update state: %v",err)
 }
}
func (a *App) finishUpdate(phase,message,detail string) {
 a.setUpdateProgress(phase,message,"",func()int{if phase=="complete"{return 100};return -1}(),-1,-1,detail)
 if detail!=""{a.updateEvent(phase,detail)}
 a.updateMu.Lock()
 if phase=="complete"{a.clearSuccessfulUpdateLocked()}
 a.updateBusy=false
 a.updateMu.Unlock()
}
func (a *App) updateProgressSnapshot() UpdateProgress {
 a.updateMu.Lock()
 defer a.updateMu.Unlock()
 if a.updateProgress.Phase==""{
  b,err:=os.ReadFile(a.updateProgressPath())
  if err==nil{_ =json.Unmarshal(b,&a.updateProgress)}
 }
 if a.updateProgress.Phase=="complete"{
  a.clearSuccessfulUpdateLocked()
  if a.updateProgress.UpdatedAt.IsZero()||time.Since(a.updateProgress.UpdatedAt)>=completedUpdateGrace{
   a.updateProgress=UpdateProgress{}
  }
 }
 if a.updateProgress.Phase==""{return UpdateProgress{Phase:"idle",Message:"Syncria is ready"}}
 return a.updateProgress
}

// The new executable confirms the restart. At that point the status file is
// removed: completed updates never remain in persistent user-facing history.
func (a *App) reconcileUpdateOnStartup() {
 a.updateMu.Lock()
 defer a.updateMu.Unlock()
 b,err:=os.ReadFile(a.updateProgressPath());if err!=nil{return}
 var p UpdateProgress
 if json.Unmarshal(b,&p)!=nil{return}
 if !p.Active&&p.Phase=="complete"{
  a.updateProgress=UpdateProgress{}
  if err:=os.Remove(a.updateProgressPath());err!=nil&&!os.IsNotExist(err){
   log.Printf("cannot remove old completed update state: %v",err)
  }
  return
 }
 if p.Active {
  if p.Phase=="restarting"&&((p.Operation=="install"&&p.Target==appVersion)||(p.Operation=="rollback"&&p.From!=appVersion)||p.Operation=="manual"){
   p.Phase="complete";p.Percent=100;p.Message="Syncria "+appVersion+" is running";p.Error="";p.Active=false
   p.Events=append(p.Events,UpdateEvent{At:time.Now().UTC(),Phase:"complete",Detail:"Verified running version "+appVersion+" after restart"})
   p.UpdatedAt=time.Now().UTC()
   a.updateProgress=p
   a.clearSuccessfulUpdateLocked()
   return
  }
  p.Phase="failed";p.Active=false
  p.Error="The update was interrupted or the expected version did not start. Inspect startup logs before trying again."
  p.Events=append(p.Events,UpdateEvent{At:time.Now().UTC(),Phase:"failed",Detail:p.Error})
  p.Message="Update did not complete"
  if err:=a.writeUpdateProgressLocked(p);err!=nil{log.Printf("cannot reconcile update progress: %v",err)}
 }else{a.updateProgress=p}
}

func (a *App) updateProgressAPI(w http.ResponseWriter,r *http.Request) {
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("X-Content-Type-Options","nosniff")
 if !a.authenticated(r){http.Error(w,"Unauthorized",http.StatusUnauthorized);return}
 if r.Method==http.MethodGet {
  w.Header().Set("Content-Type","application/json; charset=utf-8")
  _=json.NewEncoder(w).Encode(struct {
   UpdateProgress
   Current string `json:"current"`
   ElapsedSeconds int64 `json:"elapsed_seconds"`
  }{func() UpdateProgress {return a.updateProgressSnapshot()}(),appVersion,func()int64 {
    p:=a.updateProgressSnapshot()
    if p.StartedAt.IsZero(){return 0}
    end:=time.Now()
    if (p.Phase=="complete"||p.Phase=="failed")&&!p.UpdatedAt.IsZero(){end=p.UpdatedAt}
    seconds:=int64(end.Sub(p.StartedAt).Seconds())
    if seconds<0{return 0}
    return seconds
  }()})
  return
 }
 if r.Method!=http.MethodPost{w.Header().Set("Allow","GET, POST");http.Error(w,"Method not allowed",http.StatusMethodNotAllowed);return}
 if !localPost(r){http.Error(w,"Same-origin POST required",http.StatusForbidden);return}
 r.Body=http.MaxBytesReader(w,r.Body,4096)
 if err:=r.ParseForm();err!=nil{http.Error(w,"Invalid update request",http.StatusBadRequest);return}
 operation:=strings.TrimSpace(r.FormValue("operation"))
 if operation!="install"&&operation!="rollback"{http.Error(w,"Invalid update operation",http.StatusBadRequest);return}
 if err:=a.beginUpdate(operation);err!=nil{http.Error(w,err.Error(),http.StatusConflict);return}
 w.Header().Set("Content-Type","application/json; charset=utf-8")
 w.WriteHeader(http.StatusAccepted)
 _=json.NewEncoder(w).Encode(a.updateProgressSnapshot())
 go a.runUpdate(operation)
}
// Update stages represent observed actions. Only download and SHA-256 hashing
// report percentages: both have exact byte counters. Network handshakes,
// file renames and the process restart have no honest percentage estimate.
func (a *App) updateAsset(name string){
 a.updateMu.Lock();defer a.updateMu.Unlock()
 p:=a.updateProgress;p.Asset=name
 if err:=a.writeUpdateProgressLocked(p);err!=nil{log.Printf("update asset state: %v",err)}
}
func (a *App) updateEvent(phase,detail string){
 a.updateMu.Lock();defer a.updateMu.Unlock()
 p:=a.updateProgress
 p.Events=append(p.Events,UpdateEvent{At:time.Now().UTC(),Phase:phase,Detail:detail})
 if len(p.Events)>45{p.Events=p.Events[len(p.Events)-45:]}
 if err:=a.writeUpdateProgressLocked(p);err!=nil{log.Printf("update event state: %v",err)}
}
func (a *App) runUpdate(operation string) {
 defer func(){
  if x:=recover();x!=nil{a.finishUpdate("failed","Updater panicked; restart not attempted",fmt.Sprint(x))}
 }()
 ctx,cancel:=context.WithTimeout(context.Background(),15*time.Minute)
 defer cancel()
 dir:=filepath.Dir(a.file)
 if operation=="rollback" {
  a.setUpdateProgress("staging","Restoring the previously saved executable","",-1,-1,-1,"")
  if _,err:=rollbackBinary(dir);err!=nil{a.finishUpdate("failed","Rollback installation failed",err.Error());return}
  a.updateEvent("staging","Previous executable restored; requesting supervisor restart")
 }else{
  a.setUpdateProgress("checking","Reading published release metadata from GitHub","",-1,-1,-1,"")
  rel,err:=latestRelease(ctx)
  if err!=nil{a.finishUpdate("failed","Release lookup failed",err.Error());return}
  if !newerVersion(rel.Tag,appVersion){
   a.finishUpdate("complete","Syncria "+appVersion+" is already the latest published version","")
   return
  }
  a.setUpdateProgress("manifest","Release "+rel.Tag+" found; obtaining SHA256SUMS",rel.Tag,-1,0,0,"")
  a.updateEvent("manifest","Platform executable: "+assetName())
  a.updateAsset(assetName())
  _,err=updateBinaryWithProgress(ctx,dir,rel,func(phase string,bytes,total int64){
   switch phase {
   case "manifest":
    a.setUpdateProgress("manifest","Receiving SHA256SUMS from "+rel.Tag,rel.Tag,-1,-1,-1,"")
   case "downloading":
    a.setUpdateProgress("downloading","Downloading "+assetName()+" from "+rel.Tag,rel.Tag,-1,bytes,total,"")
   case "verifying":
    a.setUpdateProgress("verifying","Calculating SHA-256 of "+assetName(),rel.Tag,-1,bytes,total,"")
   case "staging":
    a.setUpdateProgress("staging","Saving rollback executable and installing "+assetName(),rel.Tag,-1,-1,-1,"")
   }
  })
  if err!=nil{a.finishUpdate("failed","Verified installation was not completed",err.Error());return}
  a.updateEvent("staging","SHA-256 matched release manifest; verified binary committed to /data/runtime/current")
 }
 a.setUpdateProgress("restarting","Verified executable staged; requesting supervised restart","",-1,-1,-1,"")
 time.Sleep(700*time.Millisecond)
 restartSyncria(dir)
 a.finishUpdate("failed","Restart request failed","The runtime did not start the staged executable. Inspect Syncria logs.")
}
