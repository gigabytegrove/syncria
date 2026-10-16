package main

import (
 "strings"
 "testing"
)

func TestBetaPublishedReleaseUsesIsolatedTag(t *testing.T) {
 if appVersion!="v0.8.6-beta.1" {
  t.Fatalf("incorrect running beta version: %s",appVersion)
 }
 if got:=computerAgentVersionTag();got!="beta-v0.8.6-beta.1"{
  t.Fatalf("server must deliver agent executables from its matching beta release, got %s",got)
 }
 if got:=releaseTagAppVersion(computerAgentVersionTag());got!=appVersion {
  t.Fatalf("beta tag normalization failed: %s",got)
 }
 if !updateVersionNewer("beta-v0.8.6-beta.1","alpha-v0.8.6-alpha.6") {
  t.Fatal("beta.1 must be a newer update for users on alpha.6")
 }
 if updateVersionNewer("alpha-v0.8.6-alpha.6","beta-v0.8.6-beta.1") {
  t.Fatal("alpha.6 must not supersede beta.1")
 }
 releases:=[]releaseInfo{
  {Tag:"alpha-v0.8.6-alpha.6",Prerelease:true},
  {Tag:"beta-v0.8.6-beta.1",Prerelease:true},
  {Tag:"v0.8.5",Prerelease:false},
 }
 for _,channel:=range []string{"beta","alpha"}{
  choice,err:=newestReleaseForChannel(releases,channel)
  if err!=nil||choice.Tag!="beta-v0.8.6-beta.1"{
   t.Fatalf("%s selected %s instead of the promoted beta: %v",channel,choice.Tag,err)
  }
 }
 stable,err:=newestReleaseForChannel(releases,"stable")
 if err!=nil||stable.Tag!="v0.8.5"{
  t.Fatalf("stable must remain v0.8.5, got %s: %v",stable.Tag,err)
 }
}

func TestBetaReleaseHasCompleteOneCommandEnrollmentAndProtectedTransfers(t *testing.T){
 for _,fragment:=range []string{
  "id=\"computerPSCommand\"",
  "id=\"computerCMDCommand\"",
  "id=\"computerLinuxCommand\"",
  "id=\"view-computer-transfers\"",
  "location.protocol==='http:'",
  "document.execCommand('copy')",
 }{
  if !strings.Contains(newUI,fragment) {t.Errorf("beta is missing %s",fragment)}
 }
}
