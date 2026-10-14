package main

import (
 "html/template"
 "strings"
 "testing"
)

func TestNewProductUIIsComplete(t *testing.T){
 if _,err:=template.New("new-ui").Parse(newUI);err!=nil{t.Fatalf("Template: %v",err)}
 screens:=[]string{"view-overview","view-sync","view-storage","view-accounts","view-activity","view-updates","view-settings"}
 for _,s:=range screens{if !strings.Contains(newUI,`id="`+s+`"`){t.Errorf("Missing screen %s",s)}}
 for _,control:=range []string{
  `id="syncWizard"`,`id="relationshipType"`,`id="sourceProvider"`,
  `id="nasTargets"`,`id="nasSourceId"`,`id="syncInterval"`,`id="scheduleMode"`,
  `id="nasDirection"`,`id="conflictPolicy"`,`id="propagateDeletions"`,
  `id="recoveryEnabled"`,`name="cross_nas_propagation"`,
  `id="accountWizard"`,`id="directNASPanel"`,`id="picker"`,
  `id="manualUpdateForm"`,`id="updateProgressBar"`,`id="updateEventTimeline"`,
  `id="updateHistoryPanel"`,`id="updateReadyPanel"`,
  `/api/update/progress`,`/api/browse/dropbox`,`/api/conflicts/resolve`,`/api/recovery/undo`,
 }{
  if !strings.Contains(newUI,control){t.Errorf("Existing feature lost: %s",control)}
 }
}
func TestNewProductUIHasItsOwnVisualIdentity(t *testing.T){
 for _,part:=range []string{
  "command-hero","sidebar-navigation","sidebar-logo","sidebar-edition",
  "flow-card-main","flow-card-actions","connection-list","provider-grid",
  "activity-scoreboard","release-panel","sync-builder","pageBreadcrumb",
 }{
  if !strings.Contains(newUI,part){t.Errorf("Missing redesigned layout: %s",part)}
 }
 for _,part:=range []string{
  "--nav:#fff","--accent:#336b92",
  ".command-hero",".flow-card-main",".provider-account",
  ".activity-scoreboard",".sync-builder",".release-well",
  "@media(max-width:800px)","@media(max-width:530px)",
 }{
  if !strings.Contains(v051Theme,part){t.Errorf("Missing style system: %s",part)}
 }
 if strings.Contains(newUI,"I WAS BUILT BY AI") {t.Fatal("Placeholder branding leaked")}
}
func TestOnlyVisualCodeChangesNoSyncPolicies(t *testing.T){
 for _,policy:=range []string{`<option value="manual">`,`<option value="newer">`,`<option value="primary">`,`<option value="secondary">`,`<option value="keep_both">`}{
  if !strings.Contains(newUI,policy){t.Fatalf("NAS conflict policy missing: %s",policy)}
 }
 for _,operation:=range []string{`name="action" value="run_job"`,`name="action" value="toggle_job"`,`name="action" value="delete_job"`}{
  if !strings.Contains(newUI,operation){t.Fatalf("Sync action missing: %s",operation)}
 }
 if !strings.Contains(newUI,`name="propagate_deletions" value="1" checked`)||
    !strings.Contains(newUI,`name="recovery_enabled" value="1" checked`){
  t.Fatal("Established deletion and 10-minute recovery defaults changed")
 }
}
