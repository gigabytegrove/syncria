package main

import (
 "context"
 "errors"
 "fmt"
 "io"
 "net"
 "path"
 "sort"
 "strings"
 "time"
)

func (a *App) checkpointBridge(j *Job)error{
 a.mu.Lock();defer a.mu.Unlock()
 existing:=a.job(j.ID)
 if existing==nil{return errors.New("Relationship was removed while synchronizing")}
 existing.PairStates=j.PairStates
 existing.PendingDeletions=j.PendingDeletions
 existing.Conflicts=j.Conflicts
 existing.ConflictOverrides=j.ConflictOverrides
 return a.save()
}
func (a *App) manualConflict(j *Job,key,rel string,l,r bridgeEntry)(string,string,error){
 if j.ConflictOverrides!=nil{
  selection:=j.ConflictOverrides[key+"|"+rel]
  if selection=="primary"||selection=="secondary"{return selection,selection,nil}
  if selection=="keep_both"{return "primary",conflictKeepBoth,nil}
 }
 found:=false
 for _,existing:=range j.Conflicts{if existing.Pair==key&&existing.Path==rel{found=true;break}}
 if !found{
  j.Conflicts=append(j.Conflicts,BridgeConflict{Pair:key,Path:rel,PrimaryModified:l.Modified,SecondaryModified:r.Modified})
 }
 return "", "",errors.New("Manual conflict resolution required; choose the version in the Syncria relationship details")
}
func (a *App) selectResolution(j *Job,key,rel,policy string,l,r bridgeEntry)(string,string,error){
 if policy==conflictManual{return a.manualConflict(j,key,rel,l,r)}
 direction,err:=pickConflict(policy,l,r)
 return direction,policy,err
}
func (a *App) clearConflict(j *Job,key,rel string){
 if j.ConflictOverrides!=nil{delete(j.ConflictOverrides,key+"|"+rel)}
 for i:=range j.Conflicts {
  if j.Conflicts[i].Pair==key&&j.Conflicts[i].Path==rel {
   j.Conflicts=append(j.Conflicts[:i],j.Conflicts[i+1:]...)
   return
  }
 }
}

// Exactly five user-approved policies. No implicit fallback that overwrites
// modified data is allowed for missing/unknown policies.
const (
 conflictNewer = "newer"
 conflictPrimary = "primary"
 conflictSecondary = "secondary"
 conflictKeepBoth = "keep_both"
 conflictManual = "manual"
)
func validNASPolicy(s string)bool{
 switch s{case conflictNewer,conflictPrimary,conflictSecondary,conflictKeepBoth,conflictManual:return true}
 return false
}
func enabledByDefault(flag *bool)bool{return flag==nil||*flag}
func isFirstCloudTarget(j *Job,key string)bool{
 if j.Provider!="bridge"{return false}
 provider:=j.CloudProvider;if provider==""{provider="google_drive"}
 account:=j.AccountID;if provider=="dropbox"{account=j.DropboxAccountID}
 dests:=j.NASDestinations
 if len(dests)==0&&j.NASShareID!=""{dests=[]NASDestination{{ShareID:j.NASShareID,Folder:j.NASFolder}}}
 if len(dests)==0{return false}
 return key==relationPairKey(provider+":"+account+":"+j.RemoteID,dests[0])
}
func recordCloudOrigin(j *Job,key,rel,fingerprint string){
 if !isFirstCloudTarget(j,key)||j.CrossNASPropagation{return}
 if j.CloudOriginHashes==nil{j.CloudOriginHashes=map[string]string{}}
 j.CloudOriginHashes[rel]=fingerprint
}
func suppressCrossNASTransfer(j *Job,key,rel string,cloud bridgeEntry,hasCloud bool,oneWay bool)bool{
 if !oneWay||j.Provider!="bridge"||j.CrossNASPropagation||isFirstCloudTarget(j,key){return false}
 previous:=j.CloudOriginHashes[rel]
 if previous==""{return false}
 if !hasCloud{return previous=="__deleted__"}
 return previous==cloud.Fingerprint
}

func relationPairKey(source string,dest NASDestination)string {
 return source+"|"+dest.ShareID+"|"+dest.Folder
}
func isChildOf(parent,rel string)bool{return rel==parent||strings.HasPrefix(rel,parent+"/")}
func isPending(j *Job,key,rel string)bool {
 for _,p:=range j.PendingDeletions {if p.Pair==key && isChildOf(p.Path,rel){return true}}
 return false
}

