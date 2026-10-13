package main

import (
 "bytes"
 "context"
 "crypto/md5"
 "encoding/hex"
 "errors"
 "io"
 "strings"
 "testing"
)

type fakeBridgeEndpoint struct{
 files map[string]string
 folders map[string]bool
}
func (m *fakeBridgeEndpoint) Close(){}
func (m *fakeBridgeEndpoint) Scan(ctx context.Context)(map[string]bridgeEntry,error){
 out:=map[string]bridgeEntry{}
 for p,v:=range m.files{
  h:=md5.Sum([]byte(v))
  out[p]=bridgeEntry{ID:"file-"+p,Size:int64(len(v)),Fingerprint:hex.EncodeToString(h[:])}
 }
 for p:=range m.folders{out[p]=bridgeEntry{ID:"folder-"+p,Folder:true}}
 return out,nil
}
func (m *fakeBridgeEndpoint) Mkdir(ctx context.Context,p string)(bridgeEntry,error){
 if m.folders==nil{m.folders=map[string]bool{}}
 m.folders[p]=true
 return bridgeEntry{ID:"folder-"+p,Folder:true},nil
}
func (m *fakeBridgeEndpoint) Open(ctx context.Context,p string)(io.ReadCloser,int64,error){
 v,ok:=m.files[p];if !ok{return nil,0,errors.New("missing test file")}
 return io.NopCloser(strings.NewReader(v)),int64(len(v)),nil
}
func (m *fakeBridgeEndpoint) Put(ctx context.Context,p string,reader io.Reader,size int64,want string,onBytes func(int64))(bridgeEntry,error){
 var buf bytes.Buffer
 n,err:=io.Copy(&buf,reader);if err!=nil{return bridgeEntry{},err}
 if n!=size{return bridgeEntry{},errors.New("invalid stream length")}
 got:=md5.Sum(buf.Bytes());digest:=hex.EncodeToString(got[:])
 if want!=""&&!strings.EqualFold(gotMD5(want),digest){return bridgeEntry{},errors.New("checksum mismatch")}
 if onBytes!=nil{onBytes(n)}
 if m.files==nil{m.files=map[string]string{}}
 m.files[p]=buf.String()
 return bridgeEntry{ID:"file-"+p,Size:n,Fingerprint:digest},nil
}
func gotMD5(h string)string{return h}
func TestDirectBridgeStreamsCloudToNAS(t *testing.T){
 cloud:=&fakeBridgeEndpoint{files:map[string]string{"family/photos/a.txt":"hello"}}
 nas:=&fakeBridgeEndpoint{files:map[string]string{}}
 a:=&App{}
 j:=&Job{ID:"a",Provider:"bridge",State:map[string]Item{}}
 if err:=a.executeBridge(context.Background(),j,cloud,nas);err!=nil{t.Fatal(err)}
 if nas.files["family/photos/a.txt"]!="hello"{t.Fatalf("NAS did not receive streamed contents")}
 if j.LastUploads!=1||j.LastUploadBytes!=5{t.Fatalf("transfer metrics wrong: %+v",j)}
 if len(j.State)!=1{t.Fatalf("missing sync checkpoint")}
}
func TestDirectBridgeStreamsNASToCloud(t *testing.T){
 cloud:=&fakeBridgeEndpoint{files:map[string]string{}}
 nas:=&fakeBridgeEndpoint{files:map[string]string{"a.txt":"remote"}}
 a:=&App{}
 j:=&Job{ID:"a",Provider:"bridge",State:map[string]Item{}}
 if err:=a.executeBridge(context.Background(),j,cloud,nas);err!=nil{t.Fatal(err)}
 if cloud.files["a.txt"]!="remote"||j.LastDownloads!=1{t.Fatalf("reverse stream did not complete")}
}
func TestDirectBridgeCanAdoptIdenticalFilesWithoutOverwriting(t *testing.T){
 cloud:=&fakeBridgeEndpoint{files:map[string]string{"a.txt":"identical"}}
 nas:=&fakeBridgeEndpoint{files:map[string]string{"a.txt":"identical"}}
 a:=&App{};j:=&Job{ID:"a",State:map[string]Item{}}
 if err:=a.executeBridge(context.Background(),j,cloud,nas);err!=nil{t.Fatal(err)}
 if j.LastUploads!=0||j.LastDownloads!=0||len(j.State)!=1{t.Fatalf("identical existing file was copied or not tracked")}
}
func TestDirectBridgeRejectsDifferentInitialFiles(t *testing.T){
 cloud:=&fakeBridgeEndpoint{files:map[string]string{"a.txt":"cloud"}}
 nas:=&fakeBridgeEndpoint{files:map[string]string{"a.txt":"nas"}}
 a:=&App{};j:=&Job{ID:"a",State:map[string]Item{}}
 if err:=a.executeBridge(context.Background(),j,cloud,nas);err==nil{t.Fatal("conflicting files overwritten")}
 if cloud.files["a.txt"]!="cloud"||nas.files["a.txt"]!="nas"{t.Fatal("conflict changed data")}
}
func TestDirectBridgeRejectsDeletions(t *testing.T){
 cloud:=&fakeBridgeEndpoint{files:map[string]string{}}
 nas:=&fakeBridgeEndpoint{files:map[string]string{"a.txt":"last synced"}}
 a:=&App{};j:=&Job{ID:"a",State:map[string]Item{"a.txt":{ID:"file-a.txt",MD5:"was-synced",LocalMD5:"synced"}}}
 if err:=a.executeBridge(context.Background(),j,cloud,nas);err==nil{t.Fatal("deleted remote file was propagated")}
 if nas.files["a.txt"]!="last synced"{t.Fatal("deletion touched NAS data")}
}
func TestNASPathsCannotEscapeShare(t *testing.T){
 for _,candidate:=range []string{"../etc","/root","a/../../x","a\\..\\..\\b"}{
  if _,err:=safeRemotePath(candidate);err==nil{t.Fatalf("accepted NAS traversal %q",candidate)}
 }
}
