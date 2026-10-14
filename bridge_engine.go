package main

import (
 "context"
 "crypto/md5"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
 "io"
 "mime"
 "net/http"
 "net/url"
 "path"
 "sort"
 "strconv"
 "strings"
 "time"
)

// A bridgeEndpoint exposes network operations without revealing local OS
// file paths. Neither endpoint is allowed to write file payloads to /data.
type bridgeEntry struct {
 ID string
 Fingerprint string
 Size int64
 Modified time.Time
 Folder bool
}
type bridgeEndpoint interface {
 Scan(context.Context)(map[string]bridgeEntry,error)
 Mkdir(context.Context,string)(bridgeEntry,error)
 Open(context.Context,string)(io.ReadCloser,int64,error)
 Put(context.Context,string,io.Reader,int64,string,func(int64))(bridgeEntry,error)
 Close()
}
type bridgeSMBEndpoint struct {
 client *directSMB
 root string
 scanHook func(string,int64)
}
func (b *bridgeSMBEndpoint) Close(){if b.client!=nil{b.client.Close()}}
func (b *bridgeSMBEndpoint) Scan(ctx context.Context)(map[string]bridgeEntry,error){
 files,err:=b.client.scanWithProgress(ctx,b.root,b.scanHook);if err!=nil{return nil,err}
 out:=make(map[string]bridgeEntry,len(files))
 for rel,f:=range files{out[rel]=bridgeEntry{Size:f.Size,Fingerprint:f.Fingerprint,Folder:!f.File,Modified:f.Modified}}
 return out,nil
}
func (b *bridgeSMBEndpoint) Mkdir(ctx context.Context,rel string)(bridgeEntry,error){
 if err:=ctx.Err();err!=nil{return bridgeEntry{},err}
 if err:=b.client.mkdir(b.root,rel);err!=nil{return bridgeEntry{},err}
 return bridgeEntry{Folder:true},nil
}
func (b *bridgeSMBEndpoint) Open(ctx context.Context,rel string)(io.ReadCloser,int64,error){
 if err:=ctx.Err();err!=nil{return nil,0,err}
 return b.client.open(b.root,rel)
}
func (b *bridgeSMBEndpoint) Put(ctx context.Context,rel string,source io.Reader,size int64,expectedMD5 string,onBytes func(int64))(bridgeEntry,error){
 if err:=b.client.put(ctx,b.root,rel,source,size,expectedMD5,onBytes);err!=nil{return bridgeEntry{},err}
 target,err:=joinRemotePath(b.root,rel);if err!=nil{return bridgeEntry{},err}
 st,err:=b.client.share.Stat(target);if err!=nil{return bridgeEntry{},err}
 return bridgeEntry{Size:st.Size(),Fingerprint:fmt.Sprintf("%d:%d",st.Size(),st.ModTime().UnixNano()),Modified:st.ModTime()},nil
}