func (a *App) executePolicyPair(ctx context.Context,j *Job,primary,secondary recoverableBridgeEndpoint,key string,oneWay bool,policy string)error{
 defer primary.Close()
 defer secondary.Close()
 left,err:=primary.Scan(ctx);if err!=nil{return fmt.Errorf("Primary scan: %w",err)}
 right,err:=secondary.Scan(ctx);if err!=nil{return fmt.Errorf("Secondary scan: %w",err)}
 if j.PairStates==nil{j.PairStates=map[string]map[string]Item{}}
 prev:=j.PairStates[key];if prev==nil{prev=map[string]Item{}}
 // Build the actual-run queue from the scans already performed; never re-read
 // file contents or re-scan endpoints for the Activity view.
 plan:=syncPreviewResult{Rows:[]syncPreviewRow{},Counts:map[string]int{},ChangesOnly:true,MaxRows:2000}
 planSyncPair(&plan,j,key,left,right,oneWay,policy)
 a.setSyncProgress(j.ID,func(p *SyncProgress){
  p.Phase="processing";p.Current="";p.Processed=0;p.Total=int64(len(left)+len(right))
  p.Plan=plan.Rows;p.PlanCounts=plan.Counts;p.PlanTruncated=plan.Truncated
 })

 next:=make(map[string]Item,len(prev));for k,v:=range prev{if !ignoredSyncPath(k){next[k]=v}}
 keys:=map[string]bool{}
 for k:=range left{keys[k]=true}
 for k:=range right{keys[k]=true}
 for k:=range prev{keys[k]=true}
 ordered:=make([]string,0,len(keys));for k:=range keys{ordered=append(ordered,k)}
 sort.Slice(ordered,func(i,k int)bool{
  di,dj:=depthKey(ordered[i]),depthKey(ordered[k]);if di!=dj{return di<dj}
  return ordered[i]<ordered[k]
 })
 var countForward,countBack int64
 for i,rel:=range ordered{
  if ignoredSyncPath(rel){continue}
  if err:=ctx.Err();err!=nil{return err}
  if isPending(j,key,rel){continue}
  // Every file still participates in metadata reconciliation, but the live
 // UI must not announce an unchanged file as if it were being transferred.
 if j.itemHook!=nil{j.itemHook(int64(i),int64(len(ordered)),"")}
  l,hasL:=left[rel];r,hasR:=right[rel];prior,tracked:=prev[rel]
  if (hasL&&l.Unsupported)||(hasR&&r.Unsupported){
   if j.operationHook!=nil{j.operationHook("unsupported Google document — kept on Drive",rel,"Google Drive","No NAS copy")}
   continue
  }
  if suppressCrossNASTransfer(j,key,rel,l,hasL,oneWay){continue}
  if !hasL&&!hasR{delete(next,rel);continue}
  if hasL&&hasR&&l.Folder!=r.Folder{return fmt.Errorf("Folder/file type conflict: %s",rel)}
  if tracked&&(!hasL||!hasR){
   if !enabledByDefault(j.PropagateDeletions){continue}
   // In one-way mode, a destination-side deletion is repaired from
   // primary, never propagated backwards.
   if oneWay&&!hasR{
    if !hasL {continue}
    var copied int64
    if l.Folder{r,err=secondary.Mkdir(ctx,rel)}else{
     sum:="";if len(l.Fingerprint)==32||len(l.Fingerprint)==64{sum=l.Fingerprint}
     r,err=bridgeCopyLogged(ctx,j,primary,secondary,rel,sum,func(n int64){
      copied=n
      if j.progressHook!=nil{j.progressHook(j.LastUploads,j.LastDownloads,j.LastUploadBytes+n,j.LastDownloadBytes)}
     })
    }
    if err!=nil{return err}
    next[rel]=indexedPairItem(l,r)
    if !l.Folder{recordTransfer(j,true,copied);countForward++;if j.operationHook!=nil{j.operationHook("copy",rel,"Primary","Secondary")}}
    continue
   }
   surviving:=secondary;survivor:=r
   missingSide:="primary"
   if hasL{surviving=primary;survivor=l;missingSide="secondary"}
   if !hasL&& !hasR{delete(next,rel);continue}
   // A concurrent edit and delete is a conflict, not a safe deletion.
   previousFP:=prior.LocalMD5
   if hasL{previousFP=prior.MD5}
   if !survivor.Folder&&previousFP!=""&&survivor.Fingerprint!=previousFP{
    return fmt.Errorf("Deletion conflicts with an edit at %s; restore or resolve manually",rel)
   }
   if enabledByDefault(j.RecoveryEnabled){
    token,e:=surviving.StageDelete(ctx,rel,j.ID)
    if e!=nil{return fmt.Errorf("Deletion recovery could not be guaranteed for %s: %w",rel,e)}
    j.PendingDeletions=append(j.PendingDeletions,PendingDeletion{
     Pair:key,Path:rel,MissingSide:missingSide,RecoveryToken:token,Expires:time.Now().Add(10*time.Minute),Folder:survivor.Folder,
    })
   }else{
    if e:=surviving.Delete(ctx,rel);e!=nil{return e}
   }
   if j.operationHook!=nil{target:="Primary";if !hasL{target="Secondary"};j.operationHook("delete",rel,missingSide,target)}
   delete(next,rel)
   j.PairStates[key]=next
   // A NAS-origin deletion removes the cloud copy on the first relationship.
   // Do not mirror it to another NAS when cross-NAS propagation is off.
   // A deletion originating in cloud (hasL=false) still propagates to every
   // explicitly selected destination that has deletion propagation enabled.
   if j.Provider=="bridge"&&hasL {recordCloudOrigin(j,key,rel,"__deleted__")}
   if err=a.checkpointBridge(j);err!=nil{return err}
   continue
  }
  if (hasL&&l.Folder)||(hasR&&r.Folder){
   if !hasL{if oneWay{continue};l,err=primary.Mkdir(ctx,rel);if err!=nil{return err}}
   if !hasR{r,err=secondary.Mkdir(ctx,rel);if err!=nil{return err}}
   next[rel]=Item{ID:l.ID,Folder:true}
   continue
  }
  if !hasL{
   if oneWay {continue}
   var bytes int64
   l,err=bridgeCopyLogged(ctx,j,secondary,primary,rel,"",func(n int64){bytes=n;if j.progressHook!=nil{j.progressHook(j.LastUploads,j.LastDownloads,j.LastUploadBytes+n,j.LastDownloadBytes)}})
   if err!=nil{return err}
   recordTransfer(j,false,bytes);countBack++;if j.operationHook!=nil{j.operationHook("copy",rel,"Secondary","Primary")}
   recordCloudOrigin(j,key,rel,l.Fingerprint)
   // Persist the origin marker before another NAS relationship can observe this cloud version.
   if err=a.checkpointBridge(j);err!=nil{return err}
  }else if !hasR{
   var bytes int64
   lChecksum:=""
   if len(l.Fingerprint)==32||len(l.Fingerprint)==64{lChecksum=l.Fingerprint}
   r,err=bridgeCopyLogged(ctx,j,primary,secondary,rel,lChecksum,func(n int64){bytes=n;if j.progressHook!=nil{j.progressHook(j.LastUploads,j.LastDownloads,j.LastUploadBytes+n,j.LastDownloadBytes)}})
   if err!=nil{return err}
   recordTransfer(j,true,bytes);countForward++;if j.operationHook!=nil{j.operationHook("copy",rel,"Primary","Secondary")}
  }else if !tracked{
   same,err:=bridgeSameFirstRun(ctx,primary,secondary,rel,l,r)
   if err!=nil{return err}
   if !same{
    decision:=decideFirstRun(j,key,rel,l,r,oneWay,policy)
    if decision.Direction==""{
     if !oneWay{
      // Preserve the existing manual resolution controls when timestamps
      // cannot identify a safe winner. Do not invent a default overwrite.
      _,_,_ =a.manualConflict(j,key,rel,l,r)
      _=a.checkpointBridge(j)
     }
     return fmt.Errorf("Initial file conflict at %s: %s",rel,decision.Reason)
    }
    if err=resolveBridgeConflict(ctx,j,primary,secondary,rel,decision.Direction,decision.Method,&l,&r);err!=nil{return err}
    if decision.Direction=="primary"{countForward++}else{countBack++}
    if j.operationHook!=nil{
     target:="Secondary";if decision.Direction=="secondary"{target="Primary"}
     j.operationHook("resolve first-run conflict",rel,decision.Direction,target)
    }
    a.clearConflict(j,key,rel)
   }
  }else{
   changedPrimary,changedSecondary:=metadataChanged(prior,l,r)
   if changedPrimary&&changedSecondary{
    if oneWay&&j.Provider=="nas_bridge"{return fmt.Errorf("Both files changed at %s; one-way NAS relationship will not overwrite them",rel)}
    if oneWay{
     if err=resolveBridgeConflict(ctx,j,primary,secondary,rel,"primary",conflictPrimary,&l,&r);err!=nil{return err}
     countForward++
     if j.operationHook!=nil{j.operationHook("one-way source wins",rel,"Primary","Secondary")}
    }else{
    direction,method,resolveErr:=a.selectResolution(j,key,rel,policy,l,r)
    if resolveErr!=nil{_ =a.checkpointBridge(j);return fmt.Errorf("File conflict at %s: %w",rel,resolveErr)}
    if err=resolveBridgeConflict(ctx,j,primary,secondary,rel,direction,method,&l,&r);err!=nil{return err}
    if direction=="primary"{countForward++}else{countBack++}
    if j.operationHook!=nil{j.operationHook("resolve conflict",rel,direction,func()string{if direction=="primary"{return "Secondary"};return "Primary"}())}
    a.clearConflict(j,key,rel)
    }
   }else if changedPrimary{
    var bytes int64
    checksum:="";if len(l.Fingerprint)==32||len(l.Fingerprint)==64{checksum=l.Fingerprint}
    r,err=bridgeCopyLogged(ctx,j,primary,secondary,rel,checksum,func(n int64){bytes=n;if j.progressHook!=nil{j.progressHook(j.LastUploads,j.LastDownloads,j.LastUploadBytes+n,j.LastDownloadBytes)}})
    if err!=nil{return err}
    recordTransfer(j,true,bytes);countForward++;if j.operationHook!=nil{j.operationHook("copy",rel,"Primary","Secondary")}
   }else if changedSecondary{
    if oneWay&&j.Provider=="nas_bridge"{return fmt.Errorf("One-way NAS destination was modified at %s; no overwrite attempted",rel)}
    if oneWay{
     checksum:="";if len(l.Fingerprint)==32||len(l.Fingerprint)==64{checksum=l.Fingerprint}
     var bytes int64
     r,err=bridgeCopyLogged(ctx,j,primary,secondary,rel,checksum,func(n int64){bytes=n;if j.progressHook!=nil{j.progressHook(j.LastUploads,j.LastDownloads,j.LastUploadBytes+n,j.LastDownloadBytes)}})
     if err!=nil{return err}
     recordTransfer(j,true,bytes);countForward++
     if j.operationHook!=nil{j.operationHook("one-way source wins",rel,"Primary","Secondary")}
    }else{
    var bytes int64
    l,err=bridgeCopyLogged(ctx,j,secondary,primary,rel,"",func(n int64){bytes=n;if j.progressHook!=nil{j.progressHook(j.LastUploads,j.LastDownloads,j.LastUploadBytes,j.LastDownloadBytes+n)}})
    if err!=nil{return err}
    recordTransfer(j,false,bytes);countBack++;if j.operationHook!=nil{j.operationHook("copy",rel,"Secondary","Primary")}
   recordCloudOrigin(j,key,rel,l.Fingerprint)
   // Persist the origin marker before another NAS relationship can observe this cloud version.
   if err=a.checkpointBridge(j);err!=nil{return err}
    }
   }
  }
  next[rel]=indexedPairItem(l,r)
  j.PairStates[key]=next
  if i%20==0 {if err=a.checkpointBridge(j);err!=nil{return err}}
  if j.itemHook!=nil{j.itemHook(int64(i+1),int64(len(ordered)),"")}
 }
 j.PairStates[key]=next
 if err=a.checkpointBridge(j);err!=nil{return err}
 j.LastSummary=fmt.Sprintf("%d files copied to secondary, %d to primary; recovery window 10 minutes",countForward,countBack)
 return nil
}
// Open a transfer only after metadata comparison proves that work is needed.
// One retry is safe for NAS destinations because Put uses a fresh staging
// file and only replaces the destination after verification. Do not retry
// cloud uploads here: a lost response to a create request can cause duplicates.
func bridgeCopyLogged(ctx context.Context,j *Job,src,dst bridgeEndpoint,rel,checksum string,notify func(int64))(bridgeEntry,error){
 if j.transferHook!=nil{j.transferHook(rel)}
 attempts:=1
 if _,safe:=dst.(*bridgeSMBEndpoint);safe{attempts=3}
 var last error
 for attempt:=1;attempt<=attempts;attempt++{
  if err:=ctx.Err();err!=nil{return bridgeEntry{},err}
  result,err:=bridgeCopy(ctx,src,dst,rel,checksum,notify)
  if err==nil{return result,nil}
  last=err
  if !retryableBridgeRead(err)||attempt==attempts {break}
  wait:=time.Duration(attempt*2)*time.Second
  select{case <-ctx.Done():return bridgeEntry{},ctx.Err();case <-time.After(wait):}
 }
 return bridgeEntry{},fmt.Errorf("Transfer %q failed after %d attempt(s): %w",rel,attempts,last)
}
func retryableBridgeRead(err error)bool{
 if err==nil{return false}
 if errors.Is(err,context.Canceled){return false}
 if errors.Is(err,context.DeadlineExceeded)||errors.Is(err,io.ErrUnexpectedEOF){return true}
 var netError net.Error
 if errors.As(err,&netError)&&netError.Timeout(){return true}
 s:=strings.ToLower(err.Error())
 return strings.Contains(s,"context deadline exceeded")||strings.Contains(s,"connection reset")||strings.Contains(s,"broken pipe")
}
func pickConflict(policy string,l,r bridgeEntry)(string,error){
 switch policy {
 case conflictPrimary, conflictKeepBoth:return "primary",nil
 case conflictSecondary:return "secondary",nil
 case conflictNewer:
  if l.Modified.IsZero()||r.Modified.IsZero()||l.Modified.Equal(r.Modified){return "",errors.New("Timestamps are missing or identical; manual resolution required")}
  if l.Modified.After(r.Modified){return "primary",nil};return "secondary",nil
 case conflictManual:return "",errors.New("Manual conflict resolution required")
 default:return "",fmt.Errorf("Unrecognized conflict policy %q",policy)
 }
}
// First-run comparisons are metadata-only. Reading a 30 GB NAS file just to
// decide whether to sync is not acceptable: modification time is the agreed
// source of truth, with size acting only as a safety check for equal times.
// File contents are streamed only when an actual transfer is needed.
func bridgeSameFirstRun(ctx context.Context,primary,secondary bridgeEndpoint,rel string,l,r bridgeEntry)(bool,error){
 if err:=ctx.Err();err!=nil{return false,err}
 if l.Modified.IsZero()||r.Modified.IsZero(){
  return false,nil // no metadata match: defer to the configured first-run conflict policy
 }
 // Account for SMB/cloud timestamp precision (often one second).
 // Do not declare equal when sizes conflict, and never open the file.
 delta:=l.Modified.Sub(r.Modified)
 if delta<time.Second&&delta >-time.Second {
  if l.Size!=r.Size{
   return false,fmt.Errorf("Cannot compare %s: matching timestamps within 2 seconds but differing sizes; manual review required",rel)
  }
  return true,nil
 }
 return false,nil
}

