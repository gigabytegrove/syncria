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
 "strconv"
 "sync"
 "strings"
 "time"
)

const tagsAPI = "https://api.github.com/repos/gigabytegrove/syncria/tags?per_page=100"
const releaseForTagAPI = "https://api.github.com/repos/gigabytegrove/syncria/releases/tags/"

type releaseAsset struct { Name string `json:"name"`; URL string `json:"url"` }
type releaseInfo struct { Tag string `json:"tag_name"`; Assets []releaseAsset `json:"assets"`; Draft bool `json:"draft"`; Prerelease bool `json:"prerelease"` }
type githubTag struct { Name string `json:"name"` }

// githubRetryAfter honors secondary limits and primary rate-limit reset headers.
func githubRetryAfter(resp *http.Response) time.Duration {
 // A generic 403 is not a rate limit. In particular a revoked optional
 // GitHub token can return 403. Treating it as a one-hour cooldown prevents
 // anonymous fallback and poisons every subsequent update check.
 if resp.StatusCode!=http.StatusForbidden && resp.StatusCode!=http.StatusTooManyRequests{return 0}
 limited:=resp.StatusCode==http.StatusTooManyRequests||resp.Header.Get("X-RateLimit-Remaining")=="0"
 seconds,hasRetry:=strconv.Atoi(resp.Header.Get("Retry-After"))
 if hasRetry==nil && seconds>0 {limited=true}
 if !limited{return 0}
 wait:=time.Duration(0)
 if hasRetry==nil && seconds>0 {wait=time.Duration(seconds)*time.Second}
 if reset,err:=strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"),10,64);err==nil&&resp.Header.Get("X-RateLimit-Remaining")=="0"{
  until:=time.Until(time.Unix(reset,0));if until>wait{wait=until}
 }
 if wait<=0 {wait=time.Minute}
 if wait>24*time.Hour{wait=24*time.Hour}
 return wait
}
var githubThrottle struct {
 sync.Mutex
 until time.Time
}
func githubBlockedUntil() time.Time {
 githubThrottle.Lock()
 defer githubThrottle.Unlock()
 return githubThrottle.until
}
func recordGitHubThrottle(resp *http.Response){
 wait:=githubRetryAfter(resp)
 if wait<=0{return}
 githubThrottle.Lock()
 until:=time.Now().Add(wait)
 if until.After(githubThrottle.until){githubThrottle.until=until}
 githubThrottle.Unlock()
}
func releaseRequest(ctx context.Context, endpoint, accept string) (*http.Response,error) {
 request:=func(token string)(*http.Response,error){
  req,err:=http.NewRequestWithContext(ctx,http.MethodGet,endpoint,nil);if err!=nil{return nil,err}
  req.Header.Set("Accept",accept)
  req.Header.Set("User-Agent","Syncria-Updater/"+appVersion)
  if token!=""{req.Header.Set("Authorization","Bearer "+token)}
  return (&http.Client{Timeout:15*time.Minute}).Do(req)
 }
 if until:=githubBlockedUntil();time.Now().Before(until){
  return nil,fmt.Errorf("GitHub API cooldown active until %s",until.Format(time.RFC3339))
 }
 token:=strings.TrimSpace(os.Getenv("SYNCRIA_GITHUB_TOKEN"))
 resp,err:=request(token)
 if err!=nil{return nil,err}
 // An invalid or insufficiently scoped optional token must not prevent
 // access to a publicly readable Syncria release. Retry once anonymously.
 if token!="" && (resp.StatusCode==http.StatusUnauthorized || resp.StatusCode==http.StatusForbidden) {
  resp.Body.Close()
  resp,err=request("")
  if err!=nil{return nil,err}
 }
 recordGitHubThrottle(resp)
 return resp,nil
}
func githubAPIError(resp *http.Response, operation string) error {
 var body struct{Message string `json:"message"`}
 _=json.NewDecoder(io.LimitReader(resp.Body,4096)).Decode(&body)
 detail:=strings.TrimSpace(body.Message)
 if resp.StatusCode==http.StatusForbidden && resp.Header.Get("X-RateLimit-Remaining")=="0" {
  detail="GitHub API rate limit reached; waiting for the reset window"
 }
 if resp.StatusCode==http.StatusTooManyRequests{detail="GitHub API throttled requests"}
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
 return latestReleaseForChannel(ctx,updateChannelStable)
}

