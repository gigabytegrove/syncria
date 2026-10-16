package main

import (
 "context"
 "crypto/rand"
 "crypto/rsa"
 "crypto/sha256"
 "crypto/subtle"
 "crypto/tls"
 "crypto/x509"
 "crypto/x509/pkix"
 "encoding/hex"
 "encoding/json"
 "encoding/pem"
 "errors"
 "fmt"
 "io"
 "log"
 "math/big"
 "net"
 "net/http"
 "net/url"
 "os"
 "path"
 "path/filepath"
 "sort"
 "strconv"
 "strings"
 "sync"
 "time"
)

type agentRootsFlag []string
func (f *agentRootsFlag) String() string { return strings.Join(*f,", ") }
func (f *agentRootsFlag) Set(s string) error { *f=append(*f,s);return nil }

type agentFile struct { Name string `json:"name"`; Directory bool `json:"directory"`; Size int64 `json:"size"`; Modified time.Time `json:"modified"` }
type agentRootInfo struct { Name string `json:"name"`; Path string `json:"-"` }
type agentTicket struct { Root,Relative string; Size int64; Modified time.Time; Expires time.Time }
type agentTransferProgress struct { Bytes int64 `json:"bytes"`; Total int64 `json:"total"`; Status string `json:"status"` }
type computerAgent struct { name,secret string; roots map[string]string; mu sync.Mutex; tickets map[string]agentTicket; progress map[string]agentTransferProgress }
type agentProgressWriter struct { dst io.Writer; onWrite func(int) }
func (w agentProgressWriter) Write(b []byte)(int,error){n,e:=w.dst.Write(b);if n>0{w.onWrite(n)};return n,e}
func (a *computerAgent) progressSet(id string,p agentTransferProgress){
 a.mu.Lock();defer a.mu.Unlock()
 if a.progress==nil{a.progress=map[string]agentTransferProgress{}}
 if len(a.progress)>1024{
  for old,entry:=range a.progress{if entry.Status!="transferring"{delete(a.progress,old)}}
 }
 a.progress[id]=p
}
func (a *computerAgent) serveProgress(w http.ResponseWriter,r *http.Request){
 if r.Method!="GET"{computerError(w,405,"GET required");return}
 a.mu.Lock();p,ok:=a.progress[r.URL.Query().Get("id")];a.mu.Unlock()
 if !ok{computerError(w,404,"Transfer progress not found");return}
 computerJSON(w,200,p)
}
type agentIdentity struct { Secret string `json:"secret"` }

