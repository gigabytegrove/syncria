package main

import (
 "context"
 "errors"
 "fmt"
 "net/http"
 "net/url"
 "os"
 "path"
 "sort"
 "strconv"
 "strings"
 "time"
)

func (a *App) directShare(id string)(StorageShare,error){
 a.mu.Lock();defer a.mu.Unlock()
 for _,s:=range a.cfg.Shares{
  if s.ID==id && s.Local=="" && s.Protocol=="smb"{return s,nil}
 }
 return StorageShare{},errors.New("Select a connected direct SMB NAS share")
}

func (a *App) syncBridge(ctx context.Context,j *Job) error {
 if j.AccountID==""||j.NASShareID==""{return errors.New("Bridge source account or NAS destination is missing")}
 share,err:=a.directShare(j.NASShareID);if err!=nil{return err}
 network,err:=a.dialSMB(ctx,share);if err!=nil{return err}
 dest:=&bridgeSMBEndpoint{client:network,root:j.NASFolder}
 cloud:=&bridgeDriveEndpoint{app:a,account:j.AccountID,root:j.RemoteID}
 return a.executeBridge(ctx,j,cloud,dest)
}

// addBridge only saves endpoint references and sync state on Syncria.
// It never creates a local file-storage path.
func (a *App) addBridge(w http.ResponseWriter,r *http.Request){
 name:=strings.TrimSpace(r.FormValue("name"))
 if name==""||len(name)>128{redirect(w,r,"Relationship name must be 1–128 characters");return}
 acc:=strings.TrimSpace(r.FormValue("account"))
 remote:=strings.TrimSpace(r.FormValue("remote"))
 if remote==""{remote="root"}
 if remote!="root"&&!validDriveID(remote){redirect(w,r,"Invalid Drive source folder");return}
 nasID:=strings.TrimSpace(r.FormValue("nas_id"))
 nasPath,err:=safeRemotePath(r.FormValue("nas_folder"))
 if err!=nil{redirect(w,r,err.Error());return}
 if _,err=a.directShare(nasID);err!=nil{redirect(w,r,err.Error());return}
 mode:=strings.TrimSpace(r.FormValue("schedule_mode"))
 if mode==""{mode="interval"}
 if mode!="interval"&&mode!="changes"{redirect(w,r,"Invalid sync schedule");return}
 interval,err:=strconv.Atoi(r.FormValue("interval"))
 if err!=nil||interval<1||interval>1440{redirect(w,r,"Invalid synchronization interval");return}
 if mode=="changes"{
  // An SMB share does not propagate filesystem inotify events to this
  // container. Polling is the only truthful option without remote change APIs.
  redirect(w,r,"SMB / Google Drive bridge connections require a polling schedule; remote changes do not generate local inotify events")
  return
 }
 a.mu.Lock()
 if a.account(acc)==nil{a.mu.Unlock();redirect(w,r,"Select a connected Google Drive account");return}
 for _,other:=range a.cfg.Jobs {
  if other.Provider=="bridge"&&other.NASShareID==nasID&&pathsOverlap(path.Join("/",other.NASFolder),path.Join("/",nasPath)){
   a.mu.Unlock();redirect(w,r,"This NAS folder is already used by another relationship");return
  }
 }
 a.mu.Unlock()
 ctx,cancel:=context.WithTimeout(r.Context(),30*time.Second);defer cancel()
 share,err:=a.directShare(nasID);if err!=nil{redirect(w,r,err.Error());return}
 smb,err:=a.dialSMB(ctx,share);if err!=nil{redirect(w,r,"NAS is not accessible: "+err.Error());return}
 _,err=smb.share.ReadDir(nasPath);smb.Close()
 if err!=nil{redirect(w,r,"NAS folder inaccessible: "+err.Error());return}
 if _,err=a.folderChildren(ctx,acc,remote);err!=nil{redirect(w,r,"Google Drive source inaccessible: "+err.Error());return}
 a.mu.Lock()
 a.cfg.Jobs=append(a.cfg.Jobs,Job{ID:newID(),Provider:"bridge",Name:name,AccountID:acc,RemoteID:remote,NASShareID:nasID,NASFolder:nasPath,Interval:interval,ScheduleMode:mode,Enabled:false,State:map[string]Item{}})
 err=a.save();a.mu.Unlock()
 if err!=nil{http.Error(w,err.Error(),500);return}
 http.Redirect(w,r,"/?msg="+url.QueryEscape("Bridge configured, paused. Files stay on Google Drive and your NAS.")+"#sync",http.StatusSeeOther)
}

