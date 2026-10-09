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
 response map[string]any
}

func cachedUpdateStatus(ctx context.Context) map[string]any {
 updateLookup.Lock()
 defer updateLookup.Unlock()
 if time.Now().Before(updateLookup.validUntil) && updateLookup.response != nil {
  return updateLookup.response
 }
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