const releasesAPI = "https://api.github.com/repos/gigabytegrove/syncria/releases?per_page=100"
const updateChannelStable = "stable"
const updateChannelBeta = "beta"
const updateChannelAlpha = "alpha"
func normalizeUpdateChannel(channel string) string {
 switch strings.ToLower(strings.TrimSpace(channel)) {
 case updateChannelAlpha:return updateChannelAlpha
 case updateChannelBeta:return updateChannelBeta
 default:return updateChannelStable
 }
}
func validUpdateChannel(channel string) bool {
 return channel==updateChannelStable||channel==updateChannelBeta||channel==updateChannelAlpha
}
type parsedUpdateVersion struct {
 Major,Minor,Patch int
 Stage string
 Sequence int
}
func parseUpdateVersion(tag string)(parsedUpdateVersion,bool){
 var v parsedUpdateVersion
 prefix:=""
 if strings.HasPrefix(tag,"alpha-"){prefix="alpha";tag=strings.TrimPrefix(tag,"alpha-")} else if strings.HasPrefix(tag,"beta-"){prefix="beta";tag=strings.TrimPrefix(tag,"beta-")}
 tag=strings.TrimPrefix(tag,"v")
 parts:=strings.SplitN(tag,"-",2)
 nums:=strings.Split(parts[0],".")
 if len(nums)!=3{return v,false}
 for i,n:=range nums{
  if n==""||len(n)>8||(len(n)>1&&n[0]=='0'){return v,false}
  for _,c:=range n{if c<'0'||c>'9'{return v,false}}
  z,e:=strconv.Atoi(n);if e!=nil{return v,false}
  switch i{case 0:v.Major=z;case 1:v.Minor=z;case 2:v.Patch=z}
 }
 if len(parts)==1{return v,prefix==""}
 pre:=strings.Split(parts[1],".")
 if len(pre)!=2||pre[1]==""||len(pre[1])>8{return v,false}
 if pre[0]!="alpha"&&pre[0]!="beta"&&pre[0]!="rc"{return v,false}
 if prefix=="alpha"&&pre[0]!="alpha"{return v,false}
 if prefix=="beta"&&pre[0]!="beta"&&pre[0]!="rc"{return v,false}
 if len(pre[1])>1&&pre[1][0]=='0'{return v,false}
 for _,c:=range pre[1]{if c<'0'||c>'9'{return v,false}}
 sequence,e:=strconv.Atoi(pre[1]);if e!=nil{return v,false}
 v.Stage=pre[0];v.Sequence=sequence
 return v,true
}
// The public GitHub tag has a protective channel prefix; the executable's
// actual version never does. Compare the normalized target after restart.
func releaseTagAppVersion(tag string)string{
 if _,ok:=parseUpdateVersion(tag);!ok{return ""}
 if strings.HasPrefix(tag,"alpha-"){return strings.TrimPrefix(tag,"alpha-")}
 if strings.HasPrefix(tag,"beta-"){return strings.TrimPrefix(tag,"beta-")}
 return tag
}
func updateVersionNewer(candidate,running string)bool{
 a,ok:=parseUpdateVersion(candidate);if !ok{return false}
 b,ok:=parseUpdateVersion(running);if !ok{return false} // never guess about a custom runtime version
 ai,bi:=[3]int{a.Major,a.Minor,a.Patch},[3]int{b.Major,b.Minor,b.Patch}
 for i:=range ai{if ai[i]!=bi[i]{return ai[i]>bi[i]}}
 priority:=func(s string)int{switch s{case "alpha":return 0;case "beta":return 1;case "rc":return 2;default:return 3}}
 if priority(a.Stage)!=priority(b.Stage){return priority(a.Stage)>priority(b.Stage)}
 return a.Sequence>b.Sequence
}
// Old v0.8.5 installations read the tag list as stable-only. Prefixing
// alpha/beta tags prevents those older clients from recognizing prereleases.
func updateReleaseEligible(r releaseInfo,channel string)bool{
 tag:=r.Tag
 v,ok:=parseUpdateVersion(tag);if !ok{return false}
 switch {
 case v.Stage=="":
  return !r.Prerelease && !strings.HasPrefix(tag,"alpha-") && !strings.HasPrefix(tag,"beta-")
 case v.Stage=="alpha":
  return channel=="alpha" && r.Prerelease && strings.HasPrefix(tag,"alpha-")
 case v.Stage=="beta"||v.Stage=="rc":
  return (channel=="beta"||channel=="alpha")&&r.Prerelease&&strings.HasPrefix(tag,"beta-")
 default:return false
 }
}
func newestReleaseForChannel(releases []releaseInfo,channel string)(releaseInfo,error){
 channel=normalizeUpdateChannel(channel)
 var best releaseInfo
 for _,r:=range releases{
  if r.Draft||!updateReleaseEligible(r,channel){continue}
  if best.Tag==""||updateVersionNewer(r.Tag,best.Tag){best=r}
 }
 if best.Tag==""{return releaseInfo{},fmt.Errorf("no published releases for the %s channel",channel)}
 return best,nil
}
func latestReleaseForChannel(ctx context.Context,channel string)(releaseInfo,error){
 resp,err:=releaseRequest(ctx,releasesAPI,"application/vnd.github+json")
 if err!=nil{return releaseInfo{},err}
 defer resp.Body.Close()
 if resp.StatusCode!=http.StatusOK{return releaseInfo{},githubAPIError(resp,"GitHub channel release lookup")}
 var releases []releaseInfo
 if err:=json.NewDecoder(io.LimitReader(resp.Body,4<<20)).Decode(&releases);err!=nil{return releaseInfo{},err}
 return newestReleaseForChannel(releases,channel)
}
func verifyReleaseAssets(r releaseInfo)error{
 if _,ok:=assetFor(r,assetName());!ok{return fmt.Errorf("%s does not have a binary for %s/%s yet",r.Tag,runtime.GOOS,runtime.GOARCH)}
 if _,ok:=assetFor(r,"SHA256SUMS");!ok{return fmt.Errorf("%s does not have its SHA256SUMS yet",r.Tag)}
 return nil
}
func assetName()string{return "syncria-"+runtime.GOOS+"-"+runtime.GOARCH+func()string{if runtime.GOOS=="windows"{return ".exe"};return ""}()}
func assetFor(r releaseInfo,name string)(releaseAsset,bool){for _,a:=range r.Assets{if a.Name==name{return a,true}};return releaseAsset{},false}
// downloadAssetWithProgress exposes actual received bytes and Content-Length;
// no extrapolation is used if the server omits the total size.
func downloadAssetWithProgress(ctx context.Context,a releaseAsset,dst string,onBytes func(int64,int64)) error {
 if a.URL==""{return errors.New("release asset has no API URL")}
 resp,err:=releaseRequest(ctx,a.URL,"application/octet-stream");if err!=nil{return err};defer resp.Body.Close()
 if resp.StatusCode!=200{return fmt.Errorf("asset download HTTP %d",resp.StatusCode)}
 if strings.Contains(resp.Header.Get("Content-Type"),"json"){return errors.New("GitHub returned JSON instead of release asset; check permissions")}
 if resp.ContentLength>1<<31{return errors.New("release asset exceeds safe size limit")}
 out,err:=os.OpenFile(dst,os.O_CREATE|os.O_EXCL|os.O_WRONLY,0700);if err!=nil{return err}
 const limit int64=1<<31
 buf:=make([]byte,128*1024)
 var received int64
 last:=time.Time{}
 total:=resp.ContentLength
 if total<0{total=0}
 if onBytes!=nil{onBytes(0,total)}
 for {
  n,readErr:=resp.Body.Read(buf)
  if n>0{
   received+=int64(n)
   if received>limit{out.Close();return errors.New("release binary exceeds safe size limit")}
   if _,err=out.Write(buf[:n]);err!=nil{out.Close();return err}
   if onBytes!=nil&&time.Since(last)>250*time.Millisecond{onBytes(received,total);last=time.Now()}
  }
  if readErr==io.EOF{break}
  if readErr!=nil{out.Close();return readErr}
 }
 if onBytes!=nil{onBytes(received,total)}
 if total>0&&received!=total{out.Close();return fmt.Errorf("download incomplete: received %d of %d bytes",received,total)}
 if received==0{out.Close();return errors.New("downloaded release asset is empty")}
 if err=out.Sync();err!=nil{out.Close();return err}
 return out.Close()
}
func downloadAsset(ctx context.Context,a releaseAsset,dst string) error {
 return downloadAssetWithProgress(ctx,a,dst,nil)
}
func expectedChecksum(b []byte,name string)(string,error){
 sc:=bufio.NewScanner(strings.NewReader(string(b)))
 for sc.Scan(){p:=strings.Fields(sc.Text());if len(p)==2&&strings.TrimPrefix(p[1],"*")==name&&len(p[0])==64{if _,err:=hex.DecodeString(p[0]);err==nil{return strings.ToLower(p[0]),nil}}}
 return "",fmt.Errorf("no valid SHA256 for %s",name)
}
type updateBinaryProgress func(phase string, downloaded,total int64)
func updateBinary(ctx context.Context, dir string, r releaseInfo)(string,error){return updateBinaryWithProgress(ctx,dir,r,nil)}
func updateBinaryWithProgress(ctx context.Context, dir string, r releaseInfo,progress updateBinaryProgress)(string,error){
 binary,ok:=assetFor(r,assetName());if !ok{return "",fmt.Errorf("release %s has no binary for %s/%s",r.Tag,runtime.GOOS,runtime.GOARCH)}
 sums,ok:=assetFor(r,"SHA256SUMS");if !ok{return "",errors.New("release lacks SHA256SUMS")}
 base:=filepath.Join(dir,"runtime");if err:=os.MkdirAll(base,0700);err!=nil{return "",err}
 tmpSum,err:=os.CreateTemp(base,".checksums-");if err!=nil{return "",err};sumPath:=tmpSum.Name();tmpSum.Close();os.Remove(sumPath);defer os.Remove(sumPath)
 if progress!=nil{progress("manifest",0,0)}
 if err=downloadAsset(ctx,sums,sumPath);err!=nil{return "",err}
 sumData,err:=os.ReadFile(sumPath);if err!=nil{return "",err}
 want,err:=expectedChecksum(sumData,binary.Name);if err!=nil{return "",err}
 tmp,err:=os.CreateTemp(base,".binary-");if err!=nil{return "",err};path:=tmp.Name();tmp.Close();os.Remove(path);defer os.Remove(path)
 if err=downloadAssetWithProgress(ctx,binary,path,func(downloaded,total int64){if progress!=nil{progress("downloading",downloaded,total)}});err!=nil{return "",err}
 if progress!=nil{progress("verifying",0,0)}
 f,err:=os.Open(path);if err!=nil{return "",err}
 fileInfo,err:=f.Stat();if err!=nil{f.Close();return "",err}
 hash:=sha256.New()
 chunk:=make([]byte,128*1024)
 var hashed int64
 lastProgress:=time.Time{}
 for {
  n,readErr:=f.Read(chunk)
  if n>0{
   if _,err=hash.Write(chunk[:n]);err!=nil{f.Close();return "",err}
   hashed+=int64(n)
   if progress!=nil&&(time.Since(lastProgress)>120*time.Millisecond||hashed==fileInfo.Size()){
    progress("verifying",hashed,fileInfo.Size())
    lastProgress=time.Now()
   }
  }
  if readErr==io.EOF{break}
  if readErr!=nil{f.Close();return "",readErr}
 }
 if err=f.Close();err!=nil{return "",err}
 if hashed!=fileInfo.Size(){return "",errors.New("Verification stopped before all binary bytes were hashed")}
 if hex.EncodeToString(hash.Sum(nil))!=want{return "",errors.New("download checksum mismatch; update rejected")}
 if err=os.Chmod(path,0700);err!=nil{return "",err}
 if progress!=nil{progress("staging",-1,-1)}
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
