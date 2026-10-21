package nasversion

import (
 "context"
 "crypto/rand"
 "crypto/sha256"
 "encoding/hex"
 "errors"
 "fmt"
 "io"
 "os"
 "path"
 "strings"
)

// RestoreRemote streams directly between NAS files; the controller does not
// store recovered content on its own disks.
type RestoreRemote interface {
 Remote
 Open(string)(io.ReadCloser,error)
 Create(string)(io.WriteCloser,error)
 Remove(string)error
}
type RestoreResult struct {
 Relative string `json:"relative"`
 Bytes int64 `json:"bytes"`
 SHA256 string `json:"sha256"`
}
func newRestoreName(relative string)(string,error){
 var nonce [16]byte
 if _,err:=rand.Read(nonce[:]);err!=nil{return "",err}
 ext:=path.Ext(relative);base:=strings.TrimSuffix(relative,ext)
 return base+" (recovered-"+hex.EncodeToString(nonce[:])+")"+ext,nil
}

// RestoreCopy publishes a verified historical copy under a UNIQUE new name.
// Existing live files and retained archives are never intentionally replaced.
// SMB Rename is not guaranteed no-replace on every NAS; a malicious concurrent
// writer can still race the final rename. Validate on real SMB before GA.
func RestoreCopy(ctx context.Context,remote RestoreRemote,root,jobID,versionID,relative string)(RestoreResult,error){
 if remote==nil||!safeRoot(root)||!validID(jobID)||!validID(versionID)||!safeRel(relative){return RestoreResult{},ErrUnsafePath}
 if err:=ctx.Err();err!=nil{return RestoreResult{},err}
 archived:=path.Join(root,".syncria-versions",jobID,versionID,relative)
 sourceInfo,err:=remote.Stat(archived)
 if err!=nil{return RestoreResult{},fmt.Errorf("historical version unavailable: %w",err)}
 if !sourceInfo.Mode().IsRegular()||sourceInfo.Size()<0{return RestoreResult{},ErrUnsafePath}
 restored,err:=newRestoreName(relative)
 if err!=nil{return RestoreResult{},err}
 if !safeRel(restored){return RestoreResult{},ErrUnsafePath}
 destination:=path.Join(root,restored)
 if _,err:=remote.Stat(destination);err==nil{return RestoreResult{},errors.New("recovery destination exists")}else if !os.IsNotExist(err){return RestoreResult{},err}
 nonce,err:=newRestoreName("temp");if err!=nil{return RestoreResult{},err}
 if err:=mkdirAll(remote,path.Dir(destination));err!=nil{return RestoreResult{},err}
 temp:=destination+".syncria-pending-"+path.Base(nonce)
 src,err:=remote.Open(archived);if err!=nil{return RestoreResult{},err}
 defer src.Close()
 dst,err:=remote.Create(temp);if err!=nil{return RestoreResult{},err}
 staged:=true
 defer func(){if staged{_ = remote.Remove(temp)}}()
 var copied int64
 h:=sha256.New()
 buf:=make([]byte,128<<10)
 for {
  if err:=ctx.Err();err!=nil{_ =dst.Close();return RestoreResult{},err}
  n,e:=src.Read(buf)
  if n>0 {
   copied+=int64(n)
   if copied>sourceInfo.Size(){_ =dst.Close();return RestoreResult{},errors.New("source archive expanded during copy")}
   _,_ = h.Write(buf[:n])
   w,we:=dst.Write(buf[:n])
   if we!=nil{_ =dst.Close();return RestoreResult{},we}
   if w!=n{_ =dst.Close();return RestoreResult{},io.ErrShortWrite}
  }
  if e==io.EOF{break}
  if e!=nil{_ =dst.Close();return RestoreResult{},e}
  if n==0{_ =dst.Close();return RestoreResult{},io.ErrNoProgress}
 }
 if err:=dst.Close();err!=nil{return RestoreResult{},err}
 if copied!=sourceInfo.Size(){return RestoreResult{},io.ErrUnexpectedEOF}
 // Re-read what landed on NAS and compare its SHA256 to source stream.
 stagedFile,err:=remote.Open(temp);if err!=nil{return RestoreResult{},err}
 actualHash:=sha256.New()
 count,copyErr:=io.CopyBuffer(actualHash,&contextReader{ctx:ctx,reader:stagedFile},make([]byte,128<<10))
 closeErr:=stagedFile.Close()
 if copyErr!=nil{return RestoreResult{},copyErr}
 if closeErr!=nil{return RestoreResult{},closeErr}
 if count!=copied||!strings.EqualFold(hex.EncodeToString(actualHash.Sum(nil)),hex.EncodeToString(h.Sum(nil))){return RestoreResult{},errors.New("staged NAS bytes failed SHA256 verification")}
 current,err:=remote.Stat(archived);if err!=nil{return RestoreResult{},err}
 if !current.Mode().IsRegular()||current.Size()!=sourceInfo.Size()||!current.ModTime().Equal(sourceInfo.ModTime()){return RestoreResult{},errors.New("archive changed during restoration")}
 if err:=ctx.Err();err!=nil{return RestoreResult{},err}
 if _,err:=remote.Stat(destination);err==nil{return RestoreResult{},errors.New("recovery destination appeared during copy")}else if !os.IsNotExist(err){return RestoreResult{},err}
 if err:=remote.Rename(temp,destination);err!=nil{return RestoreResult{},fmt.Errorf("publish NAS recovery copy: %w",err)}
 staged=false
 return RestoreResult{Relative:restored,Bytes:copied,SHA256:hex.EncodeToString(h.Sum(nil))},nil
}
type contextReader struct{ctx context.Context;reader io.Reader}
func (r *contextReader)Read(p []byte)(int,error){if err:=r.ctx.Err();err!=nil{return 0,err};return r.reader.Read(p)}
