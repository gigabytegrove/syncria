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
