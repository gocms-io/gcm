package utility_os

import (
	"os/exec"
	"syscall"
)

func SetChildProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func Kill_process(cmd *exec.Cmd) {
	syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
}