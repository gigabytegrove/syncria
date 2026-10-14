package main

import (
 "context"
 "crypto/md5"
 "encoding/hex"
 "os"
 "path/filepath"
 "strings"
 "testing"
 "time"
)

type recoveryFake struct{
 *fakeBridgeEndpoint
 trash map[string]string
 modified map[string]time.Time
}
func newRecoveryFake(files map[string]string)*recoveryFake{
 return &recoveryFake{fakeBridgeEndpoint:&fakeBridgeEndpoint{files:files},trash:map[string]string{},modified:map[string]time.Time{}}
}
func (m *recoveryFake) Scan(ctx context.Context)(map[string]bridgeEntry,error){
 list,err:=m.fakeBridgeEndpoint.Scan(ctx);if err!=nil{return nil,err}
 for rel,entry:=range list{
  entry.Modified=m.modified[rel]
  list[rel]=entry
 }
 return list,nil
}
func (m *recoveryFake) StageDelete(ctx context.Context,rel,jobID string)(string,error){
 token:=jobID+"-"+rel
 m.trash[token]=m.files[rel]
 delete(m.files,rel)
 return token,nil
}
func (m *recoveryFake) Restore(ctx context.Context,rel,token string)error{
 v,ok:=m.trash[token];if !ok{return os.ErrNotExist}
 m.files[rel]=v
 delete(m.trash,token)
 return nil
}
func (m *recoveryFake) Purge(ctx context.Context,token string)error{delete(m.trash,token);return nil}
func (m *recoveryFake) Delete(ctx context.Context,rel string)error{delete(m.files,rel);return nil}
func policyTestApp(t *testing.T,j Job)*App{
 t.Helper()
 return &App{file:filepath.Join(t.TempDir(),"state.json"),cfg:Config{Jobs:[]Job{j}}}
}
func fileState(name,source,target string)Item{
 l:=md5.Sum([]byte(source));r:=md5.Sum([]byte(target))
 return Item{ID:"file-"+name,MD5:hex.EncodeToString(l[:]),LocalMD5:hex.EncodeToString(r[:])}
}
func TestNASOneWayOnlySourceToDestination(t *testing.T){
 ctx:=context.Background()
 j:=Job{ID:"one",PairStates:map[string]map[string]Item{}}
 a:=policyTestApp(t,j)
 p:=newRecoveryFake(map[string]string{"photo.jpg":"original"})
 s:=newRecoveryFake(map[string]string{})
 if err:=a.executePolicyPair(ctx,&j,p,s,"pair",true,conflictManual);err!=nil{t.Fatal(err)}
 if s.files["photo.jpg"]!="original"{t.Fatal("one-way source did not copy to destination")}
 // Extra destination-side files should not go backwards into the source.
 p=newRecoveryFake(map[string]string{})
 s=newRecoveryFake(map[string]string{"extra.txt":"destination-only"})
 j.PairStates=map[string]map[string]Item{}
 if err:=a.executePolicyPair(ctx,&j,p,s,"extra-pair",true,conflictManual);err!=nil{t.Fatal(err)}
 if _,exists:=p.files["extra.txt"];exists{t.Fatal("one-way sync transferred a destination file backwards")}
}
func TestNASBidirectionalUploadsBothWays(t *testing.T){
 j:=Job{ID:"two",PairStates:map[string]map[string]Item{}}
 a:=policyTestApp(t,j)
 primary:=newRecoveryFake(map[string]string{})
 secondary:=newRecoveryFake(map[string]string{"from-nas2.txt":"hello"})
 if err:=a.executePolicyPair(context.Background(),&j,primary,secondary,"pair",false,conflictManual);err!=nil{t.Fatal(err)}
 if primary.files["from-nas2.txt"]!="hello"{t.Fatal("reverse NAS transfer missing")}
}
func TestNASConflictWinnerPolicies(t *testing.T){
 earlier:=time.Now().Add(-time.Hour)
 latest:=time.Now()
 for _,tc:=range []struct{policy,want string}{
  {conflictNewer,"secondary"},{conflictPrimary,"primary"},{conflictSecondary,"secondary"},
 }{
  t.Run(tc.policy,func(t *testing.T){
   j:=Job{ID:"a",PairStates:map[string]map[string]Item{"pair":{"x.txt":fileState("x.txt","previous","previous")}}}
   a:=policyTestApp(t,j)
   primary:=newRecoveryFake(map[string]string{"x.txt":"primary"})
   secondary:=newRecoveryFake(map[string]string{"x.txt":"secondary"})
   primary.modified["x.txt"]=earlier;secondary.modified["x.txt"]=latest
   if err:=a.executePolicyPair(context.Background(),&j,primary,secondary,"pair",false,tc.policy);err!=nil{t.Fatal(err)}
   if primary.files["x.txt"]!=tc.want||secondary.files["x.txt"]!=tc.want{t.Fatalf("policy %s not honored",tc.policy)}
  })
 }
}
func TestNASKeepBothRetainsSecondary(t *testing.T){
 j:=Job{ID:"a",PairStates:map[string]map[string]Item{"pair":{"x.txt":fileState("x.txt","old","old")}}}
 a:=policyTestApp(t,j)
 primary:=newRecoveryFake(map[string]string{"x.txt":"primary"})
 secondary:=newRecoveryFake(map[string]string{"x.txt":"secondary"})
 if err:=a.executePolicyPair(context.Background(),&j,primary,secondary,"pair",false,conflictKeepBoth);err!=nil{t.Fatal(err)}
 if secondary.files["x.txt"]!="primary"{t.Fatal("primary file not selected for keep both")}
 found:=false
 for name,value:=range secondary.files{if strings.Contains(name,"(conflict-")&&value=="secondary"{found=true}}
 if !found{t.Fatal("secondary conflicting version was not kept")}
}
func TestNASManualConflictIsRecordedAndResolvable(t *testing.T){
 j:=Job{ID:"a",PairStates:map[string]map[string]Item{"pair":{"x.txt":fileState("x.txt","old","old")}}}
 a:=policyTestApp(t,j)
 primary:=newRecoveryFake(map[string]string{"x.txt":"primary"})
 secondary:=newRecoveryFake(map[string]string{"x.txt":"secondary"})
 if err:=a.executePolicyPair(context.Background(),&j,primary,secondary,"pair",false,conflictManual);err==nil{t.Fatal("Manual policy silently overwrote conflict")}
 if len(j.Conflicts)!=1||len(a.cfg.Jobs[0].Conflicts)!=1{t.Fatal("Conflict not durably recorded")}
 if primary.files["x.txt"]!="primary"||secondary.files["x.txt"]!="secondary"{t.Fatal("Manual conflict changed files")}
 j.ConflictOverrides=map[string]string{"pair|x.txt":"secondary"}
 if err:=a.executePolicyPair(context.Background(),&j,primary,secondary,"pair",false,conflictManual);err!=nil{t.Fatal(err)}
 if primary.files["x.txt"]!="secondary"||len(j.Conflicts)!=0{t.Fatal("Selected manual resolution did not apply")}
}
func TestDeletionPropagationWithTenMinuteUndo(t *testing.T){
 key:="pair"
 j:=Job{ID:"recover",PairStates:map[string]map[string]Item{key:{"x.txt":fileState("x.txt","kept","kept")}}}
 a:=policyTestApp(t,j)
 primary:=newRecoveryFake(map[string]string{})
 secondary:=newRecoveryFake(map[string]string{"x.txt":"kept"})
 if err:=a.executePolicyPair(context.Background(),&j,primary,secondary,key,false,conflictManual);err!=nil{t.Fatal(err)}
 if len(j.PendingDeletions)!=1||!j.PendingDeletions[0].Expires.After(time.Now().Add(9*time.Minute)){
  t.Fatal("Ten-minute recovery window not created")
 }
 if secondary.files["x.txt"]!=""||len(secondary.trash)!=1{t.Fatal("NAS delete was not moved to NAS-side recovery")}
 if len(a.cfg.Jobs[0].PendingDeletions)!=1{t.Fatal("Recovery token was not persisted")}
}
func TestDeletionOptionsPerRelationship(t *testing.T){
 key:="pair"
 no:=false
 for _,tc:=range []struct{name string;propagate,recovery *bool;present,hasPending bool}{
  {"do-not-propagate",&no,nil,true,false},
  {"immediate-delete",nil,&no,false,false},
 }{
  t.Run(tc.name,func(t *testing.T){
   j:=Job{ID:"a",PropagateDeletions:tc.propagate,RecoveryEnabled:tc.recovery,PairStates:map[string]map[string]Item{key:{"x.txt":fileState("x.txt","x","x")}}}
   a:=policyTestApp(t,j)
   p:=newRecoveryFake(map[string]string{})
   s:=newRecoveryFake(map[string]string{"x.txt":"x"})
   if err:=a.executePolicyPair(context.Background(),&j,p,s,key,false,conflictManual);err!=nil{t.Fatal(err)}
   _,exists:=s.files["x.txt"]
   if exists!=tc.present|| (len(j.PendingDeletions)>0)!=tc.hasPending{t.Fatal("Per-relationship deletion flags not honored")}
  })
 }
}
func TestOnlyApprovedNASConflictPolicies(t *testing.T){
 for _,policy:=range []string{"newer","primary","secondary","keep_both","manual"}{if !validNASPolicy(policy){t.Fatal(policy)}}
 for _,policy:=range []string{"oldest","largest","ignore","overwrite_all",""}{if validNASPolicy(policy){t.Fatal(policy)}}
}

