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
 if !localPost(r) {http.Error(w,"Same-origin POST required",http.StatusMethodNotAllowed);return}
 if !a.authenticated(r) {http.Error(w,"Unauthorized",http.StatusUnauthorized);return}
 if runtime.GOOS=="windows" {redirect(w,r,"In-app restart is not supported on Windows");return}
 active:=false
 a.runs.Range(func(_, _ any)bool{active=true;return false})
 if active {redirect(w,r,"Update blocked while synchronization is running");return}
 r.Body=http.MaxBytesReader(w,r.Body,maxManualUpdateSize)
 if err:=r.ParseMultipartForm(2<<20);err!=nil {http.Error(w,"Invalid or oversized update upload",http.StatusBadRequest);return}
 defer r.MultipartForm.RemoveAll()
 binaryHeader:=r.MultipartForm.File["binary"]
 sumsHeader:=r.MultipartForm.File["checksums"]
 if len(binaryHeader)!=1||len(sumsHeader)!=1 {http.Error(w,"Select one binary and one SHA256SUMS",400);return}
 if binaryHeader[0].Filename!=assetName()||sumsHeader[0].Filename!="SHA256SUMS" {
  http.Error(w,fmt.Sprintf("Required files: %s and SHA256SUMS",assetName()),400);return
 }
 sums,err:=sumsHeader[0].Open()
 if err!=nil {http.Error(w,"Cannot read checksums",400);return}
 contents,err:=io.ReadAll(io.LimitReader(sums,1<<20+1))
 sums.Close()
 if err!=nil||len(contents)>1<<20 {http.Error(w,"Invalid SHA256SUMS",400);return}
 expected,err:=expectedChecksum(contents,assetName())
 if err!=nil {http.Error(w,err.Error(),400);return}
 base:=filepath.Join(filepath.Dir(a.file),"runtime")
 if err=os.MkdirAll(base,0700);err!=nil {http.Error(w,err.Error(),500);return}
 source,err:=binaryHeader[0].Open()
 if err!=nil {http.Error(w,"Cannot read uploaded binary",400);return}
 defer source.Close()
 temp,err:=os.CreateTemp(base,".manual-update-")
 if err!=nil {http.Error(w,err.Error(),500);return}
 tempName:=temp.Name()
 defer os.Remove(tempName)
 digest:=sha256.New()
 count,copyErr:=io.Copy(io.MultiWriter(temp,digest),io.LimitReader(source,70<<20+1))
 syncErr:=temp.Sync()
 closeErr:=temp.Close()
 if copyErr!=nil||syncErr!=nil||closeErr!=nil||count>70<<20||count==0 {
  http.Error(w,"Unable to save valid update binary",400);return
 }
 actual:=hex.EncodeToString(digest.Sum(nil))
 if subtle.ConstantTimeCompare([]byte(expected),[]byte(actual))!=1 {
  http.Error(w,"SHA256 verification failed; no files were installed",400);return
 }
 if err=installVerifiedManualBinary(tempName,base);err!=nil {
  http.Error(w,"Update installation failed: "+err.Error(),500);return
 }
 redirect(w,r,"Verified manual update installed; Syncria is restarting.")
 go func(){time.Sleep(500*time.Millisecond);restartSyncria(filepath.Dir(a.file))}()
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
