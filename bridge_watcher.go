package main

import (
 "bufio"
 "context"
 "errors"
 "io"
 "log"
 "os"
 "os/exec"
 "strings"
 "time"
)

// A remote SMB notification is a server-side change event. It does not
// require inotify on the Syncria host or a kernel-mounted NAS filesystem.
// Cloud changes are discovered by the scheduler's five-minute reconciliation.
func (a *App) watchBridgeChanges(ctx context.Context,id string) {
 watchCtx,cancel:=context.WithCancel(ctx)
 defer cancel()
 events:=make(chan struct{},1)
 watched:=map[string]bool{}
 ticker:=time.NewTicker(20*time.Second)
 defer ticker.Stop()

 a.mu.Lock()
 job:=a.job(id)
 if job==nil||!job.Enabled||job.ScheduleMode!="changes" {
  a.mu.Unlock();return
 }
 destinations:=append([]NASDestination(nil),job.NASDestinations...)
 if len(destinations)==0&&job.NASShareID!="" {
  destinations=append(destinations,NASDestination{ShareID:job.NASShareID,Folder:job.NASFolder})
 }
 if job.Provider=="nas_bridge"&&job.NASSourceID!=""{
  destinations=append(destinations,NASDestination{ShareID:job.NASSourceID,Folder:job.NASSourceFolder})
 }
 a.mu.Unlock()

 for _,d:=range destinations {
  if d.ShareID==""{continue}
  key:=d.ShareID+"|"+d.Folder
  if watched[key]{continue}
  watched[key]=true
  share,err:=a.directShare(d.ShareID)
  if err!=nil{log.Printf("Syncria SMB change notification setup: %v",err);continue}
  go a.monitorSMBChanges(watchCtx,share,d.Folder,events)
 }
 if len(watched)==0 {
  log.Printf("Syncria relationship %s has no NAS folder available for remote events; using periodic reconciliation",id)
  return
 }
 for {
  select {
  case <-ctx.Done():return
  case <-ticker.C:
   a.mu.Lock()
   current:=a.job(id)
   active:=current!=nil&&current.Enabled&&current.ScheduleMode=="changes"
   a.mu.Unlock()
   if !active{return}
  case <-events:
   // A burst of remote writes is one synchronization run.
   select {
   case <-ctx.Done():return
   case <-time.After(2*time.Second):
   }
   a.mu.Lock()
   current:=a.job(id)
   active:=current!=nil&&current.Enabled&&current.ScheduleMode=="changes"
   a.mu.Unlock()
   if !active{return}
   if err:=a.syncJob(ctx,id);err!=nil&&err.Error()!="job already running" {
    log.Printf("Syncria remote NAS change-triggered relationship %s: %v",id,err)
   }
  }
 }
}

func (a *App) monitorSMBChanges(ctx context.Context,share StorageShare,folder string,event chan<-struct{}) {
 if _,err:=exec.LookPath("smbclient");err!=nil {
  log.Printf("Syncria remote NAS notifications unavailable: smbclient is not installed")
  return
 }
 if share.Protocol!="smb"||share.Local!=""{return}
 remote,err:=safeRemotePath(folder);if err!=nil{return}
 credentials:=credsFile(a,share.ID)
 if _,err:=os.Stat(credentials);err!=nil{
  log.Printf("Syncria NAS change notifications: credential file is missing for %s",share.Label)
  return
 }
 for ctx.Err()==nil {
  // Directory is passed as a dedicated argument, not interpolated into an
  // smbclient shell command. Passwords are only in the private credential
  // file, never passed in command arguments.
  args:=[]string{"//"+share.Server+"/"+share.Export,"-A",credentials,"-D",remote,"-c","notify ."}
  cmd:=exec.CommandContext(ctx,"smbclient",args...)
  stdout,err:=cmd.StdoutPipe()
  if err!=nil{log.Printf("Syncria SMB notifications %s: %v",share.Label,err);return}
  cmd.Stderr=io.Discard
  if err=cmd.Start();err==nil{
   scanner:=bufio.NewScanner(stdout)
   scanner.Buffer(make([]byte,4096),128*1024)
   for scanner.Scan() {
    line:=strings.TrimSpace(scanner.Text())
    if line==""||strings.HasPrefix(line,"Domain=[")||
       strings.HasPrefix(strings.ToLower(line),"try ")||
       strings.HasPrefix(line,"smb:") {continue}
    select{case event<-struct{}{}:default:}
   }
   err=cmd.Wait()
  }
  if ctx.Err()!=nil{return}
  log.Printf("Syncria NAS change notifications disconnected on %s: %v; retrying",share.Label,err)
  select{
  case <-ctx.Done():return
  case <-time.After(15*time.Second):
  }
 }
}

var _=errors.New
