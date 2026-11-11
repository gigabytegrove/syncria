package main

import (
 "encoding/json"
 "net/http"
 "strings"
)

// Channel selection is stored with the existing Syncria configuration, never
// in browser localStorage. Stable is the implicit value for every older install.
func (a *App) updateChannelAPI(w http.ResponseWriter,r *http.Request) {
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("X-Content-Type-Options","nosniff")
 if !a.authenticated(r){http.Error(w,"Unauthorized",http.StatusUnauthorized);return}
 switch r.Method {
 case http.MethodGet:
  a.mu.Lock();channel:=normalizeUpdateChannel(a.cfg.UpdateChannel);a.mu.Unlock()
  w.Header().Set("Content-Type","application/json; charset=utf-8")
  _=json.NewEncoder(w).Encode(map[string]string{"channel":channel})
 case http.MethodPost:
  if !localPost(r){http.Error(w,"Same-origin POST required",http.StatusForbidden);return}
  r.Body=http.MaxBytesReader(w,r.Body,4096)
  var req struct{Channel string `json:"channel"`}
  decoder:=json.NewDecoder(r.Body)
  decoder.DisallowUnknownFields()
  if err:=decoder.Decode(&req);err!=nil{http.Error(w,"Invalid channel payload",http.StatusBadRequest);return}
  req.Channel=strings.ToLower(strings.TrimSpace(req.Channel))
  if !validUpdateChannel(req.Channel){http.Error(w,"Choose stable, beta, or alpha",http.StatusBadRequest);return}
  a.updateMu.Lock()
  defer a.updateMu.Unlock()
  if a.updateBusy{http.Error(w,"Cannot change channels while an update is running",http.StatusConflict);return}
  a.mu.Lock()
  previous:=a.cfg.UpdateChannel
  a.cfg.UpdateChannel=req.Channel
  err:=a.save()
  if err!=nil{a.cfg.UpdateChannel=previous}
  a.mu.Unlock()
  if err!=nil{http.Error(w,"Could not save channel: "+err.Error(),http.StatusInternalServerError);return}
  w.Header().Set("Content-Type","application/json; charset=utf-8")
  _=json.NewEncoder(w).Encode(map[string]string{"channel":req.Channel})
 default:
  w.Header().Set("Allow","GET, POST")
  http.Error(w,"Method not allowed",http.StatusMethodNotAllowed)
 }
}
