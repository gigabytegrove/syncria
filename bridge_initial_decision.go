package main

import (
 "fmt"
 "time"
)

// firstRunDecision is the single decision rule for both the read-only plan
// and the first real synchronization. It does not inspect checkpoint history.
// Actual execution must compare bytes before applying this direction.
type firstRunDecision struct {
 Direction string
 Method string
 Reason string
}

// Cloud-to-NAS relationships have historically persisted ConflictPolicy:
// "manual" even though the user did not select it. First-run decisions use
// a SEPARATE optional preference, defaulting to "newer" for existing jobs.
// This does not change the ongoing manual conflict rule.
func firstRunPolicy(j *Job, ongoingPolicy string) string {
 if j.Provider=="bridge" {
  if validNASPolicy(j.InitialConflictPolicy){return j.InitialConflictPolicy}
  return conflictNewer
 }
 return ongoingPolicy
}

func fileModifiedText(t time.Time) string {
 if t.IsZero(){return "unknown"}
 return t.UTC().Format("2006-01-02 15:04:05 UTC")
}

func decideFirstRun(j *Job,key,rel string,l,r bridgeEntry,oneWay bool,ongoingPolicy string) firstRunDecision {
 if oneWay {
  return firstRunDecision{Reason:"Both locations already contain this file. One-way synchronization cannot overwrite an existing destination without an explicit resolution."}
 }
 if override:=j.ConflictOverrides[key+"|"+rel];override!=""{
  switch override {
  case "primary","secondary":
   return firstRunDecision{Direction:override,Method:override,Reason:"Explicit file-level conflict selection; contents will be compared before any replacement."}
  case conflictKeepBoth:
   return firstRunDecision{Direction:"primary",Method:conflictKeepBoth,Reason:"Explicit choice to preserve both versions."}
  }
 }
 policy:=firstRunPolicy(j,ongoingPolicy)
 switch policy {
 case conflictPrimary:
  return firstRunDecision{Direction:"primary",Method:conflictPrimary,Reason:"Configured preference: primary wins when contents differ."}
 case conflictSecondary:
  return firstRunDecision{Direction:"secondary",Method:conflictSecondary,Reason:"Configured preference: secondary wins when contents differ."}
 case conflictKeepBoth:
  return firstRunDecision{Direction:"primary",Method:conflictKeepBoth,Reason:"Configured preference: keep both versions."}
 case conflictNewer:
  // A tolerance avoids pretending that SMB, FAT, and cloud timestamp
  // precision differences are proof of which copy is newer.
  if l.Modified.IsZero()||r.Modified.IsZero(){
   return firstRunDecision{Reason:fmt.Sprintf("First run: missing modification time (primary %s; secondary %s). Manual decision needed; no overwrite.",fileModifiedText(l.Modified),fileModifiedText(r.Modified))}
  }
  gap:=l.Modified.Sub(r.Modified)
  const minimumGap=2*time.Second
  if gap<minimumGap&&gap>-minimumGap{
   return firstRunDecision{Reason:fmt.Sprintf("First run: modification times too close to trust (primary %s; secondary %s). Manual decision needed.",fileModifiedText(l.Modified),fileModifiedText(r.Modified))}
  }
  direction:="primary"
  if gap<0{direction="secondary"}
  return firstRunDecision{Direction:direction,Method:conflictNewer,
   Reason:fmt.Sprintf("First run: newer modification timestamp wins (primary %s; secondary %s). Content is compared before execution; identical copies are not overwritten.",fileModifiedText(l.Modified),fileModifiedText(r.Modified))}
 case conflictManual:
  return firstRunDecision{Reason:fmt.Sprintf("Manual conflict policy: primary %s; secondary %s. Choose the version to keep.",fileModifiedText(l.Modified),fileModifiedText(r.Modified))}
 default:
  return firstRunDecision{Reason:"Unknown first-run conflict policy. Manual decision required; no overwrite."}
 }
}
