package utility

import (
	"os"
	"fmt"
)

// start goCMS
func StartGoCMS(destDir string, binaryCommand string) {
	done := make(chan bool)

	// change directory into gocms dir and run from that context
	err := os.Chdir(destDir)
	if err != nil {
		fmt.Printf("Error changing directory to %v: %v\n", destDir, err.Error())
		os.Exit(0)
	}
	<- done
}