type bridgeDriveEndpoint struct {
 app *App
 account,root string
 folders map[string]string
 files map[string]driveFile
 scanHook func(string,int64)
}
func (b *bridgeDriveEndpoint) Close(){}
func (b *bridgeDriveEndpoint) Scan(ctx context.Context)(map[string]bridgeEntry,error){
 files,err:=b.app.remoteTreeProgress(ctx,b.account,b.root,b.scanHook);if err!=nil{return nil,err}
 b.folders=map[string]string{"":b.root}
 b.files=files
 out:=make(map[string]bridgeEntry,len(files))
 for rel,f:=range files{
  size,_:=strconv.ParseInt(f.Size,10,64)
  if f.MimeType==folderMime{b.folders[rel]=f.ID;out[rel]=bridgeEntry{ID:f.ID,Folder:true};continue}
  if f.MD5Checksum==""{return nil,fmt.Errorf("Google Drive file %q lacks a checksum; sync safely stopped",rel)}
  modified,_:=time.Parse(time.RFC3339Nano,f.ModifiedTime)
  out[rel]=bridgeEntry{ID:f.ID,Fingerprint:f.MD5Checksum,Size:size,Modified:modified}
 }
 return out,nil
}
func (b *bridgeDriveEndpoint) Mkdir(ctx context.Context,rel string)(bridgeEntry,error){
 if b.folders==nil{b.folders=map[string]string{"":b.root}}
 if id:=b.folders[rel];id!=""{return bridgeEntry{ID:id,Folder:true},nil}
 parent:=b.folders[parentRel(rel)]
 if parent==""{return bridgeEntry{},fmt.Errorf("Google Drive parent missing for %q",rel)}
 f,err:=b.app.makeFolder(ctx,b.account,parent,originalDriveName(basename(rel)));if err!=nil{return bridgeEntry{},err}
 b.folders[rel]=f.ID
 return bridgeEntry{ID:f.ID,Folder:true},nil
}
func (b *bridgeDriveEndpoint) Open(ctx context.Context,rel string)(io.ReadCloser,int64,error){
 f,ok:=b.files[rel];if !ok||f.MimeType==folderMime{return nil,0,fmt.Errorf("Drive file %q not found",rel)}
 resp,err:=b.app.drive(ctx,b.account,http.MethodGet,"https://www.googleapis.com/drive/v3/files/"+url.PathEscape(f.ID)+"?alt=media",nil,"")
 if err!=nil{return nil,0,err}
 size,_:=strconv.ParseInt(f.Size,10,64)
 if f.Size==""{size=-1}
 return resp.Body,size,nil
}
func (b *bridgeDriveEndpoint) Put(ctx context.Context,rel string,input io.Reader,size int64,expectedMD5 string,onBytes func(int64))(bridgeEntry,error){
 if size<0{return bridgeEntry{},errors.New("Google Drive uploads require a known source file size")}
 name:=originalDriveName(basename(rel))
 parent:=b.folders[parentRel(rel)]
 if parent==""{return bridgeEntry{},fmt.Errorf("Google Drive parent missing for %s",rel)}
 meta:=map[string]any{"name":name}
 method:=http.MethodPost
 endpoint:="https://www.googleapis.com/upload/drive/v3/files?uploadType=resumable&fields=id,name,mimeType,md5Checksum,size"
 if existing,ok:=b.files[rel];ok&&!strings.EqualFold(existing.MimeType,folderMime){
  method=http.MethodPatch
  endpoint="https://www.googleapis.com/upload/drive/v3/files/"+url.PathEscape(existing.ID)+"?uploadType=resumable&fields=id,name,mimeType,md5Checksum,size"
 }else{
  meta["parents"]=[]string{parent}
 }
 data,_:=json.Marshal(meta)
 token,err:=b.app.token(ctx,b.account);if err!=nil{return bridgeEntry{},err}
 req,err:=http.NewRequestWithContext(ctx,method,endpoint,strings.NewReader(string(data)));if err!=nil{return bridgeEntry{},err}
 req.Header.Set("Authorization","Bearer "+token)
 req.Header.Set("Content-Type","application/json; charset=UTF-8")
 ctype:=mime.TypeByExtension(path.Ext(name));if ctype==""{ctype="application/octet-stream"}
 req.Header.Set("X-Upload-Content-Type",ctype)
 req.Header.Set("X-Upload-Content-Length",strconv.FormatInt(size,10))
 resp,err:=b.app.client.Do(req);if err!=nil{return bridgeEntry{},err}
 resp.Body.Close()
 if resp.StatusCode/100!=2{return bridgeEntry{},fmt.Errorf("Google Drive upload start returned HTTP %d",resp.StatusCode)}
 session:=resp.Header.Get("Location")
 if !strings.HasPrefix(session,"https://www.googleapis.com/")&&!strings.HasPrefix(session,"https://www.google.com/"){return bridgeEntry{},errors.New("Unexpected Google Drive resumable upload host")}
 digest:=md5.New()
 pr:=&bridgeProgressReader{input:io.TeeReader(input,digest),callback:onBytes}
 req,err=http.NewRequestWithContext(ctx,http.MethodPut,session,pr);if err!=nil{return bridgeEntry{},err}
 req.ContentLength=size
 req.Header.Set("Content-Type",ctype)
 req.Header.Set("Authorization","Bearer "+token)
 resp,err=b.app.client.Do(req);if err!=nil{return bridgeEntry{},err}
 defer resp.Body.Close()
 if resp.StatusCode/100!=2{
  info,_:=io.ReadAll(io.LimitReader(resp.Body,4096))
  return bridgeEntry{},fmt.Errorf("Google Drive upload HTTP %d: %s",resp.StatusCode,strings.TrimSpace(string(info)))
 }
 var f driveFile
 if err=json.NewDecoder(io.LimitReader(resp.Body,1<<20)).Decode(&f);err!=nil{return bridgeEntry{},err}
 if f.ID==""||f.MD5Checksum==""{return bridgeEntry{},errors.New("Drive did not confirm upload identity and checksum")}
 checksum:=hex.EncodeToString(digest.Sum(nil))
 if pr.read!=size||!strings.EqualFold(checksum,f.MD5Checksum){
  return bridgeEntry{},fmt.Errorf("Drive upload verification failed: local %d bytes md5 %s, Drive md5 %s",pr.read,checksum,f.MD5Checksum)
 }
 if expectedMD5!=""&&!strings.EqualFold(expectedMD5,checksum){return bridgeEntry{},errors.New("Source changed while uploading to Google Drive")}
 b.files[rel]=f
 return bridgeEntry{ID:f.ID,Fingerprint:f.MD5Checksum,Size:size},nil
}
type bridgeProgressReader struct {
 input io.Reader
 callback func(int64)
 read int64
}
func (p *bridgeProgressReader) Read(dst []byte)(int,error){
 n,e:=p.input.Read(dst)
 p.read+=int64(n)
 if n>0&&p.callback!=nil{p.callback(p.read)}
 return n,e
}

