package main

import (
 "context"
 "strings"
 "testing"
)

// Initial device/NAS collisions cannot be authenticated by timestamps or size.
func TestAgentFirstRunSameSizeTimeDoesNotAssumeIdenticalBytes(t *testing.T) {
 j:=Job{ID:"device-collision",Provider:"agent_bridge",Direction:"one_way",PairStates:map[string]map[string]Item{}}
 a:=policyTestApp(t,j)
 p:=newRecoveryFake(map[string]string{"existing.txt":"source"})
 s:=newRecoveryFake(map[string]string{"existing.txt":"target"})
 if len(p.files["existing.txt"])!=len(s.files["existing.txt"]){t.Fatal("invalid test setup")}
 err:=a.executePolicyPair(context.Background(),&j,p,s,"pair",true,conflictManual)
 if err==nil||!strings.Contains(strings.ToLower(err.Error()),"conflict"){t.Fatalf("collision was accepted: %v",err)}
 if p.files["existing.txt"]!="source"||s.files["existing.txt"]!="target"{t.Fatal("collision modified user files")}
 if len(j.PairStates["pair"])!=0{t.Fatal("unverified file was checkpointed")}
}

func TestAgentDestinationEditedIndependentlyIsNeverReplaced(t *testing.T) {
 j:=Job{ID:"device-edited",Provider:"agent_bridge",Direction:"one_way",PairStates:map[string]map[string]Item{"pair":{"photo.jpg":fileState("photo.jpg","previous","previous")}}}
 a:=policyTestApp(t,j)
 p:=newRecoveryFake(map[string]string{"photo.jpg":"previous"})
 s:=newRecoveryFake(map[string]string{"photo.jpg":"modified outside Syncria"})
 if err:=a.executePolicyPair(context.Background(),&j,p,s,"pair",true,conflictManual);err==nil{t.Fatal("independent NAS edit was overwritten")}
 if s.files["photo.jpg"]!="modified outside Syncria"{t.Fatal("NAS copy changed")}
}

func TestAgentSourceDeletionDoesNotDeleteNASFile(t *testing.T) {
 no:=false
 j:=Job{ID:"device-delete",Provider:"agent_bridge",Direction:"one_way",PropagateDeletions:&no,PairStates:map[string]map[string]Item{"pair":{"keep.txt":fileState("keep.txt","previous","previous")}}}
 a:=policyTestApp(t,j)
 p:=newRecoveryFake(map[string]string{})
 s:=newRecoveryFake(map[string]string{"keep.txt":"previous"})
 if err:=a.executePolicyPair(context.Background(),&j,p,s,"pair",true,conflictManual);err!=nil{t.Fatal(err)}
 if s.files["keep.txt"]!="previous"||len(s.trash)!=0{t.Fatal("agent-source deletion propagated to NAS")}
}

func TestAgentEmptyNASReceivesApprovedNewFile(t *testing.T) {
 j:=Job{ID:"device-new",Provider:"agent_bridge",Direction:"one_way",PairStates:map[string]map[string]Item{}}
 a:=policyTestApp(t,j)
 p:=newRecoveryFake(map[string]string{"new.txt":"approved data"})
 s:=newRecoveryFake(map[string]string{})
 if err:=a.executePolicyPair(context.Background(),&j,p,s,"pair",true,conflictManual);err!=nil{t.Fatal(err)}
 if s.files["new.txt"]!="approved data"{t.Fatal("approved file not copied")}
}