// A NAS connection is authenticated directly via SMB2/3; kernel mounts,
// /storage bind mappings and SYS_ADMIN are never used.
func (a *App) directNASAction(w http.ResponseWriter,r *http.Request){
 if !localPost(r){http.Error(w,"Same-origin POST required",http.StatusMethodNotAllowed);return}
 if !a.authenticated(r){http.Error(w,"Unauthorized",http.StatusUnauthorized);return}
 r.Body=http.MaxBytesReader(w,r.Body,16384)
 if err:=r.ParseForm();err!=nil{http.Error(w,"Invalid request",400);return}
 switch r.FormValue("operation"){
 case "connect":
  protocol:=strings.ToLower(strings.TrimSpace(r.FormValue("protocol")))
  server:=strings.TrimSpace(r.FormValue("server"))
  export:=strings.TrimSpace(r.FormValue("export"))
  label:=strings.TrimSpace(r.FormValue("label"))
  if protocol!="smb"{redirect(w,r,"NFS direct adapter is not yet implemented; choose an SMB2/3 share");return}
  if err:=validateShare(protocol,server,export,label);err!=nil{redirect(w,r,err.Error());return}
  if _,err:=validateDiscoveryIP(server);err!=nil{redirect(w,r,"NAS must have a private/local IP address");return}
  user:=strings.TrimSpace(r.FormValue("username"))
  password:=r.FormValue("password")
  if user==""||password==""||strings.ContainsAny(user+password,"\r\n\x00"){redirect(w,r,"SMB credentials are required and must not contain line breaks");return}
  a.mu.Lock()
  for _,s:=range a.cfg.Shares{if s.Protocol==protocol&&s.Server==server&&s.Export==export&&s.Local==""{a.mu.Unlock();redirect(w,r,"The NAS share is already configured");return}}
  a.mu.Unlock()
  id:=newID()
  if err:=os.MkdirAll(path.Dir(credsFile(a,id)),0700);err!=nil{http.Error(w,"Cannot save NAS credentials",500);return}
  if err:=os.WriteFile(credsFile(a,id),[]byte("username="+user+"\npassword="+password+"\n"),0600);err!=nil{http.Error(w,"Cannot save NAS credentials",500);return}
  share:=StorageShare{ID:id,Protocol:protocol,Label:label,Server:server,Export:export,Username:user}
  ctx,cancel:=context.WithTimeout(r.Context(),25*time.Second)
  network,err:=a.dialSMB(ctx,share)
  if err==nil{_,err=network.share.ReadDir(".");network.Close()}
  cancel()
  if err!=nil{_ =os.Remove(credsFile(a,id));redirect(w,r,"NAS connection failed: "+err.Error());return}
  a.mu.Lock();a.cfg.Shares=append(a.cfg.Shares,share);err=a.save();a.mu.Unlock()
  if err!=nil{_ =os.Remove(credsFile(a,id));http.Error(w,"Cannot store connection",500);return}
  http.Redirect(w,r,"/?msg="+url.QueryEscape("NAS connected directly over SMB; no mount required")+"#storage",http.StatusSeeOther)
 case "verify":
  share,err:=a.directShare(r.FormValue("id"));if err!=nil{redirect(w,r,err.Error());return}
  ctx,cancel:=context.WithTimeout(r.Context(),25*time.Second)
  network,err:=a.dialSMB(ctx,share)
  if err==nil{_,err=network.share.ReadDir(".");network.Close()}
  cancel()
  if err!=nil{redirect(w,r,"NAS verification failed: "+err.Error());return}
  http.Redirect(w,r,"/?msg="+url.QueryEscape("NAS authentication and share listing verified")+"#storage",http.StatusSeeOther)
 case "disconnect":
  id:=r.FormValue("id")
  a.mu.Lock()
  index:=-1
  for i,s:=range a.cfg.Shares{if s.ID==id&&s.Local==""{index=i;break}}
  if index<0{a.mu.Unlock();redirect(w,r,"NAS connection not found");return}
  for _,j:=range a.cfg.Jobs{if j.Provider=="bridge"&&j.NASShareID==id{a.mu.Unlock();redirect(w,r,"Remove linked sync relationships before disconnecting this NAS");return}}
  a.cfg.Shares=append(a.cfg.Shares[:index],a.cfg.Shares[index+1:]...)
  err:=a.save();a.mu.Unlock()
  if err!=nil{http.Error(w,"Unable to save connection changes",500);return}
  _ =os.Remove(credsFile(a,id))
  http.Redirect(w,r,"/?msg="+url.QueryEscape("NAS connection removed; no files were changed")+"#storage",http.StatusSeeOther)
 default:http.Error(w,"Invalid NAS operation",400)
 }
}

