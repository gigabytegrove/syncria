package main

import (
 "context"
 "sync"
 "time"
)

// GitHub metadata is shared across clients so opening multiple dashboard tabs
// cannot exhaust the unauthenticated GitHub API rate limit.
var updateLookup struct {
 sync.Mutex
 validUntil time.Time
 lastRefresh time.Time
 response map[string]any
}

func cachedUpdateStatus(ctx context.Context) map[string]any {
 return lookupUpdateStatus(ctx,false)
}
// A user pressing Check again should query the public release API instead of
// silently showing cached results. Throttle repeated manual checks to protect
// against GitHub's unauthenticated API rate limits.
func refreshedUpdateStatus(ctx context.Context) map[string]any {
 return lookupUpdateStatus(ctx,true)
}
func lookupUpdateStatus(ctx context.Context,force bool) map[string]any {
 updateLookup.Lock()
 defer updateLookup.Unlock()
 now:=time.Now()
 if updateLookup.response!=nil {
  if !force&&now.Before(updateLookup.validUntil) {return updateLookup.response}
  if force&&now.Sub(updateLookup.lastRefresh)<15*time.Second {return updateLookup.response}
 }
 if force {updateLookup.lastRefresh=now}
 c,cancel:=context.WithTimeout(ctx,15*time.Second)
 defer cancel()
 status:=map[string]any{"current":appVersion}
 ttl:=5*time.Minute
 tag,err:=latestTag(c)
 if err != nil {
  status["error"]=err.Error()
  ttl=10*time.Minute
 } else {
  status["latest"]=tag
  newer:=newerVersion(tag,appVersion)
  status["available"]=false
  status["pending"]=false
  if newer {
   if _,err=releaseForTag(c,tag);err!=nil {
    status["pending"]=true
    status["message"]=err.Error()
    ttl=10*time.Minute
   } else {status["available"]=true}
  }
 }
 updateLookup.response=status
 updateLookup.validUntil=time.Now().Add(ttl)
 return status
}
