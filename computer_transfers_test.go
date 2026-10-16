package main

import (
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "os"
 "path/filepath"
 "strings"
 "testing"
 "time"
)

func TestComputerTransferJournalSurvivesRestartWithoutSecrets(t *testing.T){
 filename:=filepath.Join(t.TempDir(),"state.json")
 app:=&App{file:filename}
 if err:=app.restoreComputerTransfers();err!=nil{t.Fatal(err)}
 job:=computerTransfer{
  ID:newID(),SourceID:"src",DestinationID:"dst",Source:"Source",Destination:"Target",
  SourceRoot:"Documents",DestinationRoot:"Documents",File:"file.txt",Target:"folder/file.txt",Move:true,
  Status:"queued",Started:time.Now(),
 }
 app.computerTransferMu.Lock()
 app.computerTransferJobs[job.ID]=job
 if err:=app.saveComputerTransfersLocked();err!=nil{t.Fatal(err)}
 app.computerTransferMu.Unlock()
 data,err:=os.ReadFile(app.computerTransferJournal());if err!=nil{t.Fatal(err)}
 if strings.Contains(string(data),"pairing_key")||strings.Contains(string(data),`"key":`){
  t.Fatal("persistent journal must never serialize pairing secrets")
 }
 permissions,err:=os.Stat(app.computerTransferJournal());if err!=nil{t.Fatal(err)}
 if permissions.Mode().Perm()!=0600{t.Fatalf("journal must be 0600: %o",permissions.Mode().Perm())}
 fresh:=&App{file:filename}
 if err:=fresh.restoreComputerTransfers();err!=nil{t.Fatal(err)}
 records:=fresh.listComputerTransfers()
 if len(records)!=1||records[0].ID!=job.ID||records[0].Status!="interrupted"{
  t.Fatalf("in-flight work must be marked interrupted after hub restart: %+v",records)
 }
 if !strings.Contains(records[0].Error,"inspect both devices"){t.Fatal("missing cautious recovery instructions")}
 fresh.setComputerTransferStatus(job.ID,computerWorkerEvent{Status:"completed"})
 if fresh.listComputerTransfers()[0].Status!="interrupted"{t.Fatal("late events overwrote terminated state")}
}

func TestComputerTransferActionRequiresAuthentication(t *testing.T){
 app:=&App{file:filepath.Join(t.TempDir(),"state.json")}
 _=app.restoreComputerTransfers()
 for _,path:=range []string{"/api/computers/transfers","/api/computers/transfers/action"}{
  request:=httptest.NewRequest("POST","http://localhost"+path,strings.NewReader(`{"id":"x","action":"cancel"}`))
  result:=httptest.NewRecorder()
  if strings.HasSuffix(path,"action"){app.computerTransferActionAPI(result,request)}else{app.computerTransfersAPI(result,request)}
  if result.Code!=http.StatusUnauthorized{t.Fatalf("%s returned %d",path,result.Code)}
 }
}
func TestTransferJournalNoSilentOverwritesAndInvalidRoot(t *testing.T){
 // A request cannot initiate file operations without two explicit roots.
 app:=&App{file:filepath.Join(t.TempDir(),"state.json"),cfg:Config{Secret:b64rand(),Computers:[]ComputerRecord{
  {ID:"a",Label:"Alpha"},{ID:"b",Label:"Beta"},
 }}}
 if err:=app.restoreComputerTransfers();err!=nil{t.Fatal(err)}
 recorder:=httptest.NewRecorder()
 app.setSession(recorder,httptest.NewRequest("GET","http://syncria.test/",nil),true)
 cookies:=recorder.Result().Cookies()
 request:=httptest.NewRequest("POST","http://syncria.test/api/computers/transfers",strings.NewReader(`{"source_id":"a","destination_id":"b","source_path":"a.txt","destination_path":"b.txt","move":true}`))
 request.AddCookie(cookies[0])
 w:=httptest.NewRecorder()
 app.computerTransfersAPI(w,request)
 if w.Code!=http.StatusBadRequest{t.Fatalf("missing shared folders accepted: %d",w.Code)}
 if len(app.listComputerTransfers())!=0{t.Fatal("invalid transfer entered queue")}
}

func TestComputerTransferCancellationChangesJournal(t *testing.T){
 app:=&App{file:filepath.Join(t.TempDir(),"state.json"),cfg:Config{Secret:b64rand()}}
 if err:=app.restoreComputerTransfers();err!=nil{t.Fatal(err)}
 job:=computerTransfer{ID:newID(),Status:"queued",Started:time.Now()}
 app.computerTransferMu.Lock()
 app.computerTransferJobs[job.ID]=job
 if err:=app.saveComputerTransfersLocked();err!=nil{t.Fatal(err)}
 app.computerTransferMu.Unlock()
 auth:=httptest.NewRecorder()
 app.setSession(auth,httptest.NewRequest("GET","http://syncria.test/",nil),true)
 request:=httptest.NewRequest("POST","http://syncria.test/api/computers/transfers/action",strings.NewReader(`{"id":"`+job.ID+`","action":"cancel"}`))
 request.AddCookie(auth.Result().Cookies()[0])
 w:=httptest.NewRecorder();app.computerTransferActionAPI(w,request)
 if w.Code!=http.StatusOK{t.Fatalf("cancel returned %d: %s",w.Code,w.Body.String())}
 fresh:=&App{file:app.file}
 if err:=fresh.restoreComputerTransfers();err!=nil{t.Fatal(err)}
 records:=fresh.listComputerTransfers()
 if len(records)!=1||records[0].Status!="canceled"{t.Fatalf("cancellation was not journaled: %+v",records)}
}
func TestComputerSeparatedWorkspaceControls(t *testing.T){
 start:=strings.Index(newUI,`id="view-computers"`)
 transfer:=strings.Index(newUI,`id="view-computer-transfers"`)
 end:=strings.Index(newUI,`id="view-settings"`)
 if start<0||transfer<0||end<0||!(start<transfer&&transfer<end){t.Fatal("computer screens are not independent pages")}
 segment:=newUI[start:transfer]
 if strings.Contains(segment,`id="computerSourceFiles"`){t.Fatal("file transfer explorer must not be stacked into My Computers")}
 for _,required:=range []string{`id="computerOnboard"`,`id="computerPSCommand"`,`id="computerCMDCommand"`,`id="computerLinuxCommand"`,`id="computerTransferList"`,
  `data-view="computer-transfers"`,`computerTransferFilter`,`changeComputerTransfer(t,action)`} {
  if !strings.Contains(newUI,required){t.Errorf("missing new direct-transfer feature %s",required)}
 }
}
func TestComputerTransferJsonRoundTrip(t *testing.T){
 job:=computerTransfer{ID:"abc",SourceID:"one",DestinationID:"two",SourceRoot:"Files",DestinationRoot:"Files",File:"a.txt",Target:"b.txt",Total:400,Bytes:200}
 data,err:=json.Marshal(job);if err!=nil{t.Fatal(err)}
 var restored computerTransfer
 if err:=json.Unmarshal(data,&restored);err!=nil{t.Fatal(err)}
 if restored.Total!=400||restored.Bytes!=200||restored.SourceID!="one"{t.Fatal("transfer progress or source identity was not preserved")}
}