// The bridge engine is provider-independent. File bytes pass from one
// io.ReadCloser to the other endpoint via bounded read buffers only.
func (a *App) executeBridge(ctx context.Context,j *Job,cloud,network bridgeEndpoint)error {
 defer cloud.Close()
 defer network.Close()
 left,err:=cloud.Scan(ctx);if err!=nil{return fmt.Errorf("Cloud scan failed: %w",err)}
 right,err:=network.Scan(ctx);if err!=nil{return fmt.Errorf("NAS scan failed: %w",err)}
 keys:=map[string]bool{}
 for rel:=range left{keys[rel]=true}
 for rel:=range right{keys[rel]=true}
 for rel:=range j.State{keys[rel]=true}
 ordered:=make([]string,0,len(keys));for rel:=range keys{ordered=append(ordered,rel)}
 sort.Slice(ordered,func(i,k int)bool{di,dj:=depthKey(ordered[i]),depthKey(ordered[k]);if di!=dj{return di<dj};return ordered[i]<ordered[k]})
 next:=make(map[string]Item,len(j.State));for rel,state:=range j.State{next[rel]=state}
 var toNAS,toCloud int64
 for index,rel:=range ordered{
  if err:=ctx.Err();err!=nil{return err}
  if j.itemHook!=nil{j.itemHook(int64(index),int64(len(ordered)),rel)}
  c,hasCloud:=left[rel];n,hasNAS:=right[rel];prior,tracked:=j.State[rel]
  if !hasCloud&&!hasNAS{delete(next,rel);continue}
  if tracked&&(!hasCloud||!hasNAS){return fmt.Errorf("File deletion detected at %s; Syncria did not delete or restore files automatically",rel)}
  if hasCloud&&hasNAS&&c.Folder!=n.Folder{return fmt.Errorf("File/folder conflict at %s",rel)}
  if hasCloud&&c.Folder || hasNAS&&n.Folder{
   if !hasCloud{c,err=cloud.Mkdir(ctx,rel);if err!=nil{return err}}
   if !hasNAS{n,err=network.Mkdir(ctx,rel);if err!=nil{return err}}
   next[rel]=Item{ID:c.ID,Folder:true}
  }else{
   switch{
   case hasCloud&&hasNAS:
    if !tracked{
     if c.Size==n.Size {
      same,checkErr:=bridgeSameContent(ctx,network,rel,c.Fingerprint)
      if checkErr!=nil{return fmt.Errorf("Cannot verify existing NAS content at %s: %w",rel,checkErr)}
      if same {
       next[rel]=Item{ID:c.ID,MD5:c.Fingerprint,LocalMD5:n.Fingerprint}
       j.State=next
       if j.itemHook!=nil{j.itemHook(int64(index+1),int64(len(ordered)),rel)}
       continue
      }
     }
     return fmt.Errorf("Initial conflict at %s; cloud and NAS contain different files. Neither will be overwritten",rel)
    }
    cloudChanged:=prior.MD5!=c.Fingerprint||prior.ID!=c.ID
    nasChanged:=prior.LocalMD5!=n.Fingerprint
    if cloudChanged&&nasChanged{return fmt.Errorf("Both sides changed %s; choose a resolution before syncing",rel)}
    if cloudChanged{
     var nbytes int64
     n,err=bridgeCopy(ctx,cloud,network,rel,c.Fingerprint,func(bytes int64){nbytes=bytes;a.setSyncProgress(j.ID,func(p *SyncProgress){p.Current=rel;p.UploadBytes=j.LastUploadBytes+bytes})})
     if err!=nil{return err};recordTransfer(j,true,nbytes);toNAS++
    }else if nasChanged{
     var nbytes int64
     c,err=bridgeCopy(ctx,network,cloud,rel,"",func(bytes int64){nbytes=bytes;a.setSyncProgress(j.ID,func(p *SyncProgress){p.Current=rel;p.DownloadBytes=j.LastDownloadBytes+bytes})})
     if err!=nil{return err};recordTransfer(j,false,nbytes);toCloud++
    }
   case hasCloud:
    var nbytes int64
    n,err=bridgeCopy(ctx,cloud,network,rel,c.Fingerprint,func(bytes int64){nbytes=bytes;a.setSyncProgress(j.ID,func(p *SyncProgress){p.Current=rel;p.UploadBytes=j.LastUploadBytes+bytes})})
    if err!=nil{return err};recordTransfer(j,true,nbytes);toNAS++
   case hasNAS:
    var nbytes int64
    c,err=bridgeCopy(ctx,network,cloud,rel,"",func(bytes int64){nbytes=bytes;a.setSyncProgress(j.ID,func(p *SyncProgress){p.Current=rel;p.DownloadBytes=j.LastDownloadBytes+bytes})})
    if err!=nil{return err};recordTransfer(j,false,nbytes);toCloud++
   }
   next[rel]=Item{ID:c.ID,MD5:c.Fingerprint,LocalMD5:n.Fingerprint}
  }
  j.State=next
  if j.itemHook!=nil{j.itemHook(int64(index+1),int64(len(ordered)),rel)}
 }
 j.State=next
 j.LastSummary=fmt.Sprintf("%d files to NAS, %d files to cloud",toNAS,toCloud)
 return nil
}
func bridgeSameContent(ctx context.Context,endpoint bridgeEndpoint,rel,want string)(bool,error){
 if want==""{return false,nil}
 reader,_,err:=endpoint.Open(ctx,rel);if err!=nil{return false,err}
 defer reader.Close()
 digest:=md5.New()
 if _,err=io.Copy(digest,reader);err!=nil{return false,err}
 return strings.EqualFold(hex.EncodeToString(digest.Sum(nil)),want),nil
}
func bridgeCopy(ctx context.Context,src,dst bridgeEndpoint,rel,checksum string,notify func(int64))(bridgeEntry,error){
 reader,size,err:=src.Open(ctx,rel);if err!=nil{return bridgeEntry{},err}
 defer reader.Close()
 return dst.Put(ctx,rel,reader,size,checksum,notify)
}
