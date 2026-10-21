package nasversion
import (
 "context"
 "os"
 "path"
 "strings"
 "testing"
)
func TestEmptyVerifiedArchiveFailsClosed(t *testing.T){
 r,id:=verifiedFixture(t)
 filename:=path.Join("folder",".syncria-versions",job,id,"x.txt")
 if e:=os.Remove(r.name(filename));e!=nil{t.Fatal(e)}
 if entries,e:=List(context.Background(),r,"folder",job);e==nil{t.Fatalf("incomplete archive listed: %v",entries)}
 verify(t,r.disk,"folder/x.txt","new")
}
type listingWithoutOpen struct{d *disk}
func (m listingWithoutOpen)ReadDir(p string)([]os.FileInfo,error){return m.d.ReadDir(p)}
func TestProofBearingArchiveCannotBeDowngradedWithoutOpen(t *testing.T){
 r,_:=verifiedFixture(t)
 if _,e:=List(context.Background(),listingWithoutOpen{d:r.disk},"folder",job);e==nil{t.Fatal("proof-bearing archive silently treated as legacy")}
}
func TestEmptyLegacyArchiveFailsClosed(t *testing.T){
 r,id:=archivedTestFile(t)
 p:=path.Join("folder",".syncria-versions",job,id,"x.txt")
 if e:=os.Remove(r.name(p));e!=nil{t.Fatal(e)}
 if _,e:=List(context.Background(),r,"folder",job);e==nil||!strings.Contains(e.Error(),"incomplete"){t.Fatalf("incomplete legacy history accepted: %v",e)}
}
