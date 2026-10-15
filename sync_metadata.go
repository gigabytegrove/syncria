package main

import "strings"

// Syncria never synchronizes metadata sidecars owned by another tool or OS.
// The excluded files remain untouched wherever they already exist.
func ignoredSyncMetadata(name string) bool {
 switch strings.ToLower(name) {
 case "sync.ffs_db", "thumbs.db", "ehthumbs.db", "ehthumbs_vista.db":
  return true
 default:
  return false
 }
}

// Ignore excluded metadata files at any depth, including old checkpoints.
// Changing the ignore policy must NEVER propagate deletion of a previously
// tracked metadata file to the cloud or NAS.
func ignoredSyncPath(rel string) bool {
 if rel=="" {return false}
 normalized:=strings.ReplaceAll(rel,"\\","/")
 return ignoredSyncMetadata(normalized[strings.LastIndex(normalized,"/")+1:])
}