func (a *App) browseNAS(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodGet{http.Error(w,"GET required",http.StatusMethodNotAllowed);return}
 if !a.authenticated(r){http.Error(w,"Unauthorized",401);return}
 share,err:=a.directShare(r.URL.Query().Get("id"));if err!=nil{browseError(w,400,err.Error());return}
 root,err:=safeRemotePath(r.URL.Query().Get("path"));if err!=nil{browseError(w,400,err.Error());return}
 ctx,cancel:=context.WithTimeout(r.Context(),25*time.Second);defer cancel()
 network,err:=a.dialSMB(ctx,share);if err!=nil{browseError(w,502,err.Error());return};defer network.Close()
 infos,err:=network.share.ReadDir(root);if err!=nil{browseError(w,502,"NAS path not accessible: "+err.Error());return}
 result:=browseResponse{Path:root,Entries:[]browseEntry{}}
 if root!="."{result.Parent=path.Dir(root)}
 for _,info:=range infos{
  if !info.IsDir()||!safeName(info.Name()){continue}
  child,err:=joinRemotePath(root,info.Name());if err!=nil{continue}
  result.Entries=append(result.Entries,browseEntry{Name:info.Name(),ID:child,Kind:"folder"})
 }
 sort.Slice(result.Entries,func(i,k int)bool{return strings.ToLower(result.Entries[i].Name)<strings.ToLower(result.Entries[k].Name)})
 sendBrowse(w,result)
}
func (a *App) createNASFolder(w http.ResponseWriter,r *http.Request){
 if !localPost(r){http.Error(w,"Same-origin POST required",405);return}
 if !a.authenticated(r){http.Error(w,"Unauthorized",401);return}
 if err:=r.ParseForm();err!=nil{http.Error(w,"Invalid request",400);return}
 share,err:=a.directShare(r.FormValue("id"));if err!=nil{http.Error(w,err.Error(),400);return}
 name:=strings.TrimSpace(r.FormValue("name"));if !safeName(name){http.Error(w,"Invalid NAS folder name",400);return}
 root,err:=safeRemotePath(r.FormValue("parent"));if err!=nil{http.Error(w,err.Error(),400);return}
 child,err:=joinRemotePath(root,name);if err!=nil{http.Error(w,err.Error(),400);return}
 ctx,cancel:=context.WithTimeout(r.Context(),25*time.Second);defer cancel()
 network,err:=a.dialSMB(ctx,share);if err!=nil{http.Error(w,err.Error(),502);return};defer network.Close()
 if err=network.mkdir(".",child);err!=nil{http.Error(w,"Could not create NAS folder: "+err.Error(),502);return}
 sendBrowse(w,browseResponse{Path:child,Parent:root,Entries:[]browseEntry{},Message:"Folder created on NAS"})
}

// Explicit documentation of the direct connector's transport semantics.
// It creates no local storage files other than credentials and config JSON.
var _ = fmt.Sprintf
