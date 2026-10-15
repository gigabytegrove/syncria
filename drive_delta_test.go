package main

import (
 "path/filepath"
 "testing"
 "time"
)

func exampleDriveSnapshot()driveTreeSnapshot{
 now:=time.Date(2026,10,14,12,0,0,0,time.UTC)
 return driveTreeSnapshot{
  Account:"account",Root:"root",RootID:"real-root-id",Cursor:"cursor-1",FullScanAt:now,
  Files:map[string]driveFile{
   "Family":{ID:"folder-id",Name:"Family",MimeType:folderMime,Parents:[]string{"real-root-id"}},
   "Family/unchanged.mov":{ID:"video-1",Name:"unchanged.mov",MimeType:"video/quicktime",Parents:[]string{"folder-id"},Size:"32212254720",ModifiedTime:"2026-10-14T12:00:00Z"},
  },
 }
}
func TestDriveDeltaNoChangesSkipsFullTree(t *testing.T){
 original:=exampleDriveSnapshot()
 result,safe:=applyDriveDelta(original,nil)
 if !safe||len(result)!=len(original.Files){t.Fatalf("Expected unchanged metadata to be reused: %v %+v",safe,result)}
 if result["Family/unchanged.mov"].Size!="32212254720"{t.Fatal("30 GiB file metadata lost")}
}
func TestDriveDeltaSingleNewFileWithinExistingFolder(t *testing.T){
 original:=exampleDriveSnapshot()
 changes:=[]driveDeltaChange{{FileID:"new-id",File:driveFile{
  ID:"new-id",Name:"new-photo.jpg",MimeType:"image/jpeg",Parents:[]string{"folder-id"},Size:"1024",ModifiedTime:"2026-10-14T19:45:00Z",
 }}}
 result,safe:=applyDriveDelta(original,changes)
 if !safe||len(result)!=3||result["Family/new-photo.jpg"].ID!="new-id"{t.Fatalf("Expected incremental file addition: safe=%v result=%+v",safe,result)}
}
func TestDriveDeltaFileRenameAndMove(t *testing.T){
 original:=exampleDriveSnapshot()
 changes:=[]driveDeltaChange{{FileID:"video-1",File:driveFile{
  ID:"video-1",Name:"renamed.mov",MimeType:"video/quicktime",Parents:[]string{"real-root-id"},Size:"32212254720",
 }}}
 result,safe:=applyDriveDelta(original,changes)
 if !safe||result["renamed.mov"].ID!="video-1"{t.Fatalf("Rename not applied: safe=%v %+v",safe,result)}
 if _,ok:=result["Family/unchanged.mov"];ok{t.Fatal("Old path retained after rename")}
}
func TestDriveDeltaFolderChangesForceFreshInventory(t *testing.T){
 snapshot:=exampleDriveSnapshot()
 for _,changes:=range [][]driveDeltaChange{
  {{FileID:"folder-id",Removed:true}},
  {{FileID:"created-folder",File:driveFile{ID:"created-folder",Name:"NewFolder",Parents:[]string{"real-root-id"},MimeType:folderMime}}},
 }{
  if _,safe:=applyDriveDelta(snapshot,changes);safe{t.Fatal("Folder change must trigger full tree reconstruction")}
 }
}
func TestDriveDeltaExclusionsAndExternalItems(t *testing.T){
 snapshot:=exampleDriveSnapshot()
 changes:=[]driveDeltaChange{
  {FileID:"thumb-id",File:driveFile{ID:"thumb-id",Name:"Thumbs.db",Parents:[]string{"folder-id"},MimeType:"application/octet-stream"}},
  {FileID:"unrelated",File:driveFile{ID:"unrelated",Name:"other.txt",Parents:[]string{"outside-folder"},MimeType:"text/plain"}},
 }
 result,safe:=applyDriveDelta(snapshot,changes)
 if !safe||len(result)!=2{t.Fatalf("Unrelated/ignored changes altered scoped Drive tree: %v %+v",safe,result)}
}
func TestDriveDeltaNamingConflictRebuildsTree(t *testing.T){
 snapshot:=exampleDriveSnapshot()
 changes:=[]driveDeltaChange{{FileID:"collision",File:driveFile{ID:"collision",Name:"unchanged.mov",Parents:[]string{"folder-id"},MimeType:"video/quicktime"}}}
 if _,safe:=applyDriveDelta(snapshot,changes);safe{t.Fatal("Duplicate basename must use full listing name resolution")}
}
func TestDriveSnapshotsPersistAcrossRestartWithCursor(t *testing.T){
 path:=filepath.Join(t.TempDir(),"syncria-index.db")
 idx,err:=loadSyncIndex(path);if err!=nil{t.Fatal(err)}
 snap:=exampleDriveSnapshot()
 if err=idx.saveDriveTree(snap);err!=nil{t.Fatal(err)}
 if err=idx.Close();err!=nil{t.Fatal(err)}
 idx,err=loadSyncIndex(path);if err!=nil{t.Fatal(err)}
 defer idx.Close()
 got,ok,err:=idx.getDriveTree(snap.Account,snap.Root)
 if err!=nil||!ok||got.Cursor!="cursor-1"||len(got.Files)!=2{t.Fatalf("Snapshot lost: ok=%v err=%v %+v",ok,err,got)}
 _,ok,err=idx.getDriveTree(snap.Account,"different-folder")
 if err!=nil||ok{t.Fatal("Root-specific cache leaked to another mapping")}
}
