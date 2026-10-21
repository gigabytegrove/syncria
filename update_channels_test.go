package main

import (
 "context"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
)
func TestChannelReleaseSelection(t *testing.T){
 published:=[]releaseInfo{
  {Tag:"v0.8.5"}, {Tag:"beta-v0.8.6-beta.1",Prerelease:true},
  {Tag:"alpha-v0.8.6-alpha.6",Prerelease:true},
  {Tag:"alpha-v0.8.6-alpha.7",Prerelease:true,Draft:true},
  {Tag:"v1.1.0",Draft:true},
  {Tag:"alpha-v1.0.0-beta.1",Prerelease:true},
 }
 for channel,want:=range map[string]string{"stable":"v0.8.5","beta":"beta-v0.8.6-beta.1","alpha":"alpha-v0.8.6-alpha.6"}{
  release,err:=choosePublishedRelease(channel,published)
  if err!=nil||release.Tag!=want{t.Fatalf("%s selected %q: %v",channel,release.Tag,err)}
 }
}
func TestChannelVersionComparison(t *testing.T){
 for _,tc:=range []struct{a,b string;newer bool}{
  {"alpha-v0.8.6-alpha.6","v0.8.5",true},
  {"beta-v0.8.6-beta.1","alpha-v0.8.6-alpha.6",true},
  {"v0.8.6","beta-v0.8.6-beta.1",true},
  {"alpha-v0.8.6-alpha.6","v0.8.6",false},
  {"alpha-v1.0.0-alpha.1","v1.0.0-alpha.1-local",false},
 }{
  if got:=newerVersion(tc.a,tc.b);got!=tc.newer{t.Errorf("%s vs %s: %v",tc.a,tc.b,got)}
 }
}
func TestInvalidReleaseChannelsRejected(t *testing.T){
 for _,ch:=range []string{"latest","dev","other","Stable"}{
  if validUpdateChannel(ch){t.Fatal(ch)}
  if _,err:=latestChannelRelease(context.Background(),ch);err==nil{t.Fatalf("%s bypassed validation",ch)}
 }
}
func TestChannelUIAndAuth(t *testing.T){
 for _,fragment:=range []string{`id="updateChannel"`,`value="stable"`,`value="beta"`,`value="alpha"`,`expected_tag:operation==='install'?selectedUpdateTag:''`,`channel:getUpdateChannel()`}{
  if !strings.Contains(newUI,fragment){t.Errorf("missing %s",fragment)}
 }
 a:=&App{}
 w:=httptest.NewRecorder()
 a.updateStatus(w,httptest.NewRequest(http.MethodGet,"/api/update/status?channel=alpha",nil))
 if w.Code!=http.StatusUnauthorized{t.Fatal(w.Code)}
}
