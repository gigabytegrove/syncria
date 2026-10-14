package main
import ("strings";"testing")
func TestPortableDriveNamesRoundTrip(t *testing.T){
 names:=[]string{"Bland, Mike - Security Camera/P2P Setup","Camera\\Setup","~syncria~custom","regular.txt","文件/摄像","sync.ffs_db.txt"}
 seen:=map[string]bool{}
 for _,name:=range names{
  encoded:=portableDriveName(name)
  if strings.ContainsAny(encoded,"/\\")||!safeName(encoded){t.Fatalf("unsafe encoded component %q",encoded)}
  if originalDriveName(encoded)!=name{t.Fatalf("did not round trip %q as %q",name,encoded)}
  if seen[strings.ToLower(encoded)]{t.Fatalf("collision for %q",name)}
  seen[strings.ToLower(encoded)]=true
 }
 if portableDriveName("regular.txt")!="regular.txt"{t.Fatal("normal names must not change")}
 if portableDriveName("A/B")==portableDriveName("A~syncria~B"){t.Fatal("collision")}
}


func TestDuplicateDriveSiblingsAreStableAndReversible(t *testing.T){
 name:="configs.xml.20260316-1302050218.backup"
 first:=portableDriveObjectName(name,"fileID_ONE_1234",true)
 second:=portableDriveObjectName(name,"fileID_TWO_5678",true)
 if first==second{t.Fatal("different Drive objects must not share a NAS path")}
 for _,item:=range []string{first,second}{
  if !safeName(item){t.Fatalf("mapped name is unsafe: %q",item)}
  if originalDriveName(item)!=name{t.Fatalf("could not recover original Drive filename: %q",item)}
 }
 if portableDriveObjectName(name,"fileID_ONE_1234",true)!=first{t.Fatal("mapping must be stable")}
}
