package main

import (
 "context"
 "encoding/json"
 "errors"
 "fmt"
 "net/http"
 "net/url"
 "sort"
 "strings"
 "time"

 bolt "go.etcd.io/bbolt"
)

// Drive's account-level change cursor is paired with a snapshot scoped to
// one account+remote root. A local index cannot replace discovery of remote
// changes. Changes API avoids a full Drive folder traversal for steady state.
// Tokens and snapshots are committed together; a partial scan is never saved.
type driveTreeSnapshot struct {
 Account string `json:"account"`
 Root string `json:"root"`
 RootID string `json:"root_id"`
 Cursor string `json:"cursor"`
 FullScanAt time.Time `json:"full_scan_at"`
 Files map[string]driveFile `json:"files"`
}
type driveDeltaChange struct {
 FileID string `json:"fileId"`
 Removed bool `json:"removed"`
 File driveFile `json:"file"`
 Trashed bool `json:"-"`
}
var driveTreesBucket=[]byte("drive-trees-v1")

func driveTreeKey(account,root string)string{return account+"\x00"+root}

func (idx *syncIndex) getDriveTree(account,root string)(driveTreeSnapshot,bool,error){
 var value driveTreeSnapshot
 var exists bool
 err:=idx.db.View(func(tx *bolt.Tx)error{
  b:=tx.Bucket(driveTreesBucket)
  if b==nil{return nil}
  raw:=b.Get([]byte(driveTreeKey(account,root)))
  if raw==nil{return nil}
  exists=true
  return json.Unmarshal(raw,&value)
 })
 if err!=nil{return value,false,fmt.Errorf("Drive index unreadable: %w",err)}
 if exists&&(value.Account!=account||value.Root!=root||value.Cursor==""||value.Files==nil||value.RootID==""){
  return value,false,errors.New("Drive index scope/cursor incomplete; refusing cached inventory")
 }
 return value,exists,nil
}
func (idx *syncIndex) saveDriveTree(s driveTreeSnapshot)error{
 if s.Account==""||s.Root==""||s.RootID==""||s.Cursor==""||s.Files==nil{return errors.New("Drive snapshot is incomplete")}
 raw,err:=json.Marshal(s);if err!=nil{return err}
 return idx.db.Update(func(tx *bolt.Tx)error{
  b,e:=tx.CreateBucketIfNotExists(driveTreesBucket);if e!=nil{return e}
  return b.Put([]byte(driveTreeKey(s.Account,s.Root)),raw)
 })
}

