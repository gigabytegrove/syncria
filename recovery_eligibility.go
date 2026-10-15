package main

import "time"

// eligiblePendingDeletions returns only recoveries whose ten-minute Undo
// deadline is still strictly in the future. Never mutate the persisted
// PendingDeletions: expired staged files remain recorded so the background
// purge worker can retry cleanup after an SMB/Drive outage.
func eligiblePendingDeletions(entries []PendingDeletion, now time.Time) []PendingDeletion {
 if len(entries)==0{return nil}
 var eligible []PendingDeletion
 for _,entry:=range entries{
  if entry.Expires.After(now){eligible=append(eligible,entry)}
 }
 return eligible
}
