package main

import (
 "os"
 "path/filepath"
 "strings"
 "testing"
)

func TestFFSMetadataIgnoredAtAnyDepth(t *testing.T) {
 for _,rel:=range []string{"sync.ffs_db","Nested/SYNC.FFS_DB","Deep/Other/sync.ffs_db"} {
  if !ignoredSyncPath(rel){t.Fatalf("metadata not ignored: %q",rel)}
 }
 for _,rel:=range []string{"","sync.ffs_db.bak","Sub/sync.ffs_db.old","sync.ffs_db/family.txt","photos.ffs_db","file.txt"} {
  if ignoredSyncPath(rel){t.Fatalf("ordinary file was ignored: %q",rel)}
 }
}
func TestLocalScannerDoesNotHashFFSDatabase(t *testing.T) {
 root:=t.TempDir()
 if err:=os.MkdirAll(filepath.Join(root,"nested"),0700);err!=nil{t.Fatal(err)}
 for _,name:=range []string{"sync.ffs_db","nested/SYNC.FFS_DB","nested/keep.txt"}{
  if err:=os.WriteFile(filepath.Join(root,filepath.FromSlash(name)),[]byte("content"),0600);err!=nil{t.Fatal(err)}
 }
 tree,err:=localTree(root)
 if err!=nil{t.Fatal(err)}
 if _,ok:=tree["sync.ffs_db"];ok{t.Fatal("root sidecar incorrectly scanned")}
 if _,ok:=tree["nested/SYNC.FFS_DB"];ok{t.Fatal("nested sidecar incorrectly scanned")}
 if _,ok:=tree["nested/keep.txt"];!ok{t.Fatal("ordinary file missing")}
}
func TestLegacyFFSCheckpointsNeverBecomeDeletionPlans(t *testing.T) {
 job:=&Job{PairStates:map[string]map[string]Item{
  "one":{"sync.ffs_db":{ID:"existing-1",MD5:"a"},"data/sync.ffs_db":{ID:"existing-2",MD5:"b"}},
 }}
 plan:=syncPreviewResult{Rows:[]syncPreviewRow{},Counts:map[string]int{}}
 planSyncPair(&plan,job,"one",map[string]bridgeEntry{},map[string]bridgeEntry{},false,conflictManual)
 if len(plan.Rows)!=0||len(plan.Counts)!=0{t.Fatalf("must not plan deletion of FreeFileSync databases: %+v",plan)}
 // The dry run must not change the actual checkpoints.
 if len(job.PairStates["one"])!=2{t.Fatal("read-only preview mutated checkpoint state")}
}
func TestFFSOldFailureClearlyLabeledAsHistorical(t *testing.T){
 for _,text:=range []string{"Last sync failed","job.running","job.last_run"} {
  if !strings.Contains(newUI,text){t.Fatalf("missing last-error context: %s",text)}
 }
}

func TestWindowsThumbnailSidecarsAreNeverTransferredOrDeleted(t *testing.T) {
 for _,p:=range []string{
  "Thumbs.db","THUMBS.DB","family/photos/Thumbs.db",
  "nested\\THUMBS.DB","Deep/ehthumbs.db","Deep/EHTHUMBS_VISTA.DB",
 } {
  if !ignoredSyncPath(p){t.Fatalf("Thumbnail cache was not excluded: %q",p)}
 }
 for _,p:=range []string{"Thumbs.db.bak","myThumbs.db","family/Thumbs.db/photos.jpg"} {
  if ignoredSyncPath(p){t.Fatalf("Regular user file incorrectly ignored: %q",p)}
 }
 j:=&Job{PairStates:map[string]map[string]Item{"pair":{
  "Thumbs.db":{ID:"old-cached-file"},
  "A/THUMBS.DB":{ID:"old-cached-file2"},
 }}}
 result:=syncPreviewResult{Counts:map[string]int{},Rows:[]syncPreviewRow{}}
 planSyncPair(&result,j,"pair",map[string]bridgeEntry{},map[string]bridgeEntry{},false,conflictManual)
 if len(result.Rows)>0||len(result.Counts)>0{t.Fatalf("Legacy excluded cache must not cause deletion: %+v",result)}
}
