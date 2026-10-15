package main

import (
 "encoding/binary"
 "encoding/json"
 "errors"
 "fmt"
 "net/http"
 "strconv"
 "strings"
 "time"

 bolt "go.etcd.io/bbolt"
)

const historyRunsLimit=5000
var runHistoryBucket=[]byte("sync-run-history-v1")
var historyMetaKey=[]byte("_summary")
var historyEventsKey=[]byte("_events")

// Run history belongs to each sync, independent of the short in-memory
// Activity feed. It records only actual operations; unchanged files are silent.
type syncRunSummary struct {
 ID string `json:"id"`
 JobID string `json:"job_id"`
 Started time.Time `json:"started"`
 Ended time.Time `json:"ended,omitempty"`
 Status string `json:"status"`
 Error string `json:"error,omitempty"`
 Summary string `json:"summary,omitempty"`
 Uploads int64 `json:"uploads"`
 Downloads int64 `json:"downloads"`
 UploadBytes int64 `json:"upload_bytes"`
 DownloadBytes int64 `json:"download_bytes"`
 EventCount uint64 `json:"event_count"`
}
type syncRunHistoryResponse struct {
 Runs []syncRunSummary `json:"runs"`
 More bool `json:"more"`
}
type syncRunEventsResponse struct {
 Run syncRunSummary `json:"run"`
 Events []SyncActionEvent `json:"events"`
 More bool `json:"more"`
}
func sequenceKey(n uint64)[]byte{
 key:=make([]byte,8);binary.BigEndian.PutUint64(key,n);return key
}
func (idx *syncIndex) beginRun(jobID string)(string,error){
 if idx==nil{return "",nil}
 if jobID==""{return "",errors.New("Missing relationship ID")}
 runID:=time.Now().UTC().Format("20060102T150405.000000000")+"-"+newID()
 record:=syncRunSummary{ID:runID,JobID:jobID,Started:time.Now().UTC(),Status:"running"}
 raw,err:=json.Marshal(record);if err!=nil{return "",err}
 err=idx.db.Update(func(tx *bolt.Tx)error{
  root,e:=tx.CreateBucketIfNotExists(runHistoryBucket);if e!=nil{return e}
  job,e:=root.CreateBucketIfNotExists([]byte(jobID));if e!=nil{return e}
  run,e:=job.CreateBucket([]byte(runID));if e!=nil{return e}
  if _,e=run.CreateBucket(historyEventsKey);e!=nil{return e}
  return run.Put(historyMetaKey,raw)
 })
 if err!=nil{return "",fmt.Errorf("Cannot start sync history: %w",err)}
 return runID,nil
}
func (idx *syncIndex) appendRunEvent(jobID,runID string,event SyncActionEvent)error{
 if idx==nil||runID==""{return nil}
 if event.Path==""{return errors.New("File operation has no path")}
 return idx.db.Update(func(tx *bolt.Tx)error{
  root:=tx.Bucket(runHistoryBucket);if root==nil{return errors.New("Run history bucket missing")}
  job:=root.Bucket([]byte(jobID));if job==nil{return errors.New("Sync history missing")}
  run:=job.Bucket([]byte(runID));if run==nil{return errors.New("Run history missing")}
  events:=run.Bucket(historyEventsKey);if events==nil{return errors.New("Event journal missing")}
  var meta syncRunSummary
  if err:=json.Unmarshal(run.Get(historyMetaKey),&meta);err!=nil{return err}
  if meta.Status!="running"{return errors.New("Cannot append to completed sync run")}
  next:=meta.EventCount+1
  body,err:=json.Marshal(event);if err!=nil{return err}
  if err=events.Put(sequenceKey(next),body);err!=nil{return err}
  meta.EventCount=next
  body,err=json.Marshal(meta);if err!=nil{return err}
  return run.Put(historyMetaKey,body)
 })
}
func (idx *syncIndex) finishRun(jobID,runID,status,summary,problem string,j Job)error{
 if idx==nil||runID==""{return nil}
 return idx.db.Update(func(tx *bolt.Tx)error{
  root:=tx.Bucket(runHistoryBucket);if root==nil{return errors.New("Missing history root")}
  job:=root.Bucket([]byte(jobID));if job==nil{return errors.New("Missing sync history")}
  run:=job.Bucket([]byte(runID));if run==nil{return errors.New("Missing run")}
  var meta syncRunSummary
  if err:=json.Unmarshal(run.Get(historyMetaKey),&meta);err!=nil{return err}
  meta.Status=status;meta.Ended=time.Now().UTC();meta.Error=problem;meta.Summary=summary
  meta.Uploads=j.LastUploads;meta.Downloads=j.LastDownloads
  meta.UploadBytes=j.LastUploadBytes;meta.DownloadBytes=j.LastDownloadBytes
  body,err:=json.Marshal(meta);if err!=nil{return err}
  return run.Put(historyMetaKey,body)
 })
}
func (idx *syncIndex) listRuns(jobID string,offset,limit int)(syncRunHistoryResponse,error){
 result:=syncRunHistoryResponse{Runs:[]syncRunSummary{}}
 if idx==nil{return result,nil}
 if offset<0||offset>100000||limit<1||limit>100{return result,errors.New("Invalid pagination")}
 err:=idx.db.View(func(tx *bolt.Tx)error{
  root:=tx.Bucket(runHistoryBucket);if root==nil{return nil}
  job:=root.Bucket([]byte(jobID));if job==nil{return nil}
  cursor:=job.Cursor();seen:=0
  for k,_:=cursor.Last();k!=nil;k,_=cursor.Prev(){
   run:=job.Bucket(k);if run==nil{continue}
   if seen<offset{seen++;continue}
   if len(result.Runs)==limit{result.More=true;break}
   var record syncRunSummary
   if err:=json.Unmarshal(run.Get(historyMetaKey),&record);err!=nil{return err}
   result.Runs=append(result.Runs,record);seen++
  }
  return nil
 })
 return result,err
}
func (idx *syncIndex) readRun(jobID,runID string,offset,limit int)(syncRunEventsResponse,error){
 result:=syncRunEventsResponse{Events:[]SyncActionEvent{}}
 if idx==nil{return result,errors.New("Run history database unavailable")}
 if offset<0||offset>1000000||limit<1||limit>250{return result,errors.New("Invalid pagination")}
 err:=idx.db.View(func(tx *bolt.Tx)error{
  root:=tx.Bucket(runHistoryBucket);if root==nil{return errors.New("Run not found")}
  job:=root.Bucket([]byte(jobID));if job==nil{return errors.New("Run not found")}
  run:=job.Bucket([]byte(runID));if run==nil{return errors.New("Run not found")}
  if err:=json.Unmarshal(run.Get(historyMetaKey),&result.Run);err!=nil{return err}
  events:=run.Bucket(historyEventsKey);if events==nil{return errors.New("Events missing")}
  for i:=offset+1;i<=offset+limit;i++{
   raw:=events.Get(sequenceKey(uint64(i)))
   if raw==nil{break}
   var item SyncActionEvent
   if err:=json.Unmarshal(raw,&item);err!=nil{return err}
   result.Events=append(result.Events,item)
  }
  result.More=uint64(offset+len(result.Events))<result.Run.EventCount
  return nil
 })
 return result,err
}
// Search is performed on the persistent journal, not merely the events
// visible on the current page. Offset is relative to the filtered results.
// No filesystem or NAS reads are required.
func (idx *syncIndex) readRunFiltered(jobID,runID string,offset,limit int,query,action string)(syncRunEventsResponse,error){
 if query==""&&action==""{return idx.readRun(jobID,runID,offset,limit)}
 result:=syncRunEventsResponse{Events:[]SyncActionEvent{}}
 if idx==nil{return result,errors.New("Run history database unavailable")}
 if offset<0||offset>1000000||limit<1||limit>250||len(query)>256||len(action)>80{
  return result,errors.New("Invalid search or pagination")
 }
 query=strings.ToLower(strings.TrimSpace(query))
 err:=idx.db.View(func(tx *bolt.Tx)error{
  root:=tx.Bucket(runHistoryBucket);if root==nil{return errors.New("Run not found")}
  job:=root.Bucket([]byte(jobID));if job==nil{return errors.New("Run not found")}
  run:=job.Bucket([]byte(runID));if run==nil{return errors.New("Run not found")}
  if err:=json.Unmarshal(run.Get(historyMetaKey),&result.Run);err!=nil{return err}
  events:=run.Bucket(historyEventsKey);if events==nil{return errors.New("Event history missing")}
  seen:=0
  return events.ForEach(func(key,raw []byte)error{
   if raw==nil{return nil}
   var e SyncActionEvent
   if err:=json.Unmarshal(raw,&e);err!=nil{return err}
   if query!=""&&!strings.Contains(strings.ToLower(e.Path),query){return nil}
   if action!=""&&!strings.EqualFold(e.Action,action){return nil}
   if seen<offset{seen++;return nil}
   if len(result.Events)>=limit{result.More=true;return nil}
   result.Events=append(result.Events,e)
   seen++
   return nil
  })
 })
 return result,err
}

