package main

import (
 "context"
 "encoding/json"
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
 SourceID string `json:"source_id"`
 DestinationID string `json:"destination_id"`
 Source string `json:"source"`
 Destination string `json:"destination"`
 SourceRoot string `json:"source_root"`
 DestinationRoot string `json:"destination_root"`
 File string `json:"file"`
 Target string `json:"target"`
 Move bool `json:"move"`
 Status string `json:"status"`
 Error string `json:"error,omitempty"`
 Bytes int64 `json:"bytes"`
 Total int64 `json:"total"`
 Started time.Time `json:"started"`
 Finished time.Time `json:"finished,omitempty"`
}
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
  a.computerTransferMu.Lock()
  for _,job:=range a.computerTransferJobs{
   if !terminalComputerTransfer(job.Status)&&(job.SourceID==body.ID||job.DestinationID==body.ID){
    a.computerTransferMu.Unlock()
    computerError(w,409,"Computer has active or queued transfers. Cancel them in Computer Transfers first.");return
   }
  }
  a.computerTransferMu.Unlock()
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
