package main

import (
 "html/template"
 "strings"
 "testing"
)

func TestReplacementUIParses(t *testing.T){
 if _,err:=template.New("application").Parse(newUI);err!=nil{t.Fatal(err)}
}
func TestReplacementUINavigation(t *testing.T){
 required:=[]string{
  "view-overview","view-sync","view-storage","view-accounts",
  "view-updates","view-settings","id=\"picker\"",
  "id=\"syncWizard\"","/api/browse/drive",
  "/api/browse/local","/api/shares/action",
  "/api/update/status","action=\"update_install\"",
 }
 for _,part:=range required{
  if !strings.Contains(newUI,part){t.Errorf("replacement UI missing %q",part)}
 }
}
func TestNoLegacyDashboard(t *testing.T){
 if strings.Contains(newUI,"Drive folder ID (leave"){t.Fatal("legacy raw folder ID input returned")}
 if strings.Contains(newUI,"Add synchronization mapping"){t.Fatal("legacy mapping section returned")}
}
