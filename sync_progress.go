package main

import (
 "os"
 "os/exec"
 "strings"
 "time"
 "context"
 "log"
)

// SyncProgress represents the currently executing relationship. Unlike
// historical statistics, it is memory-only and does not rewrite config.json
// for every file. Completed runs are persisted via the existing Job counters.
type SyncActionEvent struct {
 At time.Time `json:"at"`
 Action string `json:"action"`
 Path string `json:"path"`
 From string `json:"from,omitempty"`
 To string `json:"to,omitempty"`
}
type SyncProgress struct {
 Events []SyncActionEvent `json:"events,omitempty"`
 Phase string `json:"phase"`
 Current string `json:"current,omitempty"`
 Processed int64 `json:"processed"`
 Total int64 `json:"total"`
 Uploads int64 `json:"uploads"`
 Downloads int64 `json:"downloads"`
 UploadBytes int64 `json:"upload_bytes"`
 DownloadBytes int64 `json:"download_bytes"`
 Started time.Time `json:"started"`
 Running bool `json:"running"`
 Error string `json:"error,omitempty"`
}

func (a *App) setSyncProgress(id string,update func(*SyncProgress)) {
 a.mu.Lock()
 if a.syncProgress==nil{a.syncProgress=map[string]SyncProgress{}}
 p:=a.syncProgress[id]
 update(&p)
 a.syncProgress[id]=p
 a.mu.Unlock()
}
func (a *App) getSyncProgress(id string) SyncProgress {
 a.mu.Lock();defer a.mu.Unlock()
 return a.syncProgress[id]
}

func (a *App) watchChangedJob(ctx context.Context,id string) {
 defer a.watchers.Delete(id)
 a.mu.Lock()
 initial:=a.job(id)
 remote:=initial!=nil&&(initial.Provider=="bridge"||initial.Provider=="nas_bridge")
 a.mu.Unlock()
 if remote {
  a.watchBridgeChanges(ctx,id)
  return
 }
 // inotify notifications from local Linux paths are useful for immediate
 // sync. SMB/NFS remote changes may not propagate: a periodic five-minute
 // reconciliation still runs via the normal scheduler.
 for {
  if ctx.Err()!=nil{return}
  a.mu.Lock()
  j:=a.job(id)
  if j==nil||!j.Enabled||j.ScheduleMode!="changes"{a.mu.Unlock();return}
  paths:=[]string{j.Local}
  if j.Provider=="storage"{paths=append(paths,j.RemoteID)}
  a.mu.Unlock()
  if _,err:=exec.LookPath("inotifywait");err!=nil{
   log.Printf("Syncria relationship %s: inotifywait missing; using 5-minute reconciliation",id)
   return
  }
  args:=[]string{"-q","-r","-t","20","-e","close_write,create,delete,moved_to,moved_from","--"}
  for _,p:=range paths {
   if stat,err:=os.Stat(p);err==nil&&stat.IsDir(){args=append(args,p)}
  }
  if len(args)==7{return}
  cmd:=exec.CommandContext(ctx,"inotifywait",args...)
  err:=cmd.Run()
  if ctx.Err()!=nil{return}
  if err==nil {
   a.mu.Lock()
   j=a.job(id)
   enabled:=j!=nil&&j.Enabled&&j.ScheduleMode=="changes"
   a.mu.Unlock()
   if enabled {
    // Debounce bursty writes and protect against run overlap.
    select{case <-ctx.Done():return;case <-time.After(2*time.Second):}
    if e:=a.syncJob(ctx,id);e!=nil&&e.Error()!="job already running"{
     log.Printf("Syncria change-triggered sync %s: %v",id,e)
    }
   }
  } else if !strings.Contains(err.Error(),"exit status 2"){
   log.Printf("Syncria inotify watcher %s: %v",id,err)
   select{case <-ctx.Done():return;case <-time.After(10*time.Second):}
  }
 }
}
