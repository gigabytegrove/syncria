package main

import (
 "os"
 "path/filepath"
 "strings"
 "testing"
)

func TestManualUpdateSectionPresent(t *testing.T){
 for _,v:=range []string{"Manual update","/api/update/manual","name=\"binary\"","name=\"checksums\"","multipart/form-data"} {
  if !strings.Contains(newUI,v){t.Errorf("missing %q",v)}
 }
}
func TestVerifiedManualBinaryPromotion(t *testing.T){
 dir:=t.TempDir()
 source:=filepath.Join(dir,"upload")
 if err:=os.WriteFile(source,[]byte("new"),0600);err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(filepath.Join(dir,"current"),[]byte("old"),0700);err!=nil{t.Fatal(err)}
 if err:=installVerifiedManualBinary(source,dir);err!=nil{t.Fatal(err)}
 now,err:=os.ReadFile(filepath.Join(dir,"current"));if err!=nil{t.Fatal(err)}
 old,err:=os.ReadFile(filepath.Join(dir,"previous"));if err!=nil{t.Fatal(err)}
 if string(now)!="new"||string(old)!="old"{t.Fatalf("current %q previous %q",now,old)}
}