func computerJSON(w http.ResponseWriter,code int,v any) {
 w.Header().Set("Content-Type","application/json; charset=utf-8")
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("X-Content-Type-Options","nosniff")
 w.WriteHeader(code)
 _=json.NewEncoder(w).Encode(v)
}
func computerError(w http.ResponseWriter,code int,message string) { http.Error(w,message,code) }
func computerDecode(r *http.Request,v any) error {
 r.Body=http.MaxBytesReader(nil,r.Body,32<<10)
 d:=json.NewDecoder(r.Body);d.DisallowUnknownFields()
 if err:=d.Decode(v);err!=nil{return err}
 var extra any
 if err:=d.Decode(&extra);err!=io.EOF{return errors.New("unexpected trailing JSON")}
 return nil
}
func computerRequestAuth(r *http.Request,secret string) bool {
 got:=strings.TrimPrefix(r.Header.Get("Authorization"),"Bearer ")
 return len(got)==len(secret)&&subtle.ConstantTimeCompare([]byte(got),[]byte(secret))==1
}
func validAgentURL(value string) (*url.URL,error) {
 u,e:=url.Parse(value);if e!=nil{return nil,e}
 if u.Scheme!="https"||u.Hostname()==""||u.User!=nil||u.Path!=""||u.RawQuery!=""||u.Fragment!="" {
  return nil,errors.New("agent address must be an HTTPS origin without path or credentials")
 }
 return u,nil
}
func pinnedComputerClient(fingerprint string) (*http.Client,error) {
 fingerprint=strings.ToLower(strings.ReplaceAll(strings.TrimSpace(fingerprint),":",""))
 if len(fingerprint)!=64 {return nil,errors.New("TLS fingerprint must have 64 hexadecimal characters")}
 want,e:=hex.DecodeString(fingerprint);if e!=nil{return nil,errors.New("invalid TLS fingerprint")}
 transport:=&http.Transport{Proxy:nil,DialContext:(&net.Dialer{Timeout:10*time.Second,KeepAlive:30*time.Second}).DialContext,
  TLSClientConfig:&tls.Config{MinVersion:tls.VersionTLS12,InsecureSkipVerify:true,VerifyConnection:func(cs tls.ConnectionState)error{
   if len(cs.PeerCertificates)==0{return errors.New("agent TLS certificate absent")}
   sum:=sha256.Sum256(cs.PeerCertificates[0].Raw)
   if subtle.ConstantTimeCompare(sum[:],want)!=1{return errors.New("agent TLS fingerprint mismatch")}
   return nil
  }},ResponseHeaderTimeout:30*time.Second,IdleConnTimeout:30*time.Second}
 return &http.Client{Transport:transport,CheckRedirect:func(_ *http.Request,_ []*http.Request)error{return errors.New("agent redirects are forbidden")}},nil
}
func agentRelative(rel string) (string,error) {
 if rel==""||rel=="." {return "",nil}
 if strings.HasPrefix(rel,"/")||strings.ContainsAny(rel,"\\\x00")||path.Clean(rel)!=rel {return "",errors.New("invalid relative path")}
 for _,p:=range strings.Split(rel,"/") {
  if p==""||p=="."||p==".."||strings.ContainsAny(p,":*?\"<>|")||strings.HasPrefix(p,".syncria-"){return "",errors.New("invalid path component")}
 }
 return rel,nil
}
func (a *computerAgent) resolve(root,relative string,allowNew bool)(string,error){
 base,ok:=a.roots[root];if !ok{return "",errors.New("unknown shared folder")}
 rel,e:=agentRelative(relative);if e!=nil{return "",e}
 current:=base
 if rel=="" {return current,nil}
 parts:=strings.Split(rel,"/")
 for i,p:=range parts {
  current=filepath.Join(current,p)
  info,e:=os.Lstat(current)
  if os.IsNotExist(e)&&allowNew{return currentWithParts(current,parts[i+1:]),nil}
  if e!=nil{return "",e}
  if info.Mode()&os.ModeSymlink!=0{return "",errors.New("symbolic links are not accessible")}
  if i<len(parts)-1&&!info.IsDir(){return "",errors.New("parent is not a directory")}
 }
 return current,nil
}
func currentWithParts(current string,rest []string)string {
 for _,p:=range rest{current=filepath.Join(current,p)}
 return current
}
func (a *computerAgent) authorize(next http.HandlerFunc)http.HandlerFunc{
 return func(w http.ResponseWriter,r *http.Request){
  if !computerRequestAuth(r,a.secret){computerError(w,401,"Invalid agent key");return}
  next(w,r)
 }
}
func (a *computerAgent) serveStatus(w http.ResponseWriter,r *http.Request){
 if r.Method!="GET"{computerError(w,405,"GET required");return}
 roots:=make([]agentRootInfo,0,len(a.roots))
 for n:=range a.roots{roots=append(roots,agentRootInfo{Name:n})}
 sort.Slice(roots,func(i,j int)bool{return roots[i].Name<roots[j].Name})
 computerJSON(w,200,map[string]any{"name":a.name,"roots":roots,"status":"online"})
}
func (a *computerAgent) serveList(w http.ResponseWriter,r *http.Request){
 if r.Method!="GET"{computerError(w,405,"GET required");return}
 root,rel:=r.URL.Query().Get("root"),r.URL.Query().Get("path")
 _,e:=a.resolve(root,rel,false);if e!=nil{computerError(w,400,e.Error());return}
 handle,e:=os.OpenRoot(a.roots[root]);if e!=nil{computerError(w,500,e.Error());return};defer handle.Close()
 dirName:=rel;if dirName==""{dirName="."}
 dir,e:=handle.Open(dirName);if e!=nil{computerError(w,400,e.Error());return};defer dir.Close()
 info,e:=dir.Stat();if e!=nil||!info.IsDir(){computerError(w,400,"not a directory");return}
 entries,e:=dir.ReadDir(-1);if e!=nil{computerError(w,500,e.Error());return}
 out:=make([]agentFile,0,len(entries))
 for _,entry:=range entries{
  if strings.HasPrefix(entry.Name(),".syncria-")||entry.Type()&os.ModeSymlink!=0 {continue}
  if _,e:=agentRelative(entry.Name());e!=nil{continue}
  info,e:=entry.Info();if e!=nil||(!info.IsDir()&&!info.Mode().IsRegular()){continue}
  out=append(out,agentFile{Name:entry.Name(),Directory:info.IsDir(),Size:info.Size(),Modified:info.ModTime()})
 }
 sort.Slice(out,func(i,j int)bool{if out[i].Directory!=out[j].Directory{return out[i].Directory};return strings.ToLower(out[i].Name)<strings.ToLower(out[j].Name)})
 parent:=""
 if rel!=""{parent=path.Dir(rel);if parent=="."{parent=""}}
 computerJSON(w,200,map[string]any{"root":root,"path":rel,"parent":parent,"entries":out})
}
func (a *computerAgent) serveGrant(w http.ResponseWriter,r *http.Request){
 if r.Method!="POST"{computerError(w,405,"POST required");return}
 var body struct {Root string `json:"root"`; Path string `json:"path"`}
 if e:=computerDecode(r,&body);e!=nil{computerError(w,400,e.Error());return}
 _,e:=a.resolve(body.Root,body.Path,false);if e!=nil{computerError(w,400,e.Error());return}
 handle,e:=os.OpenRoot(a.roots[body.Root]);if e!=nil{computerError(w,400,e.Error());return};defer handle.Close()
 info,e:=handle.Lstat(body.Path);if e!=nil||!info.Mode().IsRegular(){computerError(w,400,"regular file required");return}
 ticket:=b64rand()
 grant:=agentTicket{Root:body.Root,Relative:body.Path,Size:info.Size(),Modified:info.ModTime(),Expires:time.Now().Add(30*time.Minute)}
 a.mu.Lock()
 for id,t:=range a.tickets{if time.Now().After(t.Expires){delete(a.tickets,id)}}
 if len(a.tickets)>=1024{a.mu.Unlock();computerError(w,429,"too many outstanding transfers");return}
 a.tickets[ticket]=grant;a.mu.Unlock()
 computerJSON(w,200,map[string]any{"ticket":ticket,"size":info.Size(),"modified":info.ModTime()})
}
func (a *computerAgent) getTicket(token string)(agentTicket,bool){
 a.mu.Lock();defer a.mu.Unlock()
 t,ok:=a.tickets[token];if !ok||time.Now().After(t.Expires){return agentTicket{},false}
 return t,true
}
func (a *computerAgent) serveDownload(w http.ResponseWriter,r *http.Request){
 if r.Method!="GET"{computerError(w,405,"GET required");return}
 ticket,ok:=a.getTicket(strings.TrimPrefix(r.Header.Get("Authorization"),"Bearer "))
 if !ok{computerError(w,401,"invalid or expired transfer ticket");return}
 _,e:=a.resolve(ticket.Root,ticket.Relative,false);if e!=nil{computerError(w,409,"source path changed");return}
 handle,e:=os.OpenRoot(a.roots[ticket.Root]);if e!=nil{computerError(w,409,e.Error());return};defer handle.Close()
 if info,e:=handle.Lstat(ticket.Relative);e!=nil||!info.Mode().IsRegular(){computerError(w,409,"source is not a regular file");return}
 f,e:=handle.Open(ticket.Relative);if e!=nil{computerError(w,404,"source file unavailable");return};defer f.Close()
 info,e:=f.Stat()
 if e!=nil||!info.Mode().IsRegular()||info.Size()!=ticket.Size||!info.ModTime().Equal(ticket.Modified){computerError(w,409,"source changed since transfer was approved");return}
 w.Header().Set("Content-Type","application/octet-stream")
 w.Header().Set("Trailer","X-Syncria-SHA256")
 w.Header().Set("X-Content-Type-Options","nosniff")
 h:=sha256.New()
 if n,e:=io.Copy(io.MultiWriter(w,h),f);e==nil&&n==ticket.Size {
  w.Header().Set("X-Syncria-SHA256",hex.EncodeToString(h.Sum(nil)))
  token:=strings.TrimPrefix(r.Header.Get("Authorization"),"Bearer ")
  a.mu.Lock()
  if current,ok:=a.tickets[token];ok{current.Expires=time.Now().Add(30*time.Minute);a.tickets[token]=current}
  a.mu.Unlock()
 }
}
func agentJSONPost(ctx context.Context,c *http.Client,endpoint,key string,payload any,answer any)error{
 b,e:=json.Marshal(payload);if e!=nil{return e}
 req,e:=http.NewRequestWithContext(ctx,"POST",endpoint,strings.NewReader(string(b)));if e!=nil{return e}
 req.Header.Set("Authorization","Bearer "+key);req.Header.Set("Content-Type","application/json")
 resp,e:=c.Do(req);if e!=nil{return e};defer resp.Body.Close()
 if resp.StatusCode/100!=2 {s,_:=io.ReadAll(io.LimitReader(resp.Body,1024));return fmt.Errorf("agent HTTP %d: %s",resp.StatusCode,strings.TrimSpace(string(s)))}
 if answer==nil{return nil}
 return json.NewDecoder(io.LimitReader(resp.Body,65536)).Decode(answer)
}
func (a *computerAgent) servePull(w http.ResponseWriter,r *http.Request){
 if r.Method!="POST"{computerError(w,405,"POST required");return}
 var body struct{
  Root string `json:"root"`;Path string `json:"path"`
  SourceURL string `json:"source_url"`;Fingerprint string `json:"fingerprint"`
  Ticket string `json:"ticket"`; Size int64 `json:"size"`; TransferID string `json:"transfer_id"`
 }
 if e:=computerDecode(r,&body);e!=nil{computerError(w,400,e.Error());return}
 if _,e:=validAgentURL(body.SourceURL);e!=nil{computerError(w,400,e.Error());return}
 if body.Size<0||body.Ticket==""||body.Path==""||len(body.TransferID)!=32{computerError(w,400,"invalid transfer request");return}
 if _,e:=hex.DecodeString(body.TransferID);e!=nil{computerError(w,400,"invalid transfer ID");return}
 a.progressSet(body.TransferID,agentTransferProgress{Total:body.Size,Status:"transferring"})
 succeeded:=false
 defer func(){if !succeeded{a.mu.Lock();p:=a.progress[body.TransferID];p.Status="failed";a.progress[body.TransferID]=p;a.mu.Unlock()}}()
 if body.Size>1<<50{computerError(w,400,"file exceeds maximum transfer size");return}
 if _,e:=a.resolve(body.Root,body.Path,true);e!=nil{computerError(w,400,e.Error());return}
 handle,e:=os.OpenRoot(a.roots[body.Root]);if e!=nil{computerError(w,400,e.Error());return};defer handle.Close()
 if _,e:=handle.Lstat(body.Path);e==nil{computerError(w,409,"destination already exists");return}else if !os.IsNotExist(e){computerError(w,400,e.Error());return}
 // All file mutations go through Go's traversal-resistant os.Root API.
 parentRel:=path.Dir(body.Path);if parentRel=="."{parentRel=""}
 if _,e:=a.resolve(body.Root,parentRel,true);e!=nil{computerError(w,400,e.Error());return}
 if parentRel!=""{if e:=handle.MkdirAll(parentRel,0700);e!=nil{computerError(w,500,e.Error());return}}
 if _,e:=a.resolve(body.Root,parentRel,false);e!=nil{computerError(w,400,e.Error());return}
 client,e:=pinnedComputerClient(body.Fingerprint);if e!=nil{computerError(w,400,e.Error());return}
 request,e:=http.NewRequestWithContext(r.Context(),"GET",body.SourceURL+"/v1/download",nil)
 if e!=nil{computerError(w,400,e.Error());return}
 request.Header.Set("Authorization","Bearer "+body.Ticket)
 response,e:=client.Do(request);if e!=nil{computerError(w,502,e.Error());return};defer response.Body.Close()
 if response.StatusCode!=200{b,_:=io.ReadAll(io.LimitReader(response.Body,1024));computerError(w,502,fmt.Sprintf("source HTTP %d: %s",response.StatusCode,string(b)));return}
 temp:=path.Join(parentRel,".syncria-transfer-"+newID())
 f,e:=handle.OpenFile(temp,os.O_CREATE|os.O_EXCL|os.O_WRONLY,0600);if e!=nil{computerError(w,500,e.Error());return}
 defer handle.Remove(temp)
 h:=sha256.New()
 writer:=agentProgressWriter{dst:io.MultiWriter(f,h),onWrite:func(n int){
  a.mu.Lock();p:=a.progress[body.TransferID];p.Bytes+=int64(n);a.progress[body.TransferID]=p;a.mu.Unlock()
 }}
 written,e:=io.Copy(writer,io.LimitReader(response.Body,body.Size+1))
 if e!=nil{f.Close();computerError(w,502,"transfer interrupted: "+e.Error());return}
 if written!=body.Size{f.Close();computerError(w,409,"transferred file size mismatch");return}
 digest:=hex.EncodeToString(h.Sum(nil))
 if response.Trailer.Get("X-Syncria-SHA256")!=digest{f.Close();computerError(w,409,"source file checksum mismatch");return}
 if e:=f.Sync();e!=nil{f.Close();computerError(w,500,e.Error());return}
 if e:=f.Close();e!=nil{computerError(w,500,e.Error());return}
 if _,e:=a.resolve(body.Root,parentRel,false);e!=nil{computerError(w,409,e.Error());return}
 if e:=handle.Link(temp,body.Path);e!=nil{computerError(w,409,"cannot publish without overwriting: "+e.Error());return}
 succeeded=true
 a.progressSet(body.TransferID,agentTransferProgress{Total:body.Size,Bytes:written,Status:"verified"})
 computerJSON(w,200,map[string]any{"size":written,"sha256":digest,"path":body.Path})
}
func (a *computerAgent) serveDelete(w http.ResponseWriter,r *http.Request){
 if r.Method!="POST"{computerError(w,405,"POST required");return}
 var body struct{Ticket string `json:"ticket"`;SHA256 string `json:"sha256"`}
 if e:=computerDecode(r,&body);e!=nil{computerError(w,400,e.Error());return}
 ticket,ok:=a.getTicket(body.Ticket);if !ok{computerError(w,409,"transfer ticket expired");return}
 if len(body.SHA256)!=64{computerError(w,400,"invalid checksum");return}
 _,e:=a.resolve(ticket.Root,ticket.Relative,false);if e!=nil{computerError(w,409,e.Error());return}
 handle,e:=os.OpenRoot(a.roots[ticket.Root]);if e!=nil{computerError(w,409,e.Error());return};defer handle.Close()
 if info,e:=handle.Lstat(ticket.Relative);e!=nil||!info.Mode().IsRegular(){computerError(w,409,"source is not a regular file");return}
 f,e:=handle.Open(ticket.Relative);if e!=nil{computerError(w,409,e.Error());return}
 info,e:=f.Stat()
 if e!=nil||!info.Mode().IsRegular()||info.Size()!=ticket.Size||!info.ModTime().Equal(ticket.Modified){f.Close();computerError(w,409,"source changed; copy preserved, original not deleted");return}
 h:=sha256.New();_,e=io.Copy(h,f);f.Close()
 if e!=nil||hex.EncodeToString(h.Sum(nil))!=body.SHA256{computerError(w,409,"source checksum changed; original not deleted");return}
 if _,e:=a.resolve(ticket.Root,ticket.Relative,false);e!=nil{computerError(w,409,e.Error());return}
 if e:=handle.Remove(ticket.Relative);e!=nil{computerError(w,409,e.Error());return}
 a.mu.Lock();delete(a.tickets,body.Ticket);a.mu.Unlock()
 computerJSON(w,200,map[string]any{"deleted":true})
}
func createAgentIdentity(dir string)(string,string,string,error){
 if e:=os.MkdirAll(dir,0700);e!=nil{return "","","",e}
 configFile:=filepath.Join(dir,"identity.json")
 var ident agentIdentity
 b,e:=os.ReadFile(configFile)
 if e==nil{if e=json.Unmarshal(b,&ident);e!=nil{return "","","",e}}else if !os.IsNotExist(e){return "","","",e}
 if ident.Secret=="" {
  ident.Secret=b64rand()
  b,_=json.MarshalIndent(ident,"","  ")
  f,e:=os.OpenFile(configFile,os.O_WRONLY|os.O_CREATE|os.O_EXCL,0600)
  if e!=nil{return "","","",e}
  if _,e=f.Write(b);e!=nil{f.Close();return "","","",e}
  if e=f.Close();e!=nil{return "","","",e}
 }
 certFile,keyFile:=filepath.Join(dir,"agent.crt"),filepath.Join(dir,"agent.key")
 if _,e:=os.Stat(certFile);os.IsNotExist(e){
  key,e:=rsa.GenerateKey(rand.Reader,3072);if e!=nil{return "","","",e}
  serial,e:=rand.Int(rand.Reader,new(big.Int).Lsh(big.NewInt(1),128));if e!=nil{return "","","",e}
  tmpl:=&x509.Certificate{SerialNumber:serial,Subject:pkix.Name{CommonName:"Syncria Computer Agent"},NotBefore:time.Now().Add(-time.Hour),NotAfter:time.Now().AddDate(5,0,0),KeyUsage:x509.KeyUsageDigitalSignature,ExtKeyUsage:[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},DNSNames:[]string{"localhost"}}
  raw,e:=x509.CreateCertificate(rand.Reader,tmpl,tmpl,&key.PublicKey,key);if e!=nil{return "","","",e}
  cb:=pem.EncodeToMemory(&pem.Block{Type:"CERTIFICATE",Bytes:raw})
  kb:=pem.EncodeToMemory(&pem.Block{Type:"RSA PRIVATE KEY",Bytes:x509.MarshalPKCS1PrivateKey(key)})
  if e=os.WriteFile(keyFile,kb,0600);e!=nil{return "","","",e}
  if e=os.WriteFile(certFile,cb,0600);e!=nil{return "","","",e}
 }
 pair,e:=tls.LoadX509KeyPair(certFile,keyFile);if e!=nil{return "","","",e}
 fingerprint:=sha256.Sum256(pair.Certificate[0])
 return ident.Secret,certFile,hex.EncodeToString(fingerprint[:]),nil
}
func runComputerAgent(listen,dataDir,name string,folders []string)error{
 return runComputerAgentWithEnrollment(listen,dataDir,name,folders,"","")
}
func runComputerAgentWithEnrollment(listen,dataDir,name string,folders []string,hub,code string)error{
 if len(folders)==0{return errors.New("specify at least one -agent-root Name=/absolute/path")}
 if dataDir==""{return errors.New("specify -agent-data to save the TLS identity and pairing key")}
 roots:=map[string]string{}
 for _,entry:=range folders{
  k,v,ok:=strings.Cut(entry,"=")
  if !ok||k==""||v==""||strings.ContainsAny(k,"/\\\x00") {return errors.New("-agent-root requires Name=/absolute/path")}
  if !filepath.IsAbs(v){return errors.New("agent shared folders must be absolute paths")}
  if _,exists:=roots[k];exists{return errors.New("duplicate agent root name")}
  full,e:=filepath.Abs(v);if e!=nil{return e}
  info,e:=os.Lstat(full);if e!=nil||!info.IsDir()||info.Mode()&os.ModeSymlink!=0{return fmt.Errorf("agent shared directory is unavailable: %s",k)}
  roots[k]=full
 }
 // Never allow the agent identity and private key to be shared as user files.
 identityPath,e:=filepath.Abs(dataDir);if e!=nil{return e}
 for label,folder:=range roots{
  relative,err:=filepath.Rel(folder,identityPath)
  if err==nil&&(relative=="."||(!strings.HasPrefix(relative,".."+string(filepath.Separator))&&relative!="..")){
   return fmt.Errorf("agent identity directory must not be inside shared folder %q",label)
  }
 }
 secret,certFile,fingerprint,e:=createAgentIdentity(dataDir);if e!=nil{return e}
 if name==""{name,_=os.Hostname()}
 agent:=&computerAgent{name:name,secret:secret,roots:roots,tickets:map[string]agentTicket{}}
 mux:=http.NewServeMux()
 mux.HandleFunc("/v1/status",agent.authorize(agent.serveStatus))
 mux.HandleFunc("/v1/list",agent.authorize(agent.serveList))
 mux.HandleFunc("/v1/grant",agent.authorize(agent.serveGrant))
 mux.HandleFunc("/v1/download",agent.serveDownload)
 mux.HandleFunc("/v1/pull",agent.authorize(agent.servePull))
 mux.HandleFunc("/v1/progress",agent.authorize(agent.serveProgress))
 mux.HandleFunc("/v1/delete",agent.authorize(agent.serveDelete))
 // No CORS, no wildcard file access, HTTPS-only transport with certificate pinning.
 log.Printf("Syncria Computer Agent listening at https://%s",listen)
 if hub=="" {
  log.Printf("Pairing key: %s",secret)
  log.Printf("TLS SHA-256 fingerprint: %s",fingerprint)
 } else {
  if err:=computerEnrollmentHubURL(hub);err!=nil{return err}
  if len(code)!=64{return errors.New("missing or invalid one-time enrollment code")}
  _,portText,err:=net.SplitHostPort(listen);if err!=nil{return err}
  port,err:=strconv.Atoi(portText);if err!=nil||port<1024||port>65535{return errors.New("invalid agent enrollment listen port")}
  go announceComputerToHub(hub,code,secret,fingerprint,port)
 }
 log.Printf("Shared folders: %d; no SMB/NFS required",len(roots))
 return (&http.Server{Addr:listen,Handler:mux,ReadHeaderTimeout:10*time.Second,IdleTimeout:60*time.Second,MaxHeaderBytes:1<<20}).ListenAndServeTLS(certFile,filepath.Join(dataDir,"agent.key"))
}