func (a *App) syncHistoryAPI(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodGet{http.Error(w,"GET required",http.StatusMethodNotAllowed);return}
 if !a.authenticated(r){http.Error(w,"Unauthorized",http.StatusUnauthorized);return}
 id:=r.URL.Query().Get("id")
 a.mu.Lock();known:=a.job(id)!=nil;a.mu.Unlock()
 if !known{http.Error(w,"Sync not found",http.StatusNotFound);return}
 offset:=0
 if x:=r.URL.Query().Get("offset");x!=""{
  var err error;offset,err=strconv.Atoi(x);if err!=nil{http.Error(w,"Invalid offset",400);return}
 }
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("Content-Type","application/json; charset=utf-8")
 if runID:=r.URL.Query().Get("run");runID!=""{
  query:=r.URL.Query().Get("q")
  action:=r.URL.Query().Get("action")
  output,err:=a.index.readRunFiltered(id,runID,offset,100,query,action)
  if err!=nil{http.Error(w,err.Error(),http.StatusBadRequest);return}
  _=json.NewEncoder(w).Encode(output)
  return
 }
 output,err:=a.index.listRuns(id,offset,25)
 if err!=nil{http.Error(w,err.Error(),http.StatusBadRequest);return}
 _=json.NewEncoder(w).Encode(output)
}
