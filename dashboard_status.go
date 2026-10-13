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
  Progress SyncProgress `json:"progress"`
  SuccessfulRuns int64 `json:"successful_runs"`
  FailedRuns int64 `json:"failed_runs"`
  LastDurationMillis int64 `json:"last_duration_ms"`
  LastUploads int64 `json:"last_uploads"`
  LastDownloads int64 `json:"last_downloads"`
  LastUploadBytes int64 `json:"last_upload_bytes"`
  LastDownloadBytes int64 `json:"last_download_bytes"`
  TotalUploads int64 `json:"total_uploads"`
  TotalDownloads int64 `json:"total_downloads"`
  TotalUploadBytes int64 `json:"total_upload_bytes"`
  TotalDownloadBytes int64 `json:"total_download_bytes"`
 }
 type shareState struct {
  ID string `json:"id"`
  Mounted bool `json:"mounted"`
  Direct bool `json:"direct"`
 }
 a.mu.Lock()
 jobs:=make([]jobState,0,len(a.cfg.Jobs))
 for _,j:=range a.cfg.Jobs {
  last:="Never run"
  if !j.LastRun.IsZero(){last=j.LastRun.Format("Jan 2, 2006 15:04")}
  jobs=append(jobs,jobState{ID:j.ID,Enabled:j.Enabled,Progress:a.syncProgress[j.ID],LastRun:last,Summary:j.LastSummary,Error:j.LastError,SuccessfulRuns:j.SuccessfulRuns,FailedRuns:j.FailedRuns,LastDurationMillis:j.LastDurationMillis,LastUploads:j.LastUploads,LastDownloads:j.LastDownloads,LastUploadBytes:j.LastUploadBytes,LastDownloadBytes:j.LastDownloadBytes,TotalUploads:j.TotalUploads,TotalDownloads:j.TotalDownloads,TotalUploadBytes:j.TotalUploadBytes,TotalDownloadBytes:j.TotalDownloadBytes})
 }
 shares:=append([]StorageShare(nil),a.cfg.Shares...)
 totals:=summarizeMetrics(a.cfg.Jobs)
 a.mu.Unlock()
 for i:=range jobs {_,jobs[i].Running=a.runs.Load(jobs[i].ID)}
 mounts:=make([]shareState,0,len(shares))
 for _,s:=range shares {
  direct:=s.Local==""
  mounted:=false
  if !direct{mounted=shareMounted(s.Local)}
  mounts=append(mounts,shareState{ID:s.ID,Mounted:mounted,Direct:direct})
 }
 w.Header().Set("Content-Type","application/json; charset=utf-8")
 _=json.NewEncoder(w).Encode(struct {
  Current string `json:"current"`
  Jobs []jobState `json:"jobs"`
  Shares []shareState `json:"shares"`
  Metrics DashboardTotals `json:"metrics"`
 }{appVersion,jobs,mounts,totals})
}
