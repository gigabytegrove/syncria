package nasversion

import (
 "bytes"
 "context"
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
 "io"
 "os"
 "path"
 "strings"
)

// IntegrityRemote supports destination-owned capture proofs; only bounded
// streaming buffers pass through the controller. No file payload is stored there.
type IntegrityRemote interface {
 Remote
 Open(string)(io.ReadCloser,error)
 Create(string)(io.WriteCloser,error)
}

const integrityName = ".syncria-proof.json"
const integrityLimit = 4096

// CaptureProof is SHA-256 integrity metadata, not a signature. A NAS
// administrator with write access to both data and proof can forge it.
type CaptureProof struct {
 Schema int `json:"schema"`
 JobID string `json:"job_id"`
 VersionID string `json:"version_id"`
 Relative string `json:"relative"`
 Bytes int64 `json:"bytes"`
 SHA256 string `json:"sha256"`
}

func proofPath(root,jobID,versionID string)string{
 return path.Join(root,".syncria-versions",jobID,versionID,integrityName)
}
func hashRemote(ctx context.Context,remote interface{Open(string)(io.ReadCloser,error)},name string,expected int64)(string,error){
 if expected<0||expected>1<<50{return "",ErrUnsafePath}
 f,err:=remote.Open(name);if err!=nil{return "",err}
 defer f.Close()
 h:=sha256.New()
 buf:=make([]byte,128<<10)
 var read int64
 for {
  if err:=ctx.Err();err!=nil{return "",err}
  n,e:=f.Read(buf)
  if n>0{
   read+=int64(n)
   if read>expected{return "",errors.New("archived file grew during hash")}
   _,_=h.Write(buf[:n])
  }
  if e==io.EOF{break}
  if e!=nil{return "",e}
  if n==0{return "",io.ErrNoProgress}
 }
 if read!=expected{return "",io.ErrUnexpectedEOF}
 return hex.EncodeToString(h.Sum(nil)),nil
}
func readProof(remote interface{Open(string)(io.ReadCloser,error)},root,jobID,versionID string)(CaptureProof,bool,error){
 p:=proofPath(root,jobID,versionID)
 f,err:=remote.Open(p)
 if os.IsNotExist(err){return CaptureProof{},false,nil}
 if err!=nil{return CaptureProof{},false,err}
 defer f.Close()
 b,e:=io.ReadAll(io.LimitReader(f,integrityLimit+1))
 if e!=nil{return CaptureProof{},false,e}
 if len(b)>integrityLimit{return CaptureProof{},false,errors.New("archive proof too large")}
 var proof CaptureProof
 d:=json.NewDecoder(bytes.NewReader(b));d.DisallowUnknownFields()
 if e:=d.Decode(&proof);e!=nil{return CaptureProof{},false,fmt.Errorf("bad archive proof: %w",e)}
 var extra any
 if !errors.Is(d.Decode(&extra),io.EOF){return CaptureProof{},false,errors.New("trailing archive proof bytes")}
 if proof.Schema!=1||proof.JobID!=jobID||proof.VersionID!=versionID||!safeRel(proof.Relative)||proof.Bytes<0||proof.Bytes>1<<50||len(proof.SHA256)!=64||strings.ToLower(proof.SHA256)!=proof.SHA256 {
  return CaptureProof{},false,errors.New("invalid archive proof identity")
 }
 if _,e:=hex.DecodeString(proof.SHA256);e!=nil{return CaptureProof{},false,errors.New("invalid SHA-256 proof")}
 return proof,true,nil
}
// writeProof runs AFTER preserving the previous file but BEFORE publishing
// its replacement. The caller must roll back on any failure.
func writeProof(ctx context.Context,remote IntegrityRemote,root,jobID,versionID,relative,archive string,old os.FileInfo)error{
 meta,err:=remote.Stat(archive);if err!=nil{return err}
 if !meta.Mode().IsRegular()||meta.Size()!=old.Size()||!meta.ModTime().Equal(old.ModTime()){return errors.New("archived version changed during capture")}
 sum,err:=hashRemote(ctx,remote,archive,old.Size());if err!=nil{return err}
 meta,err=remote.Stat(archive);if err!=nil{return err}
 if !meta.Mode().IsRegular()||meta.Size()!=old.Size()||!meta.ModTime().Equal(old.ModTime()){return errors.New("archived version changed during hashing")}
 proof:=CaptureProof{Schema:1,JobID:jobID,VersionID:versionID,Relative:relative,Bytes:old.Size(),SHA256:sum}
 b,err:=json.Marshal(proof);if err!=nil{return err}
 p:=proofPath(root,jobID,versionID)
 if _,err=remote.Stat(p);err==nil{return errors.New("archive proof already exists")}else if !os.IsNotExist(err){return err}
 if err:=ctx.Err();err!=nil{return err}
 f,err:=remote.Create(p);if err!=nil{return err}
 n,we:=f.Write(b);closeErr:=f.Close()
 if we!=nil{return we}
 if closeErr!=nil{return closeErr}
 if n!=len(b){return io.ErrShortWrite}
 recorded,ok,err:=readProof(remote,root,jobID,versionID)
 if err!=nil{return err}
 if !ok||recorded!=proof{return errors.New("NAS did not retain verified capture proof")}
 return nil
}
