package main
import (
 "strings"
 "context"
 "net"
 "time"
 "testing"
)
func TestNoAutomaticPageReload(t *testing.T) {
 if strings.Contains(newUI,"setInterval(checkRunningVersion") {
  t.Fatal("Background polling must not reload the page")
 }
}
func TestDirectNASConnectionControls(t *testing.T) {
 for _,part:=range []string{"id=\"shareServer\"","/api/nas/action","/api/browse/nas","/api/browse/nas/create","id=\"nasTargets\"","id=\"nasSourceId\""} {
  if !strings.Contains(newUI,part) {t.Fatalf("missing direct NAS UI %q",part)}
 }
 if strings.Contains(newUI,"Connect a new SMB / NFS mount") {t.Fatal("kernel mounting returned to the NAS connection interface")}
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
func TestDirectNASProviderAvailabilityIsAccurate(t *testing.T){
 for _,part:=range []string{"Windows-style shared folders (SMB) — recommended","Other network sharing (NFS) — not available yet","<option value=\"dropbox\">Dropbox</option>"} {
  if !strings.Contains(newUI,part){t.Errorf("missing provider status %q",part)}
 }
}
func TestDiscoveryTCPProbe(t *testing.T){
 listener,err:=net.Listen("tcp","127.0.0.1:0");if err!=nil{t.Fatal(err)}
 addr:=listener.Addr().(*net.TCPAddr)
 if err:=checkDiscoveryPort(context.Background(),"127.0.0.1",addr.Port,time.Second);err!=nil{t.Fatal(err)}
 listener.Close()
 if err:=checkDiscoveryPort(context.Background(),"127.0.0.1",addr.Port,200*time.Millisecond);err==nil{
  t.Fatal("closed discovery port should fail quickly")
 }
}
func TestNASBrowserIgnoresStaleRequests(t *testing.T){
 if !strings.Contains(newUI,"pickSeq++")||!strings.Contains(newUI,"if(sequence!==pickSeq)") {
  t.Fatal("Remote folder browser must ignore stale responses after selecting another folder")
 }
}
