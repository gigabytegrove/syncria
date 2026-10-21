package main

import (
 "context"
 "sync"
 "time"
)

type channelUpdateCache struct{validUntil,lastRefresh time.Time;response map[string]any}
// Keep the stable cache fields for existing tests and legacy API consumers.
var updateLookup struct {
 sync.Mutex
 validUntil time.Time
 lastRefresh time.Time
 response map[string]any
 channels map[string]channelUpdateCache
}
func cachedUpdateStatus(ctx context.Context)map[string]any{return lookupUpdateStatus(ctx,false)}
func refreshedUpdateStatus(ctx context.Context)map[string]any{return lookupUpdateStatus(ctx,true)}
func lookupUpdateStatus(ctx context.Context,force bool)map[string]any{return channelUpdateStatus(ctx,"stable",force)}
func channelUpdateStatus(ctx context.Context,channel string,force bool)map[string]any{
 if !validUpdateChannel(channel){return map[string]any{"error":"invalid release channel"}}
 updateLookup.Lock();defer updateLookup.Unlock()
 now:=time.Now()
 var cached channelUpdateCache
 if channel=="stable"{
  cached=channelUpdateCache{updateLookup.validUntil,updateLookup.lastRefresh,updateLookup.response}
 }else{cached=updateLookup.channels[channel]}
 if cached.response!=nil{
  if !force&&now.Before(cached.validUntil){return cached.response}
  if force&&now.Sub(cached.lastRefresh)<15*time.Second{return cached.response}
 }
 if force{cached.lastRefresh=now}
 ctx,cancel:=context.WithTimeout(ctx,15*time.Second);defer cancel()
 status:=map[string]any{"current":appVersion,"channel":channel,"available":false,"pending":false}
 ttl:=5*time.Minute
 release,err:=latestChannelRelease(ctx,channel)
 if err!=nil{
  status["error"]=err.Error();ttl=10*time.Minute
 }else{
  status["latest"]=release.Tag
  if newerVersion(release.Tag,appVersion){
   if err=requireUpdateAssets(release);err!=nil{
    status["pending"]=true;status["message"]=err.Error();ttl=10*time.Minute
   }else{status["available"]=true}
  }
 }
 cached.response=status
 cached.validUntil=time.Now().Add(ttl)
 if channel=="stable"{
  updateLookup.response=cached.response
  updateLookup.validUntil=cached.validUntil
  updateLookup.lastRefresh=cached.lastRefresh
 }else{
  if updateLookup.channels==nil{updateLookup.channels=map[string]channelUpdateCache{}}
  updateLookup.channels[channel]=cached
 }
 return status
}
