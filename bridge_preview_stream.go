package main

import (
 "context"
 "encoding/json"
 "fmt"
 "net/http"
 "time"
)

type previewScanEvent struct{
 Endpoint string `json:"endpoint"`
 Path string `json:"path"`
 Items int64 `json:"items"`
}
type previewCompletion struct{
 Result syncPreviewResult `json:"result"`
 Error string `json:"error,omitempty"`
}
func (a *App) previewStreamAPI(w http.ResponseWriter,r *http.Request){
 w.Header().Set("Cache-Control","no-store, no-transform")
 w.Header().Set("X-Accel-Buffering","no")
 if r.Method!=http.MethodGet{http.Error(w,"GET required",405);return}
 if !a.authenticated(r){http.Error(w,"Unauthorized",401);return}
 a.mu.Lock()
 job:=a.job(r.URL.Query().Get("id"))
 if job==nil{a.mu.Unlock();http.Error(w,"Sync not found",404);return}
 raw,err:=json.Marshal(job)
 a.mu.Unlock()
 if err!=nil{http.Error(w,"Unable to snapshot sync",500);return}
 var snapshot Job
 if err=json.Unmarshal(raw,&snapshot);err!=nil{http.Error(w,"Unable to read sync",500);return}
 if snapshot.Provider!="bridge"&&snapshot.Provider!="nas_bridge"{http.Error(w,"Unsupported sync type",400);return}
 flush,ok:=w.(http.Flusher);if !ok{http.Error(w,"Streaming not supported",500);return}
 w.Header().Set("Content-Type","text/event-stream; charset=utf-8")
 fmt.Fprint(w,"retry: 2000\n\n");flush.Flush()
 ctx,cancel:=context.WithTimeout(r.Context(),30*time.Minute);defer cancel()
 progress:=make(chan previewScanEvent,32)
 completed:=make(chan previewCompletion,1)
 go func(){
  last:=time.Time{}
  var endpoint string
  result,runErr:=a.previewBridgeProgress(ctx,&snapshot,func(e,path string,count int64){
   now:=time.Now()
   if e==endpoint && count!=0 && now.Sub(last)<150*time.Millisecond{return}
   endpoint=e;last=now
   select{case progress<-previewScanEvent{Endpoint:e,Path:path,Items:count}:default:}
  })
  c:=previewCompletion{Result:result}
  if runErr!=nil{c.Error=runErr.Error()}
  completed<-c
 }()
 send:=func(event string,v any)bool{
  b,e:=json.Marshal(v);if e!=nil{return false}
  if _,e=fmt.Fprintf(w,"event: %s\ndata: %s\n\n",event,b);e!=nil{return false}
  flush.Flush();return true
 }
 for{
  select{
  case <-ctx.Done():return
  case item:=<-progress:if !send("scan",item){return}
  case done:=<-completed:
   if done.Error!=""{send("failed",map[string]string{"error":done.Error})}else{send("complete",done.Result)}
   return
  }
 }
}
