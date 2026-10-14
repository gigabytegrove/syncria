package main

import (
 "context"
 "encoding/json"
 "errors"
 "fmt"
 "net/http"
 "sort"
 "time"
)

// SyncPreview describes the current scan, not a guarantee that files will
// remain unchanged between preview and execution. It never writes to a source.
type syncPreviewRow struct {
 Path string `json:"path"`
 Action string `json:"action"`
 From string `json:"from,omitempty"`
 To string `json:"to,omitempty"`
 Reason string `json:"reason"`
 PrimaryModified string `json:"primary_modified,omitempty"`
 SecondaryModified string `json:"secondary_modified,omitempty"`
 FirstRun bool `json:"first_run,omitempty"`
 Bytes int64 `json:"bytes,omitempty"`
}
type syncPreviewResult struct {
 Job string `json:"job"`
 ReadOnly bool `json:"read_only"`
 Generated time.Time `json:"generated"`
 Truncated bool `json:"truncated"`
 Rows []syncPreviewRow `json:"rows"`
 Counts map[string]int `json:"counts"`
 Warnings []string `json:"warnings"`
}
func addPreviewRow(p *syncPreviewResult,row syncPreviewRow) {
 p.Counts[row.Action]++
 if len(p.Rows)<500 {p.Rows=append(p.Rows,row)}else{p.Truncated=true}
}
func planSyncPair(p *syncPreviewResult,j *Job,key string,left,right map[string]bridgeEntry,oneWay bool,policy string){
 prev:=j.PairStates[key]
 keys:=make(map[string]bool,len(left)+len(right))
 for x:=range left{keys[x]=true};for x:=range right{keys[x]=true};for x:=range prev{keys[x]=true}
 names:=make([]string,0,len(keys));for x:=range keys{names=append(names,x)};sort.Strings(names)
 for _,name:=range names{
  if ignoredSyncPath(name){continue}
  l,hasL:=left[name];r,hasR:=right[name];old,tracked:=prev[name]
  row:=syncPreviewRow{Path:name}
  if hasL&&!l.Modified.IsZero(){row.PrimaryModified=l.Modified.UTC().Format(time.RFC3339)}
  if hasR&&!r.Modified.IsZero(){row.SecondaryModified=r.Modified.UTC().Format(time.RFC3339)}
  switch {
  case (hasL&&l.Unsupported)||(hasR&&r.Unsupported):
   row.Action="skip";row.Reason="Google-native item cannot be exported by this connector; preserved at source, no copies or deletions attempted"
  case !hasL&&!hasR:continue
  case hasL&&hasR&&l.Folder!=r.Folder:
   row.Action="blocked";row.Reason="File/folder mismatch; no overwrite is safe"
  case tracked&&(!hasL||!hasR):
   if !enabledByDefault(j.PropagateDeletions){row.Action="skip";row.Reason="Deletion sharing is disabled"
   }else if oneWay&&!hasR&&hasL{
    row.Action="copy";row.From="Primary";row.To="Secondary";row.Bytes=l.Size;row.Reason="Repair a missing destination item"
   }else{
    survivor:=r; if hasL{survivor=l}
    original:=old.LocalMD5;if hasL{original=old.MD5}
    if !survivor.Folder&&original!=""&&survivor.Fingerprint!=original {
     row.Action="blocked";row.Reason="A deletion conflicts with a newer edit"
    }else{row.Action="delete";if hasL{row.From="Primary"}else{row.From="Secondary"};row.Reason="Deletion propagation (recovery settings apply)"}
   }
  case l.Folder||r.Folder:
   if !hasL&&oneWay{row.Action="skip";row.Reason="One-way relationship ignores secondary-only folders"
   }else if !hasL{row.Action="create folder";row.From="Secondary";row.To="Primary"
   }else if !hasR{row.Action="create folder";row.From="Primary";row.To="Secondary"
   }else{row.Action="unchanged";row.Reason="Folder exists on both sides"}
  case !hasL:
   if oneWay{row.Action="skip";row.Reason="One-way relationship leaves secondary-only files alone"
   }else{row.Action="copy";row.From="Secondary";row.To="Primary";row.Bytes=r.Size;row.Reason="Missing on primary"}
  case !hasR:
   row.Action="copy";row.From="Primary";row.To="Secondary";row.Bytes=l.Size;row.Reason="Missing on secondary"
  case !tracked:
   row.FirstRun=true
   if l.Modified.IsZero()||r.Modified.IsZero(){
    if oneWay && j.Provider=="bridge" {
     // The source is explicitly authoritative in cloud-to-NAS one-way mode,
     // even when a provider does not expose modification times.
     decision:=decideFirstRun(j,key,name,l,r,oneWay,policy)
     if decision.Direction=="primary"{
      row.Action="replace";row.From="Primary";row.To="Secondary"
      row.Bytes=l.Size;row.Reason="One-way cloud source is authoritative; timestamp unavailable"
     }else{row.Action="review";row.Reason="A modification timestamp is missing; manual review required"}
    }else{
     row.Action="review";row.Reason="A modification timestamp is missing; no automatic replacement"
    }
   }else if l.Modified.Equal(r.Modified){
    if l.Size==r.Size{
     row.Action="unchanged";row.Reason="Modification times and sizes match; no file contents read"
    }else{
     row.Action="review";row.Reason="Modification times match but sizes differ; manual review required"
    }
   }else{
    decision:=decideFirstRun(j,key,name,l,r,oneWay,policy)
    if decision.Direction==""{
     row.Action="review"
     row.Reason=decision.Reason
    }else{
     row.Action="replace"
     row.From="Primary";row.To="Secondary";row.Bytes=l.Size
     if decision.Direction=="secondary"{row.From="Secondary";row.To="Primary";row.Bytes=r.Size}
     row.Reason=decision.Reason
    }
   }
  default:
   changeL:=old.MD5!=l.Fingerprint||(old.ID!=""&&l.ID!=""&&old.ID!=l.ID)
   changeR:=old.LocalMD5!=r.Fingerprint
   switch {
   case changeL&&changeR:
    if oneWay&&j.Provider=="nas_bridge"{row.Action="blocked";row.Reason="Both files changed in one-way NAS relationship"
    }else if oneWay {row.Action="replace";row.From="Primary";row.To="Secondary";row.Bytes=l.Size;row.Reason="One-way: source wins when both changed; NAS changes are not uploaded"
    }else if policy==conflictPrimary||policy==conflictSecondary{
     row.Action="replace";row.From="Primary";row.To="Secondary";if policy==conflictSecondary{row.From="Secondary";row.To="Primary"}
     row.Reason="Both changed; configured conflict preference determines winner"
    }else if policy==conflictNewer&& !l.Modified.IsZero()&&!r.Modified.IsZero()&&!l.Modified.Equal(r.Modified){
     row.Action="replace";row.From="Primary";row.To="Secondary";if r.Modified.After(l.Modified){row.From="Secondary";row.To="Primary"}
     row.Reason="Both changed; newer timestamp wins"
    }else{row.Action="review";row.Reason="Both files changed; a conflict decision is required"}
   case changeL:row.Action="replace";row.From="Primary";row.To="Secondary";row.Bytes=l.Size;row.Reason="Primary changed since last sync"
   case changeR:
    if oneWay&&j.Provider=="nas_bridge"{row.Action="blocked";row.Reason="One-way NAS destination changed; no overwrite attempted"
    }else if oneWay{row.Action="replace";row.From="Primary";row.To="Secondary";row.Bytes=l.Size;row.Reason="One-way: restore the source copy to the destination"
    }else{row.Action="replace";row.From="Secondary";row.To="Primary";row.Bytes=r.Size;row.Reason="Secondary changed since last sync"}
   default:row.Action="unchanged";row.Reason="No recorded changes"
   }
  }
  addPreviewRow(p,row)
 }
}
func (a *App) previewBridge(ctx context.Context,j *Job)(syncPreviewResult,error){return a.previewBridgeProgress(ctx,j,nil)}
func (a *App) previewBridgeProgress(ctx context.Context,j *Job,notify func(string,string,int64))(syncPreviewResult,error){
 result:=syncPreviewResult{Job:j.Name,ReadOnly:true,Generated:time.Now(),Rows:[]syncPreviewRow{},Counts:map[string]int{},Warnings:[]string{
  "This is a read-only scan, not an execution. Files can change after preview.",
  "First-run matches and transfer decisions use modification timestamps, not full-file checksums. Matching timestamps and sizes are skipped; missing or ambiguous metadata requires review.",
  "FreeFileSync database file sync.ffs_db is excluded on both sides; existing source copies remain untouched.",
 }}
 destinations:=append([]NASDestination(nil),j.NASDestinations...)
 if len(destinations)==0&&j.NASShareID!=""{destinations=[]NASDestination{{ShareID:j.NASShareID,Folder:j.NASFolder}}}
 if len(destinations)==0{return result,errors.New("No NAS destination selected")}
 // Mirror live syncBridge routing: when an enabled NAS-to-NAS job manages
 // the downstream devices, the cloud relationship only targets its first NAS.
 if j.Provider=="bridge"&&len(destinations)>1{
  a.mu.Lock()
  related:=false
  for _,other:=range a.cfg.Jobs{
   if other.ID==j.ID||other.Provider!="nas_bridge"||!other.Enabled{continue}
   for _,d:=range destinations{
    if d.ShareID==other.NASSourceID{related=true;break}
    for _,child:=range other.NASDestinations{if d.ShareID==child.ShareID{related=true;break}}
    if related{break}
   }
   if related{break}
  }
  a.mu.Unlock()
  if related{
   destinations=destinations[:1]
   result.Warnings=append(result.Warnings,"Only the designated first NAS is part of this cloud relationship; a separate NAS-to-NAS relationship handles downstream copies.")
  }
 }
 for i,d:=range destinations{
  var primary recoverableBridgeEndpoint
  var err error
  sourceKey:=""
  oneWay:=false
  policy:=conflictManual
  if j.Provider=="nas_bridge"{
   primary,err=a.openNAS(ctx,NASDestination{ShareID:j.NASSourceID,Folder:j.NASSourceFolder})
   sourceKey="nas:"+j.NASSourceID+":"+j.NASSourceFolder
   oneWay=j.Direction=="one_way";policy=j.ConflictPolicy;if policy==""{policy=conflictManual}
  }else{
   primary,err=a.cloudEndpoint(j)
   prov:=j.CloudProvider;if prov==""{prov="google_drive"}
   account:=j.AccountID;if prov=="dropbox"{account=j.DropboxAccountID}
   sourceKey=prov+":"+account+":"+j.RemoteID
   oneWay=j.Direction=="one_way"||(i>0&&!j.CrossNASPropagation)
  }
  if err!=nil{return result,fmt.Errorf("Primary connection: %w",err)}
  secondary,err:=a.openNAS(ctx,d)
  if err!=nil{primary.Close();return result,fmt.Errorf("Secondary connection: %w",err)}
  primaryName:="Google Drive"
  if j.Provider=="nas_bridge"{
   primaryName=j.NASSourceID
   if share,e:=a.directShare(j.NASSourceID);e==nil&&share.Label!=""{primaryName=share.Label}
  }else if j.CloudProvider=="dropbox"{primaryName="Dropbox"}
  secondaryName:=d.ShareID
  if share,e:=a.directShare(d.ShareID);e==nil&&share.Label!=""{secondaryName=share.Label}
  if notify!=nil{
   if ep,ok:=primary.(*bridgeDriveEndpoint);ok{ep.scanHook=func(rel string,n int64){notify(primaryName,rel,n)}}
   if ep,ok:=primary.(*bridgeSMBEndpoint);ok{ep.scanHook=func(rel string,n int64){notify(primaryName,rel,n)}}
   if ep,ok:=secondary.(*bridgeSMBEndpoint);ok{ep.scanHook=func(rel string,n int64){notify(secondaryName,rel,n)}}
   notify(primaryName,"Starting source scan",0)
  }
  left,err:=primary.Scan(ctx)
  if err!=nil{secondary.Close();primary.Close();return result,fmt.Errorf("Primary scan: %w",err)}
  if notify!=nil{notify(secondaryName,"Starting destination scan",0)}
  right,err:=secondary.Scan(ctx)
  secondary.Close();primary.Close()
  if err!=nil{return result,fmt.Errorf("Secondary scan: %w",err)}
  if notify!=nil{notify("Planning changes","Comparing files",int64(len(left)+len(right)))}
  key:=relationPairKey(sourceKey,d)
  label:=d.ShareID
  if share,e:=a.directShare(d.ShareID);e==nil&&share.Label!=""{label=share.Label}
  if d.Folder!=""{label+="/"+d.Folder}
  primaryName="Google Drive"
  if j.Provider=="nas_bridge"{
    primaryName=j.NASSourceID
    if share,e:=a.directShare(j.NASSourceID);e==nil&&share.Label!=""{primaryName=share.Label}
    if j.NASSourceFolder!=""{primaryName+="/"+j.NASSourceFolder}
  }else if j.CloudProvider=="dropbox"{primaryName="Dropbox"}
  start:=len(result.Rows)
  planSyncPair(&result,j,key,left,right,oneWay,policy)
  for k:=start;k<len(result.Rows);k++{
    row:=&result.Rows[k]
    if row.From=="Primary"{row.From=primaryName}else if row.From=="Secondary"{row.From=label}
    if row.To=="Primary"{row.To=primaryName}else if row.To=="Secondary"{row.To=label}
    row.Reason=fmt.Sprintf("[%s] %s",label,row.Reason)
  }
  if oneWay{result.Warnings=append(result.Warnings,fmt.Sprintf("%s: source %s wins; destination-side edits never overwrite it",label,primaryName))}
 }
 return result,nil
}
func (a *App) previewSyncAPI(w http.ResponseWriter,r *http.Request){
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("X-Content-Type-Options","nosniff")
 if r.Method!=http.MethodGet{http.Error(w,"GET required",405);return}
 if !a.authenticated(r){http.Error(w,"Unauthorized",401);return}
 id:=r.URL.Query().Get("id")
 a.mu.Lock()
 original:=a.job(id)
 if original==nil{a.mu.Unlock();http.Error(w,"Sync not found",404);return}
 // Snapshot state; never change the live job during preview.
 blob,_:=json.Marshal(original)
 a.mu.Unlock()
 var j Job
 if err:=json.Unmarshal(blob,&j);err!=nil{http.Error(w,"Unable to read sync configuration",500);return}
 if j.Provider!="bridge"&&j.Provider!="nas_bridge"{http.Error(w,"Preview currently supports NAS and cloud/NAS relationships",400);return}
 ctx,cancel:=context.WithTimeout(r.Context(),2*time.Minute);defer cancel()
 preview,err:=a.previewBridge(ctx,&j)
 w.Header().Set("Content-Type","application/json; charset=utf-8")
 if err!=nil{w.WriteHeader(http.StatusUnprocessableEntity);_ =json.NewEncoder(w).Encode(map[string]string{"error":err.Error()});return}
 _=json.NewEncoder(w).Encode(preview)
}
