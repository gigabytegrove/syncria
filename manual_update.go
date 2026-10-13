package main

import (
 "crypto/sha256"
 "crypto/subtle"
 "encoding/hex"
 "errors"
 "fmt"
 "io"
 "net/http"
 "os"
 "path/filepath"
 "runtime"
 "time"
)

const maxManualUpdateSize = 80 << 20

// manualUpdate allows recovery while GitHub's API is unavailable.
// The admin supplies both the platform binary and its release SHA256SUMS.
func (a *App) manualUpdate(w http.ResponseWriter,r *http.Request) {
 if !localPost(r){http.Error(w,"Same-origin POST required",http.StatusMethodNotAllowed);return}
 if !a.authenticated(r){http.Error(w,"Unauthorized",http.StatusUnauthorized);return}
 if runtime.GOOS=="windows"{http.Error(w,"In-app binary replacement unsupported on Windows",400);return}
 if err:=a.beginUpdate("manual");err!=nil{http.Error(w,err.Error(),http.StatusConflict);return}
 fail:=func(status int,step string,err error){
  a.finishUpdate("failed",step,err.Error())
  http.Error(w,step+": "+err.Error(),status)
 }
 r.Body=http.MaxBytesReader(w,r.Body,maxManualUpdateSize)
 a.setUpdateProgress("receiving","Receiving the executable and checksum files in this Syncria instance","",-1,-1,-1,"")
 if err:=r.ParseMultipartForm(2<<20);err!=nil{
  fail(400,"Unable to receive supplied update",err);return
 }
 defer r.MultipartForm.RemoveAll()
 binaries:=r.MultipartForm.File["binary"]
 sumsHeaders:=r.MultipartForm.File["checksums"]
 if len(binaries)!=1||len(sumsHeaders)!=1 {
  fail(400,"Invalid manual update input",errors.New("one executable and one SHA256SUMS are required"));return
 }
 if binaries[0].Filename!=assetName()||sumsHeaders[0].Filename!="SHA256SUMS"{
  fail(400,"Incorrect files",fmt.Errorf("provide exactly %s and SHA256SUMS",assetName()));return
 }
 a.updateAsset(assetName())
 a.updateEvent("receiving","Received "+assetName()+" and SHA256SUMS multipart upload")
 sums,err:=sumsHeaders[0].Open();if err!=nil{fail(400,"Unable to read SHA256SUMS",err);return}
 checksumContent,err:=io.ReadAll(io.LimitReader(sums,(1<<20)+1));sums.Close()
 if err!=nil||len(checksumContent)>(1<<20){
  fail(400,"Invalid SHA256SUMS",errors.New("checksum manifest unreadable or larger than one MiB"));return
 }
 expected,err:=expectedChecksum(checksumContent,assetName())
 if err!=nil{fail(400,"Missing checksum for platform binary",err);return}
 a.updateEvent("manifest","Found expected SHA-256 for "+assetName())
 base:=filepath.Join(filepath.Dir(a.file),"runtime")
 if err=os.MkdirAll(base,0700);err!=nil{fail(500,"Cannot prepare updater directory",err);return}
 source,err:=binaries[0].Open()
 if err!=nil{fail(400,"Cannot open update executable",err);return}
 defer source.Close()
 tmp,err:=os.CreateTemp(base,".manual-update-")
 if err!=nil{fail(500,"Cannot create candidate binary",err);return}
 tempName:=tmp.Name();defer os.Remove(tempName)
 fileSize:=binaries[0].Size
 if fileSize<=0||fileSize>70<<20{
  tmp.Close();fail(400,"Invalid executable size",errors.New("executable must be between 1 byte and 70 MiB"));return
 }
 a.setUpdateProgress("verifying","Computing SHA-256 for the supplied "+assetName(),"",-1,0,fileSize,"")
 digest:=sha256.New()
 const blockSize=128*1024
 buf:=make([]byte,blockSize)
 var count int64
 for {
  n,readErr:=source.Read(buf)
  if n>0{
   count+=int64(n)
   if count>fileSize||count>70<<20{tmp.Close();fail(400,"Executable size changed",errors.New("received more bytes than declared"));return}
   if _,err=tmp.Write(buf[:n]);err!=nil{tmp.Close();fail(500,"Unable to stage executable",err);return}
   if _,err=digest.Write(buf[:n]);err!=nil{tmp.Close();fail(500,"Unable to hash executable",err);return}
   a.setUpdateProgress("verifying","Computing SHA-256 for "+assetName(),"",-1,count,fileSize,"")
  }
  if readErr==io.EOF{break}
  if readErr!=nil{tmp.Close();fail(400,"Uploaded executable could not be read",readErr);return}
 }
 if err=tmp.Sync();err!=nil{tmp.Close();fail(500,"Unable to flush executable",err);return}
 if err=tmp.Close();err!=nil{fail(500,"Unable to close executable",err);return}
 if count!=fileSize {
  fail(400,"Uploaded size mismatch",fmt.Errorf("expected %d bytes, verified %d bytes",fileSize,count));return
 }
 actual:=hex.EncodeToString(digest.Sum(nil))
 if subtle.ConstantTimeCompare([]byte(expected),[]byte(actual))!=1 {
  fail(400,"SHA-256 mismatch",errors.New("computed hash does not match SHA256SUMS; current binary was not replaced"));return
 }
 a.updateEvent("verifying","SHA-256 matched SHA256SUMS for "+assetName()+" ("+fmt.Sprint(count)+" bytes)")
 a.setUpdateProgress("staging","Saving rollback executable and replacing current binary","",-1,-1,-1,"")
 if err=installVerifiedManualBinary(tempName,base);err!=nil {
  fail(500,"Unable to install verified executable",err);return
 }
 a.updateEvent("staging","Verified manual executable installed in /data/runtime/current; previous binary saved for rollback")
 a.setUpdateProgress("restarting","Manual executable verified and staged; supervisor restart requested","",-1,-1,-1,"")
 w.Header().Set("Content-Type","application/json; charset=utf-8")
 w.Header().Set("Cache-Control","no-store")
 w.WriteHeader(http.StatusAccepted)
 _,_=io.WriteString(w,`{"accepted":true,"phase":"restarting"}`)
 if flusher,ok:=w.(http.Flusher);ok{flusher.Flush()}
 go func(){
  time.Sleep(1200*time.Millisecond)
  restartSyncria(filepath.Dir(a.file))
  a.finishUpdate("failed","Manual restart failed","The installed executable was staged but no restart process was started")
 }()
}

