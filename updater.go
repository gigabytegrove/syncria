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

const tagsAPI = "https://api.github.com/repos/gigabytegrove/syncria/tags?per_page=100"
const releaseForTagAPI = "https://api.github.com/repos/gigabytegrove/syncria/releases/tags/"

type releaseAsset struct { Name string `json:"name"`; URL string `json:"url"` }
type releaseInfo struct { Tag string `json:"tag_name"`; Assets []releaseAsset `json:"assets"` }
type githubTag struct { Name string `json:"name"` }

func releaseRequest(ctx context.Context, endpoint, accept string) (*http.Response,error) {
 // Public releases should not require credentials. An incorrectly configured
 // token can cause GitHub 403 responses even for otherwise public content.
 request:=func(token string)(*http.Response,error){
  req,err:=http.NewRequestWithContext(ctx,http.MethodGet,endpoint,nil);if err!=nil{return nil,err}
  req.Header.Set("Accept",accept)
  req.Header.Set("User-Agent","Syncria-Updater/"+appVersion)
  if token!=""{req.Header.Set("Authorization","Bearer "+token)}
  return (&http.Client{Timeout:15*time.Minute}).Do(req)
 }
 token:=strings.TrimSpace(os.Getenv("SYNCRIA_GITHUB_TOKEN"))
 resp,err:=request(token)
 if err!=nil{return nil,err}
 if token!=""&&(resp.StatusCode==http.StatusUnauthorized||resp.StatusCode==http.StatusForbidden){
  resp.Body.Close()
  return request("")
 }
 return resp,nil
}
func githubAPIError(resp *http.Response, operation string) error {
 var body struct{Message string `json:"message"`}
 _=json.NewDecoder(io.LimitReader(resp.Body,4096)).Decode(&body)
 detail:=strings.TrimSpace(body.Message)
 if resp.StatusCode==http.StatusForbidden && resp.Header.Get("X-RateLimit-Remaining")=="0"{
  detail="GitHub API rate limit reached; retry after the limit resets"
 }
 if detail==""{return fmt.Errorf("%s returned HTTP %d",operation,resp.StatusCode)}
 return fmt.Errorf("%s returned HTTP %d: %s",operation,resp.StatusCode,detail)
}
func versionParts(tag string)([3]int,bool){
 var p [3]int
 tag=strings.TrimPrefix(tag,"v")
 if _,err:=fmt.Sscanf(tag,"%d.%d.%d",&p[0],&p[1],&p[2]);err!=nil{return p,false}
 return p,true
}
func newerVersion(a,b string)bool{
 x,ok:=versionParts(a);if !ok{return false}
 y,ok:=versionParts(b);if !ok{return true}
 for i:=0;i<3;i++{if x[i]!=y[i]{return x[i]>y[i]}}
 return false
}
func latestTag(ctx context.Context)(string,error){
 resp,err:=releaseRequest(ctx,tagsAPI,"application/vnd.github+json");if err!=nil{return "",err};defer resp.Body.Close()
 if resp.StatusCode!=200{return "",githubAPIError(resp,"GitHub tag lookup")}
 var tags []githubTag
 if err=json.NewDecoder(io.LimitReader(resp.Body,1<<20)).Decode(&tags);err!=nil{return "",err}
 best:=""
 for _,tag:=range tags{if newerVersion(tag.Name,best){best=tag.Name}}
 if best==""{return "",errors.New("no version tags found")}
 return best,nil
}
func releaseForTag(ctx context.Context,tag string)(releaseInfo,error){
 resp,err:=releaseRequest(ctx,releaseForTagAPI+tag,"application/vnd.github+json");if err!=nil{return releaseInfo{},err};defer resp.Body.Close()
 if resp.StatusCode==404{return releaseInfo{},fmt.Errorf("%s tag exists, but its release binaries are not published yet",tag)}
 if resp.StatusCode!=200{return releaseInfo{},githubAPIError(resp,"GitHub release lookup")}
 var r releaseInfo
 if err=json.NewDecoder(io.LimitReader(resp.Body,2<<20)).Decode(&r);err!=nil{return r,err}
 if _,ok:=assetFor(r,assetName());!ok{return r,fmt.Errorf("%s has no binary for this platform",tag)}
 if _,ok:=assetFor(r,"SHA256SUMS");!ok{return r,fmt.Errorf("%s has no checksum file",tag)}
 return r,nil
}
func latestRelease(ctx context.Context)(releaseInfo,error){
 tag,err:=latestTag(ctx);if err!=nil{return releaseInfo{},err}
 return releaseForTag(ctx,tag)
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
