package main

import (
 "context"
 "encoding/json"
 "errors"
 "io"
 "net/http"
 "os"
 "strings"
 "time"
 "github.com/gigabytegrove/syncria/nasversion"
)

// Existing remote NAS credentials are reused. File content is streamed directly
// between NAS-side archived files and NAS-side temporary files.
type nasVersionShare struct{nas *directSMB}
func (v nasVersionShare)Stat(p string)(os.FileInfo,error){return v.nas.share.Stat(p)}
func (v nasVersionShare)ReadDir(p string)([]os.FileInfo,error){return v.nas.share.ReadDir(p)}
func (v nasVersionShare)Mkdir(p string,m os.FileMode)error{return v.nas.share.Mkdir(p,m)}
func (v nasVersionShare)Rename(a,b string)error{return v.nas.share.Rename(a,b)}
func (v nasVersionShare)Remove(p string)error{return v.nas.share.Remove(p)}
func (v nasVersionShare)Open(p string)(io.ReadCloser,error){return v.nas.share.Open(p)}
func (v nasVersionShare)Create(p string)(io.WriteCloser,error){return v.nas.share.Create(p)}
type nasVersionRemote interface{nasversion.RestoreRemote;nasversion.RetentionRemote}

func (a *App) authorizedNASVersionJob(w http.ResponseWriter,r *http.Request,jobID string)(Job,bool){
 if !a.authenticated(r){http.Error(w,"unauthorized",401);return Job{},false}
 if len(jobID)!=32{http.Error(w,"agent sync not found",404);return Job{},false}
 a.mu.Lock();existing:=a.job(jobID);var job Job
 if existing!=nil{job=*existing};a.mu.Unlock()
 if existing==nil||job.Provider!="agent_bridge"||len(job.NASDestinations)!=1{http.Error(w,"agent sync not found",404);return Job{},false}
 return job,true
}
func (a *App) openVersionNAS(ctx context.Context,job Job)(nasVersionRemote,func(),error){
 endpoint,err:=a.openNAS(ctx,job.NASDestinations[0])
 if err!=nil{return nil,nil,err}
 smb,ok:=endpoint.(*bridgeSMBEndpoint)
 if !ok{endpoint.Close();return nil,nil,errors.New("unsupported NAS endpoint")}
 return nasVersionShare{nas:smb.client},endpoint.Close,nil
}
func respondNASVersions(w http.ResponseWriter,status int,payload any){
 w.Header().Set("Content-Type","application/json; charset=utf-8")
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("X-Content-Type-Options","nosniff")
 w.WriteHeader(status);_ =json.NewEncoder(w).Encode(payload)
}

