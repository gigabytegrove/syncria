package main
import "errors"
var (
 errSyncMustBePaused=errors.New("Pause automatic sync before editing; nothing has been changed")
 errSyncTypeImmutable=errors.New("Changing a cloud relationship into a NAS relationship (or the reverse) requires a new sync")
 errSyncPendingRecovery=errors.New("Restore or finish pending deletion recovery before changing folders")
 errSyncResetRequired=errors.New("Changing source or destination folders resets the comparison baseline. Tick the acknowledgment and review a new dry run")
)