func TestFirstRunMetadataDoesNotReadThirtyGigabytes(t *testing.T){
 ctx:=context.Background()
 primary:=newRecoveryFake(map[string]string{})
 secondary:=newRecoveryFake(map[string]string{})
 stamp:=time.Date(2026,10,14,12,0,0,0,time.UTC)
 huge:=int64(30)*1024*1024*1024
 same,err:=bridgeSameFirstRun(ctx,primary,secondary,"huge.mov",bridgeEntry{Size:huge,Modified:stamp},bridgeEntry{Size:huge,Modified:stamp})
 if err!=nil||!same{t.Fatalf("Equal metadata must skip even massive files: same=%v err=%v",same,err)}
 same,err=bridgeSameFirstRun(ctx,primary,secondary,"huge.mov",bridgeEntry{Size:huge,Modified:stamp.Add(time.Hour)},bridgeEntry{Size:huge,Modified:stamp})
 if err!=nil||same{t.Fatalf("Different timestamps should require sync decision: same=%v err=%v",same,err)}
 if _,err=bridgeSameFirstRun(ctx,primary,secondary,"huge.mov",bridgeEntry{Size:huge,Modified:stamp},bridgeEntry{Size:huge-1,Modified:stamp});err==nil{t.Fatal("Equal timestamp with different sizes must be reviewed")}
 if _,err=bridgeSameFirstRun(ctx,primary,secondary,"huge.mov",bridgeEntry{Size:huge},bridgeEntry{Size:huge});err==nil{t.Fatal("Missing timestamps must never be assumed identical")}
}
