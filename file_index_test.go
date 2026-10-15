package main

import (
 "os"
 "path/filepath"
 "strconv"
 "testing"
 "time"
)

func TestIndexMigratesLegacyCheckpointsAndSurvivesRestart(t *testing.T){
 path:=filepath.Join(t.TempDir(),"syncria-index.db")
 index,err:=loadSyncIndex(path);if err!=nil{t.Fatal(err)}
 old:=[]Job{{ID:"job-a",PairStates:map[string]map[string]Item{"account|nas|folder":{
  "Personal Data/Kaleb/video.mp4":{ID:"id",MD5:"cloudHash",LocalMD5:"100:123"},
 }}}}
 if err=index.hydratePairs(old);err!=nil{t.Fatal(err)}
 if err=index.Close();err!=nil{t.Fatal(err)}
 info,err:=os.Stat(path);if err!=nil{t.Fatal(err)}
 if info.Mode().Perm()&0077!=0{t.Fatalf("Index must be private: %o",info.Mode().Perm())}
 reopened,err:=loadSyncIndex(path);if err!=nil{t.Fatal(err)}
 defer reopened.Close()
 blank:=[]Job{{ID:"job-a"}}
 if err=reopened.hydratePairs(blank);err!=nil{t.Fatal(err)}
 record:=blank[0].PairStates["account|nas|folder"]["Personal Data/Kaleb/video.mp4"]
 if record.MD5!="cloudHash"||record.LocalMD5!="100:123"{t.Fatal("Migrated records weren't restored from index")}
}

func TestIndexAtomicallyAdvancesOnlyCommittedChanges(t *testing.T){
 path:=filepath.Join(t.TempDir(),"index.db")
 idx,err:=loadSyncIndex(path);if err!=nil{t.Fatal(err)}
 jobs:=[]Job{{ID:"j",PairStates:map[string]map[string]Item{
  "pair":{"old.txt":{MD5:"A",LocalMD5:"B"}},
 }}}
 if err=idx.hydratePairs(jobs);err!=nil{t.Fatal(err)}
 jobs[0].PairStates["pair"]["new.txt"]=Item{MD5:"new-file",LocalMD5:"new-nas"}
 if err=idx.commitJobs(jobs);err!=nil{t.Fatal(err)}
 delete(jobs[0].PairStates["pair"],"old.txt")
 if err=idx.commitJobs(jobs);err!=nil{t.Fatal(err)}
 if err=idx.Close();err!=nil{t.Fatal(err)}
 idx,err=loadSyncIndex(path);if err!=nil{t.Fatal(err)}
 defer idx.Close()
 restore:=[]Job{{ID:"j"}}
 if err=idx.hydratePairs(restore);err!=nil{t.Fatal(err)}
 pair:=restore[0].PairStates["pair"]
 if len(pair)!=1||pair["new.txt"].MD5!="new-file"{t.Fatalf("Wrong persisted transaction: %+v",pair)}
}
func TestIndexMappingResetDoesNotResurrectOldPair(t *testing.T){
 path:=filepath.Join(t.TempDir(),"index.db")
 idx,err:=loadSyncIndex(path);if err!=nil{t.Fatal(err)}
 defer idx.Close()
 jobs:=[]Job{{ID:"j",PairStates:map[string]map[string]Item{"OLD_SHARE":{"a.txt":{MD5:"old"}}}}}
 if err=idx.hydratePairs(jobs);err!=nil{t.Fatal(err)}
 jobs[0].PairStates=map[string]map[string]Item{"NEW_SHARE":{}}
 if err=idx.commitJobs(jobs);err!=nil{t.Fatal(err)}
 if _,exists:=idx.last["j"]["OLD_SHARE"];exists{t.Fatal("Old mapping index was not invalidated")}
}

func TestIndexMetadataDecisionsIgnoreUnchangedContentHashes(t *testing.T){
 stamp:=time.Date(2026,10,14,18,0,0,0,time.UTC)
 left:=bridgeEntry{ID:"file",Fingerprint:"cloud-md5-old",Size:30*1024*1024*1024,Modified:stamp}
 right:=bridgeEntry{Fingerprint:"size:modtime",Size:left.Size,Modified:stamp}
 indexed:=indexedPairItem(left,right)
 left.Fingerprint="different-provider-hash"
 right.Fingerprint="different-metadata-hash"
 changedL,changedR:=metadataChanged(indexed,left,right)
 if changedL||changedR{t.Fatal("Unchanged timestamps/sizes must not transfer 30 GB video")}
 right.Modified=stamp.Add(10*time.Second)
 _,changedR=metadataChanged(indexed,left,right)
 if !changedR{t.Fatal("Changed NAS timestamp should be detected")}
 right.Modified=stamp;right.Size++
 _,changedR=metadataChanged(indexed,left,right)
 if !changedR{t.Fatal("Changed file size should be detected")}
}

func TestIndexCanCommitFiveThousandEntriesWithoutPayloadReads(t *testing.T){
 path:=filepath.Join(t.TempDir(),"index.db")
 idx,err:=loadSyncIndex(path);if err!=nil{t.Fatal(err)}
 defer idx.Close()
 files:=map[string]Item{}
 for i:=0;i<5800;i++{files["file-"+strconv.Itoa(i)]=Item{MD5:"hash",LocalMD5:"nas",SourceSize:123,DestinationSize:123,IndexVersion:1}}
 jobs:=[]Job{{ID:"bulk",PairStates:map[string]map[string]Item{"pair":files}}}
 if err=idx.hydratePairs(jobs);err!=nil{t.Fatal(err)}
 if len(idx.last["bulk"]["pair"])!=5800{t.Fatal("Index lost files")}
 if err=idx.commitJobs(jobs);err!=nil{t.Fatal(err)}
}
