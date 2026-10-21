package nasversion

import (
 "context"
 "errors"
 "os"
 "path"
 "strings"
 "testing"
 "time"
)

type retentionDisk struct{*disk}
func (d *retentionDisk)Remove(p string)error{return os.Remove(d.name(p))}
func (d *retentionDisk)ReadDir(p string)([]os.FileInfo,error){return d.disk.ReadDir(p)}

func retentionFixture(t *testing.T)(*retentionDisk,[]Version){
 t.Helper()
 d,live,temp:=setup(t)
 for i,old:=range []string{"oldest","middle","newest"} {
  if e:=os.WriteFile(d.name(live),[]byte(old),0600);e!=nil{t.Fatal(e)}
  if e:=os.WriteFile(d.name(temp),[]byte("next"),0600);e!=nil{t.Fatal(e)}
  v,e:=Promote(context.Background(),d,"folder","x.txt",temp,job)
  if e!=nil{t.Fatal(e)}
  stamp:=time.Date(2026,10,i+1,12,0,0,0,time.UTC)
  if e:=os.Chtimes(d.name(path.Dir(v.VersionPath)),stamp,stamp);e!=nil{t.Fatal(e)}
 }
 items,e:=List(context.Background(),d,"folder",job)
 if e!=nil||len(items)!=3{t.Fatalf("archive inventory %v: %v",items,e)}
 return &retentionDisk{disk:d},items
}
func TestPreviewAndExplicitRetentionRequiresMatchingDigest(t *testing.T){
 d,items:=retentionFixture(t)
 plan,e:=PreviewRetention(context.Background(),d,"folder",job,1,1<<20)
 if e!=nil{t.Fatal(e)}
 if len(plan.Delete)!=2||len(plan.Digest)!=64{t.Fatalf("unsafe plan %#v",plan)}
 listed,e:=List(context.Background(),d,"folder",job)
 if e!=nil||len(listed)!=3{t.Fatal("preview changed existing history")}
 if n,e:=ApplyRetention(context.Background(),d,"folder",job,1,1<<20,strings.Repeat("0",64));e==nil||n!=0{t.Fatal("invalid preview digest accepted")}
 removed,e:=ApplyRetention(context.Background(),d,"folder",job,1,1<<20,plan.Digest)
 if e!=nil||removed!=2{t.Fatalf("removed=%d err=%v",removed,e)}
 after,e:=List(context.Background(),d,"folder",job)
 if e!=nil||len(after)!=1||after[0].VersionID!=items[0].VersionID{t.Fatalf("newest snapshot not preserved: %v %v",after,e)}
 verify(t,d.disk,"folder/x.txt","next")
 verify(t,d.disk,after[0].Path,"newest")
}
func TestRetentionRefusesRemovingLastRecoveryVersion(t *testing.T){
 d,_:=retentionFixture(t)
 if _,e:=PreviewRetention(context.Background(),d,"folder",job,1,1);e==nil{t.Fatal("last recoverable version would be deleted")}
 if listed,e:=List(context.Background(),d,"folder",job);e!=nil||len(listed)!=3{t.Fatal("quota check deleted data")}
}
func TestChangedArchiveOrCanceledPruneFailsClosed(t *testing.T){
 d,items:=retentionFixture(t)
 plan,e:=PreviewRetention(context.Background(),d,"folder",job,1,1<<20);if e!=nil{t.Fatal(e)}
 if e:=os.WriteFile(d.name(items[1].Path),[]byte("tampered and resized"),0600);e!=nil{t.Fatal(e)}
 if removed,e:=ApplyRetention(context.Background(),d,"folder",job,1,1<<20,plan.Digest);e==nil||removed!=0{t.Fatal("stale snapshot accepted")}
 ctx,cancel:=context.WithCancel(context.Background());cancel()
 if _,e:=PreviewRetention(ctx,d,"folder",job,1,1<<20);!errors.Is(e,context.Canceled){t.Fatal(e)}
}
func TestRetentionRejectsPathEscape(t *testing.T){
 d,_:=retentionFixture(t)
 if _,e:=PreviewRetention(context.Background(),d,"../outside",job,1,1<<20);!errors.Is(e,ErrUnsafePath){t.Fatal(e)}
 if n,e:=ApplyRetention(context.Background(),d,"folder",job,1,1<<20,"bad");!errors.Is(e,ErrUnsafePath)||n!=0{t.Fatal(e)}
}
