package main

import (
 "strings"
 "testing"
)

func TestUpdaterStatusFirstLayout(t *testing.T) {
 required:=[]string{
  `id="updateStatusCard"`,`id="updateStatusHeadline"`,`id="updateAvailableVersion"`,
  `id="updateReleaseNotes"`,`id="installUpdateButton"`,`id="updateNavIndicator"`,
  `id="updateFlow"`,`id="updateHistoryPanel"`,`id="manualUpdateForm"`,
  `id="rollbackUpdateButton"`,`data-update-state="checking"`,
  "Update available","Syncria is up to date","Unable to check for updates",
 }
 for _,item:=range required {if !strings.Contains(newUI,item){t.Errorf("Missing updater capability: %s",item)}}
 if strings.Contains(newUI,`class="update-aside"`) {t.Error("Updater must not return to a permanent right sidebar")}
 if strings.Contains(newUI,`id="updateReadyPanel"`){t.Error("Obsolete explanatory panel must not be shown")}
 status:=strings.Index(newUI,`id="updateStatusCard"`)
 flow:=strings.Index(newUI,`id="updateFlow"`)
 log:=strings.Index(newUI,`id="updateHistoryPanel"`)
 advanced:=strings.Index(newUI,`class="update-advanced-section"`)
 if status<0||flow<status||log<flow||advanced<log {t.Fatal("Updater sections are not in status → progress → log → advanced order")}
 if !strings.Contains(v051Theme,".update-primary[data-update-state=available]") {t.Fatal("Update-available status lacks distinct styling")}
 if !strings.Contains(v051Theme,".update-console") {t.Fatal("Single-column updater styling missing")}
}
