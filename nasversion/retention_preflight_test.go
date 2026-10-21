package nasversion
import (
 "context"
 "errors"
 "io"
 "os"
 "path"
 "strings"
 "testing"
 "time"
)
func verifiedCleanupFixture(t *testing.T)(*restorationShare,[]Version){
 t.Helper()
 d,live,temp:=setup(t)
 r:=&restorationShare{disk:d}
 for i,data:=range []string{"old-v1","old-v2","old-v3"}{
  if e:=os.WriteFile(d.name(live),[]byte(data),0600);e!=nil{t.Fatal(e)}
  if e:=os.WriteFile(d.name(temp),[]byte("next-v"),0600);e!=nil{t.Fatal(e)}
  v,e:=PromoteVerified(context.Background(),r,"folder","x.txt",temp,job);if e!=nil{t.Fatal(e)}
  stamp:=time.Date(2026,10,10+i,12,0,0,0,time.UTC)
  if e:=os.Chtimes(d.name(path.Dir(v.VersionPath)),stamp,stamp);e!=nil{t.Fatal(e)}
 }
 copies,e:=List(context.Background(),r,"folder",job)
 if e!=nil||len(copies)!=3{t.Fatalf("versions %#v %v",copies,e)}
 for _,v:=range copies{if !v.CaptureProofPresent{t.Fatal("capture proof missing")}}
 return r,copies
}
func TestRetentionPreflightCorruptLastCandidateRemovesNothing(t *testing.T){
 r,items:=verifiedCleanupFixture(t)
 plan,e:=PreviewRetention(context.Background(),r,"folder",job,1,1<<20)
 if e!=nil||len(plan.Delete)!=2{t.Fatalf("plan %#v %v",plan,e)}
 last:=plan.Delete[len(plan.Delete)-1]
 st,e:=r.Stat(last.Path);if e!=nil{t.Fatal(e)}
 if e:=os.WriteFile(r.name(last.Path),[]byte(strings.Repeat("!",int(last.Bytes))),0600);e!=nil{t.Fatal(e)}
 if e:=os.Chtimes(r.name(last.Path),st.ModTime(),st.ModTime());e!=nil{t.Fatal(e)}
 n,e:=ApplyRetention(context.Background(),r,"folder",job,1,1<<20,plan.Digest)
 if e==nil||n!=0{t.Fatalf("partial deletion: %d %v",n,e)}
 for _,v:=range items{if _,e:=r.Stat(v.Path);e!=nil{t.Fatalf("archive %s lost: %v",v.VersionID,e)}}
 verify(t,r.disk,"folder/x.txt","next-v")
}
type failedLaterRead struct{*restorationShare;fail string}
func (r *failedLaterRead)Open(p string)(io.ReadCloser,error){
 if p==r.fail{return nil,errors.New("injected SMB read failure")}
 return r.restorationShare.Open(p)
}
func TestRetentionPreflightLaterReadFailureRemovesNothing(t *testing.T){
 r,items:=verifiedCleanupFixture(t)
 plan,e:=PreviewRetention(context.Background(),r,"folder",job,1,1<<20)
 if e!=nil||len(plan.Delete)!=2{t.Fatalf("plan %v",e)}
 faulty:=&failedLaterRead{restorationShare:r,fail:plan.Delete[len(plan.Delete)-1].Path}
 n,e:=ApplyRetention(context.Background(),faulty,"folder",job,1,1<<20,plan.Digest)
 if e==nil||n!=0{t.Fatalf("partial deletion: %d %v",n,e)}
 for _,v:=range items{if _,e:=r.Stat(v.Path);e!=nil{t.Fatal(e)}}
}
func TestRetentionPreflightCancelledPreservesArchive(t *testing.T){
 r,items:=verifiedCleanupFixture(t)
 plan,e:=PreviewRetention(context.Background(),r,"folder",job,1,1<<20);if e!=nil{t.Fatal(e)}
 ctx,cancel:=context.WithCancel(context.Background());cancel()
 n,e:=ApplyRetention(ctx,r,"folder",job,1,1<<20,plan.Digest)
 if !errors.Is(e,context.Canceled)||n!=0{t.Fatalf("cancel %d %v",n,e)}
 for _,v:=range items{if _,e:=r.Stat(v.Path);e!=nil{t.Fatal(e)}}
}
