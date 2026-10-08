package main

import (
 "net/http"
 "net/http/httptest"
 "testing"
)

func TestNetworkShareValidation(t *testing.T) {
 valid:=[]struct{protocol,server,export,label string}{
  {"nfs","192.168.0.20","/export/media","NAS Media"},
  {"smb","nas.local","Shared","Family Storage"},
 }
 for _,v:=range valid{
  if err:=validateShare(v.protocol,v.server,v.export,v.label);err!=nil{t.Fatalf("expected valid %v: %v",v,err)}
 }
 bad:=[]struct{protocol,server,export,label string}{
  {"ftp","nas.local","Shared","NAS"},
  {"nfs","nas.local","relative/path","NAS"},
  {"smb","nas.local","\\\\host\\share","NAS"},
  {"smb","--evil","Shared","NAS"},
 }
 for _,v:=range bad {
  if err:=validateShare(v.protocol,v.server,v.export,v.label);err==nil{t.Fatalf("expected invalid: %v",v)}
 }
}
func TestShareManagementRequiresAuthentication(t *testing.T){
 app:=&App{}
 req:=httptest.NewRequest(http.MethodPost,"/api/shares/action",nil)
 w:=httptest.NewRecorder()
 app.shareAction(w,req)
 if w.Code!=http.StatusUnauthorized{t.Fatalf("Expected 401, got %d",w.Code)}
}
func TestManagedNetworkMountsDisabledByDefault(t *testing.T){
 t.Setenv("SYNCRIA_ENABLE_MOUNTS","0")
 app:=&App{}
 if err:=app.mountShare(StorageShare{Protocol:"nfs"});err==nil{t.Fatal("mount should be disabled")}
}
