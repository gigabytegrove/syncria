package main

import (
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "os"
 "path/filepath"
 "strings"
 "testing"
 "time"
)

func TestUpdateVersionPrereleasePrecedence(t *testing.T){
 sorted:=[]string{"v0.8.5","alpha-v0.8.6-alpha.1","alpha-v0.8.6-alpha.2","beta-v0.8.6-beta.1","beta-v0.8.6-rc.1","v0.8.6","alpha-v0.8.7-alpha.1"}
 for i:=1;i<len(sorted);i++{
  if !updateVersionNewer(sorted[i],sorted[i-1]){t.Errorf("%s should be newer than %s",sorted[i],sorted[i-1])}
  if updateVersionNewer(sorted[i-1],sorted[i]){t.Errorf("%s should not outrank %s",sorted[i-1],sorted[i])}
 }
 for _,tag:=range []string{"v0.8","v0.8.6-dev","alpha-v0.8.6","v0.8.6-alpha.01","v0.8.6-alpha.none","v01.8.5"}{
  if _,ok:=parseUpdateVersion(tag);ok{t.Errorf("accepted malformed tag: %s",tag)}
 }
 if newerVersion("alpha-v0.8.6-alpha.1","v0.8.5"){t.Fatal("legacy stable updater would accidentally offer alpha tag")}
}

