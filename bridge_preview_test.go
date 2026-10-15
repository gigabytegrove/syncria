package main

import (
 "strings"
 "testing"
 "time"
)
func TestSyncMetadataExclusionIsNarrow(t *testing.T){
 for _,name:=range []string{"sync.ffs_db","SYNC.FFS_DB","Sync.Ffs_Db"}{
  if !ignoredSyncMetadata(name){t.Errorf("must ignore FreeFileSync sidecar %q",name)}
 }
 for _,name:=range []string{"photos.ffs_db","sync.ffs_db.backup","sync.ffs_db.txt","sync.ffs_db-folder","important.txt",""}{
  if ignoredSyncMetadata(name){t.Errorf("ordinary file incorrectly omitted %q",name)}
 }
}
func TestPreviewPlansDirectionAndNeverMutatesState(t *testing.T){
 original:=map[string]map[string]Item{"p": {"old.txt":{MD5:"oldhash",LocalMD5:"same"}}}
 j:=&Job{PairStates:original,Direction:"bidirectional"}
 left:=map[string]bridgeEntry{
 "new.txt":{Size:45,Fingerprint:"new"},
 "changed.txt":{Size:11,Fingerprint:"leftnew"},
 "old.txt":{Size:3,Fingerprint:"oldhash"},
 }
 right:=map[string]bridgeEntry{
 "back.txt":{Size:9,Fingerprint:"back"},
 "changed.txt":{Size:10,Fingerprint:"rightnew"},
 "old.txt":{Size:3,Fingerprint:"same"},
 }
 j.PairStates["p"]["changed.txt"]=Item{MD5:"leftold",LocalMD5:"rightold"}
 plan:=syncPreviewResult{Counts:map[string]int{},Rows:[]syncPreviewRow{}}
 planSyncPair(&plan,j,"p",left,right,false,conflictManual)
 got:=map[string]syncPreviewRow{}
 for _,row:=range plan.Rows{got[row.Path]=row}
 if got["new.txt"].From!="Primary"||got["new.txt"].To!="Secondary"{t.Fatal("forward file plan missing")}
 if got["back.txt"].From!="Secondary"||got["back.txt"].To!="Primary"{t.Fatal("reverse file plan missing")}
 if got["changed.txt"].Action!="review"{t.Fatal("manual conflict incorrectly auto-resolved")}
 if plan.Counts["unchanged"]!=1{t.Fatal("unchanged tracked file not counted")}
 if _,visible:=got["old.txt"];visible{t.Fatal("unchanged tracked file must not appear in actionable preview")}
 if j.PairStates["p"]["changed.txt"].MD5!="leftold"{t.Fatal("preview mutated checkpoint")}
}
func TestPreviewCollisionAndSkipMessages(t *testing.T){
 j:=&Job{Direction:"one_way",PairStates:map[string]map[string]Item{"x":{}}}
 plan:=syncPreviewResult{Counts:map[string]int{},Rows:[]syncPreviewRow{}}
 planSyncPair(&plan,j,"x",map[string]bridgeEntry{},map[string]bridgeEntry{"secondary.txt":{Size:8,Modified:time.Now()}},true,conflictManual)
 if len(plan.Rows)!=1||plan.Rows[0].Action!="skip"||!strings.Contains(plan.Rows[0].Reason,"One-way"){t.Fatal("one-way preview must disclose skipped item")}
}
