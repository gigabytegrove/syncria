package main

import (
 "context"
 "errors"
 "os/exec"
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

func (a *App) cloudEndpoint(j *Job)(recoverableBridgeEndpoint,error){
 provider:=j.CloudProvider
 if provider==""{provider="google_drive"}
 switch provider{
 case "google_drive":
  if j.AccountID==""{return nil,errors.New("Google Drive account not selected")}
  a.mu.Lock();exists:=a.account(j.AccountID)!=nil;a.mu.Unlock()
  if !exists{return nil,errors.New("Google Drive account unavailable")}
  return &bridgeDriveEndpoint{app:a,account:j.AccountID,root:j.RemoteID,scanHook:func(rel string,count int64){if j.scanHook!=nil{j.scanHook("Google Drive",rel,count)}}},nil
 case "dropbox":
  if j.DropboxAccountID==""{return nil,errors.New("Dropbox account not selected")}
  a.mu.Lock();exists:=a.dropboxAccount(j.DropboxAccountID)!=nil;a.mu.Unlock()
  if !exists{return nil,errors.New("Dropbox account unavailable")}
  return &bridgeDropboxEndpoint{app:a,account:j.DropboxAccountID,root:j.RemoteID},nil
 default:return nil,fmt.Errorf("Cloud provider %q is not supported",provider)
 }
}
func (a *App) openNAS(ctx context.Context,d NASDestination)(recoverableBridgeEndpoint,error){
 share,err:=a.directShare(d.ShareID);if err!=nil{return nil,err}
 conn,err:=a.dialSMB(ctx,share);if err!=nil{return nil,err}
 return &bridgeSMBEndpoint{client:conn,root:d.Folder},nil
}
func (a *App) labelSyncOperation(j *Job,source,secondary string) func(){
 original:=j.operationHook
 if original==nil{return func(){}}
 j.operationHook=func(action,rel,from,to string){
  if from=="Primary"||from=="primary"{from=source}else if from=="Secondary"||from=="secondary"{from=secondary}
  if to=="Primary"||to=="primary"{to=source}else if to=="Secondary"||to=="secondary"{to=secondary}
  original(action,rel,from,to)
 }
 return func(){j.operationHook=original}
}
func (a *App) syncBridge(ctx context.Context,j *Job) error {
 destinations:=append([]NASDestination(nil),j.NASDestinations...)
 if len(destinations)==0&&j.NASShareID!=""{
  destinations=[]NASDestination{{ShareID:j.NASShareID,Folder:j.NASFolder}}
 }
 if len(destinations)==0{return errors.New("No destination NAS configured")}
 if j.Provider=="nas_bridge"{
  if j.NASSourceID==""{return errors.New("Select the primary NAS for the relationship")}
  source:=NASDestination{ShareID:j.NASSourceID,Folder:j.NASSourceFolder}
  policy:=j.ConflictPolicy
  if policy==""{policy=conflictManual}
  if j.Direction=="bidirectional"&&!validNASPolicy(policy){return errors.New("A valid approved NAS conflict policy is required")}
  if j.Direction!="one_way"&&j.Direction!="bidirectional"{return errors.New("Choose one-way or bidirectional NAS sync")}
  for _,destination:=range destinations{
   primary,err:=a.openNAS(ctx,source);if err!=nil{return fmt.Errorf("Primary NAS: %w",err)}
   secondary,err:=a.openNAS(ctx,destination);if err!=nil{primary.Close();return fmt.Errorf("Destination NAS: %w",err)}
   key:=relationPairKey("nas:"+source.ShareID+":"+source.Folder,destination)
   sourceLabel:=source.ShareID
   if share,e:=a.directShare(source.ShareID);e==nil&&share.Label!=""{sourceLabel=share.Label}
   destLabel:=destination.ShareID
   if share,e:=a.directShare(destination.ShareID);e==nil&&share.Label!=""{destLabel=share.Label}
   restore:=a.labelSyncOperation(j,sourceLabel,destLabel)
   err=a.executePolicyPair(ctx,j,primary,secondary,key,j.Direction=="one_way",policy)
   restore()
   if err!=nil{return err}
  }
  return nil
 }
 if j.Provider!="bridge"{return errors.New("Unsupported relationship type")}
 // When a separate NAS-to-NAS relationship owns the selected NAS group,
 // use the designated cloud-connected NAS only and let NAS-to-NAS do the
 // rest. Otherwise every explicitly configured destination remains active.
 a.mu.Lock();related:=false
 for _,other:=range a.cfg.Jobs {
  if other.ID==j.ID||other.Provider!="nas_bridge"||!other.Enabled{continue}
  for _,d:=range destinations{
   if d.ShareID==other.NASSourceID{related=true;break}
   for _,child:=range other.NASDestinations{if d.ShareID==child.ShareID{related=true;break}}
  }
  if related{break}
 }
 a.mu.Unlock()
 if related&&len(destinations)>1{destinations=destinations[:1]}
 for i,d:=range destinations{
  cloud,err:=a.cloudEndpoint(j);if err!=nil{return err}
  nas,err:=a.openNAS(ctx,d);if err!=nil{cloud.Close();return err}
  sourceName:=j.CloudProvider;if sourceName==""{sourceName="google_drive"}
  account:=j.AccountID;if sourceName=="dropbox"{account=j.DropboxAccountID}
  key:=relationPairKey(sourceName+":"+account+":"+j.RemoteID,d)
  // The designated first NAS is the bidirectional cloud counterpart.
  // Remaining NAS destinations are mirrors of the cloud only; they do not
  // feed changes into the cloud unless explicitly configured as a separate
  // cloud relationship. This avoids accidental NAS-to-NAS propagation.
  oneWay:=i>0&&!j.CrossNASPropagation
  primaryLabel:="Google Drive"
  if sourceName=="dropbox"{primaryLabel="Dropbox"}
  destLabel:=d.ShareID
  if share,e:=a.directShare(d.ShareID);e==nil&&share.Label!=""{destLabel=share.Label}
  restore:=a.labelSyncOperation(j,primaryLabel,destLabel)
  err=a.executePolicyPair(ctx,j,cloud,nas,key,oneWay,conflictManual)
  restore()
  if err!=nil{return err}
 }
 return nil
}

// addBridge only saves endpoint references and sync state on Syncria.
// It never creates a local file-storage path.
func relationshipOverlap(left,right NASDestination)bool{
 if left.ShareID!=right.ShareID{return false}
 a:=path.Join("/",left.Folder);b:=path.Join("/",right.Folder)
 return pathsOverlap(a,b)
}
// An intentionally designated cloud NAS may also be the primary/source
// of an independent NAS-to-NAS relationship. No other overlap is permitted.
func sameNASFolder(x,y NASDestination)bool{
 return x.ShareID==y.ShareID&&path.Clean("/"+x.Folder)==path.Clean("/"+y.Folder)
}
func allowedDesignatedNASOverlap(newJob,existing Job,folder NASDestination)bool{
 if newJob.Provider=="bridge"&&existing.Provider=="nas_bridge"{
  return len(newJob.NASDestinations)>0&&sameNASFolder(newJob.NASDestinations[0],folder)&&
   sameNASFolder(NASDestination{ShareID:existing.NASSourceID,Folder:existing.NASSourceFolder},folder)
 }
 if newJob.Provider=="nas_bridge"&&existing.Provider=="bridge"{
  existingDest:=existing.NASDestinations
  if len(existingDest)==0&&existing.NASShareID!="" {existingDest=[]NASDestination{{ShareID:existing.NASShareID,Folder:existing.NASFolder}}}
  return len(existingDest)>0&&sameNASFolder(existingDest[0],folder)&&
   sameNASFolder(NASDestination{ShareID:newJob.NASSourceID,Folder:newJob.NASSourceFolder},folder)
 }
 return false
}
func (a *App) verifyNASFolder(ctx context.Context,d NASDestination)error{
 nas,err:=a.openNAS(ctx,d);if err!=nil{return err}
 defer nas.Close()
 endpoint,ok:=nas.(*bridgeSMBEndpoint)
 if !ok{return errors.New("NAS connection cannot browse selected folder")}
 _,err=endpoint.client.share.ReadDir(endpoint.root)
 return err
}
func (a *App) addBridge(w http.ResponseWriter,r *http.Request){
 name:=strings.TrimSpace(r.FormValue("name"))
 if name==""||len(name)>128{redirect(w,r,"Relationship name must be 1–128 characters");return}
 relationship:=r.FormValue("relationship_type")
 if relationship==""{relationship="cloud"}
 if relationship!="cloud"&&relationship!="nas"{redirect(w,r,"Choose a cloud ↔ NAS or NAS ↔ NAS relationship");return}
 interval,err:=strconv.Atoi(r.FormValue("interval"))
 if err!=nil||interval<1||interval>1440{redirect(w,r,"Interval must be between 1 and 1440 minutes");return}
 mode:=r.FormValue("schedule_mode")
 if mode==""{mode="interval"}
 if mode!="interval"&&mode!="changes"{redirect(w,r,"Choose a scheduled interval or on-change NAS notifications");return}
 if mode=="changes"{
  if _,err:=exec.LookPath("smbclient");err!=nil{
   redirect(w,r,"Remote NAS change notifications require smbclient in Syncria's container")
   return
  }
 }
 var selected []NASDestination
 used:=map[string]bool{}
 for _,id:=range r.Form["nas_ids"]{
  if id==""||used[id]{continue}
  used[id]=true
  if _,err=a.directShare(id);err!=nil{redirect(w,r,"One selected NAS connection was not found");return}
  folder,err:=safeRemotePath(r.FormValue("nas_folder_"+id));if err!=nil{redirect(w,r,err.Error());return}
  d:=NASDestination{ShareID:id,Folder:folder}
  for _,other:=range selected{if relationshipOverlap(other,d){redirect(w,r,"Overlapping NAS destinations are not allowed");return}}
  selected=append(selected,d)
 }
 // Backward-compatible new relationship form submission.
 if len(selected)==0&&r.FormValue("nas_id")!=""{
  id:=r.FormValue("nas_id")
  if _,err=a.directShare(id);err!=nil{redirect(w,r,err.Error());return}
  folder,err:=safeRemotePath(r.FormValue("nas_folder"));if err!=nil{redirect(w,r,err.Error());return}
  selected=append(selected,NASDestination{ShareID:id,Folder:folder})
 }
 if len(selected)==0{redirect(w,r,"Select at least one NAS destination");return}
 j:=Job{ID:newID(),Name:name,Interval:interval,ScheduleMode:mode,Enabled:false,
  NASDestinations:selected,Direction:"bidirectional",ConflictPolicy:conflictManual,
  State:map[string]Item{},PairStates:map[string]map[string]Item{},
  PropagateDeletions:boolPtr(r.FormValue("propagate_deletions")=="1"),
  RecoveryEnabled:boolPtr(r.FormValue("recovery_enabled")=="1"),
 }
 if relationship=="nas"{
  j.Provider="nas_bridge";j.RelationshipType="nas"
  j.NASSourceID=r.FormValue("nas_source_id")
  folder,err:=safeRemotePath(r.FormValue("nas_source_folder"));if err!=nil{redirect(w,r,err.Error());return}
  j.NASSourceFolder=folder
  if _,err=a.directShare(j.NASSourceID);err!=nil{redirect(w,r,"Primary NAS connection is missing");return}
  j.Direction=r.FormValue("direction")
  if j.Direction!="one_way"&&j.Direction!="bidirectional"{redirect(w,r,"Choose one-way or bidirectional NAS sync");return}
  j.ConflictPolicy=r.FormValue("conflict_policy")
  if j.Direction=="bidirectional"&&!validNASPolicy(j.ConflictPolicy){redirect(w,r,"Choose one of the five approved conflict policies");return}
  origin:=NASDestination{ShareID:j.NASSourceID,Folder:j.NASSourceFolder}
  for _,d:=range selected{if relationshipOverlap(origin,d){redirect(w,r,"Primary NAS source cannot overlap a destination folder");return}}
 }else{
  j.Provider="bridge";j.RelationshipType="cloud"
  j.CloudProvider=r.FormValue("cloud_provider")
  if j.CloudProvider==""{j.CloudProvider="google_drive"}
  j.RemoteID=strings.TrimSpace(r.FormValue("remote"))
  switch j.CloudProvider{
  case "google_drive":
   j.AccountID=r.FormValue("account")
   if j.RemoteID==""{j.RemoteID="root"}
   if j.RemoteID!="root"&&!validDriveID(j.RemoteID){redirect(w,r,"Choose a valid Google Drive folder");return}
   a.mu.Lock();exists:=a.account(j.AccountID)!=nil;a.mu.Unlock()
   if !exists{redirect(w,r,"Google account not connected");return}
  case "dropbox":
   j.DropboxAccountID=r.FormValue("dropbox_account")
   root,err:=cleanDropboxRoot(j.RemoteID);if err!=nil{redirect(w,r,err.Error());return}
   j.RemoteID=root
   a.mu.Lock();exists:=a.dropboxAccount(j.DropboxAccountID)!=nil;a.mu.Unlock()
   if !exists{redirect(w,r,"Dropbox account not connected");return}
  default:redirect(w,r,"Select Google Drive or Dropbox");return
  }
  j.CrossNASPropagation=r.FormValue("cross_nas_propagation")=="1"
 }
 // Make sure the selected endpoint folders actually exist and are
 // accessible, without mounting or copying anything into Syncria.
 ctx,cancel:=context.WithTimeout(r.Context(),90*time.Second);defer cancel()
 if j.Provider=="nas_bridge"{
  if err:=a.verifyNASFolder(ctx,NASDestination{ShareID:j.NASSourceID,Folder:j.NASSourceFolder});err!=nil{
   redirect(w,r,"Primary NAS folder unavailable: "+err.Error());return
  }
 }else if j.CloudProvider=="google_drive" {
  if _,err:=a.folderChildren(ctx,j.AccountID,j.RemoteID);err!=nil{
   redirect(w,r,"Google Drive folder unavailable: "+err.Error());return
  }
 }else {
  if _,err:=a.dropboxList(ctx,j.DropboxAccountID,j.RemoteID,false);err!=nil{
   redirect(w,r,"Dropbox folder unavailable: "+err.Error());return
  }
 }
 for _,d:=range selected{
  if err:=a.verifyNASFolder(ctx,d);err!=nil{
   redirect(w,r,"Destination NAS folder unavailable: "+err.Error());return
  }
 }
 a.mu.Lock()
 // Overlapping storage is allowed ONLY for a cloud relationship's
 // designated NAS and the source NAS of its explicit NAS-to-NAS relationship.
 // Every other overlap is rejected to prevent simultaneous writers.
 allNew:=append([]NASDestination(nil),selected...)
 if j.Provider=="nas_bridge" {allNew=append(allNew,NASDestination{ShareID:j.NASSourceID,Folder:j.NASSourceFolder})}
 for _,existing:=range a.cfg.Jobs{
  if existing.Provider!="bridge"&&existing.Provider!="nas_bridge"{continue}
  previous:=append([]NASDestination(nil),existing.NASDestinations...)
  if len(previous)==0&&existing.NASShareID!=""{previous=append(previous,NASDestination{ShareID:existing.NASShareID,Folder:existing.NASFolder})}
  if existing.Provider=="nas_bridge"{previous=append(previous,NASDestination{ShareID:existing.NASSourceID,Folder:existing.NASSourceFolder})}
  for _,d:=range allNew {
   for _,p:=range previous{
    if !relationshipOverlap(d,p){continue}
    if sameNASFolder(d,p)&&allowedDesignatedNASOverlap(j,existing,d){continue}
    a.mu.Unlock()
    redirect(w,r,"A NAS folder overlaps an existing relationship other than its explicitly designated NAS-to-NAS source")
    return
   }
  }
 }
 a.cfg.Jobs=append(a.cfg.Jobs,j)
 err=a.save();a.mu.Unlock()
 if err!=nil{http.Error(w,"Could not save relationship",500);return}
 http.Redirect(w,r,"/?msg="+url.QueryEscape("Synchronization relationship created and paused. No files were copied.")+"#sync",http.StatusSeeOther)
}
func boolPtr(value bool)*bool{return &value}

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
  for _,j:=range a.cfg.Jobs {
   if j.NASShareID==id||j.NASSourceID==id {
    a.mu.Unlock();redirect(w,r,"This NAS is used by a synchronization relationship; remove the relationship before disconnecting");return
   }
   for _,d:=range j.NASDestinations{
    if d.ShareID==id{a.mu.Unlock();redirect(w,r,"This NAS is used by a synchronization relationship; remove the relationship before disconnecting");return}
   }
  }
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
