package utility_os

import (
	"os/exec"
	"syscall"
)

func SetChildProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}

func Kill_process(cmd *exec.Cmd) {
	cmd.Process.Kill()
}