func (a *App) driveChangeCursor(ctx context.Context,account string)(string,error){
 var response struct{StartPageToken string `json:"startPageToken"`}
 if err:=a.apiJSON(ctx,account,http.MethodGet,"https://www.googleapis.com/drive/v3/changes/startPageToken",nil,&response);err!=nil{return "",err}
 if response.StartPageToken==""{return "",errors.New("Drive did not return a start page token")}
 return response.StartPageToken,nil
}
func (a *App) driveRootID(ctx context.Context,account,root string)(string,error){
 if root!="root"{return root,nil}
 var folder struct{ID string `json:"id"`}
 if err:=a.apiJSON(ctx,account,http.MethodGet,"https://www.googleapis.com/drive/v3/files/root?fields=id",nil,&folder);err!=nil{return "",err}
 if folder.ID==""{return "",errors.New("Drive root has no stable ID")}
 return folder.ID,nil
}
func (a *App) driveChanges(ctx context.Context,account,cursor string)([]driveDeltaChange,string,error){
 if cursor==""{return nil,"",errors.New("Missing Drive change cursor")}
 changes:=make([]driveDeltaChange,0,16)
 page:=cursor
 for pageNum:=0;pageNum<200;pageNum++{
  query:=url.Values{}
  query.Set("pageToken",page)
  query.Set("pageSize","1000")
  query.Set("includeRemoved","true")
  query.Set("fields","nextPageToken,newStartPageToken,changes(fileId,removed,file(id,name,mimeType,md5Checksum,parents,size,modifiedTime,trashed))")
  var response struct{
   NextPageToken string `json:"nextPageToken"`
   NewStartPageToken string `json:"newStartPageToken"`
   Changes []struct{
    FileID string `json:"fileId"`
    Removed bool `json:"removed"`
    File struct{
     driveFile
     Trashed bool `json:"trashed"`
    } `json:"file"`
   } `json:"changes"`
  }
  if err:=a.apiJSON(ctx,account,http.MethodGet,"https://www.googleapis.com/drive/v3/changes?"+query.Encode(),nil,&response);err!=nil{return nil,"",fmt.Errorf("Drive delta failed (no cursor advanced): %w",err)}
  for _,change:=range response.Changes{
   if change.FileID==""{return nil,"",errors.New("Drive change missing file identity; full scan required")}
   changes=append(changes,driveDeltaChange{FileID:change.FileID,Removed:change.Removed||change.File.Trashed,File:change.File.driveFile})
  }
  if len(changes)>5000{return nil,"",errors.New("Drive delta exceeds safe limit; full scan required")}
  if response.NextPageToken!=""{
   if response.NextPageToken==page{return nil,"",errors.New("Drive returned a repeated page cursor")}
   page=response.NextPageToken
   continue
  }
  if response.NewStartPageToken==""{return nil,"",errors.New("Drive delta lacks final cursor; full scan required")}
  return changes,response.NewStartPageToken,nil
 }
 return nil,"",errors.New("Drive change pagination exceeded safe limit; full scan required")
}

// Applying changes is pure and conservative. Folder structure, duplicate
// names, or uncertain ancestry require full inventory to keep remote deletes,
// moves, and collisions safe. Unknown file changes outside the selected root
// can be ignored, because the feed includes changes across the whole account.
func applyDriveDelta(snapshot driveTreeSnapshot,changes []driveDeltaChange)(map[string]driveFile,bool){
 next:=make(map[string]driveFile,len(snapshot.Files))
 byID:=make(map[string]string,len(snapshot.Files))
 folderIDs:=map[string]string{snapshot.RootID:""}
 for rel,f:=range snapshot.Files{
  next[rel]=f
  if f.ID==""{return nil,false}
  if _,used:=byID[f.ID];used{return nil,false}
  byID[f.ID]=rel
  if f.MimeType==folderMime{folderIDs[f.ID]=rel}
 }
 for _,change:=range changes{
  existingPath,known:=byID[change.FileID]
  old:=next[existingPath]
  if known&&old.MimeType==folderMime{return nil,false}
  if !change.Removed&&change.File.MimeType==folderMime{
   // Ignore unrelated account folders; only a folder in this subtree
   // requires rebuilding child paths and descendant membership.
   for _,parent:=range change.File.Parents{if _,inside:=folderIDs[parent];inside{return nil,false}}
   continue
  }
  // Removed can also mean loss of access, not necessarily user deletion.
  // Reinventory before planning deletions on either endpoint.
  if change.Removed && known{return nil,false}
  if known{delete(next,existingPath);delete(byID,change.FileID)}
  if change.Removed {continue}
  f:=change.File
  if f.ID==""||f.ID!=change.FileID||f.Name==""||len(f.Parents)!=1{return nil,false}
  if ignoredSyncMetadata(f.Name){continue}
  parent,inside:=folderIDs[f.Parents[0]]
  if !inside{continue} // file outside tracked root
  component:=portableDriveObjectName(f.Name,f.ID,false)
  if !safeName(component){return nil,false}
  newPath:=component
  if parent!=""{newPath=parent+"/"+component}
  if occupied,exists:=next[newPath];exists&&occupied.ID!=f.ID{return nil,false}
  // A rename can change Drive's duplicate-ID filename suffix handling.
  if known&&existingPath!=newPath{
   for _,oldFile:=range next{
    if oldFile.ID==f.ID{return nil,false}
   }
  }
  next[newPath]=f
  byID[f.ID]=newPath
 }
 // Detect case collisions on files even when unique exact names are used.
 names:=make(map[string]string,len(next))
 for p:=range next{
  folded:=strings.ToLower(p)
  if previous,exists:=names[folded];exists&&previous!=p{return nil,false}
  names[folded]=p
 }
 return next,true
}

