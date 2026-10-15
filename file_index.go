package main

import (
 "encoding/json"
 "errors"
 "fmt"
 "os"
 "sync"
 "time"

 bolt "go.etcd.io/bbolt"
)

// syncIndex is Syncria's persistent local per-relationship index. It is not
// stored on either synchronized endpoint. Bolt uses atomic write transactions
// and a single local database file, with no server or cgo dependency.
type syncIndex struct {
 mu sync.Mutex
 db *bolt.DB
 last map[string]map[string]map[string]Item
}
var syncIndexBucket=[]byte("sync-pairs-v1")

func clonePairStates(in map[string]map[string]Item)map[string]map[string]Item {
 out:=make(map[string]map[string]Item,len(in))
 for pair,files:=range in{
  entries:=make(map[string]Item,len(files))
  for path,item:=range files{entries[path]=item}
  out[pair]=entries
 }
 return out
}
func loadSyncIndex(path string)(*syncIndex,error) {
 db,err:=bolt.Open(path,0600,&bolt.Options{Timeout:5*time.Second})
 if err!=nil{return nil,fmt.Errorf("open Syncria index: %w",err)}
 if err=os.Chmod(path,0600);err!=nil{db.Close();return nil,err}
 index:=&syncIndex{db:db,last:map[string]map[string]map[string]Item{}}
 err=db.View(func(tx *bolt.Tx)error{
  root:=tx.Bucket(syncIndexBucket)
  if root==nil{return nil}
  return root.ForEach(func(jobID,value []byte)error{
   if value!=nil{return errors.New("Invalid index record: job is not a bucket")}
   job:=root.Bucket(jobID)
   if job==nil{return errors.New("Missing index job bucket")}
   pairs:=map[string]map[string]Item{}
   err:=job.ForEach(func(pairKey,value []byte)error{
    if value!=nil{return errors.New("Invalid index record: relationship is not a bucket")}
    pair:=job.Bucket(pairKey)
    if pair==nil{return errors.New("Missing index pair bucket")}
    items:=map[string]Item{}
    if err:=pair.ForEach(func(path,body []byte)error{
     if body==nil{return errors.New("Unexpected nested file bucket")}
     var item Item
     if err:=json.Unmarshal(body,&item);err!=nil{return fmt.Errorf("Corrupt indexed file: %w",err)}
     items[string(path)]=item
     return nil
    });err!=nil{return err}
    pairs[string(pairKey)]=items
    return nil
   })
   if err!=nil{return err}
   index.last[string(jobID)]=pairs
   return nil
  })
 })
 if err!=nil{db.Close();return nil,fmt.Errorf("Syncria index integrity check: %w",err)}
 return index,nil
}
func (idx *syncIndex) Close()error{return idx.db.Close()}

func sameIndexRecords(left,right map[string]map[string]Item)bool {
 if len(left)!=len(right){return false}
 for key,files:=range left{
  existing,ok:=right[key];if !ok||len(files)!=len(existing){return false}
  for path,value:=range files{other,ok:=existing[path];if !ok||value!=other{return false}}
 }
 return true
}

// hydratePairs imports any existing JSON-based PairStates exactly once.
// Once the database holds a job, it is the authoritative checkpoint source.
// Missing/corrupt index data never causes blind transfer/delete decisions.
func (idx *syncIndex) hydratePairs(jobs []Job)error {
 idx.mu.Lock()
 for i:=range jobs {
  if existing,ok:=idx.last[jobs[i].ID];ok {
   jobs[i].PairStates=clonePairStates(existing)
  }
 }
 idx.mu.Unlock()
 return idx.commitJobs(jobs)
}

// commitJobs writes only changed or deleted index entries, in one atomic
// transaction per checkpoint. A failed DB transaction leaves the last
// committed index intact and prevents the calling sync from continuing.
func (idx *syncIndex) commitJobs(jobs []Job)error {
 if idx==nil{return nil}
 idx.mu.Lock()
 defer idx.mu.Unlock()
 next:=make(map[string]map[string]map[string]Item,len(jobs))
 changed:=len(idx.last)!=len(jobs)
 for _,j:=range jobs{
  next[j.ID]=clonePairStates(j.PairStates)
  if old,ok:=idx.last[j.ID];!ok||!sameIndexRecords(old,j.PairStates){changed=true}
 }
 if !changed{return nil}
 err:=idx.db.Update(func(tx *bolt.Tx)error{
  root,err:=tx.CreateBucketIfNotExists(syncIndexBucket);if err!=nil{return err}
  for oldID:=range idx.last{if _,exists:=next[oldID];!exists{if err=root.DeleteBucket([]byte(oldID));err!=nil{return err}}}
  for jobID,pairs:=range next{
   bucket,err:=root.CreateBucketIfNotExists([]byte(jobID));if err!=nil{return err}
   prev:=idx.last[jobID]
   for pair:=range prev{if _,ok:=pairs[pair];!ok{if err=bucket.DeleteBucket([]byte(pair));err!=nil{return err}}}
   for pair,items:=range pairs{
    existing,err:=bucket.CreateBucketIfNotExists([]byte(pair));if err!=nil{return err}
    old:=prev[pair]
    for path:=range old{if _,ok:=items[path];!ok{if err=existing.Delete([]byte(path));err!=nil{return err}}}
    for path,item:=range items{
     if previous,ok:=old[path];ok&&previous==item{continue}
     b,err:=json.Marshal(item);if err!=nil{return err}
     if err=existing.Put([]byte(path),b);err!=nil{return err}
    }
   }
  }
  return nil
 })
 if err!=nil{return fmt.Errorf("Syncria could not commit file index (sync safely stopped): %w",err)}
 idx.last=next
 return nil
}
