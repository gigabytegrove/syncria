package main

import (
 "bufio"
 "context"
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
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

const releaseAPI = "https://api.github.com/repos/gigabytegrove/syncria/releases/latest"

type releaseAsset struct { Name string `json:"name"`; URL string `json:"url"` }
type releaseInfo struct { Tag string `json:"tag_name"`; Assets []releaseAsset `json:"assets"` }

func releaseRequest(ctx context.Context, url, accept string) (*http.Response,error) {
 req,err:=http.NewRequestWithContext(ctx,"GET",url,nil); if err!=nil{return nil,err}
 req.Header.Set("Accept",accept)
 req.Header.Set("User-Agent","Syncria-Updater/"+appVersion)
 if tok:=os.Getenv("SYNCRIA_GITHUB_TOKEN");tok!="" { req.Header.Set("Authorization","Bearer "+tok) }
 return (&http.Client{Timeout:15*time.Minute}).Do(req)
}
func latestRelease(ctx context.Context)(releaseInfo,error){
 resp,err:=releaseRequest(ctx,releaseAPI,"application/vnd.github+json");if err!=nil{return releaseInfo{},err};defer resp.Body.Close()
 if resp.StatusCode!=200{return releaseInfo{},fmt.Errorf("GitHub release lookup returned HTTP %d",resp.StatusCode)}
 var r releaseInfo;err=json.NewDecoder(io.LimitReader(resp.Body,2<<20)).Decode(&r)
 if err!=nil{return r,err};if r.Tag==""||len(r.Assets)==0{return r,errors.New("release has no assets")};return r,nil
}
func assetName()string{return "syncria-"+runtime.GOOS+"-"+runtime.GOARCH+func()string{if runtime.GOOS=="windows"{return ".exe"};return ""}()}
func assetFor(r releaseInfo,name string)(releaseAsset,bool){for _,a:=range r.Assets{if a.Name==name{return a,true}};return releaseAsset{},false}
func downloadAsset(ctx context.Context,a releaseAsset, dst string)error{
 if a.URL==""{return errors.New("release asset has no API URL")}
 resp,err:=releaseRequest(ctx,a.URL,"application/octet-stream");if err!=nil{return err};defer resp.Body.Close()
 if resp.StatusCode!=200{return fmt.Errorf("asset download HTTP %d",resp.StatusCode)}
 if strings.Contains(resp.Header.Get("Content-Type"),"json") {return errors.New("GitHub returned JSON instead of release asset; check permissions")}
 out,err:=os.OpenFile(dst,os.O_CREATE|os.O_EXCL|os.O_WRONLY,0700);if err!=nil{return err}
 _,copyErr:=io.Copy(out,io.LimitReader(resp.Body,1<<31));syncErr:=out.Sync();closeErr:=out.Close()
 if copyErr!=nil{return copyErr};if syncErr!=nil{return syncErr};return closeErr
}
func expectedChecksum(b []byte,name string)(string,error){
 sc:=bufio.NewScanner(strings.NewReader(string(b)))
 for sc.Scan(){p:=strings.Fields(sc.Text());if len(p)==2&&strings.TrimPrefix(p[1],"*")==name&&len(p[0])==64{if _,err:=hex.DecodeString(p[0]);err==nil{return strings.ToLower(p[0]),nil}}}
 return "",fmt.Errorf("no valid SHA256 for %s",name)
}
func updateBinary(ctx context.Context, dir string, r releaseInfo)(string,error){
 binary,ok:=assetFor(r,assetName());if !ok{return "",fmt.Errorf("release %s has no binary for %s/%s",r.Tag,runtime.GOOS,runtime.GOARCH)}
 sums,ok:=assetFor(r,"SHA256SUMS");if !ok{return "",errors.New("release lacks SHA256SUMS")}
 base:=filepath.Join(dir,"runtime");if err:=os.MkdirAll(base,0700);err!=nil{return "",err}
 tmpSum,err:=os.CreateTemp(base,".checksums-");if err!=nil{return "",err};sumPath:=tmpSum.Name();tmpSum.Close();os.Remove(sumPath);defer os.Remove(sumPath)
 if err=downloadAsset(ctx,sums,sumPath);err!=nil{return "",err};sumData,err:=os.ReadFile(sumPath);if err!=nil{return "",err}
 want,err:=expectedChecksum(sumData,binary.Name);if err!=nil{return "",err}
 tmp,err:=os.CreateTemp(base,".binary-");if err!=nil{return "",err};path:=tmp.Name();tmp.Close();os.Remove(path);defer os.Remove(path)
 if err=downloadAsset(ctx,binary,path);err!=nil{return "",err}
 f,err:=os.Open(path);if err!=nil{return "",err};hash:=sha256.New();_,err=io.Copy(hash,f);f.Close();if err!=nil{return "",err}
 if hex.EncodeToString(hash.Sum(nil))!=want{return "",errors.New("download checksum mismatch; update rejected")}
 if err=os.Chmod(path,0700);err!=nil{return "",err}
 ext:="";if runtime.GOOS=="windows"{ext=".exe"}
 current:=filepath.Join(base,"current"+ext);previous:=filepath.Join(base,"previous"+ext)
 if _,err=os.Stat(current);err==nil{os.Remove(previous);if err=os.Rename(current,previous);err!=nil{return "",err}}else if !os.IsNotExist(err){return "",err}else{
  exe,err:=os.Executable();if err==nil {if b,err:=os.ReadFile(exe);err==nil {os.WriteFile(previous,b,0700)}}
 }
 if err=os.Rename(path,current);err!=nil{os.Rename(previous,current);return "",err}
 return current,nil
}
func rollbackBinary(dir string)(string,error){
 base:=filepath.Join(dir,"runtime");ext:="";if runtime.GOOS=="windows"{ext=".exe"}
 current:=filepath.Join(base,"current"+ext);previous:=filepath.Join(base,"previous"+ext)
 if _,err:=os.Stat(previous);err!=nil{return "",errors.New("no previous binary available")}
 old:=filepath.Join(base,".rollback-"+ext);os.Remove(old)
 if err:=os.Rename(current,old);err!=nil{return "",err}
 if err:=os.Rename(previous,current);err!=nil{os.Rename(old,current);return "",err}
 os.Rename(old,previous);return current,nil
}
