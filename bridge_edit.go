package main

import (
 "encoding/json"
 "net/http"
 "strings"
)

// The editor only exposes configuration references, never OAuth secrets,
// API tokens, NAS passwords, or internal checkpoint maps.
func (a *App) syncEditorConfig(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodGet{http.Error(w,"GET required",http.StatusMethodNotAllowed);return}
 if !a.authenticated(r){http.Error(w,"Unauthorized",http.StatusUnauthorized);return}
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("Content-Type","application/json; charset=utf-8")
 a.mu.Lock()
 j:=a.job(strings.TrimSpace(r.URL.Query().Get("id")))
 if j==nil{a.mu.Unlock();http.Error(w,"Sync not found",http.StatusNotFound);return}
 if j.Provider!="bridge"&&j.Provider!="nas_bridge"{a.mu.Unlock();http.Error(w,"This sync type does not support the relationship editor",http.StatusBadRequest);return}
 directions:=j.Direction
 if directions==""{directions="bidirectional"} // Existing cloud relationships defaulted to both ways.
 destinations:=append([]NASDestination(nil),j.NASDestinations...)
 if len(destinations)==0&&j.NASShareID!=""{destinations=[]NASDestination{{ShareID:j.NASShareID,Folder:j.NASFolder}}}
 response:=struct{
  ID string `json:"id"`
  Name string `json:"name"`
  RelationshipType string `json:"relationship_type"`
  CloudProvider string `json:"cloud_provider"`
  AccountID string `json:"account_id"`
  DropboxAccountID string `json:"dropbox_account_id"`
  RemoteID string `json:"remote_id"`
  NASSourceID string `json:"nas_source_id"`
  NASSourceFolder string `json:"nas_source_folder"`
  NASDestinations []NASDestination `json:"nas_destinations"`
  Direction string `json:"direction"`
  Interval int `json:"interval"`
  ScheduleMode string `json:"schedule_mode"`
  ConflictPolicy string `json:"conflict_policy"`
  InitialConflictPolicy string `json:"initial_conflict_policy"`
  CrossNASPropagation bool `json:"cross_nas_propagation"`
  PropagateDeletions bool `json:"propagate_deletions"`
  RecoveryEnabled bool `json:"recovery_enabled"`
  Enabled bool `json:"enabled"`
 }{
  ID:j.ID,Name:j.Name,RelationshipType:j.RelationshipType,
  CloudProvider:j.CloudProvider,AccountID:j.AccountID,DropboxAccountID:j.DropboxAccountID,RemoteID:j.RemoteID,
  NASSourceID:j.NASSourceID,NASSourceFolder:j.NASSourceFolder,NASDestinations:destinations,
  Direction:directions,Interval:j.Interval,ScheduleMode:j.ScheduleMode,ConflictPolicy:j.ConflictPolicy,
  InitialConflictPolicy:j.InitialConflictPolicy,CrossNASPropagation:j.CrossNASPropagation,
  PropagateDeletions:enabledByDefault(j.PropagateDeletions),RecoveryEnabled:enabledByDefault(j.RecoveryEnabled),Enabled:j.Enabled,
 }
 if response.RelationshipType==""{if j.Provider=="nas_bridge"{response.RelationshipType="nas"}else{response.RelationshipType="cloud"}}
 if response.CloudProvider==""{response.CloudProvider="google_drive"}
 if response.InitialConflictPolicy==""{response.InitialConflictPolicy=conflictNewer}
 if response.ScheduleMode==""{response.ScheduleMode="interval"}
 if response.Interval<=0{response.Interval=5}
 a.mu.Unlock()
 _=json.NewEncoder(w).Encode(response)
}

func sameSyncEndpoints(x,y Job)bool{
 cloud:=func(j Job)string{if j.Provider=="bridge" && j.CloudProvider==""{return "google_drive"};return j.CloudProvider}
 if x.Provider!=y.Provider||cloud(x)!=cloud(y)||x.AccountID!=y.AccountID||
 x.DropboxAccountID!=y.DropboxAccountID||x.RemoteID!=y.RemoteID||x.NASSourceID!=y.NASSourceID||x.NASSourceFolder!=y.NASSourceFolder{return false}
 xd:=x.NASDestinations;yd:=y.NASDestinations
 if len(xd)==0&&x.NASShareID!=""{xd=[]NASDestination{{ShareID:x.NASShareID,Folder:x.NASFolder}}}
 if len(yd)==0&&y.NASShareID!=""{yd=[]NASDestination{{ShareID:y.NASShareID,Folder:y.NASFolder}}}
 if len(xd)!=len(yd){return false}
 // Destination ordering matters: the first NAS is the cloud's designated peer.
 for i:=range xd{if !sameNASFolder(xd[i],yd[i]){return false}}
 return true
}

// Preserve established history when editing rules. If any endpoint changes,
// old pair checkpoints cannot be reused against new folders. Do not allow
// that reset without user acknowledgment, and never touch actual files here.
func carryEditedSync(newJob *Job,old Job,allowReset bool)error{
 if old.Enabled{return errSyncMustBePaused}
 if old.Provider!=newJob.Provider{return errSyncTypeImmutable}
 newJob.ID=old.ID
 newJob.Enabled=false
 newJob.LastRun=old.LastRun
 newJob.LastError=old.LastError
 newJob.LastSummary=old.LastSummary
 newJob.RecentEvents=old.RecentEvents
 newJob.SuccessfulRuns=old.SuccessfulRuns
 newJob.FailedRuns=old.FailedRuns
 newJob.LastDurationMillis=old.LastDurationMillis
 newJob.LastUploads=old.LastUploads
 newJob.LastDownloads=old.LastDownloads
 newJob.LastUploadBytes=old.LastUploadBytes
 newJob.LastDownloadBytes=old.LastDownloadBytes
 newJob.TotalUploads=old.TotalUploads
 newJob.TotalDownloads=old.TotalDownloads
 newJob.TotalUploadBytes=old.TotalUploadBytes
 newJob.TotalDownloadBytes=old.TotalDownloadBytes
 if sameSyncEndpoints(*newJob,old){
  newJob.State=old.State
  newJob.PairStates=old.PairStates
  newJob.PendingDeletions=old.PendingDeletions
  newJob.Conflicts=old.Conflicts
  newJob.ConflictOverrides=old.ConflictOverrides
  newJob.CloudOriginHashes=old.CloudOriginHashes
  return nil
 }
 if len(old.PendingDeletions)>0{return errSyncPendingRecovery}
 if !allowReset{return errSyncResetRequired}
 newJob.LastError=""
 newJob.LastSummary="Connection changed; previous file comparison history is retained in run statistics only. Run a new preview before syncing."
 newJob.RecentEvents=nil
 return nil
}
