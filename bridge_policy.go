package main

import (
 "context"
 "crypto/md5"
 "errors"
 "fmt"
 "io"
 "path"
 "sort"
 "strings"
 "time"
)

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
 next:=make(map[string]Item,len(prev));for k,v:=range prev{next[k]=v}
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
  if err:=ctx.Err();err!=nil{return err}
  if isPending(j,key,rel){continue}
  if j.itemHook!=nil{j.itemHook(int64(i),int64(len(ordered)),rel)}
  l,hasL:=left[rel];r,hasR:=right[rel];prior,tracked:=prev[rel]
  if !hasL&&!hasR{delete(next,rel);continue}
  if hasL&&hasR&&l.Folder!=r.Folder{return fmt.Errorf("Folder/file type conflict: %s",rel)}
  if tracked&&(!hasL||!hasR){
   if !enabledByDefault(j.PropagateDeletions){continue}
   // In one-way mode, a destination-side deletion is repaired from
   // primary, never propagated backwards.
   if oneWay&&!hasR{
    if !hasL {continue}
    if l.Folder{r,err=secondary.Mkdir(ctx,rel)}else{r,err=bridgeCopy(ctx,primary,secondary,rel,"",nil)}
    if err!=nil{return err}
    next[rel]=Item{ID:l.ID,MD5:l.Fingerprint,LocalMD5:r.Fingerprint,Folder:l.Folder}
    countForward++
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
   delete(next,rel)
   j.PairStates[key]=next
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
   l,err=bridgeCopy(ctx,secondary,primary,rel,"",func(n int64){bytes=n})
   if err!=nil{return err}
   recordTransfer(j,false,bytes);countBack++
  }else if !hasR{
   var bytes int64
   lChecksum:=""
   if len(l.Fingerprint)==32{lChecksum=l.Fingerprint}
   r,err=bridgeCopy(ctx,primary,secondary,rel,lChecksum,func(n int64){bytes=n})
   if err!=nil{return err}
   recordTransfer(j,true,bytes);countForward++
  }else if !tracked{
   same,err:=bridgeSameStreams(ctx,primary,secondary,rel)
   if err!=nil{return err}
   if !same{
    if oneWay{return fmt.Errorf("Existing different files at %s: choose a safe resolution",rel)}
    direction,err:=pickConflict(policy,l,r)
    if err!=nil{return fmt.Errorf("Initial file conflict at %s: %w",rel,err)}
    if err=resolveBridgeConflict(ctx,j,primary,secondary,rel,direction,policy,&l,&r);err!=nil{return err}
   }
  }else{
   changedPrimary:=prior.MD5!=l.Fingerprint|| (prior.ID!=""&&l.ID!=""&&prior.ID!=l.ID)
   changedSecondary:=prior.LocalMD5!=r.Fingerprint
   if changedPrimary&&changedSecondary{
    if oneWay{return fmt.Errorf("Both files changed at %s; one-way relationship will not overwrite them",rel)}
    direction,err:=pickConflict(policy,l,r)
    if err!=nil{return fmt.Errorf("File conflict at %s: %w",rel,err)}
    if err=resolveBridgeConflict(ctx,j,primary,secondary,rel,direction,policy,&l,&r);err!=nil{return err}
   }else if changedPrimary{
    var bytes int64
    checksum:="";if len(l.Fingerprint)==32{checksum=l.Fingerprint}
    r,err=bridgeCopy(ctx,primary,secondary,rel,checksum,func(n int64){bytes=n})
    if err!=nil{return err}
    recordTransfer(j,true,bytes);countForward++
   }else if changedSecondary{
    if oneWay{return fmt.Errorf("One-way destination was modified at %s; no overwrite attempted",rel)}
    var bytes int64
    l,err=bridgeCopy(ctx,secondary,primary,rel,"",func(n int64){bytes=n})
    if err!=nil{return err}
    recordTransfer(j,false,bytes);countBack++
   }
  }
  next[rel]=Item{ID:l.ID,MD5:l.Fingerprint,LocalMD5:r.Fingerprint}
  j.PairStates[key]=next
  if j.itemHook!=nil{j.itemHook(int64(i+1),int64(len(ordered)),rel)}
 }
 j.PairStates[key]=next
 j.LastSummary=fmt.Sprintf("%d files copied to secondary, %d to primary; recovery window 10 minutes",countForward,countBack)
 return nil
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
func bridgeSameStreams(ctx context.Context,a,b bridgeEndpoint,rel string)(bool,error){
 ar,as,err:=a.Open(ctx,rel);if err!=nil{return false,err};defer ar.Close()
 br,bs,err:=b.Open(ctx,rel);if err!=nil{return false,err};defer br.Close()
 if as!=bs{return false,nil}
 ah:=md5.New();bh:=md5.New()
 if _,err=io.Copy(ah,ar);err!=nil{return false,err}
 if _,err=io.Copy(bh,br);err!=nil{return false,err}
 return string(ah.Sum(nil))==string(bh.Sum(nil)),nil
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
  sum:="";if len(l.Fingerprint)==32{sum=l.Fingerprint}
  *r,err=bridgeCopy(ctx,p,s,rel,sum,func(n int64){bytes=n})
  if err==nil{recordTransfer(j,true,bytes)}
 }else{
  *l,err=bridgeCopy(ctx,s,p,rel,"",func(n int64){bytes=n})
  if err==nil{recordTransfer(j,false,bytes)}
 }
 return err
}
