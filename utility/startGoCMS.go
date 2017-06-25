package utility

import (
	"bufio"
	"fmt"
	"github.com/gocms-io/gcm/config/config_os"
	"os"
	"os/exec"
)

// start goCMS
func StartGoCMS(destDir string, goCMSDevMode bool) {

	// change directory into gocms dir and run from that context
	err := os.Chdir(destDir)
	if err != nil {
		fmt.Printf("Error changing directory to %v: %v\n", destDir, err.Error())
		os.Exit(0)
	}

	// build command

	var cmd *exec.Cmd
	if goCMSDevMode {
		cmd = exec.Command("go", "run", "main.go")
	} else {
		cmd = exec.Command(config_os.BINARY_FILE)
	}

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
}
