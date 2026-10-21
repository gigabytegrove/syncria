package nasversion

import (
 "context"
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
 "os"
 "path"
 "sort"
 "strings"
 "time"
)

// RetentionPlan previews bounded archival cleanup. No deletion is automatic.
type RetentionPlan struct {
 JobID string `json:"job_id"`
 KeepPerFile int `json:"keep_per_file"`
 MaxBytes int64 `json:"max_bytes"`
 TotalBytes int64 `json:"total_bytes"`
 ReclaimBytes int64 `json:"reclaim_bytes"`
 Delete []Version `json:"delete"`
 Digest string `json:"digest"`
}

type RetentionRemote interface {
 Lister
 Stat(string)(os.FileInfo,error)
 Remove(string)error
}
func retentionDigest(p RetentionPlan)(string,error){
 p.Digest=""
 b,e:=json.Marshal(p);if e!=nil{return "",e}
 sum:=sha256.Sum256(b)
 return hex.EncodeToString(sum[:]),nil
}

// PreviewRetention is scoped to one NAS job. A missing/inaccessible archive
// is never treated as proof that old versions are safe to delete.
func PreviewRetention(ctx context.Context,remote RetentionRemote,root,jobID string,keepPerFile int,maxBytes int64)(RetentionPlan,error){
 if remote==nil||!safeRoot(root)||!validID(jobID)||keepPerFile<1||keepPerFile>1000||maxBytes<1||maxBytes>1<<50{return RetentionPlan{},ErrUnsafePath}
 versions,err:=List(ctx,remote,root,jobID);if err!=nil{return RetentionPlan{},err}
 if len(versions)>10000{return RetentionPlan{},errors.New("archive inventory exceeds safe limits")}
 stamps:=map[string]time.Time{}
 for _,v:=range versions {
  if err:=ctx.Err();err!=nil{return RetentionPlan{},err}
  if _,ok:=stamps[v.VersionID];ok{continue}
  info,e:=remote.Stat(path.Join(root,".syncria-versions",jobID,v.VersionID))
  if e!=nil||!info.IsDir()||info.ModTime().IsZero(){return RetentionPlan{},errors.New("version capture timestamp unavailable")}
  if !info.ModTime().Equal(v.Captured){return RetentionPlan{},errors.New("version changed during inventory")}
  stamps[v.VersionID]=info.ModTime()
 }
 sort.Slice(versions,func(i,j int)bool{
  l,r:=stamps[versions[i].VersionID],stamps[versions[j].VersionID]
  if l.Equal(r){return versions[i].Path<versions[j].Path}
  return l.After(r)
 })
 plan:=RetentionPlan{JobID:jobID,KeepPerFile:keepPerFile,MaxBytes:maxBytes,Delete:[]Version{}}
 kept:=make([]Version,0,len(versions))
 count:=map[string]int{}
 for _,v:=range versions{
  if v.Bytes<0||v.Bytes>1<<50||plan.TotalBytes>1<<50-v.Bytes{return RetentionPlan{},errors.New("archive size overflow")}
  plan.TotalBytes+=v.Bytes;count[v.Relative]++
  if count[v.Relative]>keepPerFile{plan.Delete=append(plan.Delete,v);plan.ReclaimBytes+=v.Bytes}else{kept=append(kept,v)}
 }
 remaining:=plan.TotalBytes-plan.ReclaimBytes
 survivors:=map[string]int{}
 for _,v:=range kept{survivors[v.Relative]++}
 for i:=len(kept)-1;i>=0&&remaining>maxBytes;i--{
  v:=kept[i];if survivors[v.Relative]<=1{continue}
  survivors[v.Relative]--;plan.Delete=append(plan.Delete,v)
  plan.ReclaimBytes+=v.Bytes;remaining-=v.Bytes
 }
 if remaining>maxBytes{return RetentionPlan{},fmt.Errorf("quota would require deleting the final recovery version of a file")}
 for i:=1;i<len(versions);i++{
  if versions[i].Relative==versions[i-1].Relative&&stamps[versions[i].VersionID].Equal(stamps[versions[i-1].VersionID])&&len(plan.Delete)>0{
   return RetentionPlan{},errors.New("archive capture chronology ambiguous; manual review required")
  }
 }
 sort.Slice(plan.Delete,func(i,j int)bool{return plan.Delete[i].Path<plan.Delete[j].Path})
 plan.Digest,err=retentionDigest(plan)
 return plan,err
}

// ApplyRetention only acts after a matching administrator-reviewed preview.
// Paths are constrained to the dedicated archive; live files are never removed.
func ApplyRetention(ctx context.Context,remote RetentionRemote,root,jobID string,keepPerFile int,maxBytes int64,digest string)(int,error){
 if len(digest)!=64{return 0,ErrUnsafePath}
 plan,err:=PreviewRetention(ctx,remote,root,jobID,keepPerFile,maxBytes)
 if err!=nil{return 0,err}
 if !strings.EqualFold(plan.Digest,digest){return 0,errors.New("archive changed since preview; nothing removed")}
 removed:=0
 for _,v:=range plan.Delete{
  if err:=ctx.Err();err!=nil{return removed,err}
  p:=path.Join(root,".syncria-versions",jobID,v.VersionID,v.Relative)
  if p!=v.Path||!validID(v.VersionID)||!safeRel(v.Relative){return removed,ErrUnsafePath}
  info,e:=remote.Stat(p);if e!=nil{return removed,e}
  if !info.Mode().IsRegular()||info.Size()!=v.Bytes||!info.ModTime().Equal(v.Modified){return removed,errors.New("archived file changed; remaining versions preserved")}
  if e:=remote.Remove(p);e!=nil{return removed,fmt.Errorf("remove archived version: %w",e)}
  removed++
 }
 return removed,nil
}
