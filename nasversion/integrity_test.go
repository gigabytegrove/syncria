package nasversion

import (
 "context"
 "crypto/sha256"
 "encoding/hex"
 "errors"
 "os"
 "path"
 "strings"
 "testing"
 "time"
)

func verifiedFixture(t *testing.T)(*restorationShare,string){
 t.Helper()
 d,dest,temp:=setup(t)
 if e:=os.WriteFile(d.name(dest),[]byte("archived original"),0600);e!=nil{t.Fatal(e)}
 r:=&restorationShare{disk:d}
 v,e:=PromoteVerified(context.Background(),r,"folder","x.txt",temp,job)
 if e!=nil||!v.Replaced{t.Fatalf("version capture: %v %+v",e,v)}
 return r,path.Base(path.Dir(v.VersionPath))
}
func TestCaptureChecksumAndVerifiedRestore(t *testing.T){
 d,id:=verifiedFixture(t)
 proof,ok,e:=readProof(d,"folder",job,id)
 if e!=nil||!ok{t.Fatalf("proof missing %v %v",ok,e)}
 h:=sha256.Sum256([]byte("archived original"))
 if proof.SHA256!=hex.EncodeToString(h[:])||proof.Relative!="x.txt"{t.Fatalf("wrong proof %+v",proof)}
 items,e:=List(context.Background(),d,"folder",job)
 if e!=nil||len(items)!=1||!items[0].CaptureProofPresent||items[0].CaptureSHA256!=proof.SHA256{t.Fatalf("listing %+v %v",items,e)}
 restored,e:=RestoreCopyVerified(context.Background(),d,"folder",job,id,"x.txt")
 if e!=nil||!restored.CaptureVerified{t.Fatalf("unverified restore %+v %v",restored,e)}
 verify(t,d.disk,"folder/"+restored.Relative,"archived original")
 verify(t,d.disk,"folder/x.txt","new")
}
func TestArchivedContentTamperSameSizeAndTimestamp(t *testing.T){
 d,id:=verifiedFixture(t)
 name:=path.Join("folder",".syncria-versions",job,id,"x.txt")
 st,e:=d.Stat(name);if e!=nil{t.Fatal(e)}
 if e=os.WriteFile(d.name(name),[]byte("altered! original"),0600);e!=nil{t.Fatal(e)}
 if e=os.Chtimes(d.name(name),st.ModTime(),st.ModTime());e!=nil{t.Fatal(e)}
 if _,e=RestoreCopyVerified(context.Background(),d,"folder",job,id,"x.txt");e==nil||!strings.Contains(e.Error(),"capture-time SHA-256"){t.Fatalf("tampering accepted: %v",e)}
 verify(t,d.disk,"folder/x.txt","new")
}
func TestProofMissingOrCorruptFailsClosed(t *testing.T){
 d,id:=verifiedFixture(t)
 proof:=proofPath("folder",job,id)
 if e:=os.Remove(d.name(proof));e!=nil{t.Fatal(e)}
 if _,e:=RestoreCopyVerified(context.Background(),d,"folder",job,id,"x.txt");e==nil{t.Fatal("missing proof accepted")}
 if e:=os.WriteFile(d.name(proof),[]byte("bad proof"),0600);e!=nil{t.Fatal(e)}
 if _,e:=List(context.Background(),d,"folder",job);e==nil{t.Fatal("malformed proof accepted")}
 if _,e:=RestoreCopyVerified(context.Background(),d,"folder",job,id,"x.txt");e==nil{t.Fatal("bad proof accepted")}
}
func TestProofCreationFailureRestoresOriginal(t *testing.T){
 d,dest,temp:=setup(t)
 if e:=os.WriteFile(d.name(dest),[]byte("old file"),0600);e!=nil{t.Fatal(e)}
 r:=&restorationShare{disk:d,failCreate:true}
 if _,e:=PromoteVerified(context.Background(),r,"folder","x.txt",temp,job);e==nil{t.Fatal("accepted proof failure")}
 verify(t,d,dest,"old file")
 verify(t,d,temp,"new")
}
func TestVerifiedRetentionPreventsDeletionAfterTamper(t *testing.T){
 d,olderID:=verifiedFixture(t)
 dest:="folder/x.txt"
 if e:=os.WriteFile(d.name(dest),[]byte("new original"),0600);e!=nil{t.Fatal(e)}
 temp:=dest+".syncria-pending-"+strings.Repeat("b",32)
 if e:=os.WriteFile(d.name(temp),[]byte("newest"),0600);e!=nil{t.Fatal(e)}
 result,e:=PromoteVerified(context.Background(),d,"folder","x.txt",temp,job);if e!=nil{t.Fatal(e)}
 latestID:=path.Base(path.Dir(result.VersionPath))
 olderDir:=path.Join("folder",".syncria-versions",job,olderID)
 latestDir:=path.Join("folder",".syncria-versions",job,latestID)
 oldClock:=time.Date(2026,10,18,10,0,0,0,time.UTC)
 if e:=os.Chtimes(d.name(olderDir),oldClock,oldClock);e!=nil{t.Fatal(e)}
 newClock:=oldClock.Add(24*time.Hour)
 if e:=os.Chtimes(d.name(latestDir),newClock,newClock);e!=nil{t.Fatal(e)}
 plan,e:=PreviewRetention(context.Background(),d,"folder",job,1,1<<20)
 if e!=nil||len(plan.Delete)!=1||plan.Delete[0].VersionID!=olderID{t.Fatalf("plan %+v %v",plan,e)}
 archive:=path.Join(olderDir,"x.txt")
 stat,e:=d.Stat(archive);if e!=nil{t.Fatal(e)}
 if e:=os.WriteFile(d.name(archive),[]byte("altered! original"),0600);e!=nil{t.Fatal(e)}
 if e:=os.Chtimes(d.name(archive),stat.ModTime(),stat.ModTime());e!=nil{t.Fatal(e)}
 if n,e:=ApplyRetention(context.Background(),d,"folder",job,1,1<<20,plan.Digest);n!=0||e==nil{t.Fatalf("deleted tampered data: %d %v",n,e)}
 verify(t,d.disk,dest,"newest")
 if e:=os.WriteFile(d.name(archive),[]byte("archived original"),0600);e!=nil{t.Fatal(e)}
 if e:=os.Chtimes(d.name(archive),stat.ModTime(),stat.ModTime());e!=nil{t.Fatal(e)}
 plan,e=PreviewRetention(context.Background(),d,"folder",job,1,1<<20);if e!=nil{t.Fatal(e)}
 n,e:=ApplyRetention(context.Background(),d,"folder",job,1,1<<20,plan.Digest)
 if e!=nil||n!=1{t.Fatalf("prune %d %v",n,e)}
 if _,e:=os.Stat(d.name(olderDir));!errors.Is(e,os.ErrNotExist){t.Fatalf("orphan dir %v",e)}
 versions,e:=List(context.Background(),d,"folder",job)
 if e!=nil||len(versions)!=1||versions[0].VersionID!=latestID{t.Fatalf("postprune %+v %v",versions,e)}
}
func TestLegacyArchiveNotReportedVerified(t *testing.T){
 d,id:=archivedTestFile(t)
 versions,e:=List(context.Background(),d,"folder",job)
 if e!=nil||len(versions)!=1||versions[0].CaptureProofPresent{t.Fatalf("legacy falsely labeled %+v %v",versions,e)}
 if _,e:=RestoreCopyVerified(context.Background(),d,"folder",job,id,"x.txt");e==nil{t.Fatal("controller restored unverified legacy archive")}
 legacy,e:=RestoreCopy(context.Background(),d,"folder",job,id,"x.txt")
 if e!=nil||legacy.CaptureVerified{t.Fatalf("legacy falsely verified %+v %v",legacy,e)}
}