// A verified file is promoted only after the previous executable is preserved.
func installVerifiedManualBinary(source,base string)error {
 if err:=os.Chmod(source,0700);err!=nil{return err}
 current:=filepath.Join(base,"current")
 previous:=filepath.Join(base,"previous")
 backup:=filepath.Join(base,".manual-previous")
 _=os.Remove(backup)
 if _,err:=os.Stat(current);err==nil {
  if err=os.Rename(current,backup);err!=nil{return err}
 }else if errors.Is(err,os.ErrNotExist){
  executable,err:=os.Executable()
  if err!=nil{return err}
  in,err:=os.Open(executable)
  if err!=nil{return err}
  out,err:=os.OpenFile(backup,os.O_CREATE|os.O_EXCL|os.O_WRONLY,0700)
  if err!=nil{in.Close();return err}
  _,copyErr:=io.Copy(out,in)
  closeErr:=out.Close()
  in.Close()
  if copyErr!=nil||closeErr!=nil {_=os.Remove(backup);return errors.New("could not preserve current binary")}
 }else{return err}
 if err:=os.Rename(source,current);err!=nil {
  if _,statErr:=os.Stat(current);errors.Is(statErr,os.ErrNotExist){_ =os.Rename(backup,current)}
  return err
 }
 _=os.Remove(previous)
 return os.Rename(backup,previous)
}
