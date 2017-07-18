package utility_os

import (
	"os/exec"
	"syscall"
)

func SetChildProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func Kill_process(cmd *exec.Cmd) {
	cmd.Process.Kill()
}