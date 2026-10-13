package main

import (
 "bytes"
 "context"
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
 "io"
 "net/http"
 "net/url"
 "path"
 "strconv"
 "strings"
 "time"
)

type dropboxFile struct {
 Tag string `json:".tag"`
 ID string `json:"id"`
 Name string `json:"name"`
 PathDisplay string `json:"path_display"`
 Size int64 `json:"size"`
 ContentHash string `json:"content_hash"`
 Rev string `json:"rev"`
 ServerModified string `json:"server_modified"`
}
type dropboxFolderList struct {
 Entries []dropboxFile `json:"entries"`
 Cursor string `json:"cursor"`
 HasMore bool `json:"has_more"`
}
func (a *App) dropboxAccount(id string)*DropboxAccount{
 for i:=range a.cfg.DropboxAccounts {
  if a.cfg.DropboxAccounts[i].ID==id{return &a.cfg.DropboxAccounts[i]}
 }
 return nil
}
func (a *App) dropboxToken(ctx context.Context,id string)(string,error){
 a.mu.Lock()
 account:=a.dropboxAccount(id)
 if account==nil{a.mu.Unlock();return "",errors.New("Dropbox account not found")}
 if account.AccessToken!=""&&time.Until(account.Expires)>2*time.Minute{token:=account.AccessToken;a.mu.Unlock();return token,nil}
 refresh:=account.RefreshToken;clientID:=a.cfg.DropboxOAuthID;secret:=a.cfg.DropboxOAuthSecret
 a.mu.Unlock()
 if refresh==""||clientID==""||secret==""{return "",errors.New("Dropbox OAuth credentials incomplete")}
 body:=url.Values{"grant_type":{"refresh_token"},"refresh_token":{refresh},"client_id":{clientID},"client_secret":{secret}}
 req,err:=http.NewRequestWithContext(ctx,"POST","https://api.dropbox.com/oauth2/token",strings.NewReader(body.Encode()))
 if err!=nil{return "",err}
 req.Header.Set("Content-Type","application/x-www-form-urlencoded")
 resp,err:=a.client.Do(req);if err!=nil{return "",err};defer resp.Body.Close()
 if resp.StatusCode/100!=2{
  b,_:=io.ReadAll(io.LimitReader(resp.Body,2048));return "",fmt.Errorf("Dropbox token refresh HTTP %d: %s",resp.StatusCode,b)
 }
 var token struct{AccessToken string `json:"access_token"`;ExpiresIn int64 `json:"expires_in"`}
 if err=json.NewDecoder(io.LimitReader(resp.Body,1<<20)).Decode(&token);err!=nil{return "",err}
 if token.AccessToken==""{return "",errors.New("Dropbox did not return an access token")}
 a.mu.Lock()
 if acct:=a.dropboxAccount(id);acct!=nil{
  acct.AccessToken=token.AccessToken;acct.Expires=time.Now().Add(time.Duration(token.ExpiresIn)*time.Second)
  _=a.save()
 }
 a.mu.Unlock()
 return token.AccessToken,nil
}
func (a *App) dropboxAPI(ctx context.Context,account,endpoint string,input any,out any)error{
 token,err:=a.dropboxToken(ctx,account);if err!=nil{return err}
 data,err:=json.Marshal(input);if err!=nil{return err}
 req,err:=http.NewRequestWithContext(ctx,"POST","https://api.dropboxapi.com/2/"+endpoint,bytes.NewReader(data))
 if err!=nil{return err}
 req.Header.Set("Authorization","Bearer "+token)
 req.Header.Set("Content-Type","application/json")
 resp,err:=a.client.Do(req);if err!=nil{return err};defer resp.Body.Close()
 if resp.StatusCode/100!=2{
  info,_:=io.ReadAll(io.LimitReader(resp.Body,4096))
  return fmt.Errorf("Dropbox %s HTTP %d: %s",endpoint,resp.StatusCode,strings.TrimSpace(string(info)))
 }
 if out==nil{_,err=io.Copy(io.Discard,resp.Body);return err}
 return json.NewDecoder(io.LimitReader(resp.Body,16<<20)).Decode(out)
}
func (a *App) dropboxList(ctx context.Context,account,folder string,recursive bool)([]dropboxFile,error){
 var page dropboxFolderList
 if err:=a.dropboxAPI(ctx,account,"files/list_folder",map[string]any{"path":folder,"recursive":recursive,"include_deleted":false,"limit":1000},&page);err!=nil{return nil,err}
 entries:=append([]dropboxFile(nil),page.Entries...)
 for page.HasMore{
  if len(entries)>200000{return nil,errors.New("Dropbox folder too large for a single safe scan")}
  if err:=a.dropboxAPI(ctx,account,"files/list_folder/continue",map[string]string{"cursor":page.Cursor},&page);err!=nil{return nil,err}
  entries=append(entries,page.Entries...)
 }
 return entries,nil
}
type bridgeDropboxEndpoint struct{
 app *App
 account,root string
 files map[string]dropboxFile
}
func (b *bridgeDropboxEndpoint) Close(){}
func cleanDropboxRoot(value string)(string,error){
 if value==""||value=="/"{return "",nil}
 normalized,err:=safeRemotePath(strings.TrimPrefix(value,"/"));if err!=nil{return "",err}
 if normalized=="."{return "",nil}
 return "/"+normalized,nil
}
func (b *bridgeDropboxEndpoint) fullPath(rel string)(string,error){
 root,err:=cleanDropboxRoot(b.root);if err!=nil{return "",err}
 if rel==""{return root,nil}
 value,err:=safeRemotePath(rel);if err!=nil{return "",err}
 if value=="."{return root,nil}
 return root+"/"+value,nil
}
func (b *bridgeDropboxEndpoint) Scan(ctx context.Context)(map[string]bridgeEntry,error){
 root,err:=cleanDropboxRoot(b.root);if err!=nil{return nil,err}
 entries,err:=b.app.dropboxList(ctx,b.account,root,true);if err!=nil{return nil,err}
 b.files=map[string]dropboxFile{}
 out:=map[string]bridgeEntry{}
 seen:=map[string]bool{}
 for _,f:=range entries{
  if f.Tag!="file"&&f.Tag!="folder"{continue}
  actual:=f.PathDisplay
  if root!="" {
   if !strings.HasPrefix(strings.ToLower(actual),strings.ToLower(root)+"/"){return nil,errors.New("Dropbox item escaped selected root")}
   actual=actual[len(root):]
  }
  actual=strings.TrimPrefix(actual,"/")
  if actual==""{continue}
  rel,err:=safeRemotePath(actual);if err!=nil{return nil,fmt.Errorf("Dropbox entry %q: %w",actual,err)}
  if strings.HasPrefix(rel,".syncria-recovery/"){continue}
  lower:=strings.ToLower(rel)
  if seen[lower]{return nil,fmt.Errorf("Dropbox case collision at %s",rel)}
  seen[lower]=true
  modified,_:=time.Parse(time.RFC3339Nano,f.ServerModified)
  b.files[rel]=f
  out[rel]=bridgeEntry{ID:f.ID,Size:f.Size,Fingerprint:f.ContentHash,Modified:modified,Folder:f.Tag=="folder"}
 }
 return out,nil
}
func (b *bridgeDropboxEndpoint) Mkdir(ctx context.Context,rel string)(bridgeEntry,error){
 full,err:=b.fullPath(rel);if err!=nil{return bridgeEntry{},err}
 var output struct{Metadata dropboxFile `json:"metadata"`}
 if err=b.app.dropboxAPI(ctx,b.account,"files/create_folder_v2",map[string]any{"path":full,"autorename":false},&output);err!=nil{return bridgeEntry{},err}
 if b.files==nil{b.files=map[string]dropboxFile{}}
 b.files[rel]=output.Metadata
 return bridgeEntry{ID:output.Metadata.ID,Folder:true},nil
}
func (b *bridgeDropboxEndpoint) Open(ctx context.Context,rel string)(io.ReadCloser,int64,error){
 f,exists:=b.files[rel];if !exists||f.Tag!="file"{return nil,0,errors.New("Dropbox file not in current scan")}
 token,err:=b.app.dropboxToken(ctx,b.account);if err!=nil{return nil,0,err}
 value,_:=json.Marshal(map[string]string{"path":f.ID})
 req,err:=http.NewRequestWithContext(ctx,"POST","https://content.dropboxapi.com/2/files/download",nil)
 if err!=nil{return nil,0,err}
 req.Header.Set("Authorization","Bearer "+token)
 req.Header.Set("Dropbox-API-Arg",string(value))
 resp,err:=b.app.client.Do(req);if err!=nil{return nil,0,err}
 if resp.StatusCode!=200{
  text,_:=io.ReadAll(io.LimitReader(resp.Body,4096));resp.Body.Close()
  return nil,0,fmt.Errorf("Dropbox download HTTP %d: %s",resp.StatusCode,text)
 }
 return resp.Body,f.Size,nil
}
func (b *bridgeDropboxEndpoint) uploadChunk(ctx context.Context,endpoint string,account string,metadata any,source io.Reader,size int64)([]byte,error){
 token,err:=b.app.dropboxToken(ctx,account);if err!=nil{return nil,err}
 value,err:=json.Marshal(metadata);if err!=nil{return nil,err}
 req,err:=http.NewRequestWithContext(ctx,"POST","https://content.dropboxapi.com/2/"+endpoint,source)
 if err!=nil{return nil,err}
 req.Header.Set("Authorization","Bearer "+token)
 req.Header.Set("Dropbox-API-Arg",string(value))
 req.Header.Set("Content-Type","application/octet-stream")
 req.ContentLength=size
 resp,err:=b.app.client.Do(req);if err!=nil{return nil,err};defer resp.Body.Close()
 payload,err:=io.ReadAll(io.LimitReader(resp.Body,1<<20))
 if err!=nil{return nil,err}
 if resp.StatusCode/100!=2{return nil,fmt.Errorf("Dropbox %s HTTP %d: %s",endpoint,resp.StatusCode,payload)}
 return payload,nil
}
func (b *bridgeDropboxEndpoint) Put(ctx context.Context,rel string,input io.Reader,size int64,expectedMD5 string,progress func(int64))(bridgeEntry,error){
 if size<0{return bridgeEntry{},errors.New("Dropbox upload source size unknown")}
 target,err:=b.fullPath(rel);if err!=nil{return bridgeEntry{},err}
 const maxSimple int64=150*1024*1024
 var metadata dropboxFile
 counter:=&bridgeProgressReader{input:input,callback:progress}
 mode:=map[string]string{".tag":"overwrite"}
 if size<=maxSimple{
  payload,err:=b.uploadChunk(ctx,"files/upload",b.account,map[string]any{"path":target,"mode":mode,"autorename":false,"mute":true},counter,size)
  if err!=nil{return bridgeEntry{},err}
  if err=json.Unmarshal(payload,&metadata);err!=nil{return bridgeEntry{},err}
 }else{
  // Eight-megabyte upload-session chunks stay in RAM only; never on /data.
  const chunk int64=8*1024*1024
  part:=make([]byte,chunk)
  var session struct{SessionID string `json:"session_id"`}
  for sent:=int64(0);sent<size;{
   want:=chunk;if size-sent<want{want=size-sent}
   n,e:=io.ReadFull(counter,part[:int(want)]);if e!=nil{return bridgeEntry{},e}
   body:=bytes.NewReader(part[:n])
   if sent==0{
    payload,e:=b.uploadChunk(ctx,"files/upload_session/start",b.account,map[string]any{"close":false},body,int64(n))
    if e!=nil{return bridgeEntry{},e}
    if e=json.Unmarshal(payload,&session);e!=nil||session.SessionID==""{return bridgeEntry{},errors.New("Dropbox upload session could not start")}
   }else if sent+int64(n)<size{
    _,e=b.uploadChunk(ctx,"files/upload_session/append_v2",b.account,
     map[string]any{"cursor":map[string]any{"session_id":session.SessionID,"offset":sent},"close":false},body,int64(n))
    if e!=nil{return bridgeEntry{},e}
   }else{
    payload,e:=b.uploadChunk(ctx,"files/upload_session/finish",b.account,
     map[string]any{"cursor":map[string]any{"session_id":session.SessionID,"offset":sent},"commit":map[string]any{"path":target,"mode":mode,"autorename":false,"mute":true}},
     body,int64(n))
    if e!=nil{return bridgeEntry{},e}
    if e=json.Unmarshal(payload,&metadata);e!=nil{return bridgeEntry{},e}
   }
   sent+=int64(n)
  }
 }
 if counter.read!=size||metadata.ID==""||metadata.Size!=size{return bridgeEntry{},errors.New("Dropbox upload could not verify file size and identity")}
 if b.files==nil{b.files=map[string]dropboxFile{}}
 b.files[rel]=metadata
 modified,_:=time.Parse(time.RFC3339Nano,metadata.ServerModified)
 return bridgeEntry{ID:metadata.ID,Size:size,Fingerprint:metadata.ContentHash,Modified:modified},nil
}
func (b *bridgeDropboxEndpoint) StageDelete(ctx context.Context,rel,jobID string)(string,error){
 f,ok:=b.files[rel];if !ok{return "",errors.New("Dropbox item unavailable")}
 if f.Tag=="folder"{return "",errors.New("Dropbox folder recovery needs provider support; no deletion performed")}
 if f.Rev==""{return "",errors.New("Dropbox revision unavailable; cannot guarantee restore")}
 full,err:=b.fullPath(rel);if err!=nil{return "",err}
 if err=b.app.dropboxAPI(ctx,b.account,"files/delete_v2",map[string]any{"path":full},nil);err!=nil{return "",err}
 return f.Rev,nil
}
func (b *bridgeDropboxEndpoint) Restore(ctx context.Context,rel,token string)error{
 full,err:=b.fullPath(rel);if err!=nil{return err}
 return b.app.dropboxAPI(ctx,b.account,"files/restore",map[string]any{"path":full,"rev":token},nil)
}
func (b *bridgeDropboxEndpoint) Purge(ctx context.Context,token string)error{
 // Dropbox retains deleted revisions per the account's own retention policy.
 // Syncria's ten-minute *Undo* authorization expires independently.
 return nil
}
func (b *bridgeDropboxEndpoint) Delete(ctx context.Context,rel string)error{
 full,err:=b.fullPath(rel);if err!=nil{return err}
 return b.app.dropboxAPI(ctx,b.account,"files/delete_v2",map[string]any{"path":full},nil)
}
func (a *App) dropboxOAuthStart(w http.ResponseWriter,r *http.Request){
 if !a.authenticated(r){http.Error(w,"Unauthorized",401);return}
 a.mu.Lock();client:=a.cfg.DropboxOAuthID;secret:=a.cfg.DropboxOAuthSecret;base:=a.cfg.BaseURL;a.mu.Unlock()
 if client==""||secret==""||base==""{redirect(w,r,"Configure Dropbox OAuth under Settings first");return}
 state:="dropbox-"+b64rand()
 a.mu.Lock();if a.oauth==nil{a.oauth=map[string]time.Time{}};a.oauth[state]=time.Now();a.mu.Unlock()
 u:=url.Values{"client_id":{client},"redirect_uri":{base+"/oauth/dropbox/callback"},"response_type":{"code"},"token_access_type":{"offline"},"state":{state}}
 http.Redirect(w,r,"https://www.dropbox.com/oauth2/authorize?"+u.Encode(),302)
}
func (a *App) dropboxOAuthCallback(w http.ResponseWriter,r *http.Request){
 state:=r.URL.Query().Get("state")
 a.mu.Lock();created,ok:=a.oauth[state];delete(a.oauth,state)
 client:=a.cfg.DropboxOAuthID;secret:=a.cfg.DropboxOAuthSecret;base:=a.cfg.BaseURL;a.mu.Unlock()
 if !ok||!strings.HasPrefix(state,"dropbox-")||time.Since(created)>10*time.Minute||!a.authenticated(r){http.Error(w,"Dropbox OAuth state expired",403);return}
 if r.URL.Query().Get("error")!=""{redirect(w,r,"Dropbox authorization was declined");return}
 code:=r.URL.Query().Get("code");if code==""{redirect(w,r,"Dropbox authorization code missing");return}
 body:=url.Values{"code":{code},"grant_type":{"authorization_code"},"client_id":{client},"client_secret":{secret},"redirect_uri":{base+"/oauth/dropbox/callback"}}
 req,e:=http.NewRequestWithContext(r.Context(),"POST","https://api.dropbox.com/oauth2/token",strings.NewReader(body.Encode()));if e!=nil{http.Error(w,"OAuth request failed",500);return}
 req.Header.Set("Content-Type","application/x-www-form-urlencoded")
 resp,e:=a.client.Do(req);if e!=nil{redirect(w,r,"Dropbox token exchange unavailable");return}
 var token struct{AccessToken string `json:"access_token"`;RefreshToken string `json:"refresh_token"`;ExpiresIn int64 `json:"expires_in"`;AccountID string `json:"account_id"`}
 if resp.StatusCode/100!=2{resp.Body.Close();redirect(w,r,"Dropbox token exchange rejected");return}
 e=json.NewDecoder(io.LimitReader(resp.Body,1<<20)).Decode(&token);resp.Body.Close()
 if e!=nil||token.RefreshToken==""||token.AccessToken==""{redirect(w,r,"Dropbox did not authorize offline synchronization");return}
 req,e=http.NewRequestWithContext(r.Context(),"POST","https://api.dropboxapi.com/2/users/get_current_account",strings.NewReader("null"));if e!=nil{http.Error(w,"Request failed",500);return}
 req.Header.Set("Authorization","Bearer "+token.AccessToken);req.Header.Set("Content-Type","application/json")
 resp,e=a.client.Do(req);if e!=nil{redirect(w,r,"Dropbox account verification unavailable");return}
 var who struct{Email string `json:"email"`;AccountID string `json:"account_id"`}
 if resp.StatusCode/100!=2{resp.Body.Close();redirect(w,r,"Dropbox account verification rejected");return}
 e=json.NewDecoder(io.LimitReader(resp.Body,1<<20)).Decode(&who);resp.Body.Close()
 if e!=nil||who.AccountID==""{redirect(w,r,"Dropbox account identity unavailable");return}
 record:=DropboxAccount{ID:who.AccountID,Email:who.Email,RefreshToken:token.RefreshToken,AccessToken:token.AccessToken,Expires:time.Now().Add(time.Duration(token.ExpiresIn)*time.Second)}
 a.mu.Lock();found:=false
 for i:=range a.cfg.DropboxAccounts{if a.cfg.DropboxAccounts[i].ID==record.ID{a.cfg.DropboxAccounts[i]=record;found=true;break}}
 if !found{a.cfg.DropboxAccounts=append(a.cfg.DropboxAccounts,record)}
 e=a.save();a.mu.Unlock()
 if e!=nil{http.Error(w,"Could not save Dropbox authorization",500);return}
 http.Redirect(w,r,"/?msg="+url.QueryEscape("Dropbox connected as "+who.Email)+"#accounts",303)
}
func (a *App) browseDropbox(w http.ResponseWriter,r *http.Request){
 if r.Method!="GET"{http.Error(w,"GET required",405);return}
 if !a.authenticated(r){http.Error(w,"Unauthorized",401);return}
 id:=r.URL.Query().Get("account");root,err:=cleanDropboxRoot(r.URL.Query().Get("path"))
 if err!=nil{http.Error(w,err.Error(),400);return}
 ctx,cancel:=context.WithTimeout(r.Context(),30*time.Second);defer cancel()
 files,err:=a.dropboxList(ctx,id,root,false);if err!=nil{http.Error(w,err.Error(),502);return}
 response:=browseResponse{Path:root,Entries:[]browseEntry{}}
 if root!=""{parent:=path.Dir(root);if parent=="/"{parent=""};response.Parent=parent}
 for _,f:=range files{
  if f.Tag!="folder"||!safeName(f.Name){continue}
  response.Entries=append(response.Entries,browseEntry{Name:f.Name,ID:f.PathDisplay,Kind:"folder"})
 }
 sendBrowse(w,response)
}
var _=sha256.New
var _=hex.EncodeToString
var _=strconv.Itoa
