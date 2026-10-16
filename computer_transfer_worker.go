package main

// A direct transfer has its own OS process, with its own lifecycle and memory.
// File payloads stay between the peer agents; Syncria only dispatches metadata.
// Pairing secrets cross the process boundary over private stdin, never arguments.
import (
 "context"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
 "io"
 "os"
 "time"
)

type computerTransferTask struct {
 Job computerTransfer `json:"job"`
 Source ComputerRecord `json:"source"`
 Destination ComputerRecord `json:"destination"`
}
type computerWorkerEvent struct {
 Status string `json:"status"`
 Bytes int64 `json:"bytes,omitempty"`
 Total int64 `json:"total,omitempty"`
 Error string `json:"error,omitempty"`
}
func runComputerTransferWorker()error {
 var task computerTransferTask
 decode:=json.NewDecoder(io.LimitReader(os.Stdin,32<<10))
 decode.DisallowUnknownFields()
 if err:=decode.Decode(&task);err!=nil{return fmt.Errorf("bad worker request: %w",err)}
 var rest any
 if err:=decode.Decode(&rest);err!=io.EOF{return errors.New("trailing worker request")}
 if task.Job.ID==""||task.Source.ID!=task.Job.SourceID||task.Destination.ID!=task.Job.DestinationID{
  return errors.New("invalid transfer worker identifiers")
 }
 ctx,cancel:=context.WithTimeout(context.Background(),24*time.Hour)
 defer cancel()
 encoder:=json.NewEncoder(os.Stdout)
 emit:=func(event computerWorkerEvent){_=encoder.Encode(event)}
 err:=executeComputerTransfer(ctx,task,emit)
 if err!=nil {
  emit(computerWorkerEvent{Status:"failed",Error:err.Error()})
  return err
 }
 emit(computerWorkerEvent{Status:"completed",Bytes:task.Job.Total,Total:task.Job.Total})
 return nil
}
func executeComputerTransfer(ctx context.Context,task computerTransferTask,emit func(computerWorkerEvent))error{
 job:=task.Job
 src,dst:=task.Source,task.Destination
 sourceClient,err:=pinnedComputerClient(src.Fingerprint);if err!=nil{return err}
 emit(computerWorkerEvent{Status:"preparing"})
 var grant struct{Ticket string `json:"ticket"`;Size int64 `json:"size"`}
 if err:=agentJSONPost(ctx,sourceClient,src.URL+"/v1/grant",src.Key,map[string]any{
  "root":job.SourceRoot,"path":job.File,
 },&grant);err!=nil{return fmt.Errorf("source: %w",err)}
 if grant.Ticket==""||grant.Size<0{return errors.New("source returned an invalid transfer grant")}
 emit(computerWorkerEvent{Status:"transferring",Total:grant.Size})
 destClient,err:=pinnedComputerClient(dst.Fingerprint);if err!=nil{return err}
 var result struct{Size int64 `json:"size"`;SHA256 string `json:"sha256"`}
 if err:=agentJSONPost(ctx,destClient,dst.URL+"/v1/pull",dst.Key,map[string]any{
  "root":job.DestinationRoot,"path":job.Target,
  "source_url":src.URL,"fingerprint":src.Fingerprint,"ticket":grant.Ticket,
  "size":grant.Size,"transfer_id":job.ID,
 },&result);err!=nil{return fmt.Errorf("destination: %w",err)}
 if result.Size!=grant.Size||len(result.SHA256)!=64{return errors.New("destination did not verify all received bytes")}
 if _,err=hex.DecodeString(result.SHA256);err!=nil{return errors.New("destination returned an invalid SHA-256 digest")}
 emit(computerWorkerEvent{Status:"verified",Bytes:result.Size,Total:grant.Size})
 if err:=ctx.Err();err!=nil{return err}
 if job.Move {
  emit(computerWorkerEvent{Status:"removing source",Bytes:result.Size,Total:grant.Size})
  if err:=agentJSONPost(ctx,sourceClient,src.URL+"/v1/delete",src.Key,map[string]any{
   "ticket":grant.Ticket,"sha256":result.SHA256,
  },nil);err!=nil{return fmt.Errorf("destination copy verified, but source was not deleted: %w",err)}
 }
 return nil
}
