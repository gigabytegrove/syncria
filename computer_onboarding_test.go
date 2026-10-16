package main

import (
 "encoding/json"
 "fmt"
 "net/http"
 "net/http/httptest"
 "os"
 "path/filepath"
 "strings"
 "testing"
 "time"
)

func inviteFixture(t *testing.T)(*App,*http.Cookie){
 t.Helper()
 a:=&App{file:filepath.Join(t.TempDir(),"state.json"),cfg:Config{Secret:b64rand()}}
 auth:=httptest.NewRecorder()
 a.setSession(auth,httptest.NewRequest("GET","https://syncria.example/",nil),true)
 return a,auth.Result().Cookies()[0]
}
func TestComputerSelfHostedInviteAndBootstrap(t *testing.T){
 app,cookie:=inviteFixture(t)
 req:=httptest.NewRequest("POST","https://syncria.example/api/computers/invite",strings.NewReader(`{"label":"Office desktop"}`))
 unauth:=httptest.NewRecorder()
 app.computerInviteAPI(unauth,req)
 if unauth.Code!=401{t.Fatalf("unauthenticated invitation accepted: %d",unauth.Code)}
 req=httptest.NewRequest("POST","https://syncria.example/api/computers/invite",strings.NewReader(`{"label":"Office desktop"}`))
 req.AddCookie(cookie)
 w:=httptest.NewRecorder();app.computerInviteAPI(w,req)
 if w.Code!=201{t.Fatalf("invite creation failed: %d %s",w.Code,w.Body.String())}
 var invitation struct{Code string `json:"code"`;ExpiresAt time.Time `json:"expires_at"`}
 if err:=json.Unmarshal(w.Body.Bytes(),&invitation);err!=nil{t.Fatal(err)}
 if len(invitation.Code)!=64||time.Until(invitation.ExpiresAt)>21*time.Minute{t.Fatal("invalid invitation expiry or entropy")}
 for _,platform:=range []string{"windows","linux"}{
  script:="https://syncria.example/api/computers/bootstrap?code="+invitation.Code+"&platform="+platform+"&hub=https%3A%2F%2Fsyncria.example"
  res:=httptest.NewRecorder()
  app.computerBootstrapAPI(res,httptest.NewRequest("GET",script,nil))
  if res.Code!=200{t.Fatalf("%s bootstrap failed: %d %s",platform,res.Code,res.Body.String())}
  if !strings.Contains(res.Body.String(),invitation.Code)||!strings.Contains(res.Body.String(),"syncria.example")||
   !strings.Contains(res.Body.String(),"SHA256SUMS")||
   !strings.Contains(res.Body.String(),"pending-enrollment.json"){
   t.Fatalf("%s missing authenticated self-hosted install/pairing/verification",platform)
  }
  if platform=="windows"&&!strings.Contains(res.Body.String(),"CreateShortcut"){t.Fatal("Windows login startup missing")}
  if platform=="linux"&&!strings.Contains(res.Body.String(),"systemctl --user enable --now"){t.Fatal("Linux user unit missing")}
 }
 bad:=httptest.NewRecorder()
 app.computerBootstrapAPI(bad,httptest.NewRequest("GET",
 "https://syncria.example/api/computers/bootstrap?code="+invitation.Code+"&platform=linux&hub=https%3A%2F%2Fevil.example",nil))
 if bad.Code!=400{t.Fatalf("bootstrap host injection accepted: %d",bad.Code)}
 bad=httptest.NewRecorder()
 app.computerBootstrapAPI(bad,httptest.NewRequest("GET",
 "https://syncria.example/api/computers/bootstrap?code="+invitation.Code+"&platform=linux&hub=http%3A%2F%2Fsyncria.example",nil))
 if bad.Code!=400{t.Fatalf("unencrypted bootstrap hub accepted: %d",bad.Code)}
}
func TestNativeAgentAutomaticallyPairsAndInviteCannotBeReused(t *testing.T){
 src:=newAgentTestFixture(t,"source")
 app,cookie:=inviteFixture(t)
 req:=httptest.NewRequest("POST","https://syncria.example/api/computers/invite",strings.NewReader(`{"label":"Auto connected"}`));req.AddCookie(cookie)
 created:=httptest.NewRecorder();app.computerInviteAPI(created,req)
 if created.Code!=201{t.Fatalf("invite: %d",created.Code)}
 var invitation struct{Code string `json:"code"`}
 _=json.Unmarshal(created.Body.Bytes(),&invitation)
 // Real pinned HTTPS probe against the running native agent.
 var port int
 if _,err:=fmt.Sscanf(src.server.URL,"https://127.0.0.1:%d",&port);err!=nil{t.Fatal(err)}
 payload,_:=json.Marshal(map[string]any{
  "code":invitation.Code,"key":src.record.Key,"fingerprint":src.record.Fingerprint,
  "port":port,"addresses":[]string{"127.0.0.1"},
 })
 request:=httptest.NewRequest("POST","https://syncria.example/api/computers/enroll",strings.NewReader(string(payload)))
 result:=httptest.NewRecorder();app.computerEnrollAPI(result,request)
 if result.Code!=201{t.Fatalf("auto enrollment failed: %d %s",result.Code,result.Body.String())}
 if len(app.cfg.Computers)!=1||app.cfg.Computers[0].Label!="Auto connected"||app.cfg.Computers[0].Fingerprint!=src.record.Fingerprint{
  t.Fatalf("agent not correctly persisted: %+v",app.cfg.Computers)
 }
 again:=httptest.NewRecorder();app.computerEnrollAPI(again,httptest.NewRequest("POST","https://syncria.example/api/computers/enroll",strings.NewReader(string(payload))))
 if again.Code!=409{t.Fatalf("replayed one-time enrollment accepted: %d",again.Code)}
 script:=httptest.NewRecorder()
 app.computerBootstrapAPI(script,httptest.NewRequest("GET",
 "https://syncria.example/api/computers/bootstrap?platform=linux&code="+invitation.Code+"&hub=https%3A%2F%2Fsyncria.example",nil))
 if script.Code!=403{t.Fatalf("used invite could still download bootstrap script: %d",script.Code)}
 saved,err:=os.ReadFile(app.file);if err!=nil{t.Fatal(err)}
 if !strings.Contains(string(saved),src.record.Fingerprint){t.Fatal("verified fingerprint not stored")}
}
func TestAgentDistributionSafetyAllowlist(t *testing.T){
 for _,name:=range []string{"../../state.json","syncria-linux-386","syncria-windows.exe","agent.key",".env"}{
  if validComputerAgentAsset(name){t.Fatalf("unsafe asset approved: %s",name)}
 }
 for _,name:=range []string{"SHA256SUMS","syncria-linux-amd64","syncria-linux-arm64","syncria-linux-arm","syncria-windows-amd64.exe","syncria-windows-arm64.exe"}{
  if !validComputerAgentAsset(name){t.Fatalf("supported asset rejected: %s",name)}
 }
 if safeComputerAddress("8.8.8.8")||safeComputerAddress("0.0.0.0")||safeComputerAddress("example.com"){t.Fatal("external or invalid computer endpoint accepted")}
 if !safeComputerAddress("192.168.0.55")||!safeComputerAddress("127.0.0.1"){t.Fatal("LAN address unexpectedly rejected")}
}
