package main

import (
 "context"
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "os"
 "path/filepath"
 "strings"
 "testing"
 "time"
)

type agentTestFixture struct{ server *httptest.Server; record ComputerRecord; folder string }
func newAgentTestFixture(t *testing.T,alias string)agentTestFixture{
 t.Helper()
 dir:=t.TempDir()
 shared:=filepath.Join(dir,"shared")
 if e:=os.Mkdir(shared,0700);e!=nil{t.Fatal(e)}
 agent:=&computerAgent{name:alias,secret:b64rand(),roots:map[string]string{"Files":shared},tickets:map[string]agentTicket{}}
 mux:=http.NewServeMux()
 mux.HandleFunc("/v1/status",agent.authorize(agent.serveStatus))
 mux.HandleFunc("/v1/list",agent.authorize(agent.serveList))
 mux.HandleFunc("/v1/grant",agent.authorize(agent.serveGrant))
 mux.HandleFunc("/v1/download",agent.serveDownload)
 mux.HandleFunc("/v1/pull",agent.authorize(agent.servePull))
 mux.HandleFunc("/v1/delete",agent.authorize(agent.serveDelete))
 server:=httptest.NewTLSServer(mux)
 t.Cleanup(server.Close)
 digest:=sha256.Sum256(server.TLS.Certificates[0].Certificate[0])
 return agentTestFixture{server:server,record:ComputerRecord{ID:alias,Label:alias,URL:server.URL,Key:agent.secret,Fingerprint:hex.EncodeToString(digest[:])},folder:shared}
}
func TestComputerAgentPathsRejectTraversal(t *testing.T){
 for _,p:=range []string{"../private","/etc/passwd","a/../b","a//b","a\\b","C:/Windows","foo/.syncria-transfer-1","one/./two"}{
  if _,e:=agentRelative(p);e==nil{t.Errorf("unsafe agent path accepted: %q",p)}
 }
 for _,p:=range []string{"","ok","subdir/file.txt",".env","filename with spaces.txt"}{
  if _,e:=agentRelative(p);e!=nil{t.Errorf("valid agent path rejected: %q: %v",p,e)}
 }
}
func TestAgentDirectPeerCopyAndMove(t *testing.T){
 src,dst:=newAgentTestFixture(t,"source"),newAgentTestFixture(t,"destination")
 file:=filepath.Join(src.folder,"sample.txt")
 if e:=os.WriteFile(file,[]byte("Syncria peer-to-peer transfer works"),0600);e!=nil{t.Fatal(e)}
 transfer:=computerTransfer{ID:newID(),File:"sample.txt",Started:time.Now(),Move:false}
 runComputerTransfer(src.record,dst.record,"Files","Files","sample.txt","nested/sample.txt",transfer)
 computerTransferLock.Lock();result:=computerTransfers[transfer.ID];computerTransferLock.Unlock()
 if result.Status!="completed"{t.Fatalf("copy failed: %+v",result)}
 contents,e:=os.ReadFile(filepath.Join(dst.folder,"nested","sample.txt"))
 if e!=nil||string(contents)!="Syncria peer-to-peer transfer works"{t.Fatalf("destination bytes mismatched: %q %v",contents,e)}
 if _,e=os.Stat(file);e!=nil{t.Fatalf("copy removed source: %v",e)}
 move:=computerTransfer{ID:newID(),File:"sample.txt",Started:time.Now(),Move:true}
 runComputerTransfer(src.record,dst.record,"Files","Files","sample.txt","sample-moved.txt",move)
 computerTransferLock.Lock();afterMove:=computerTransfers[move.ID];computerTransferLock.Unlock()
 if afterMove.Status!="completed"{t.Fatalf("move failed: %+v",afterMove)}
 if _,e=os.Stat(file);!os.IsNotExist(e){t.Fatalf("source was not removed after verified move: %v",e)}
 if bytes,e:=os.ReadFile(filepath.Join(dst.folder,"sample-moved.txt"));e!=nil||string(bytes)!="Syncria peer-to-peer transfer works"{t.Fatalf("moved bytes invalid: %q %v",bytes,e)}
}
func TestComputerTransferNeverOverwrites(t *testing.T){
 src,dst:=newAgentTestFixture(t,"source"),newAgentTestFixture(t,"destination")
 if e:=os.WriteFile(filepath.Join(src.folder,"a.txt"),[]byte("new"),0600);e!=nil{t.Fatal(e)}
 if e:=os.WriteFile(filepath.Join(dst.folder,"a.txt"),[]byte("existing"),0600);e!=nil{t.Fatal(e)}
 transfer:=computerTransfer{ID:newID(),File:"a.txt",Started:time.Now(),Move:true}
 runComputerTransfer(src.record,dst.record,"Files","Files","a.txt","a.txt",transfer)
 computerTransferLock.Lock();result:=computerTransfers[transfer.ID];computerTransferLock.Unlock()
 if result.Status!="failed"{t.Fatalf("overwrite unexpectedly permitted: %+v",result)}
 dest,e:=os.ReadFile(filepath.Join(dst.folder,"a.txt"));if e!=nil||string(dest)!="existing"{t.Fatal("destination modified",e)}
 source,e:=os.ReadFile(filepath.Join(src.folder,"a.txt"));if e!=nil||string(source)!="new"{t.Fatal("source removed",e)}
}
func TestComputerAgentPinnedTLSAndTicketAuth(t *testing.T){
 src:=newAgentTestFixture(t,"source")
 wrong:=strings.Repeat("0",64)
 if wrong==src.record.Fingerprint{wrong=strings.Repeat("1",64)}
 client,e:=pinnedComputerClient(wrong);if e!=nil{t.Fatal(e)}
 req,_:=http.NewRequest("GET",src.record.URL+"/v1/status",nil)
 req.Header.Set("Authorization","Bearer "+src.record.Key)
 if resp,e:=client.Do(req);e==nil{resp.Body.Close();t.Fatal("incorrect certificate fingerprint accepted")}
 if _,e:=validAgentURL("http://localhost:9765");e==nil{t.Fatal("plaintext URL accepted")}
 if _,e:=validAgentURL("https://localhost:9765/unsafe");e==nil{t.Fatal("URL path accepted")}
 client,e=pinnedComputerClient(src.record.Fingerprint);if e!=nil{t.Fatal(e)}
 response,e:=client.Get(src.record.URL+"/v1/download")
 if e!=nil{t.Fatal(e)}
 response.Body.Close()
 if response.StatusCode!=401{t.Fatalf("unauthorized download returned %d",response.StatusCode)}
}
func TestComputerAgentRejectsSymlink(t *testing.T){
 src:=newAgentTestFixture(t,"source")
 external:=filepath.Join(t.TempDir(),"secret.txt")
 if e:=os.WriteFile(external,[]byte("private"),0600);e!=nil{t.Fatal(e)}
 if e:=os.Symlink(external,filepath.Join(src.folder,"link"));e!=nil{t.Skipf("symlinks unavailable: %v",e)}
 client,e:=pinnedComputerClient(src.record.Fingerprint);if e!=nil{t.Fatal(e)}
 if e:=agentJSONPost(context.Background(),client,src.record.URL+"/v1/grant",src.record.Key,map[string]any{"root":"Files","path":"link"},nil);e==nil{t.Fatal("agent accepted symlink as a file")}
}
func TestComputerHubRegistrationAndNoSecretLeak(t *testing.T){
 src:=newAgentTestFixture(t,"source")
 state:=filepath.Join(t.TempDir(),"state.json")
 app:=&App{file:state,cfg:Config{Secret:b64rand()}}
 auth:=httptest.NewRecorder()
 app.setSession(auth,httptest.NewRequest("GET","http://syncria.test/",nil),true)
 cookie:=auth.Result().Cookies()[0]
 payload,_:=json.Marshal(map[string]any{"label":"Desk","url":src.record.URL,"key":src.record.Key,"fingerprint":src.record.Fingerprint})
 request:=httptest.NewRequest("POST","http://syncria.test/api/computers/register",strings.NewReader(string(payload)))
 request.AddCookie(cookie)
 response:=httptest.NewRecorder();app.computerRegisterAPI(response,request)
 if response.Code!=201{t.Fatalf("register returned %d: %s",response.Code,response.Body.String())}
 request=httptest.NewRequest("GET","http://syncria.test/api/computers",nil);request.AddCookie(cookie)
 response=httptest.NewRecorder();app.computersAPI(response,request)
 if response.Code!=200{t.Fatalf("list returned %d: %s",response.Code,response.Body.String())}
 if !strings.Contains(response.Body.String(),"\"online\":true"){t.Fatal("registered agent not online",response.Body.String())}
 if strings.Contains(response.Body.String(),src.record.Key){t.Fatal("list API disclosed pairing key")}
 persisted,e:=os.ReadFile(state);if e!=nil||!strings.Contains(string(persisted),src.record.Key){t.Fatal("agent credential was not persisted securely",e)}
}