func (a *App) remoteTreeIndexed(ctx context.Context,account,root string,onScan func(string,int64))(map[string]driveFile,error){
 if a.index==nil{return a.remoteTreeProgress(ctx,account,root,onScan)}
 cached,ok,err:=a.index.getDriveTree(account,root)
 if err!=nil{return nil,err}
 if ok&&time.Since(cached.FullScanAt)<24*time.Hour{
  changes,nextCursor,err:=a.driveChanges(ctx,account,cached.Cursor)
  if err==nil{
   next,safe:=applyDriveDelta(cached,changes)
   if safe{
    cached.Files=next;cached.Cursor=nextCursor
    if err=a.index.saveDriveTree(cached);err!=nil{return nil,fmt.Errorf("Cannot commit Drive delta to index: %w",err)}
    if onScan!=nil{onScan("Drive indexed changes checked",int64(len(changes)))}
    return next,nil
   }
  }
  // Transient network/API errors must not cause a massive fallback scan,
  // nor should an incomplete response be interpreted as deletions.
  if err!=nil&&!strings.Contains(err.Error(),"HTTP 410")&&!strings.Contains(err.Error(),"full scan required"){
   return nil,err
  }
 }
 // Take a cursor BEFORE walking the tree; concurrent changes are replayed
 // in the next delta so they cannot fall through the initial snapshot.
 cursor,err:=a.driveChangeCursor(ctx,account);if err!=nil{return nil,err}
 rootID,err:=a.driveRootID(ctx,account,root);if err!=nil{return nil,err}
 files,err:=a.remoteTreeProgress(ctx,account,root,onScan);if err!=nil{return nil,err}
 snap:=driveTreeSnapshot{Account:account,Root:root,RootID:rootID,Cursor:cursor,FullScanAt:time.Now().UTC(),Files:files}
 if err=a.index.saveDriveTree(snap);err!=nil{return nil,fmt.Errorf("Cannot commit Drive baseline: %w",err)}
 return files,nil
}

// Preview may reuse an existing Drive index and request a read-only delta,
// but must never advance its persisted cursor or write any checkpoint.
// A later live sync will replay those changes from the previous cursor.
func (a *App) remoteTreePreviewIndexed(ctx context.Context,account,root string,onScan func(string,int64))(map[string]driveFile,error){
 if a.index==nil{return a.remoteTreeProgress(ctx,account,root,onScan)}
 snapshot,exists,err:=a.index.getDriveTree(account,root)
 if err!=nil{return nil,err}
 if !exists||time.Since(snapshot.FullScanAt)>=24*time.Hour{
  return a.remoteTreeProgress(ctx,account,root,onScan)
 }
 changes,_,err:=a.driveChanges(ctx,account,snapshot.Cursor)
 if err!=nil{
  if strings.Contains(err.Error(),"HTTP 410")||strings.Contains(err.Error(),"full scan required"){
   return a.remoteTreeProgress(ctx,account,root,onScan)
  }
  return nil,err
 }
 files,safe:=applyDriveDelta(snapshot,changes)
 if !safe{return a.remoteTreeProgress(ctx,account,root,onScan)}
 if onScan!=nil{onScan("Drive read-only index",int64(len(changes)))}
 return files,nil
}

// Sorted filenames help tests ensure deterministic collision handling.
func sortedDrivePaths(files map[string]driveFile)[]string{
 names:=make([]string,0,len(files));for n:=range files{names=append(names,n)};sort.Strings(names);return names
}
