package main

import (
 "bytes"
 "context"
 "encoding/json"
 "errors"
 "fmt"
 "io"
 "log"
 "net/http"
 "net/url"
 "os"
 "os/exec"
 "path/filepath"
 "sort"
 "strings"
 "time"
)

const maxComputerTransferRecords=400
const maxConcurrentComputerTransfers=3

func (a *App) computerTransferJournal()string{
 return filepath.Join(filepath.Dir(a.file),"runtime","computer-transfers.json")
}
func terminalComputerTransfer(status string)bool{
 return status=="completed"||status=="failed"||status=="canceled"||status=="interrupted"
}
func (a *App) restoreComputerTransfers()error{
 a.computerTransferMu.Lock()
 defer a.computerTransferMu.Unlock()
 a.computerTransferJobs=map[string]computerTransfer{}
 a.computerTransferCancels=map[string]context.CancelFunc{}
 a.computerTransferSlots=make(chan struct{},maxConcurrentComputerTransfers)
 data,err:=os.ReadFile(a.computerTransferJournal())
 if os.IsNotExist(err){return nil}
 if err!=nil{return err}
 var jobs []computerTransfer
 if err:=json.Unmarshal(data,&jobs);err!=nil{return fmt.Errorf("cannot parse saved transfer journal: %w",err)}
 for _,job:=range jobs {
  if job.ID==""{continue}
  if !terminalComputerTransfer(job.Status) {
   job.Status="interrupted";job.Error="The Syncria hub restarted. Original files are preserved unless a move had already completed; inspect both devices before retrying."
   job.Finished=time.Now()
  }
  a.computerTransferJobs[job.ID]=job
 }
 return a.saveComputerTransfersLocked()
}
func (a *App) saveComputerTransfersLocked()error{
 base:=filepath.Dir(a.computerTransferJournal())
 if err:=os.MkdirAll(base,0700);err!=nil{return err}
 records:=make([]computerTransfer,0,len(a.computerTransferJobs))
 for _,job:=range a.computerTransferJobs{records=append(records,job)}
 sort.Slice(records,func(i,j int)bool{return records[i].Started.After(records[j].Started)})
 if len(records)>maxComputerTransferRecords{
  keep:=records[:maxComputerTransferRecords]
  for _,drop:=range records[maxComputerTransferRecords:]{if terminalComputerTransfer(drop.Status){delete(a.computerTransferJobs,drop.ID)}}
  records=keep
 }
 data,err:=json.MarshalIndent(records,"","  ");if err!=nil{return err}
 f,err:=os.CreateTemp(base,".computer-transfers-");if err!=nil{return err}
 name:=f.Name();defer os.Remove(name)
 if err:=f.Chmod(0600);err!=nil{f.Close();return err}
 if _,err:=f.Write(data);err!=nil{f.Close();return err}
 if err:=f.Sync();err!=nil{f.Close();return err}
 if err:=f.Close();err!=nil{return err}
 return os.Rename(name,a.computerTransferJournal())
}
func (a *App) setComputerTransferStatus(id string,event computerWorkerEvent){
 a.computerTransferMu.Lock()
 defer a.computerTransferMu.Unlock()
 job,ok:=a.computerTransferJobs[id];if !ok||terminalComputerTransfer(job.Status){return}
 if event.Status!=""{job.Status=event.Status}
 if event.Total>0{job.Total=event.Total}
 if event.Bytes>=job.Bytes{job.Bytes=event.Bytes}
 if event.Error!=""{job.Error=event.Error}
 if terminalComputerTransfer(job.Status){job.Finished=time.Now()}
 a.computerTransferJobs[id]=job
 if err:=a.saveComputerTransfersLocked();err!=nil{log.Printf("Cannot save computer transfer: %v",err)}
}
func (a *App) listComputerTransfers()[]computerTransfer{
 a.computerTransferMu.Lock()
 result:=make([]computerTransfer,0,len(a.computerTransferJobs))
 for _,job:=range a.computerTransferJobs{result=append(result,job)}
 a.computerTransferMu.Unlock()
 sort.Slice(result,func(i,j int)bool{return result[i].Started.After(result[j].Started)})
 if len(result)>maxComputerTransferRecords{result=result[:maxComputerTransferRecords]}
 return result
}
func (a *App) enqueueComputerTransfer(job computerTransfer)error{
 a.computerTransferMu.Lock()
 defer a.computerTransferMu.Unlock()
 if len(a.computerTransferJobs)>=maxComputerTransferRecords{
  ids:=make([]computerTransfer,0,len(a.computerTransferJobs))
  for _,v:=range a.computerTransferJobs{if terminalComputerTransfer(v.Status){ids=append(ids,v)}}
  sort.Slice(ids,func(i,j int)bool{return ids[i].Started.Before(ids[j].Started)})
  for _,v:=range ids{if len(a.computerTransferJobs)<maxComputerTransferRecords{break};delete(a.computerTransferJobs,v.ID)}
 }
 if len(a.computerTransferJobs)>=maxComputerTransferRecords{return errors.New("transfer journal is full of active jobs")}
 ctx,cancel:=context.WithCancel(context.Background())
 a.computerTransferCancels[job.ID]=cancel
 a.computerTransferJobs[job.ID]=job
 if err:=a.saveComputerTransfersLocked();err!=nil{
  delete(a.computerTransferJobs,job.ID);delete(a.computerTransferCancels,job.ID);cancel();return err
 }
 go a.runIsolatedComputerTransfer(ctx,job.ID)
 return nil
}
func (a *App) runIsolatedComputerTransfer(ctx context.Context,id string){
 defer func(){
  a.computerTransferMu.Lock()
  delete(a.computerTransferCancels,id)
  a.computerTransferMu.Unlock()
 }()
 select{
 case a.computerTransferSlots<-struct{}{}:
  defer func(){<-a.computerTransferSlots}()
 case <-ctx.Done():
  a.setComputerTransferStatus(id,computerWorkerEvent{Status:"canceled",Error:"Canceled before transfer started"})
  return
 }
 if ctx.Err()!=nil{return}
 a.computerTransferMu.Lock();job,ok:=a.computerTransferJobs[id];a.computerTransferMu.Unlock()
 if !ok||terminalComputerTransfer(job.Status){return}
 src,hasSrc:=a.computerByID(job.SourceID)
 dst,hasDst:=a.computerByID(job.DestinationID)
 if !hasSrc||!hasDst{a.setComputerTransferStatus(id,computerWorkerEvent{Status:"failed",Error:"A computer was removed; reconnect it before retrying"});return}
 executable,err:=os.Executable()
 if err!=nil{a.setComputerTransferStatus(id,computerWorkerEvent{Status:"failed",Error:"Cannot find the transfer worker executable"});return}
 task,err:=json.Marshal(computerTransferTask{Job:job,Source:src,Destination:dst})
 if err!=nil{a.setComputerTransferStatus(id,computerWorkerEvent{Status:"failed",Error:"Cannot encode transfer request"});return}
 cmd:=exec.CommandContext(ctx,executable,"-transfer-worker")
 configureTransferWorker(cmd)
 cmd.Stdin=bytes.NewReader(task)
 stdout,err:=cmd.StdoutPipe()
 if err!=nil{a.setComputerTransferStatus(id,computerWorkerEvent{Status:"failed",Error:"Cannot open worker status channel"});return}
 var stderr bytes.Buffer
 cmd.Stderr=&stderr
 if err:=cmd.Start();err!=nil{a.setComputerTransferStatus(id,computerWorkerEvent{Status:"failed",Error:"Cannot start isolated worker: "+err.Error()});return}
 a.setComputerTransferStatus(id,computerWorkerEvent{Status:"preparing"})
 decoder:=json.NewDecoder(io.LimitReader(stdout,1<<20))
 for {
  var event computerWorkerEvent
  if err:=decoder.Decode(&event);err!=nil{break}
  switch event.Status{
  case "preparing","transferring","verified","removing source","completed","failed":
   a.setComputerTransferStatus(id,event)
  }
 }
 err=cmd.Wait()
 if ctx.Err()!=nil {
  a.setComputerTransferStatus(id,computerWorkerEvent{Status:"canceled",Error:"Transfer canceled. If the copy was already verified, it may remain on the destination."})
  return
 }
 a.computerTransferMu.Lock();current:=a.computerTransferJobs[id];a.computerTransferMu.Unlock()
 if terminalComputerTransfer(current.Status){return}
 if err!=nil{
  detail:=strings.TrimSpace(stderr.String())
  if len(detail)>500{detail=detail[:500]}
  if detail==""{detail=err.Error()}
  a.setComputerTransferStatus(id,computerWorkerEvent{Status:"failed",Error:detail})
 }else{
  a.setComputerTransferStatus(id,computerWorkerEvent{Status:"completed"})
 }
}
func (a *App) shutdownComputerTransfers(){
 a.computerTransferMu.Lock()
 defer a.computerTransferMu.Unlock()
 for _,cancel:=range a.computerTransferCancels{cancel()}
}
func (a *App) computerTransferActionAPI(w http.ResponseWriter,r *http.Request){
 if !a.authenticated(r){computerError(w,401,"Sign in to Syncria");return}
 if !localPost(r){computerError(w,405,"Same-origin POST required");return}
 var body struct{ID string `json:"id"`;Action string `json:"action"`}
 if err:=computerDecode(r,&body);err!=nil{computerError(w,400,err.Error());return}
 a.computerTransferMu.Lock()
 job,exists:=a.computerTransferJobs[body.ID]
 if !exists{a.computerTransferMu.Unlock();computerError(w,404,"Transfer not found");return}
 switch body.Action{
 case "cancel":
  if terminalComputerTransfer(job.Status){a.computerTransferMu.Unlock();computerError(w,409,"Transfer already finished");return}
  cancel:=a.computerTransferCancels[job.ID]
  job.Status="canceled";job.Error="Canceled by operator; inspect destination if any data was already copied.";job.Finished=time.Now()
  a.computerTransferJobs[job.ID]=job
  err:=a.saveComputerTransfersLocked()
  a.computerTransferMu.Unlock()
  if cancel!=nil{cancel()}
  if err!=nil{computerError(w,500,err.Error());return}
  computerJSON(w,200,map[string]any{"status":"canceled"});return
 case "retry":
  if !terminalComputerTransfer(job.Status)||job.Status=="completed"{a.computerTransferMu.Unlock();computerError(w,409,"Only failed, interrupted or canceled transfers can be retried");return}
  a.computerTransferMu.Unlock()
  source,ok:=a.computerByID(job.SourceID);if !ok{computerError(w,409,"Source computer was removed");return}
  dest,ok:=a.computerByID(job.DestinationID);if !ok{computerError(w,409,"Destination computer was removed");return}
  newJob:=job
  newJob.ID=newID();newJob.Source=source.Label;newJob.Destination=dest.Label
  newJob.Status="queued";newJob.Error="";newJob.Bytes=0;newJob.Total=0
  newJob.Started=time.Now();newJob.Finished=time.Time{}
  if err:=a.enqueueComputerTransfer(newJob);err!=nil{computerError(w,500,err.Error());return}
  computerJSON(w,202,map[string]any{"id":newJob.ID,"status":"queued"});return
 case "clear":
  if !terminalComputerTransfer(job.Status){a.computerTransferMu.Unlock();computerError(w,409,"Cannot clear an active transfer");return}
  delete(a.computerTransferJobs,job.ID)
  err:=a.saveComputerTransfersLocked()
  a.computerTransferMu.Unlock()
  if err!=nil{computerError(w,500,err.Error());return}
  computerJSON(w,200,map[string]any{"removed":true});return
 default:
  a.computerTransferMu.Unlock();computerError(w,400,"Unknown transfer action");return
 }
}
func (a *App) computerTransfersAPI(w http.ResponseWriter,r *http.Request){
 if !a.authenticated(r){computerError(w,401,"Sign in to Syncria");return}
 if r.Method=="GET"{
  records:=a.listComputerTransfers()
  // Read actual bytes written on the receiving agent, not estimated progress.
  // Limit checks to three concurrent transfers; never block the UI on offline peers.
  for i:=range records{
   if records[i].Status!="transferring"{continue}
   dst,ok:=a.computerByID(records[i].DestinationID);if !ok{continue}
   ctx,cancel:=context.WithTimeout(r.Context(),1100*time.Millisecond)
   var progress agentTransferProgress
   err:=agentGET(ctx,dst,"/v1/progress?id="+url.QueryEscape(records[i].ID),&progress)
   cancel()
   if err==nil&&progress.Bytes>=records[i].Bytes{
    if progress.Bytes>records[i].Bytes || (progress.Total>0&&records[i].Total==0){
     a.setComputerTransferStatus(records[i].ID,computerWorkerEvent{Bytes:progress.Bytes,Total:progress.Total})
    }
    records[i].Bytes=progress.Bytes
    if records[i].Total==0{records[i].Total=progress.Total}
   }
  }
  active,queued,finished:=0,0,0
  for _,j:=range records{if j.Status=="queued"{queued++}else if terminalComputerTransfer(j.Status){finished++}else{active++}}
  computerJSON(w,200,map[string]any{"transfers":records,"active":active,"queued":queued,"finished":finished,"max_concurrent":maxConcurrentComputerTransfers});return
 }
 if !localPost(r){computerError(w,405,"Same-origin POST required");return}
 var body struct{
  SourceID string `json:"source_id"`;DestinationID string `json:"destination_id"`
  SourceRoot string `json:"source_root"`;DestinationRoot string `json:"destination_root"`
  SourcePath string `json:"source_path"`;DestinationPath string `json:"destination_path"`
  Move bool `json:"move"`
 }
 if err:=computerDecode(r,&body);err!=nil{computerError(w,400,err.Error());return}
 src,found:=a.computerByID(body.SourceID);if !found{computerError(w,400,"Source computer missing");return}
 dst,found:=a.computerByID(body.DestinationID);if !found{computerError(w,400,"Destination computer missing");return}
 if src.ID==dst.ID{computerError(w,400,"Choose two different computers");return}
 if body.SourceRoot==""||body.DestinationRoot==""||body.SourcePath==""||body.DestinationPath==""{
  computerError(w,400,"Choose a source file and destination folder");return
 }
 if _,err:=agentRelative(body.SourcePath);err!=nil{computerError(w,400,err.Error());return}
 if _,err:=agentRelative(body.DestinationPath);err!=nil{computerError(w,400,err.Error());return}
 job:=computerTransfer{
  ID:newID(),SourceID:src.ID,DestinationID:dst.ID,Source:src.Label,Destination:dst.Label,
  SourceRoot:body.SourceRoot,DestinationRoot:body.DestinationRoot,
  File:body.SourcePath,Target:body.DestinationPath,Move:body.Move,
  Status:"queued",Started:time.Now(),
 }
 if err:=a.enqueueComputerTransfer(job);err!=nil{computerError(w,503,err.Error());return}
 computerJSON(w,202,map[string]any{"id":job.ID,"status":"queued"})
}
