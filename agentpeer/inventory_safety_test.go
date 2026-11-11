package agentpeer

import (
 "bytes"
 "context"
 "crypto/sha256"
 "encoding/base64"
 "encoding/hex"
 "fmt"
 "io"
 "net/http"
 "net/http/httptest"
 "path/filepath"
 "strings"
 "sync/atomic"
 "testing"
 "github.com/gigabytegrove/syncria/agenttransfer"
)

func TestFailClosedOnMalformedOrIncompleteInventories(t *testing.T) {
 cases:=[]struct{name,body string;good bool}{
  {"empty-array","{\"entries\":[]}",true},
  {"valid-record","{\"entries\":[{\"path\":\"photos/a.jpg\",\"size\":1,\"modified\":\"2026-10-20T12:00:00Z\",\"directory\":false}]}",true},
  {"empty-response","",false},
  {"missing-entries","{}",false},
  {"null-entries","{\"entries\":null}",false},
  {"truncated","{\"entries\":[",false},
  {"trailing-json","{\"entries\":[]}{\"entries\":[]}",false},
  {"extra-field","{\"entries\":[],\"unrecognized\":true}",false},
  {"trailing-text","{\"entries\":[]} bad",false},
 }
 grant:=strings.Repeat("a",32)
 key:=base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{'z'},32))
 for _,tc:=range cases {
  t.Run(tc.name,func(t *testing.T){
   server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
    if r.URL.Path!="/v1/roots/"+grant+"/inventory"{t.Errorf("wrong endpoint: %q",r.URL.Path)}
    _,_=io.WriteString(w,tc.body)
   }))
   defer server.Close()
   _,err:=(Client{Server:server.URL,Token:key}).Inventory(context.Background(),grant)
   if (err==nil)!=tc.good{t.Fatalf("unexpected inventory parse result: %v",err)}
  })
 }
}

func TestMalformedGrantAndPathNeverReachPeer(t *testing.T) {
 var calls atomic.Int64
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){calls.Add(1);w.WriteHeader(http.StatusNoContent)}))
 defer server.Close()
 key:=base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{'x'},32))
 client:=Client{Server:server.URL,Token:key}
 store,err:=agenttransfer.Open(filepath.Join(t.TempDir(),"stage"),4<<20)
 if err!=nil{t.Fatal(err)}
 for _,grant:=range []string{"","../etc",strings.Repeat("A",32),strings.Repeat("a",31)}{
  if _,err:=client.Inventory(context.Background(),grant);err==nil{t.Fatalf("bad grant %q accepted",grant)}
 }
 for _,path:=range []string{"../escape","x/../../escape","x//y","x\\y","",".syncria-private"}{
  if _,err:=client.Stage(context.Background(),strings.Repeat("a",32),path,"device",store);err==nil{t.Fatalf("bad path %q accepted",path)}
 }
 if got:=calls.Load();got!=0{t.Fatalf("rejected input caused %d network requests",got)}
}

func TestStagePreservesReservedCharactersInSelectedFilename(t *testing.T){
 content:=[]byte("verified approved file")
 sum:=sha256.Sum256(content)
 digest:=hex.EncodeToString(sum[:])
 grant:=strings.Repeat("b",32)
 rel:="photos/what? #1.jpg"
 key:=base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{'x'},32))
 var gets atomic.Int64
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if r.URL.Path!="/v1/roots/"+grant+"/files/"+rel||r.URL.RawQuery!=""{http.Error(w,"filename became URL syntax",400);return}
  if r.Header.Get("Authorization")!="Bearer "+key{http.Error(w,"denied",401);return}
  w.Header().Set("X-Syncria-SHA256",digest)
  w.Header().Set("ETag", "\"" +digest+"\"")
  w.Header().Set("Content-Length",fmt.Sprint(len(content)))
  if r.Method=="GET"{
   gets.Add(1)
   if r.Header.Get("If-Match")!="\""+digest+"\""{t.Error("missing source revision condition")}
   _,_=w.Write(content)
  }
 }))
 defer server.Close()
 store,err:=agenttransfer.Open(filepath.Join(t.TempDir(),"stage"),4<<20)
 if err!=nil{t.Fatal(err)}
 done,err:=(Client{Server:server.URL,Token:key}).Stage(context.Background(),grant,rel,"device",store)
 if err!=nil{t.Fatal(err)}
 file,_,err:=store.OpenCompleted(done.ID);if err!=nil{t.Fatal(err)}
 result,err:=io.ReadAll(file);_=file.Close()
 if err!=nil||!bytes.Equal(result,content){t.Fatalf("wrong transfer bytes: %v",err)}
 if gets.Load()!=1{t.Fatalf("GET calls: %d",gets.Load())}
}
