package main

import (
    "os"
    "path/filepath"
    "runtime"
    "strings"
    "testing"
)
func TestChecksumManifest(t *testing.T) {
    name := assetName()
    hash := strings.Repeat("a",64)
    got,err:=expectedChecksum([]byte(hash+"  "+name+"\n"),name)
    if err!=nil || got!=hash {t.Fatalf("valid checksum: %s, %v",got,err)}
    if _,err:=expectedChecksum([]byte(hash+"  something-else\n"),name);err==nil {t.Fatal("mismatched filename accepted")}
    if _,err:=expectedChecksum([]byte("not-a-sha  "+name+"\n"),name);err==nil {t.Fatal("invalid digest accepted")}
}
func TestRollback(t *testing.T){
    dir:=t.TempDir();runtimeDir:=filepath.Join(dir,"runtime")
    if err:=os.MkdirAll(runtimeDir,0700);err!=nil {t.Fatal(err)}
    ext:="";if runtime.GOOS=="windows"{ext=".exe"}
    if err:=os.WriteFile(filepath.Join(runtimeDir,"current"+ext),[]byte("new"),0700);err!=nil{t.Fatal(err)}
    if err:=os.WriteFile(filepath.Join(runtimeDir,"previous"+ext),[]byte("old"),0700);err!=nil{t.Fatal(err)}
    _,err:=rollbackBinary(dir);if err!=nil{t.Fatal(err)}
    got,err:=os.ReadFile(filepath.Join(runtimeDir,"current"+ext));if err!=nil || string(got)!="old"{t.Fatalf("rollback failed %q: %v",got,err)}
}
