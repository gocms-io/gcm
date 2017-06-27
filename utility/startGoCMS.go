package utility

import (
	"bufio"
	"fmt"
	"github.com/gocms-io/gcm/config/config_os"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// start goCMS
func StartGoCMS(destDir string, goCMSDevMode bool, doneChan chan bool) {

	// if dev mode first build gocms
	if goCMSDevMode {
		// guild gocms first
		goCMSBuildCMD := exec.Command("go", "build", "-o", config_os.BINARY_FILE, "main.go")
		goCMSBuildCMD.Dir = destDir
		out, err := goCMSBuildCMD.CombinedOutput()
		if err != nil {
			fmt.Printf("Error building gocms: %v\n", err.Error())
		}
		fmt.Printf("GOCMS Build Output: %v\n ", out)
	}

	// build command
	var cmd *exec.Cmd
	commandString := filepath.FromSlash("./" + config_os.BINARY_FILE)
	cmd = exec.Command(commandString)
	cmd.Dir = destDir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// set stdout to pipe
	cmdStdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
		os.Exit(0)
	}

	// setup stdout to scan continuously
	stdOutScanner := bufio.NewScanner(cmdStdoutReader)
	go func() {
		for stdOutScanner.Scan() {
			fmt.Printf("%s\n", stdOutScanner.Text())
		}
	}()

	// set stderr to pipe
	cmdStderrReader, err := cmd.StderrPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
		os.Exit(0)
	}

	// setup stderr to scan continuously
	stdErrScanner := bufio.NewScanner(cmdStderrReader)
	go func() {
		for stdErrScanner.Scan() {
			fmt.Printf("%s\n", stdErrScanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		fmt.Printf("Error starting gocms: %v\n", err.Error())
		os.Exit(0)
	}

	select {
	case <-doneChan:
		//cmd.Process.Kill()
		syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	}
}
