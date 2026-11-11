package main
import ("strings";"testing")
func TestUnsupportedGoogleNativeItemsCannotBeTransferredOrDeleted(t *testing.T){
 j:=&Job{PairStates:map[string]map[string]Item{"pair":{"native-map":{ID:"old",MD5:"old",LocalMD5:"old"}}}}
 primary:=map[string]bridgeEntry{"native-map":{ID:"native",Unsupported:true}}
 secondary:=map[string]bridgeEntry{"native-map":{Fingerprint:"existing",Size:35}}
 plan:=syncPreviewResult{Counts:map[string]int{},Rows:[]syncPreviewRow{}}
 planSyncPair(&plan,j,"pair",primary,secondary,false,conflictPrimary)
 if len(plan.Rows)!=1||plan.Rows[0].Action!="skip"{t.Fatalf("native item must be skipped without overwrites: %+v",plan.Rows)}
 if !strings.Contains(plan.Rows[0].Reason,"preserved"){t.Fatal("Missing safety explanation")}
}
func TestUpdateVerifiedReload(t *testing.T){
 for _,part:=range []string{"function finishVerifiedUpdate(version)","syncria-update-expected","window.location.replace(","p.current===runtimeVersionOfRelease(expectedVersion)"}{
  if !strings.Contains(newUI,part){t.Errorf("Missing automatic verified version handoff %q",part)}
 }
 if strings.Contains(newUI,`id="updateReloadButton"`){t.Fatal("Obsolete manual View updated button still displayed")}
}