func resolveBridgeConflict(ctx context.Context,j *Job,p,s recoverableBridgeEndpoint,rel,direction,policy string,l,r *bridgeEntry)error{
 if policy==conflictKeepBoth{
  // Retain the secondary version under an explicitly distinguishable name
  // *on the NAS itself*, not in Syncria. It becomes independently syncable
  // on the next run.
  suffix:=" (conflict-"+newID()[:8]+")"
  extension:=path.Ext(rel)
  backup:=strings.TrimSuffix(rel,extension)+suffix+extension
  data,sz,err:=s.Open(ctx,rel);if err!=nil{return err}
  _,err=s.Put(ctx,backup,data,sz,"",nil);data.Close();if err!=nil{return err}
 }
 var bytes int64
 var err error
 if direction=="primary"{
  sum:="";if len(l.Fingerprint)==32||len(l.Fingerprint)==64{sum=l.Fingerprint}
  *r,err=bridgeCopyLogged(ctx,j,p,s,rel,sum,func(n int64){bytes=n;if j.progressHook!=nil{j.progressHook(j.LastUploads,j.LastDownloads,j.LastUploadBytes+n,j.LastDownloadBytes)}})
  if err==nil{recordTransfer(j,true,bytes)}
 }else{
  *l,err=bridgeCopyLogged(ctx,j,s,p,rel,"",func(n int64){bytes=n;if j.progressHook!=nil{j.progressHook(j.LastUploads,j.LastDownloads,j.LastUploadBytes,j.LastDownloadBytes+n)}})
  if err==nil{recordTransfer(j,false,bytes)}
 }
 return err
}
