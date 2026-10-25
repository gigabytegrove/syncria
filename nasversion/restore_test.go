package nasversion

import (
 "context"
 "crypto/sha256"
 "encoding/hex"
 "errors"
 "io"
 "os"
 "path"
 "strings"
 "testing"
)
type restorationShare struct{*disk;corrupt,failCreate,failStagedRead bool}
func (s *restorationShare)Open(p string)(io.ReadCloser,error){
 if s.failStagedRead&&strings.Contains(p,".syncria-pending-"){return nil,errors.New("injected staged read failure")}
 return os.Open(s.name(p))
}
func (s *restorationShare)Create(p string)(io.WriteCloser,error){
 if s.failCreate{return nil,errors.New("injected create failure")}
 if s.corrupt{return &corruptNASWriter{path:s.name(p)},nil}
 return os.Create(s.name(p))
}
func (s *restorationShare)Remove(p string)error{return os.Remove(s.name(p))}
type corruptNASWriter struct{path string;file *os.File}
func (w *corruptNASWriter)Write(p []byte)(int,error){
 if w.file==nil{f,e:=os.Create(w.path);if e!=nil{return 0,e};w.file=f}
 data:=append([]byte(nil),p...);for i:=range data{data[i]^=0xff}
 return w.file.Write(data)
}
func (w *corruptNASWriter)Close()error{if w.file==nil{return nil};return w.file.Close()}
func archivedTestFile(t *testing.T)(*restorationShare,string){
 t.Helper();d,dest,temp:=setup(t)
 if e:=os.WriteFile(d.name(dest),[]byte("original bytes"),0600);e!=nil{t.Fatal(e)}
 v,e:=Promote(context.Background(),d,"folder","x.txt",temp,job)
 if e!=nil{t.Fatal(e)}
 return &restorationShare{disk:d},path.Base(path.Dir(v.VersionPath))
}
func TestRestoreKeepsLiveAndArchivedCopies(t *testing.T){
 d,id:=archivedTestFile(t)
 result,e:=RestoreCopy(context.Background(),d,"folder",job,id,"x.txt")
 if e!=nil{t.Fatal(e)}
 verify(t,d.disk,"folder/x.txt","new")
 verify(t,d.disk,"folder/"+result.Relative,"original bytes")
 verify(t,d.disk,path.Join("folder",".syncria-versions",job,id,"x.txt"),"original bytes")
 hash:=sha256.Sum256([]byte("original bytes"))
 if result.Bytes!=14||result.SHA256!=hex.EncodeToString(hash[:]){t.Fatalf("result %#v",result)}
 if _,e=RestoreCopy(context.Background(),d,"folder",job,id,"x.txt");e!=nil{t.Fatal(e)}
}
func TestRestoreRejectsInvalidAndCanceled(t *testing.T){
 d,id:=archivedTestFile(t)
 for _,r:=range []string{"../escape",".syncria-hidden/x","folder/../../x"}{
  if _,e:=RestoreCopy(context.Background(),d,"folder",job,id,r);!errors.Is(e,ErrUnsafePath){t.Errorf("accepted path %q: %v",r,e)}
 }
 ctx,cancel:=context.WithCancel(context.Background());cancel()
 if _,e:=RestoreCopy(ctx,d,"folder",job,id,"x.txt");!errors.Is(e,context.Canceled){t.Fatal(e)}
 verify(t,d.disk,"folder/x.txt","new")
}
func TestRestoreNeverPublishesCorruptedOrIncompleteStaging(t *testing.T){
 for _,kind:=range []string{"corrupt","create","readback"}{
  t.Run(kind,func(t *testing.T){
   d,id:=archivedTestFile(t)
   switch kind{case "corrupt":d.corrupt=true;case "create":d.failCreate=true;case "readback":d.failStagedRead=true}
   if _,e:=RestoreCopy(context.Background(),d,"folder",job,id,"x.txt");e==nil{t.Fatal("unsafe restoration succeeded")}
   verify(t,d.disk,"folder/x.txt","new")
   entries,e:=os.ReadDir(d.name("folder"));if e!=nil{t.Fatal(e)}
   for _,v:=range entries{if strings.Contains(v.Name(),"(recovered-"){t.Fatalf("left unrecoverable output: %s",v.Name())}}
  })
 }
}
