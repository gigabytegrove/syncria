package main

import (
 "context"
 "errors"
 "fmt"
 "time"

 "github.com/gigabytegrove/syncria/agentpeer"
)

// previewAgentBridgeProgress reuses the main read-only Syncria preview plan
// and exactly the same approved device source and NAS destination as an agent
// sync run. It never executes a Put, Delete or destination Mkdir operation.
func (a *App) previewAgentBridgeProgress(ctx context.Context,j *Job,notify func(string,string,int64))(syncPreviewResult,error){
 result:=syncPreviewResult{
  Job:j.Name,ReadOnly:true,Generated:time.Now(),
  Rows:[]syncPreviewRow{},Counts:map[string]int{},
  Warnings:[]string{
   "Device-to-NAS alpha previews are read-only; do not rely on their result after files change.",
   "An existing destination filename blocks a device's first sync, even when sizes and modification times match.",
   "Source deletions do not propagate to NAS for agent-to-NAS alpha jobs.",
  },
 }
 if j.Provider!="agent_bridge"||j.Direction!="one_way"||j.AgentDeviceID==""||j.AgentGrantID==""||j.AgentPeerURL==""||len(j.NASDestinations)!=1 {
  return result,errors.New("invalid device-to-NAS relationship")
 }
 token,err:=readAgentPeerCredential(a.file,j.ID)
 if err!=nil{return result,fmt.Errorf("approved device credential unavailable: %w",err)}
 primary:=&agentReadEndpoint{peer:agentpeer.Client{Server:j.AgentPeerURL,Token:token},grant:j.AgentGrantID}
 if notify!=nil{primary.scanHook=func(rel string,count int64){notify("Device",rel,count)}}
 if notify!=nil{notify("Device","Scanning selected files",0)}
 left,err:=primary.Scan(ctx)
 if err!=nil{return result,fmt.Errorf("approved device scan: %w",err)}
 destination:=j.NASDestinations[0]
 nas,err:=a.openNAS(ctx,destination)
 if err!=nil{return result,fmt.Errorf("NAS destination unavailable: %w",err)}
 defer nas.Close()
 if smb,ok:=nas.(*bridgeSMBEndpoint);ok&&notify!=nil{
  smb.scanHook=func(rel string,count int64){notify("NAS",rel,count)}
 }
 if notify!=nil{notify("NAS","Scanning destination files",0)}
 right,err:=nas.Scan(ctx)
 if err!=nil{return result,fmt.Errorf("NAS destination scan: %w",err)}
 key:=relationPairKey("device:"+j.AgentDeviceID+":"+j.AgentGrantID,destination)
 if notify!=nil{notify("Planning changes","Comparing approved files",int64(len(left)+len(right)))}
 // The shared preview planner reads the existing indexed pair-state snapshot.
 // No guessed state is persisted, and no file bodies are transferred.
 copy:=*j
 copy.PropagateDeletions=boolPtr(false)
 planSyncPair(&result,&copy,key,left,right,true,conflictManual)
 return result,nil
}
