package main

import (
 "bytes"
 "encoding/base64"
 "fmt"
 "os"
 "path/filepath"
 "strings"
 "testing"
 "time"
)

// This test only creates a synthetic, offline HTML preview when explicitly
// requested by CI. It never loads real Syncria accounts or credentials.
func TestProduceUIBrowserPreview(t *testing.T){
 if os.Getenv("SYNCRIA_VISUAL_PREVIEW")!="1"{t.Skip("preview only")}
 example:=struct{
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
  Version:"v0.8.2",Configured:true,Logged:true,
  Accounts:[]Account{{ID:"google-demo",Email:"home@example.com"}},
  DropboxAccounts:[]DropboxAccount{{ID:"dropbox-demo",Email:"family@example.com"}},
  Shares:[]StorageShare{
   {ID:"nas-one",Label:"Family NAS",Protocol:"smb",Server:"192.168.1.20",Export:"Photos"},
   {ID:"nas-two",Label:"Backup NAS",Protocol:"smb",Server:"192.168.1.30",Export:"Family"},
  },
  Jobs:[]Job{
   {ID:"photos",Provider:"bridge",CloudProvider:"google_drive",Name:"Family Photos",Enabled:true,Interval:5,ScheduleMode:"interval",SuccessfulRuns:142,TotalUploads:242,TotalDownloads:36,LastSummary:"Everything is up to date",LastRun:time.Now().Add(-12*time.Minute),NASDestinations:[]NASDestination{{ShareID:"nas-one",Folder:"Pictures"}},PendingDeletions:[]PendingDeletion{
    {Pair:"photos-nas-one",Path:"Family Photos/2025/Vacation/Photo-01.jpg",Expires:time.Now().Add(8*time.Minute)},
    {Pair:"photos-nas-one",Path:"Family Photos/2025/Vacation/Photo-02.jpg",Expires:time.Now().Add(7*time.Minute)},
    {Pair:"photos-nas-one",Path:"Family Photos/2025/Vacation/Video-03.mov",Expires:time.Now().Add(6*time.Minute)},
   }},
   {ID:"backup",Provider:"nas_bridge",Name:"NAS Backup",Enabled:false,Interval:30,Direction:"one_way",SuccessfulRuns:29,LastSummary:"Paused by user",NASDestinations:[]NASDestination{{ShareID:"nas-two",Folder:"Documents"}}},
  },
 }
 example.Metrics=summarizeMetrics(example.Jobs)
 var html bytes.Buffer
 if err:=ui.Execute(&html,example);err!=nil{t.Fatal(err)}
 logo,err:=os.ReadFile("assets/syncria-logo.svg");if err!=nil{t.Fatal(err)}
 embedded:="data:image/svg+xml;base64,"+base64.StdEncoding.EncodeToString(logo)
 rendered:=strings.ReplaceAll(html.String(),"/assets/syncria-logo.svg",embedded)
 folder:=os.Getenv("RUNNER_TEMP");if folder==""{folder=os.TempDir()}
 output:=filepath.Join(folder,"syncria-preview.html")
 if err:=os.WriteFile(output,[]byte(rendered),0600);err!=nil{t.Fatal(err)}
 fmt.Println("OFFLINE_UI_PREVIEW="+output)
}
