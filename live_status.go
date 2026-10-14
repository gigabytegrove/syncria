package main

import (
 "context"
 "fmt"
 "net/http"
 "net/http/httptest"
 "time"
)

// /api/live is an authenticated, same-origin Server-Sent Events stream.
// Keep the existing /api/dashboard/status endpoint as a fallback for clients
// or proxies that don't support streaming responses.
func (a *App) liveStatus(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodGet{http.Error(w,"GET required",http.StatusMethodNotAllowed);return}
 if !a.authenticated(r){http.Error(w,"Unauthorized",http.StatusUnauthorized);return}
 flush,ok:=w.(http.Flusher)
 if !ok{http.Error(w,"Streaming unsupported",http.StatusInternalServerError);return}
 w.Header().Set("Content-Type","text/event-stream; charset=utf-8")
 w.Header().Set("Cache-Control","no-cache, no-store, no-transform")
 w.Header().Set("X-Accel-Buffering","no")
 w.Header().Set("Connection","keep-alive")
 fmt.Fprint(w,"retry: 2000\n\n")
 flush.Flush()
 ticker:=time.NewTicker(time.Second)
 defer ticker.Stop()
 // Every tick carries an authoritative snapshot including the current
 // transfer and the last 100 executed operations per relationship.
 send:=func()error{
  rec:=httptest.NewRecorder()
  a.dashboardStatus(rec,r.WithContext(context.WithValue(r.Context(),liveStreamKey{},true)))
  if rec.Code!=http.StatusOK{return fmt.Errorf("dashboard snapshot HTTP %d",rec.Code)}
  _,err:=fmt.Fprintf(w,"event: status\ndata: %s\n\n",rec.Body.Bytes())
  if err==nil{flush.Flush()}
  return err
 }
 if send()!=nil{return}
 for{
  select{
  case <-r.Context().Done():return
  case <-ticker.C:if send()!=nil{return}
  }
 }
}
type liveStreamKey struct{}
