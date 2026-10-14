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
