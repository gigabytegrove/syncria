package main

import (
 "context"
 "io"
 "net/http"
 "path/filepath"
 "strings"
 "testing"
 "time"
)

type driveTestTransport func(*http.Request)(*http.Response,error)
func (f driveTestTransport)RoundTrip(r *http.Request)(*http.Response,error){return f(r)}
func driveTestResponse(status int,body string)*http.Response{
 return &http.Response{StatusCode:status,Body:io.NopCloser(strings.NewReader(body)),Header:make(http.Header)}
}

func TestDriveChangeCursorAvoidsSecondFullListing(t *testing.T){
 idx,err:=loadSyncIndex(filepath.Join(t.TempDir(),"index.db"));if err!=nil{t.Fatal(err)}
 defer idx.Close()
 nLists,nDelta,nStarts:=0,0,0
 mode:=0
 app:=&App{index:idx,cfg:Config{Accounts:[]Account{{
  ID:"acc",AccessToken:"fake-access",Expires:time.Now().Add(time.Hour),
 }}}}
 app.client=&http.Client{Transport:driveTestTransport(func(req *http.Request)(*http.Response,error){
  if !strings.Contains(req.URL.Host,"googleapis.com"){t.Fatalf("Unexpected external request: %s",req.URL);return nil,nil}
  switch req.URL.Path{
  case "/drive/v3/changes/startPageToken":
   nStarts++;return driveTestResponse(200,`{"startPageToken":"100"}`),nil
  case "/drive/v3/files":
   nLists++;return driveTestResponse(200,`{"files":[{"id":"video-id","name":"huge.mov","mimeType":"video/quicktime","size":"32212254720","modifiedTime":"2026-10-14T18:00:00Z","md5Checksum":"abc","parents":["folder-root"]}]}`),nil
  case "/drive/v3/changes":
   nDelta++
   if mode==0{return driveTestResponse(200,`{"newStartPageToken":"101","changes":[]}`),nil}
   return driveTestResponse(200,`{"newStartPageToken":"102","changes":[{"fileId":"video-id","file":{"id":"video-id","name":"huge.mov","mimeType":"video/quicktime","size":"32212254720","modifiedTime":"2026-10-14T19:00:00Z","md5Checksum":"def","parents":["folder-root"]}}]}`),nil
  default:t.Fatalf("Unexpected API URL %s",req.URL.String());return nil,nil
  }
 })}
 ctx:=context.Background()
 first,err:=app.remoteTreeIndexed(ctx,"acc","folder-root",nil)
 if err!=nil||len(first)!=1||nLists!=1||nStarts!=1{t.Fatalf("First scan wrong: err=%v files=%+v full=%d start=%d",err,first,nLists,nStarts)}
 second,err:=app.remoteTreeIndexed(ctx,"acc","folder-root",nil)
 if err!=nil||len(second)!=1||nLists!=1||nDelta!=1{t.Fatalf("Second scan MUST use change token: err=%v files=%+v full=%d delta=%d",err,second,nLists,nDelta)}
 mode=1
 third,err:=app.remoteTreeIndexed(ctx,"acc","folder-root",nil)
 if err!=nil||third["huge.mov"].ModifiedTime!="2026-10-14T19:00:00Z"||nLists!=1||nDelta!=2{
  t.Fatalf("Changed file wasn't updated from delta: err=%v files=%+v full=%d delta=%d",err,third,nLists,nDelta)
 }
 snapshot,ok,err:=idx.getDriveTree("acc","folder-root")
 if err!=nil||!ok||snapshot.Cursor!="102"{t.Fatalf("Change cursor was not persisted atomically: %+v %v",snapshot,err)}
}

func TestDrivePreviewUsesCacheWithoutAdvancingStoredCursor(t *testing.T){
 idx,err:=loadSyncIndex(filepath.Join(t.TempDir(),"idx.db"));if err!=nil{t.Fatal(err)}
 defer idx.Close()
 snap:=driveTreeSnapshot{
  Account:"acc",Root:"subfolder",RootID:"subfolder",Cursor:"baseline-token",FullScanAt:time.Now(),
  Files:map[string]driveFile{"old.txt":{ID:"file1",Name:"old.txt",Parents:[]string{"subfolder"},MimeType:"text/plain",Size:"7"}},
 }
 if err=idx.saveDriveTree(snap);err!=nil{t.Fatal(err)}
 listings:=0
 app:=&App{index:idx,cfg:Config{Accounts:[]Account{{ID:"acc",AccessToken:"fake",Expires:time.Now().Add(time.Hour)}}}}
 app.client=&http.Client{Transport:driveTestTransport(func(req *http.Request)(*http.Response,error){
  if req.URL.Path=="/drive/v3/files"{listings++;return driveTestResponse(500,"no full scan expected"),nil}
  return driveTestResponse(200,`{"newStartPageToken":"next-token","changes":[{"fileId":"file2","file":{"id":"file2","name":"new.txt","mimeType":"text/plain","size":"3","parents":["subfolder"]}}]}`),nil
 })}
 files,err:=app.remoteTreePreviewIndexed(context.Background(),"acc","subfolder",nil)
 if err!=nil||len(files)!=2||listings!=0{t.Fatalf("Read-only preview unexpectedly listed full tree: err=%v count=%d listings=%d",err,len(files),listings)}
 saved,_,err:=idx.getDriveTree("acc","subfolder")
 if err!=nil||saved.Cursor!="baseline-token"||len(saved.Files)!=1{t.Fatal("Preview must never update the persisted change token or cache")}
}

func TestDriveDeltaFailureKeepsLastKnownCursor(t *testing.T){
 idx,err:=loadSyncIndex(filepath.Join(t.TempDir(),"index.db"));if err!=nil{t.Fatal(err)}
 defer idx.Close()
 s:=exampleDriveSnapshot();s.FullScanAt=time.Now()
 if err=idx.saveDriveTree(s);err!=nil{t.Fatal(err)}
 app:=&App{index:idx,cfg:Config{Accounts:[]Account{{ID:"account",AccessToken:"fake",Expires:time.Now().Add(time.Hour)}}}}
 app.client=&http.Client{Transport:driveTestTransport(func(req *http.Request)(*http.Response,error){
  return driveTestResponse(503,`{"error":"Drive temporarily unavailable"}`),nil
 })}
 if _,err=app.remoteTreeIndexed(context.Background(),"account","root",nil);err==nil{t.Fatal("Transient API failure must stop sync, not claim files deleted")}
 stored,_,err:=idx.getDriveTree("account","root")
 if err!=nil||stored.Cursor!="cursor-1"{t.Fatal("Partial failed delta must not advance token")}
}
