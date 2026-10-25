package main

import "testing"

// The entire published alpha asset set must support the updater platform
// naming scheme previously shipped with v0.8.6.
func TestAlphaOneIsSelectableWithoutChangingOtherChannels(t *testing.T) {
 rels:=[]releaseInfo{
  {Tag:"v0.8.5"},
  {Tag:"beta-v0.8.6-beta.1",Prerelease:true},
  {Tag:"alpha-v0.8.6-alpha.6",Prerelease:true},
  {Tag:"alpha-v1.0.0-alpha.1",Prerelease:true},
 }
 for _,tc:=range []struct{channel,want string}{
  {"stable","v0.8.5"},
  {"beta","beta-v0.8.6-beta.1"},
  {"alpha","alpha-v1.0.0-alpha.1"},
 }{
  found,err:=newestReleaseForChannel(rels,tc.channel)
  if err!=nil||found.Tag!=tc.want{t.Fatalf("%s: got %+v %v",tc.channel,found,err)}
 }
 if got:=releaseTagAppVersion("alpha-v1.0.0-alpha.1");got!="v1.0.0-alpha.1"{t.Fatalf("wrong restart version: %q",got)}
 if !updateVersionNewer("alpha-v1.0.0-alpha.1","v0.8.6-alpha.6"){t.Fatal("alpha 1.0 must upgrade alpha 0.8.6")}
 if updateVersionNewer("alpha-v1.0.0-alpha.1","v1.0.0-alpha.1-local"){t.Fatal("same alpha local build is not a newer release")}
}
