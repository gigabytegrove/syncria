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
type UpdateProgress struct {
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
 p.Phase=phase;p.Message=message
 if target!=""{p.Target=target}
 if percent>=0{p.Percent=percent}
 if downloaded>=0{p.Downloaded=downloaded}
 if total>=0{p.Total=total}
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
 p:=UpdateProgress{Phase:"checking",Operation:operation,From:appVersion,Message:"Starting update",Percent:2,Active:true,StartedAt:time.Now().UTC()}
 if operation=="rollback"{p.Phase="staging";p.Message="Preparing previous verified binary";p.Percent=70}
 if err:=a.writeUpdateProgressLocked(p);err!=nil{return fmt.Errorf("cannot persist update progress: %w",err)}
 a.updateBusy=true
 return nil
}
func (a *App) finishUpdate(phase,message,detail string) {
 a.setUpdateProgress(phase,message,"",func()int{if phase=="complete"{return 100};return -1}(),-1,-1,detail)
 a.updateMu.Lock()
 a.updateBusy=false
 a.updateMu.Unlock()
}
func (a *App) updateProgressSnapshot() UpdateProgress {
 a.updateMu.Lock()
 defer a.updateMu.Unlock()
 if a.updateProgress.Phase!="" {return a.updateProgress}
 b,err:=os.ReadFile(a.updateProgressPath())
 if err==nil{_ =json.Unmarshal(b,&a.updateProgress)}
 if a.updateProgress.Phase==""{return UpdateProgress{Phase:"idle",Message:"No update in progress"}}
 return a.updateProgress
}

// Called by the newly launched Syncria binary. No separate host service,
// Docker socket, update container or fixed-time page reload is necessary.
func (a *App) reconcileUpdateOnStartup() {
 a.updateMu.Lock()
 defer a.updateMu.Unlock()
 b,err:=os.ReadFile(a.updateProgressPath());if err!=nil{return}
 var p UpdateProgress
 if json.Unmarshal(b,&p)!=nil{return}
 if p.Active {
  if p.Phase=="restarting" && ((p.Operation=="install"&&p.Target==appVersion)||(p.Operation=="rollback"&&p.From!=appVersion)) {
   p.Phase="complete";p.Percent=100;p.Message="Syncria "+appVersion+" started successfully";p.Error="";p.Active=false
  }else{
   p.Phase="failed";p.Active=false
   p.Error="The update was interrupted or the expected version did not start."
   p.Message="Update did not complete"
  }
  if err:=a.writeUpdateProgressLocked(p);err!=nil{log.Printf("cannot reconcile update progress: %v",err)}
 }else{a.updateProgress=p}
}

func (a *App) updateProgressAPI(w http.ResponseWriter,r *http.Request) {
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("X-Content-Type-Options","nosniff")
 if !a.authenticated(r){http.Error(w,"Unauthorized",http.StatusUnauthorized);return}
 if r.Method==http.MethodGet {
  w.Header().Set("Content-Type","application/json; charset=utf-8")
  _=json.NewEncoder(w).Encode(a.updateProgressSnapshot())
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
func (a *App) runUpdate(operation string) {
 defer func(){
  if x:=recover();x!=nil{a.finishUpdate("failed","Update interrupted",fmt.Sprint(x))}
 }()
 ctx,cancel:=context.WithTimeout(context.Background(),15*time.Minute)
 defer cancel()
 dir:=filepath.Dir(a.file)
 if operation=="rollback" {
  if _,err:=rollbackBinary(dir);err!=nil{a.finishUpdate("failed","Rollback failed",err.Error());return}
 }else{
  a.setUpdateProgress("checking","Checking latest published release","",5,0,0,"")
  rel,err:=latestRelease(ctx)
  if err!=nil{a.finishUpdate("failed","Release check failed",err.Error());return}
  if !newerVersion(rel.Tag,appVersion){
   a.finishUpdate("complete","Syncria is already running the latest published release","")
   return
  }
  a.setUpdateProgress("downloading","Downloading "+rel.Tag,rel.Tag,10,0,0,"")
  _,err=updateBinaryWithProgress(ctx,dir,rel,func(phase string,downloaded,total int64){
   switch phase {
   case "downloading":
    percent:=10
    if total>0{percent+=int(downloaded*60/total);if percent>70{percent=70}}
    a.setUpdateProgress(phase,"Downloading "+rel.Tag,rel.Tag,percent,downloaded,total,"")
   case "verifying":
    a.setUpdateProgress(phase,"Checking SHA-256 integrity",rel.Tag,80,-1,-1,"")
   case "staging":
    a.setUpdateProgress(phase,"Saving rollback and installing verified binary",rel.Tag,92,-1,-1,"")
   }
  })
  if err!=nil{a.finishUpdate("failed","Update failed",err.Error());return}
 }
 a.setUpdateProgress("restarting","Syncria is restarting and confirming the new version","",97,-1,-1,"")
 time.Sleep(700*time.Millisecond)
 restartSyncria(dir)
 // A supervised or successfully spawned process exits and cannot return.
 a.finishUpdate("failed","Restart failed","Syncria could not start the staged executable")
}