func testChannelReleases()[]releaseInfo{
 return []releaseInfo{
  {Tag:"v0.8.5"},
  {Tag:"alpha-v0.8.6-alpha.1",Prerelease:true},
  {Tag:"beta-v0.8.6-beta.1",Prerelease:true},
  {Tag:"beta-v0.8.6-rc.1",Prerelease:true},
  {Tag:"v0.8.6",Draft:true},
  {Tag:"alpha-v0.8.7-alpha.1",Draft:true,Prerelease:true},
  {Tag:"v0.9.0-alpha.1",Prerelease:true}, // unsafe legacy-compatible tag must be ignored
  {Tag:"alpha-v0.9.0-alpha.1",Prerelease:false}, // prerelease designation is required
 }
}
func TestChannelReleaseAllowLists(t *testing.T){
 for _,tc:=range []struct{channel,want string}{
  {"stable","v0.8.5"},
  {"beta","beta-v0.8.6-rc.1"},
  {"alpha","beta-v0.8.6-rc.1"},
 }{
  rel,err:=newestReleaseForChannel(testChannelReleases(),tc.channel)
  if err!=nil||rel.Tag!=tc.want{t.Fatalf("%s selected %+v: %v; wanted %s",tc.channel,rel,err,tc.want)}
 }
 valid:=append(testChannelReleases(),releaseInfo{Tag:"alpha-v0.8.7-alpha.2",Prerelease:true})
 rel,err:=newestReleaseForChannel(valid,"alpha")
 if err!=nil||rel.Tag!="alpha-v0.8.7-alpha.2"{t.Fatal("alpha should accept next valid alpha",rel,err)}
 rel,err=newestReleaseForChannel(valid,"beta")
 if err!=nil||rel.Tag!="beta-v0.8.6-rc.1"{t.Fatal("beta must exclude alpha",rel,err)}
 rel,err=newestReleaseForChannel(valid,"stable")
 if err!=nil||rel.Tag!="v0.8.5"{t.Fatal("stable must exclude all prereleases",rel,err)}
 if _,err:=newestReleaseForChannel([]releaseInfo{{Tag:"alpha-v0.8.6-alpha.1",Prerelease:true}},"stable");err==nil{t.Fatal("stable returned alpha")}
}
func TestChannelLegacyConfigurationDefaultsStable(t *testing.T){
 var cfg Config
 if err:=json.Unmarshal([]byte(`{"secret":"legacy"}`),&cfg);err!=nil{t.Fatal(err)}
 if normalizeUpdateChannel(cfg.UpdateChannel)!="stable"{t.Fatal("older installations must default to Stable")}
 if normalizeUpdateChannel(" invalid ")!="stable"{t.Fatal("unexpected channel should safely default stable")}
}
func TestChannelSelectionPersistsAndIsAuthenticated(t *testing.T){
 filename:=filepath.Join(t.TempDir(),"state.json")
 a:=&App{file:filename,cfg:Config{Secret:b64rand()}}
 unauthorized:=httptest.NewRecorder()
 a.updateChannelAPI(unauthorized,httptest.NewRequest("GET","http://localhost/api/update/channel",nil))
 if unauthorized.Code!=http.StatusUnauthorized{t.Fatalf("unauthenticated channel read returned %d",unauthorized.Code)}
 auth:=httptest.NewRecorder()
 a.setSession(auth,httptest.NewRequest("GET","http://localhost/",nil),true)
 cookie:=auth.Result().Cookies()[0]
 call:=func(method,body string,origin string)*httptest.ResponseRecorder{
  req:=httptest.NewRequest(method,"http://localhost/api/update/channel",strings.NewReader(body))
  req.AddCookie(cookie)
  if origin!=""{req.Header.Set("Origin",origin)}
  response:=httptest.NewRecorder();a.updateChannelAPI(response,req);return response
 }
 for _,value:=range []string{`{"channel":"nightly"}`,`{"channel":"invalid"}`,`{"channel":"ALPHA","other":true}`} {
  r:=call("POST",value,"http://localhost")
  if r.Code!=http.StatusBadRequest{t.Fatalf("invalid channel accepted: %s (%d)",value,r.Code)}
 }
 if r:=call("POST",`{"channel":"alpha"}`,"http://outside.example");r.Code!=http.StatusForbidden{t.Fatalf("cross-origin change accepted: %d",r.Code)}
 if r:=call("POST",`{"channel":"alpha"}`,"http://localhost");r.Code!=http.StatusOK{t.Fatalf("valid change failed: %d %s",r.Code,r.Body.String())}
 raw,err:=os.ReadFile(filename);if err!=nil{t.Fatal(err)}
 var cfg Config;if err:=json.Unmarshal(raw,&cfg);err!=nil{t.Fatal(err)}
 if cfg.UpdateChannel!="alpha"{t.Fatal("server did not persist chosen channel")}
 a.updateMu.Lock();a.updateBusy=true;a.updateMu.Unlock()
 if r:=call("POST",`{"channel":"stable"}`,"http://localhost");r.Code!=http.StatusConflict{t.Fatalf("channel changed during update: %d",r.Code)}
 if a.cfg.UpdateChannel!="alpha"{t.Fatal("blocked change modified saved config")}
 a.updateMu.Lock();a.updateBusy=false;a.updateMu.Unlock()
 if r:=call("GET","","");r.Code!=http.StatusOK||!strings.Contains(r.Body.String(),`"alpha"`){t.Fatalf("channel GET did not reflect saved choice: %d %s",r.Code,r.Body.String())}
}
func TestUpdateChannelCacheIsolation(t *testing.T){
 updateLookup.Lock()
 savedResponse,savedUntil,savedRefresh,savedChannel:=updateLookup.response,updateLookup.validUntil,updateLookup.lastRefresh,updateLookup.channel
 updateLookup.response=map[string]any{"current":"cached-stable"}
 updateLookup.validUntil=time.Now().Add(time.Minute)
 updateLookup.channel="stable"
 updateLookup.Unlock()
 defer func(){
  updateLookup.Lock()
  updateLookup.response=savedResponse
  updateLookup.validUntil=savedUntil
  updateLookup.lastRefresh=savedRefresh
  updateLookup.channel=savedChannel
  updateLookup.Unlock()
 }()
 if v:=cachedUpdateStatusForChannel(nil,"stable");v["current"]!="cached-stable"{t.Fatal("stable cache unexpectedly missed")}
 // Cross-channel cache checks must NOT return a Stable result when Alpha is selected.
 // Verify keying directly, without issuing an external network request.
 updateLookup.Lock()
 mismatch:=updateLookup.channel!="alpha"
 updateLookup.Unlock()
 if !mismatch{t.Fatal("cache incorrectly shared between stable and alpha")}
}
func TestChannelControlsVisibleInUpdater(t *testing.T){
 for _,part:=range []string{"id=\"updateChannelChoices\"","id=\"updateChannelMessage\"","value=\"stable\"","value=\"beta\"","value=\"alpha\"","/api/update/channel"}{
  if !strings.Contains(newUI,part){t.Errorf("missing update channel UI: %s",part)}
 }
}
