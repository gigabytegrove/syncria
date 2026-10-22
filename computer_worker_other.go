//go:build !linux

package main

import "os/exec"

// Other platforms use CommandContext cancellation. Linux also enforces
// parent-death behavior at the kernel level.
func configureTransferWorker(cmd *exec.Cmd){}
