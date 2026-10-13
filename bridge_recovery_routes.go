package main

import (
 "context"
 "errors"
 "fmt"
 "net/http"
 "net/url"
 "strings"
 "time"
)

func (a *App) recoveryPair(ctx context.Context,j *Job,key string)(recoverableBridgeEndpoint,recoverableBridgeEndpoint,error){
 destinations:=append([]NASDestination(nil),j.NASDestinations...)
 if len(destinations)==0&&j.NASShareID!=""{destinations=append(destinations,NASDestination{ShareID:j.NASShareID,Folder:j.NASFolder})}
 for _,dest:=range destinations{
  var source recoverableBridgeEndpoint
  var e error
  var expected string
  if j.Provider=="nas_bridge"{
   expected=relationPairKey("nas:"+j.NASSourceID+":"+j.NASSourceFolder,dest)
   if expected!=key{continue}
   source,e=a.openNAS(ctx,NASDestination{ShareID:j.NASSourceID,Folder:j.NASSourceFolder})
  }else{
   provider:=j.CloudProvider;if provider==""{provider="google_drive"}
   account:=j.AccountID;if provider=="dropbox"{account=j.DropboxAccountID}
   expected=relationPairKey(provider+":"+account+":"+j.RemoteID,dest)
   if expected!=key{continue}
   source,e=a.cloudEndpoint(j)
  }
  if e!=nil{return nil,nil,e}
  target,e:=a.openNAS(ctx,dest)
  if e!=nil{source.Close();return nil,nil,e}
  return source,target,nil
 }
 return nil,nil,errors.New("Recovery relationship endpoints are unavailable")
}
func findPending(j *Job,key,rel string)(int,PendingDeletion){
 for i,p:=range j.PendingDeletions{if p.Pair==key&&p.Path==rel{return i,p}}
 return -1,PendingDeletion{}
}
func (a *App) restorePendingDeletion(ctx context.Context,id,key,rel string)error{
 if _,busy:=a.runs.Load(id);busy{return errors.New("Wait for the current sync to finish before restoring")}
 a.mu.Lock();job:=a.job(id)
 if job==nil{a.mu.Unlock();return errors.New("Relationship not found")}
 pendingIndex,pending:=findPending(job,key,rel)
 if pendingIndex<0{a.mu.Unlock();return errors.New("Recovery record not found")}
 if time.Now().After(pending.Expires){a.mu.Unlock();return errors.New("The ten-minute recovery window has expired")}
 copyJob:=*job
 copyJob.PairStates=make(map[string]map[string]Item,len(job.PairStates))
 for k,v:=range job.PairStates{
  m:=make(map[string]Item,len(v));for item,s:=range v{m[item]=s};copyJob.PairStates[k]=m
 }
 copyJob.PendingDeletions=append([]PendingDeletion(nil),job.PendingDeletions...)
 a.mu.Unlock()
 primary,secondary,err:=a.recoveryPair(ctx,&copyJob,key)
 if err!=nil{return err}
 surviving:=primary;missing:=secondary
 if pending.MissingSide=="primary"{surviving=secondary;missing=primary}
 if pending.MissingSide!="primary"&&pending.MissingSide!="secondary"{primary.Close();secondary.Close();return errors.New("Invalid recovery state")}
 // Never overwrite a newly-created file after an accidental deletion.
 existing,err:=missing.Scan(ctx)
 if err!=nil{primary.Close();secondary.Close();return err}
 if _,exists:=existing[rel];exists{primary.Close();secondary.Close();return errors.New("A new file exists at the deleted location; no restore was performed")}
 if err=surviving.Restore(ctx,rel,pending.RecoveryToken);err!=nil{primary.Close();secondary.Close();return err}
 // The recovery file is now back at its original remote location.
 // The next sync can recreate its missing counterpart; for one-way syncs
 // where the primary was deleted, explicitly restore the primary as well.
 latest,err:=surviving.Scan(ctx)
 if err!=nil{primary.Close();secondary.Close();return err}
 entry,exists:=latest[rel]
 if !exists{primary.Close();secondary.Close();return errors.New("Recovered endpoint did not expose restored item")}
 if entry.Folder{
  if _,err=missing.Mkdir(ctx,rel);err!=nil{primary.Close();secondary.Close();return err}
  // Descendants are recovered by the following normal sync.
 }else{
  _,err=bridgeCopy(ctx,surviving,missing,rel,"",nil)
  if err!=nil{primary.Close();secondary.Close();return fmt.Errorf("Recovery succeeded on one endpoint but second copy failed: %w",err)}
 }
 primary.Close();secondary.Close()
 a.mu.Lock()
 if current:=a.job(id);current!=nil{
  idx,p:=findPending(current,key,rel)
  if idx>=0&&p.RecoveryToken==pending.RecoveryToken{
   current.PendingDeletions=append(current.PendingDeletions[:idx],current.PendingDeletions[idx+1:]...)
   if current.PairStates!=nil{delete(current.PairStates,key)}
   err=a.save()
  }
 }
 a.mu.Unlock()
 return err
}
func (a *App) recoveryAction(w http.ResponseWriter,r *http.Request){
 if !localPost(r){http.Error(w,"Same-origin POST required",405);return}
 if !a.authenticated(r){http.Error(w,"Unauthorized",401);return}
 r.Body=http.MaxBytesReader(w,r.Body,8192)
 if err:=r.ParseForm();err!=nil{http.Error(w,"Invalid form",400);return}
 if r.FormValue("operation")!="undo"{http.Error(w,"Unknown recovery action",400);return}
 ctx,cancel:=context.WithTimeout(r.Context(),2*time.Minute);defer cancel()
 if err:=a.restorePendingDeletion(ctx,r.FormValue("id"),r.FormValue("pair"),r.FormValue("path"));err!=nil{
  redirect(w,r,"Recovery failed: "+err.Error());return
 }
 http.Redirect(w,r,"/?msg="+url.QueryEscape("File restored to its sync endpoints")+"#sync",303)
}
func (a *App) expireRecoveries(ctx context.Context){
 a.mu.Lock()
 jobs:=append([]Job(nil),a.cfg.Jobs...)
 a.mu.Unlock()
 for _,job:=range jobs {
  if len(job.PendingDeletions)==0{continue}
  if _,running:=a.runs.Load(job.ID);running{continue}
  for _,p:=range job.PendingDeletions{
   if time.Now().Before(p.Expires){continue}
   runCtx,cancel:=context.WithTimeout(ctx,30*time.Second)
   primary,secondary,err:=a.recoveryPair(runCtx,&job,p.Pair)
   if err!=nil{cancel();continue}
   surviving:=primary
   if p.MissingSide=="primary"{surviving=secondary}
   err=surviving.Purge(runCtx,p.RecoveryToken)
   primary.Close();secondary.Close();cancel()
   if err!=nil{continue}
   a.mu.Lock()
   actual:=a.job(job.ID)
   if actual!=nil{
    idx,current:=findPending(actual,p.Pair,p.Path)
    if idx>=0&&current.RecoveryToken==p.RecoveryToken{
     actual.PendingDeletions=append(actual.PendingDeletions[:idx],actual.PendingDeletions[idx+1:]...)
     _=a.save()
    }
   }
   a.mu.Unlock()
  }
 }
}

var _=strings.TrimSpace
