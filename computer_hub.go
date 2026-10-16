package main

import (
 "context"
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
 "io"
 "net/http"
 "net/url"
 "strings"
 "sync"
 "time"
)

type ComputerRecord struct {
 ID string `json:"id"`
 Label string `json:"label"`
 URL string `json:"url"`
 Key string `json:"key"`
 Fingerprint string `json:"fingerprint"`
}
type computerPublic struct {
 ID string `json:"id"`
 Label string `json:"label"`
 URL string `json:"url"`
 Online bool `json:"online"`
 Error string `json:"error,omitempty"`
 Roots []agentRootInfo `json:"roots"`
}
type computerTransfer struct {
 ID string `json:"id"`
 Source string `json:"source"`
 Destination string `json:"destination"`
 File string `json:"file"`
 Target string `json:"target"`
 Move bool `json:"move"`
 Status string `json:"status"`
 Error string `json:"error,omitempty"`
 Bytes int64 `json:"bytes"`
 Started time.Time `json:"started"`
 Finished time.Time `json:"finished,omitempty"`
}
var computerTransferLock sync.Mutex
var computerTransfers=map[string]computerTransfer{}

func (a *App) computerByID(id string)(ComputerRecord,bool){
 a.mu.Lock();defer a.mu.Unlock()
 for _,v:=range a.cfg.Computers {if v.ID==id{return v,true}}
 return ComputerRecord{},false
}
func agentGET(ctx context.Context,rec ComputerRecord,route string,result any)error {
 c,e:=pinnedComputerClient(rec.Fingerprint);if e!=nil{return e}
 req,e:=http.NewRequestWithContext(ctx,"GET",rec.URL+route,nil);if e!=nil{return e}
 req.Header.Set("Authorization","Bearer "+rec.Key)
 resp,e:=c.Do(req);if e!=nil{return e}
 defer resp.Body.Close()
 if resp.StatusCode!=200 {
  b,_:=io.ReadAll(io.LimitReader(resp.Body,1024))
  return fmt.Errorf("agent HTTP %d: %s",resp.StatusCode,strings.TrimSpace(string(b)))
 }
 return json.NewDecoder(io.LimitReader(resp.Body,2<<20)).Decode(result)
}
func (a *App) computersAPI(w http.ResponseWriter,r *http.Request){
 if !a.authenticated(r){computerError(w,401,"Sign in to Syncria");return}
 if r.Method!="GET"{computerError(w,405,"GET required");return}
 a.mu.Lock();configured:=append([]ComputerRecord(nil),a.cfg.Computers...);a.mu.Unlock()
 output:=make([]computerPublic,len(configured))
 var wg sync.WaitGroup
 for i,rec:=range configured{
  output[i]=computerPublic{ID:rec.ID,Label:rec.Label,URL:rec.URL,Roots:[]agentRootInfo{}}
  wg.Add(1)
  go func(i int,rec ComputerRecord){
   defer wg.Done()
   ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second);defer cancel()
   var status struct{Name string `json:"name"`;Roots []agentRootInfo `json:"roots"`}
   if e:=agentGET(ctx,rec,"/v1/status",&status);e!=nil{output[i].Error=e.Error();return}
   output[i].Online=true;output[i].Roots=status.Roots
  }(i,rec)
 }
 wg.Wait()
 computerJSON(w,200,map[string]any{"computers":output})
}
func (a *App) computerRegisterAPI(w http.ResponseWriter,r *http.Request){
 if !localPost(r){computerError(w,405,"Same-origin POST required");return}
 if !a.authenticated(r){computerError(w,401,"Sign in to Syncria");return}
 var body struct{ Label string `json:"label"`; URL string `json:"url"`;Key string `json:"key"`;Fingerprint string `json:"fingerprint"` }
 if e:=computerDecode(r,&body);e!=nil{computerError(w,400,e.Error());return}
 body.Label=strings.TrimSpace(body.Label);body.URL=strings.TrimSuffix(strings.TrimSpace(body.URL),"/")
 if len(body.Label)==0||len(body.Label)>80||len(body.Key)<32||len(body.Key)>256{computerError(w,400,"Enter a computer name and its complete pairing key");return}
 if _,e:=validAgentURL(body.URL);e!=nil{computerError(w,400,e.Error());return}
 body.Fingerprint=strings.ToLower(strings.ReplaceAll(strings.TrimSpace(body.Fingerprint),":",""))
 if _,e:=pinnedComputerClient(body.Fingerprint);e!=nil{computerError(w,400,e.Error());return}
 rec:=ComputerRecord{ID:newID(),Label:body.Label,URL:body.URL,Key:body.Key,Fingerprint:body.Fingerprint}
 ctx,cancel:=context.WithTimeout(r.Context(),10*time.Second);defer cancel()
 var result struct{Roots []agentRootInfo `json:"roots"`}
 if e:=agentGET(ctx,rec,"/v1/status",&result);e!=nil{computerError(w,400,"Cannot securely connect to that agent: "+e.Error());return}
 if len(result.Roots)==0{computerError(w,400,"Agent has no shared folders");return}
 a.mu.Lock()
 defer a.mu.Unlock()
 for _,existing:=range a.cfg.Computers{if existing.URL==rec.URL{computerError(w,409,"This computer is already connected");return}}
 a.cfg.Computers=append(a.cfg.Computers,rec)
 if e:=a.save();e!=nil{a.cfg.Computers=a.cfg.Computers[:len(a.cfg.Computers)-1];computerError(w,500,"Could not save the computer");return}
 computerJSON(w,201,map[string]any{"id":rec.ID,"label":rec.Label})
}
func (a *App) computerRemoveAPI(w http.ResponseWriter,r *http.Request){
 if !localPost(r){computerError(w,405,"Same-origin POST required");return}
 if !a.authenticated(r){computerError(w,401,"Sign in to Syncria");return}
 var body struct{ID string `json:"id"`}
 if e:=computerDecode(r,&body);e!=nil{computerError(w,400,e.Error());return}
 a.mu.Lock();defer a.mu.Unlock()
 for i,c:=range a.cfg.Computers{if c.ID==body.ID{
  old:=append([]ComputerRecord(nil),a.cfg.Computers...)
  a.cfg.Computers=append(a.cfg.Computers[:i],a.cfg.Computers[i+1:]...)
  if e:=a.save();e!=nil{a.cfg.Computers=old;computerError(w,500,e.Error());return}
  computerJSON(w,200,map[string]any{"removed":true});return
 }}
 computerError(w,404,"Computer not found")
}
func (a *App) computerBrowseAPI(w http.ResponseWriter,r *http.Request){
 if r.Method!="GET"{computerError(w,405,"GET required");return}
 if !a.authenticated(r){computerError(w,401,"Sign in to Syncria");return}
 rec,ok:=a.computerByID(r.URL.Query().Get("id"));if !ok{computerError(w,404,"Computer not found");return}
 root,rel:=r.URL.Query().Get("root"),r.URL.Query().Get("path")
 if _,e:=agentRelative(rel);e!=nil{computerError(w,400,e.Error());return}
 ctx,cancel:=context.WithTimeout(r.Context(),15*time.Second);defer cancel()
 var result struct{
  Root string `json:"root"`; Path string `json:"path"`; Parent string `json:"parent"`; Entries []agentFile `json:"entries"`
 }
 query:=url.Values{"root":{root},"path":{rel}}
 if e:=agentGET(ctx,rec,"/v1/list?"+query.Encode(),&result);e!=nil{computerError(w,502,e.Error());return}
 computerJSON(w,200,result)
}
func (a *App) computerTransfersAPI(w http.ResponseWriter,r *http.Request){
 if !a.authenticated(r){computerError(w,401,"Sign in to Syncria");return}
 if r.Method=="GET"{
  computerTransferLock.Lock();result:=make([]computerTransfer,0,len(computerTransfers))
  for _,job:=range computerTransfers{result=append(result,job)}
  computerTransferLock.Unlock()
  // Browser displays newest-first.
  for i:=0;i<len(result);i++{for j:=i+1;j<len(result);j++{if result[j].Started.After(result[i].Started){result[i],result[j]=result[j],result[i]}}}
  if len(result)>100{result=result[:100]}
  computerJSON(w,200,map[string]any{"transfers":result});return
 }
 if !localPost(r){computerError(w,405,"Same-origin POST required");return}
 var body struct{
  SourceID string `json:"source_id"`; DestinationID string `json:"destination_id"`
  SourceRoot string `json:"source_root"`;DestinationRoot string `json:"destination_root"`
  SourcePath string `json:"source_path"`;DestinationPath string `json:"destination_path"`
  Move bool `json:"move"`
 }
 if e:=computerDecode(r,&body);e!=nil{computerError(w,400,e.Error());return}
 source,ok:=a.computerByID(body.SourceID);if !ok{computerError(w,400,"Source computer missing");return}
 dest,ok:=a.computerByID(body.DestinationID);if !ok{computerError(w,400,"Destination computer missing");return}
 if source.ID==dest.ID{computerError(w,400,"Select two different computers");return}
 if body.SourcePath==""||body.DestinationPath==""{computerError(w,400,"Select a file and destination folder");return}
 if _,e:=agentRelative(body.SourcePath);e!=nil{computerError(w,400,e.Error());return}
 if _,e:=agentRelative(body.DestinationPath);e!=nil{computerError(w,400,e.Error());return}
 transfer:=computerTransfer{ID:newID(),Source:source.Label,Destination:dest.Label,File:body.SourcePath,Target:body.DestinationPath,Move:body.Move,Status:"preparing",Started:time.Now()}
 computerTransferLock.Lock()
 if len(computerTransfers)>500{
  for id,j:=range computerTransfers{if j.Status=="completed"||j.Status=="failed"{delete(computerTransfers,id)}}
 }
 computerTransfers[transfer.ID]=transfer;computerTransferLock.Unlock()
 // Background job continues if the browser closes. The response contains a real job ID.
 go runComputerTransfer(source,dest,body.SourceRoot,body.DestinationRoot,body.SourcePath,body.DestinationPath,transfer)
 computerJSON(w,202,map[string]any{"id":transfer.ID,"status":"preparing"})
}
func updateComputerTransfer(t computerTransfer){
 computerTransferLock.Lock();computerTransfers[t.ID]=t;computerTransferLock.Unlock()
}
func runComputerTransfer(src,dst ComputerRecord,sr,dr,sourcePath,destinationPath string,t computerTransfer){
 defer func(){if v:=recover();v!=nil{t.Status="failed";t.Error="Unexpected transfer failure";t.Finished=time.Now();updateComputerTransfer(t)}}()
 ctx,cancel:=context.WithCancel(context.Background());defer cancel()
 sourceClient,e:=pinnedComputerClient(src.Fingerprint)
 if e!=nil {t.Status="failed";t.Error=e.Error();t.Finished=time.Now();updateComputerTransfer(t);return}
 var grant struct{Ticket string `json:"ticket"`;Size int64 `json:"size"`}
 e=agentJSONPost(ctx,sourceClient,src.URL+"/v1/grant",src.Key,map[string]any{"root":sr,"path":sourcePath},&grant)
 if e!=nil{t.Status="failed";t.Error="Source: "+e.Error();t.Finished=time.Now();updateComputerTransfer(t);return}
 t.Status="transferring";updateComputerTransfer(t)
 destClient,e:=pinnedComputerClient(dst.Fingerprint)
 if e!=nil{t.Status="failed";t.Error=e.Error();t.Finished=time.Now();updateComputerTransfer(t);return}
 var result struct{Size int64 `json:"size"`;SHA256 string `json:"sha256"`}
 e=agentJSONPost(ctx,destClient,dst.URL+"/v1/pull",dst.Key,map[string]any{
  "root":dr,"path":destinationPath,"source_url":src.URL,"fingerprint":src.Fingerprint,"ticket":grant.Ticket,"size":grant.Size,
 },&result)
 if e!=nil{t.Status="failed";t.Error="Destination: "+e.Error();t.Finished=time.Now();updateComputerTransfer(t);return}
 if result.Size!=grant.Size||len(result.SHA256)!=64{
  t.Status="failed";t.Error="Destination did not verify transferred bytes";t.Finished=time.Now();updateComputerTransfer(t);return
 }
 if _,e=hex.DecodeString(result.SHA256);e!=nil{t.Status="failed";t.Error="Destination checksum invalid";t.Finished=time.Now();updateComputerTransfer(t);return}
 t.Bytes=result.Size
 if t.Move{
  t.Status="removing source";updateComputerTransfer(t)
  e=agentJSONPost(ctx,sourceClient,src.URL+"/v1/delete",src.Key,map[string]any{"ticket":grant.Ticket,"sha256":result.SHA256},nil)
  if e!=nil{t.Status="failed";t.Error="File copied and verified, but source was NOT deleted: "+e.Error();t.Finished=time.Now();updateComputerTransfer(t);return}
 }
 t.Status="completed";t.Finished=time.Now();updateComputerTransfer(t)
}
var _ = sha256.Size
var _ = errors.New
