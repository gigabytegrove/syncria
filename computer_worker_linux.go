package main

import (
 "os/exec"
 "syscall"
)

// If the Syncria hub is restarted, its child transfer workers must not
// continue independently and delete a source after an unsupervised move.
func configureTransferWorker(cmd *exec.Cmd){
 cmd.SysProcAttr=&syscall.SysProcAttr{Pdeathsig:syscall.SIGKILL}
}
