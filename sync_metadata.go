package main

import "strings"

// FreeFileSync uses sync.ffs_db to record its own file database. It is not
// user content and multiple copies often exist in Google Drive with the same
// name. Replicating it between devices is unsafe: it can corrupt the
// other program's tracking state and duplicate names cannot map to one NAS
// pathname. Exclude this exact sidecar, preserving every copy at its source.
func ignoredSyncMetadata(name string) bool {
 return strings.EqualFold(name,"sync.ffs_db")
}

// Ignore an FFS database at any depth, including entries retained in
// checkpoints from an earlier Syncria version. Removing its checkpoint
// must never be interpreted as a remote/local deletion instruction.
func ignoredSyncPath(rel string) bool {
 if rel=="" {return false}
 return ignoredSyncMetadata(rel[strings.LastIndex(rel,"/")+1:])
}
