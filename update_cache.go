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
 channel string
}

func cachedUpdateStatus(ctx context.Context) map[string]any {
 return cachedUpdateStatusForChannel(ctx,"stable")
}
func cachedUpdateStatusForChannel(ctx context.Context,channel string)map[string]any {
 return lookupUpdateStatus(ctx,false,channel)
}
// A user pressing Check again should query the public release API instead of
// silently showing cached results. Throttle repeated manual checks to protect
// against GitHub's unauthenticated API rate limits.
func refreshedUpdateStatus(ctx context.Context) map[string]any {
 return refreshedUpdateStatusForChannel(ctx,"stable")
}
func refreshedUpdateStatusForChannel(ctx context.Context,channel string)map[string]any{
 return lookupUpdateStatus(ctx,true,channel)
}
func lookupUpdateStatus(ctx context.Context,force bool,channel string) map[string]any {
 channel=normalizeUpdateChannel(channel)
 updateLookup.Lock()
 defer updateLookup.Unlock()
 now:=time.Now()
 if updateLookup.response!=nil&&(updateLookup.channel==channel||(updateLookup.channel==""&&channel=="stable")) {
  if !force&&now.Before(updateLookup.validUntil) {return updateLookup.response}
  if force&&now.Sub(updateLookup.lastRefresh)<15*time.Second {return updateLookup.response}
 }
 if force {updateLookup.lastRefresh=now}
 c,cancel:=context.WithTimeout(ctx,15*time.Second)
 defer cancel()
 status:=map[string]any{"current":appVersion,"channel":channel,"available":false,"pending":false}
 ttl:=5*time.Minute
 release,err:=latestReleaseForChannel(c,channel)
 if err!=nil {
  status["error"]=err.Error()
  ttl=10*time.Minute
 } else {
  status["latest"]=release.Tag
  if updateVersionNewer(release.Tag,appVersion) {
   if err:=verifyReleaseAssets(release);err!=nil{
    status["pending"]=true
    status["message"]=err.Error()
    ttl=10*time.Minute
   }else{status["available"]=true}
  }
 }
 updateLookup.response=status
 updateLookup.channel=channel
 updateLookup.validUntil=time.Now().Add(ttl)
 return status
}
