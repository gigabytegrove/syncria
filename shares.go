package main

import (
 "context"
 "errors"
 "fmt"
 "net"
 "net/http"
 "os"
 "os/exec"
 "path/filepath"
 "regexp"
 "runtime"
 "strings"
 "time"
)

// StorageShare is a managed mount. SMB credentials are persisted in a private
// file under the application data directory, never in the HTML or state JSON.
type StorageShare struct {
 ID string `json:"id"`
 Label string `json:"label"`
 Protocol string `json:"protocol"`
 Server string `json:"server"`
 Export string `json:"export"`
 Username string `json:"username,omitempty"`
 Local string `json:"local"`
 Mounted bool `json:"-"`
}

var shareHost = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.:-]{0,252}$`)
var shareName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._ -]{0,127}$`)

func validateShare(protocol,server,export,label string)error{
 if protocol!="nfs"&&protocol!="smb"{return errors.New("Select NFS or SMB")}
 if !shareHost.MatchString(server)||net.ParseIP(server)==nil&&strings.Contains(server,":"){return errors.New("Invalid host name or IP address")}
 if !shareName.MatchString(label){return errors.New("Use a short connection name with letters, numbers, spaces or dashes")}
 if protocol=="nfs"{
  if !strings.HasPrefix(export,"/")||strings.ContainsAny(export,"\x00\n\r")||strings.Contains(export,".."){return errors.New("NFS export must be an absolute export path")}
 }else{
  if !shareName.MatchString(export){return errors.New("SMB share must be a share name, not a path")}
 }
 return nil
}
func credsFile(a *App,id string)string{return filepath.Join(filepath.Dir(a.file),"connections",id+".credentials")}
func shareMounted(path string)bool{
 c, cancel:=context.WithTimeout(context.Background(),3*time.Second);defer cancel()
 return exec.CommandContext(c,"mountpoint","-q",path).Run()==nil
}
func (a *App) mountShare(s StorageShare)error{
 if runtime.GOOS!="linux"{return errors.New("Managed NFS/SMB mounts currently require Linux")}
 if os.Getenv("SYNCRIA_ENABLE_MOUNTS")!="1"{return errors.New("Network mounting is disabled. Set SYNCRIA_ENABLE_MOUNTS=1 on a Linux host with mount privileges")}
 if shareMounted(s.Local){return nil}
 root,e:=resolvedPath(browseStorageRoot());if e!=nil{return fmt.Errorf("storage root: %w",e)}
 dest:=filepath.Join(root,"syncria-shares",s.ID)
 if !contained(root,dest)||dest!=s.Local{return errors.New("Invalid mount destination")}
 if e=os.MkdirAll(dest,0750);e!=nil{return e}
 ctx,cancel:=context.WithTimeout(context.Background(),30*time.Second);defer cancel()
 var args []string
 if s.Protocol=="nfs"{
  args=[]string{"-t","nfs","-o","nosuid,nodev",s.Server+":"+s.Export,dest}
 }else{
  creds:=credsFile(a,s.ID)
  if _,e=os.Stat(creds);e!=nil{return errors.New("SMB credentials missing; reconnect share")}
  args=[]string{"-t","cifs","-o","nosuid,nodev,credentials="+creds,"//"+s.Server+"/"+s.Export,dest}
 }
 out,e:=exec.CommandContext(ctx,"mount",args...).CombinedOutput()
 if e!=nil{return fmt.Errorf("Mount failed: %s (%w). Verify network access, mount helpers and Linux mount permissions",strings.TrimSpace(string(out)),e)}
 return nil
}
func (a *App) removeShare(w http.ResponseWriter,r *http.Request,id string){
 a.mu.Lock()
 index:=-1
 var share StorageShare
 for i,s:=range a.cfg.Shares{if s.ID==id{index=i;share=s;break}}
 if index<0{a.mu.Unlock();redirect(w,r,"Connection not found");return}
 for _,j:=range a.cfg.Jobs{if j.Local==share.Local||strings.HasPrefix(j.Local,share.Local+string(os.PathSeparator)){a.mu.Unlock();redirect(w,r,"Remove synchronization jobs using this connection first");return}}
 a.mu.Unlock()
 if shareMounted(share.Local){
  ctx,cancel:=context.WithTimeout(r.Context(),20*time.Second)
  out,e:=exec.CommandContext(ctx,"umount",share.Local).CombinedOutput();cancel()
  if e!=nil{redirect(w,r,"Cannot disconnect mounted share: "+strings.TrimSpace(string(out)));return}
 }
 a.mu.Lock()
 for i,s:=range a.cfg.Shares{if s.ID==id{a.cfg.Shares=append(a.cfg.Shares[:i],a.cfg.Shares[i+1:]...);break}}
 e:=a.save();a.mu.Unlock()
 if e!=nil{http.Error(w,e.Error(),500);return}
 _=os.Remove(credsFile(a,id))
 redirect(w,r,"Storage connection disconnected")
}
func (a *App) shareAction(w http.ResponseWriter,r *http.Request){
 if !localPost(r){http.Error(w,"Same-origin POST required",405);return}
 if !a.authenticated(r){http.Error(w,"Unauthorized",401);return}
 if e:=r.ParseForm();e!=nil{http.Error(w,"Bad form data",400);return}
 switch r.FormValue("operation"){
 case "disconnect":a.removeShare(w,r,r.FormValue("id"));return
 case "reconnect":
  id:=r.FormValue("id")
  a.mu.Lock();var share StorageShare;ok:=false
  for _,s:=range a.cfg.Shares{if s.ID==id{share=s;ok=true;break}}
  a.mu.Unlock()
  if !ok{redirect(w,r,"Connection not found");return}
  if e:=a.mountShare(share);e!=nil{redirect(w,r,e.Error());return}
  redirect(w,r,"Storage connected");return
 case "connect":
  protocol:=strings.ToLower(strings.TrimSpace(r.FormValue("protocol")))
  server:=strings.TrimSpace(r.FormValue("server"))
  export:=strings.TrimSpace(r.FormValue("export"))
  label:=strings.TrimSpace(r.FormValue("label"))
  if e:=validateShare(protocol,server,export,label);e!=nil{redirect(w,r,e.Error());return}
  if os.Getenv("SYNCRIA_ENABLE_MOUNTS")!="1"{redirect(w,r,"Mounting disabled on this host; set SYNCRIA_ENABLE_MOUNTS=1 and grant Linux mount capability");return}
  a.mu.Lock()
  for _,s:=range a.cfg.Shares{if s.Protocol==protocol&&s.Server==server&&s.Export==export{a.mu.Unlock();redirect(w,r,"That network share is already configured");return}}
  a.mu.Unlock()
  id:=newID()
  root,e:=resolvedPath(browseStorageRoot());if e!=nil{redirect(w,r,"Storage root unavailable: "+e.Error());return}
  share:=StorageShare{ID:id,Protocol:protocol,Server:server,Export:export,Label:label,Username:strings.TrimSpace(r.FormValue("username")),Local:filepath.Join(root,"syncria-shares",id)}
  if protocol=="smb"{
   password:=r.FormValue("password")
   if share.Username==""||password==""||strings.ContainsAny(password,"\n\r")||strings.ContainsAny(share.Username,"\n\r"){redirect(w,r,"SMB username and password required without line breaks");return}
   if e=os.MkdirAll(filepath.Join(filepath.Dir(a.file),"connections"),0700);e!=nil{redirect(w,r,e.Error());return}
   creds:="username="+share.Username+"\npassword="+password+"\n"
   if e=os.WriteFile(credsFile(a,id),[]byte(creds),0600);e!=nil{redirect(w,r,e.Error());return}
  }
  if e=a.mountShare(share);e!=nil{
   _=os.Remove(credsFile(a,id))
   redirect(w,r,e.Error());return
  }
  a.mu.Lock();a.cfg.Shares=append(a.cfg.Shares,share);e=a.save();a.mu.Unlock()
  if e!=nil{redirect(w,r,"Mounted but unable to save connection: "+e.Error());return}
  redirect(w,r,"Network storage connected and available in folder browser");return
 default:http.Error(w,"Unknown storage operation",400)
 }
}
func (a *App) restoreShares(){
 a.mu.Lock();shares:=append([]StorageShare(nil),a.cfg.Shares...);a.mu.Unlock()
 for _,s:=range shares{if err:=a.mountShare(s);err!=nil{fmt.Fprintf(os.Stderr,"Syncria share %s: %v\n",s.Label,err)}}
}
