package main

import (
 "context"
 "crypto/sha256"
 "encoding/hex"
 "errors"
 "fmt"
 "io"
 "net/http"
 "os"
 "path/filepath"
 "runtime"
 "strings"
 "time"
)

func computerAgentVersionTag()string{
 if strings.Contains(appVersion,"-alpha."){return "alpha-"+appVersion}
 if strings.Contains(appVersion,"-beta.")||strings.Contains(appVersion,"-rc."){return "beta-"+appVersion}
 return appVersion
}
func validComputerAgentAsset(filename string)bool{
 switch filename {
 case "SHA256SUMS","syncria-windows-amd64.exe","syncria-windows-arm64.exe",
  "syncria-linux-amd64","syncria-linux-arm64","syncria-linux-arm":
  return true
 }
 return false
}
func (a *App) ensureComputerAgentAsset(filename string)(string,string,error){
 if !validComputerAgentAsset(filename){return "","",errors.New("unsupported agent platform")}
 a.computerAssetMu.Lock();defer a.computerAssetMu.Unlock()
 folder:=filepath.Join(filepath.Dir(a.file),"runtime","agent-downloads",computerAgentVersionTag())
 if err:=os.MkdirAll(folder,0700);err!=nil{return "","",err}
 manifest:=filepath.Join(folder,"SHA256SUMS")
 tag:=computerAgentVersionTag()
 base:="https://github.com/gigabytegrove/syncria/releases/download/"+tag+"/"
 download:=func(name,path string)error{
  ctx,cancel:=context.WithTimeout(context.Background(),4*time.Minute);defer cancel()
  req,err:=http.NewRequestWithContext(ctx,http.MethodGet,base+name,nil);if err!=nil{return err}
  req.Header.Set("User-Agent","Syncria-Agent-Server/"+appVersion)
  resp,err:=(&http.Client{Timeout:4*time.Minute}).Do(req);if err!=nil{return err}
  defer resp.Body.Close()
  if resp.StatusCode!=200{return fmt.Errorf("agent release %s returned HTTP %d",name,resp.StatusCode)}
  temp,err:=os.CreateTemp(folder,".agent-download-");if err!=nil{return err}
  defer os.Remove(temp.Name())
  if err:=temp.Chmod(0600);err!=nil{temp.Close();return err}
  written,err:=io.Copy(temp,io.LimitReader(resp.Body,30<<20+1))
  if err!=nil{temp.Close();return err}
  if written>30<<20{temp.Close();return errors.New("agent download exceeded 30 MiB")}
  if err:=temp.Sync();err!=nil{temp.Close();return err}
  if err:=temp.Close();err!=nil{return err}
  return os.Rename(temp.Name(),path)
 }
 if _,err:=os.Stat(manifest);os.IsNotExist(err){if err:=download("SHA256SUMS",manifest);err!=nil{return "","",err}}
 sums,err:=os.ReadFile(manifest);if err!=nil{return "","",err}
 if len(sums)>8192{return "","",errors.New("invalid SHA256 manifest length")}
 if filename=="SHA256SUMS"{return manifest,"",nil}
 want:=""
 for _,line:=range strings.Split(string(sums),"\n"){
  fields:=strings.Fields(line)
  if len(fields)==2&&fields[1]==filename&&len(fields[0])==64{
   if _,err:=hex.DecodeString(fields[0]);err==nil{want=strings.ToLower(fields[0])}
  }
 }
 if want==""{return "","",errors.New("published manifest missing requested agent binary")}
 location:=filepath.Join(folder,filename)
 check:=func()bool{
  f,err:=os.Open(location);if err!=nil{return false};defer f.Close()
  h:=sha256.New();n,err:=io.Copy(h,io.LimitReader(f,30<<20+1))
  return err==nil&&n<=30<<20&&hex.EncodeToString(h.Sum(nil))==want
 }
 if !check(){
  if err:=download(filename,location);err!=nil{return "","",err}
  if !check(){os.Remove(location);return "","",errors.New("downloaded agent failed GitHub release SHA-256 verification")}
 }
 return location,want,nil
}
func (a *App) computerAgentAssetAPI(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodGet{computerError(w,405,"GET required");return}
 code:=r.URL.Query().Get("code")
 invite,ok:=a.getComputerInvite(code)
 if !ok||invite.State=="connected"{computerError(w,403,"Enrollment invitation expired or used");return}
 filename:=r.URL.Query().Get("file")
 if !validComputerAgentAsset(filename){computerError(w,400,"Unknown agent asset");return}
 path,sum,err:=a.ensureComputerAgentAsset(filename)
 if err!=nil{computerError(w,503,"Agent package is not ready: "+err.Error());return}
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("X-Content-Type-Options","nosniff")
 if sum!=""{w.Header().Set("X-Syncria-SHA256",sum)}
 w.Header().Set("Content-Disposition",fmt.Sprintf("attachment; filename=%q",filename))
 if filename=="SHA256SUMS"{w.Header().Set("Content-Type","text/plain; charset=utf-8")}else{w.Header().Set("Content-Type","application/octet-stream")}
 http.ServeFile(w,r,path)
}
// Embedded SHA-256 is checked by Syncria before hosting a release file;
// installers independently verify the served executable with SHA256SUMS.
// The server always serves the exact version it is currently running.
func servedAgentArchitecture()string{
 return runtime.GOARCH
}