// Authenticated metadata-only history, never unrestricted browsing.
func (a *App) agentVersionAPI(w http.ResponseWriter,r *http.Request){
 w.Header().Set("Cache-Control","no-store")
 if !a.authenticated(r){http.Error(w,"unauthorized",401);return}
 if r.Method!=http.MethodGet{w.Header().Set("Allow","GET");http.Error(w,"method not allowed",405);return}
 job,ok:=a.authorizedNASVersionJob(w,r,r.URL.Query().Get("job_id"));if !ok{return}
 a.bridgeSyncMu.Lock();defer a.bridgeSyncMu.Unlock()
 ctx,cancel:=context.WithTimeout(r.Context(),30*time.Second);defer cancel()
 remote,closeFn,err:=a.openVersionNAS(ctx,job)
 if err!=nil{http.Error(w,"NAS version listing unavailable",503);return};defer closeFn()
 items,err:=nasversion.List(ctx,remote,job.NASDestinations[0].Folder,job.ID)
 if err!=nil{http.Error(w,"NAS archive listing failed",409);return}
 respondNASVersions(w,200,map[string]any{"job_id":job.ID,"versions":items})
}
type agentVersionRestoreRequest struct{
 JobID string `json:"job_id"`
 VersionID string `json:"version_id"`
 Relative string `json:"relative"`
}
type agentVersionRetentionRequest struct{
 JobID string `json:"job_id"`
 KeepPerFile int `json:"keep_per_file"`
 MaxBytes int64 `json:"max_bytes"`
 Confirm bool `json:"confirm"`
 Digest string `json:"expected_digest,omitempty"`
}
func parseNASVersionPost(w http.ResponseWriter,r *http.Request,out any)bool{
 if r.Method!=http.MethodPost{w.Header().Set("Allow","POST");http.Error(w,"method not allowed",405);return false}
 if !localPost(r){http.Error(w,"same-origin POST required",403);return false}
 if !strings.HasPrefix(r.Header.Get("Content-Type"),"application/json"){http.Error(w,"JSON required",415);return false}
 r.Body=http.MaxBytesReader(w,r.Body,4096)
 dec:=json.NewDecoder(r.Body);dec.DisallowUnknownFields()
 if err:=dec.Decode(out);err!=nil{http.Error(w,"invalid JSON",400);return false}
 var extra any
 if !errors.Is(dec.Decode(&extra),io.EOF){http.Error(w,"trailing JSON",400);return false}
 return true
}
func (a *App)agentVersionRestoreAPI(w http.ResponseWriter,r *http.Request){
 w.Header().Set("Cache-Control","no-store")
 if !a.authenticated(r){http.Error(w,"unauthorized",401);return}
 var req agentVersionRestoreRequest
 if !parseNASVersionPost(w,r,&req){return}
 job,ok:=a.authorizedNASVersionJob(w,r,req.JobID);if !ok{return}
 a.bridgeSyncMu.Lock();defer a.bridgeSyncMu.Unlock()
 ctx,cancel:=context.WithTimeout(r.Context(),60*time.Minute);defer cancel()
 remote,closeFn,err:=a.openVersionNAS(ctx,job)
 if err!=nil{http.Error(w,"NAS archive unavailable",503);return};defer closeFn()
 restored,err:=nasversion.RestoreCopy(ctx,remote,job.NASDestinations[0].Folder,job.ID,req.VersionID,req.Relative)
 if err!=nil{http.Error(w,"NAS restore failed; existing live files were not intentionally replaced",409);return}
 respondNASVersions(w,201,map[string]any{"job_id":job.ID,"restored":restored,"note":"Recovered copy has a new filename; historical and live files remain."})
}
func (a *App)agentVersionRetentionAPI(w http.ResponseWriter,r *http.Request){
 w.Header().Set("Cache-Control","no-store")
 if !a.authenticated(r){http.Error(w,"unauthorized",401);return}
 var req agentVersionRetentionRequest
 if !parseNASVersionPost(w,r,&req){return}
 job,ok:=a.authorizedNASVersionJob(w,r,req.JobID);if !ok{return}
 a.bridgeSyncMu.Lock();defer a.bridgeSyncMu.Unlock()
 ctx,cancel:=context.WithTimeout(r.Context(),2*time.Minute);defer cancel()
 remote,closeFn,err:=a.openVersionNAS(ctx,job)
 if err!=nil{http.Error(w,"NAS archive unavailable",503);return};defer closeFn()
 if !req.Confirm {
  if req.Digest!=""{http.Error(w,"preview must omit expected_digest",400);return}
  plan,e:=nasversion.PreviewRetention(ctx,remote,job.NASDestinations[0].Folder,job.ID,req.KeepPerFile,req.MaxBytes)
  if e!=nil{http.Error(w,"cannot calculate safe retention plan",409);return}
  respondNASVersions(w,200,map[string]any{"preview":plan,"requires_confirmation":true});return
 }
 if len(req.Digest)!=64{http.Error(w,"preview digest required",400);return}
 removed,e:=nasversion.ApplyRetention(ctx,remote,job.NASDestinations[0].Folder,job.ID,req.KeepPerFile,req.MaxBytes,req.Digest)
 if e!=nil{respondNASVersions(w,409,map[string]any{"error":"retention failed or archive changed","removed_before_failure":removed});return}
 respondNASVersions(w,200,map[string]any{"job_id":job.ID,"removed_archived_versions":removed})
}
