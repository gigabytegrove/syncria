package main
import (
 "strings"
 "context"
 "testing"
)
func TestNoAutomaticPageReload(t *testing.T) {
 if strings.Contains(newUI,"location.replace(")||strings.Contains(newUI,"setInterval(checkRunningVersion") {
  t.Fatal("Background update watcher must not reload the entire page")
 }
}
func TestShareDiscoveryControl(t *testing.T) {
 for _,part:=range []string{"id=\"shareServer\"","id=\"discoveredShares\"","/api/shares/discover"} {
  if !strings.Contains(newUI,part) {t.Fatalf("missing discovery UI %q",part)}
 }
}
func TestSMBShareDiscoveryRejectsPublicIP(t *testing.T) {
 for _,host:=range []string{"8.8.8.8","example.com","", "not-an-ip"} {
  if _,err:=discoverSMBShares(context.Background(),host);err==nil {t.Fatalf("accepted untrusted discovery target %q",host)}
 }
}

func TestNFSDiscoveryRejectsPublicAndMalformedTargets(t *testing.T){
 for _,host:=range []string{"8.8.8.8","example.com","", "10.0.0.999"}{
  if _,err:=discoverNFSExports(context.Background(),host);err==nil{t.Fatalf("accepted untrusted target %q",host)}
 }
}
func TestShareDiscoverySupportsBothProtocols(t *testing.T){
 for _,part:=range []string{"Discover shares / exports","Discover with credentials","protocol:protocol","method:'POST'"}{
  if !strings.Contains(newUI,part){t.Errorf("UI missing %s",part)}
 }
}
