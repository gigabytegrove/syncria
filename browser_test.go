package main

import (
 "net/http"
 "net/http/httptest"
 "path/filepath"
 "runtime"
 "strings"
 "testing"
)

func TestBrowseContainment(t *testing.T) {
 root:=filepath.Join(string(filepath.Separator),"srv","syncria")
 if !contained(root,root) {t.Fatal("root should be included")}
 if !contained(root,filepath.Join(root,"nas","folder")) {t.Fatal("child should be included")}
 if contained(root,filepath.Join(root,"..","private")) {t.Fatal("parent escape accepted")}
 if contained(root,root+"-other") {t.Fatal("prefix collision accepted")}
}
func TestBrowseRequiresAuthentication(t *testing.T) {
 a:=&App{}
 for _,path:=range []string{"/api/browse/local","/api/browse/drive?account=example&parent=root"} {
  req:=httptest.NewRequest(http.MethodGet,path,nil)
  w:=httptest.NewRecorder()
  if strings.Contains(path,"/local") {a.browseLocal(w,req)}else{a.browseDrive(w,req)}
  if w.Code!=http.StatusUnauthorized{t.Fatalf("%s expected 401, got %d",path,w.Code)}
 }
}
func TestBrowseStorageRootOverride(t *testing.T) {
 if runtime.GOOS=="windows"{t.Skip("Unix-specific test root")}
 root:=t.TempDir();t.Setenv("SYNCRIA_STORAGE_ROOT",root)
 if got:=browseStorageRoot();got!=root {t.Fatalf("got %q want %q",got,root)}
}
