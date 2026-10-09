package main

import (
 "encoding/json"
 "net/http"
)

// dashboardStatus returns a small snapshot for updating existing labels.
// It never rebuilds the page, resets forms, or initiates navigation.
func (a *App) dashboardStatus(w http.ResponseWriter,r *http.Request) {
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("X-Content-Type-Options","nosniff")
 if r.Method!=http.MethodGet {http.Error(w,"GET required",http.StatusMethodNotAllowed);return}
 if !a.authenticated(r) {http.Error(w,"Unauthorized",http.StatusUnauthorized);return}
 type jobState struct {
  ID string `json:"id"`
  Enabled bool `json:"enabled"`
  LastRun string `json:"last_run"`
  Summary string `json:"summary"`
  Error string `json:"error"`
  Running bool `json:"running"`
 }
 type shareState struct {
  ID string `json:"id"`
  Mounted bool `json:"mounted"`
 }
 a.mu.Lock()
 jobs:=make([]jobState,0,len(a.cfg.Jobs))
 for _,j:=range a.cfg.Jobs {
  last:="Never run"
  if !j.LastRun.IsZero(){last=j.LastRun.Format("Jan 2, 2006 15:04")}
  jobs=append(jobs,jobState{ID:j.ID,Enabled:j.Enabled,LastRun:last,Summary:j.LastSummary,Error:j.LastError})
 }
 shares:=append([]StorageShare(nil),a.cfg.Shares...)
 a.mu.Unlock()
 for i:=range jobs {_,jobs[i].Running=a.runs.Load(jobs[i].ID)}
 mounts:=make([]shareState,0,len(shares))
 for _,s:=range shares {mounts=append(mounts,shareState{ID:s.ID,Mounted:shareMounted(s.Local)})}
 w.Header().Set("Content-Type","application/json; charset=utf-8")
 _=json.NewEncoder(w).Encode(struct {
  Current string `json:"current"`
  Jobs []jobState `json:"jobs"`
  Shares []shareState `json:"shares"`
 }{appVersion,jobs,mounts})
}
