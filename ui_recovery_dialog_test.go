package main

import (
 "bytes"
 "strings"
 "testing"
 "time"
)

func TestRecoveryDialogIsCompactAndUndoStillPostsExactFile(t *testing.T){
 data:=struct{
  Version,Message,ClientID,BaseURL,Callback string
  DropboxClientID,DropboxCallback string
  Configured,Logged bool
  Accounts []Account
  DropboxAccounts []DropboxAccount
  Jobs []Job
  Shares []StorageShare
  Metrics DashboardTotals
  MountingEnabled bool
 }{
  Version:"v0.8.3",Configured:true,Logged:true,
  Jobs:[]Job{{
   ID:"recovery-case",Name:"NAS family photos",
   PendingDeletions:[]PendingDeletion{
    {Pair:"one",Path:"Family/Christmas/Kaleb.mov",Expires:time.Now().Add(8*time.Minute)},
    {Pair:"two",Path:"Family/2026/Photo.jpg",Expires:time.Now().Add(9*time.Minute)},
   },
  }},
 }
 var rendered bytes.Buffer
 if err:=ui.Execute(&rendered,data);err!=nil{t.Fatal(err)}
 html:=rendered.String()
 for _,wanted:=range []string{
  `class="recovery-summary"`,`class="recovery-count">2</span>`,
  `<dialog class="recovery-dialog" id="recovery-recovery-case"`,
  `aria-labelledby="recovery-title-recovery-case"`,
  `role="status">2 files eligible for recovery</span>`,
  `data-recovery-path="Family/Christmas/Kaleb.mov"`,
  `data-recovery-path="Family/2026/Photo.jpg"`,
  `name="operation" value="undo"`,
  `name="id" value="recovery-case"`,
  `name="pair" value="one"`,
  `name="path" value="Family/Christmas/Kaleb.mov"`,
  `action="/api/recovery/undo"`,
  `function filterRecoveryFiles(id)`,
  `dialog.showModal()`,`dialog.close()`,
 }{
  if !strings.Contains(html,wanted){t.Errorf("Recovery dialog has lost required markup or undo action: %s",wanted)}
 }
 if strings.Contains(html,`class="relationship-alert recovery-alert"`){t.Fatal("Old full-height in-card deleted-file list returned")}
}

func TestCloudWizardAndModalPolishIsScopedAndMobileFriendly(t *testing.T){
 for _,selector:=range []string{
  "#view-accounts .provider-grid",
  "#view-accounts .provider-add{width:100%",
  "#view-accounts .provider-account",
  "#view-accounts .connection-drawer>.step-indicator",
  "#view-accounts .connection-drawer>.step{padding:",
  "#view-accounts .connection-drawer>.step>.field",
  "#view-storage .connection-drawer",
  ".recovery-dialog::backdrop",
  ".recovery-dialog-items",
  ".recovery-dialog-item-copy strong",
  "@media(max-width:700px)",
 }{
  if !strings.Contains(syncriaDesignCSS,selector){t.Errorf("Missing scoped design rule: %s",selector)}
 }
 for _,required:=range []string{
  "openRecoveryDialog(selectedSyncID)",
  "filterRecoveryFiles('{{.ID}}')",
  "aria-labelledby=\"recovery-title-{{.ID}}\"",
  "id=\"syncDetailRecoveryBar\"",
  "parts[0]==='accounts'&&parts[1]==='connect'",
 }{
  if !strings.Contains(newUI,required){t.Errorf("Missing UX interaction: %s",required)}
 }
}
