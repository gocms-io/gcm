package utility

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"path/filepath"
)

type WatchFileContext struct {
	Verbose         bool
	SourceBase      string
	DestinationBase string
	IgnorePaths     []string
	Chmod           func(c *WatchFileContext, eventPath string)
	Removed         func(c *WatchFileContext, eventPath string)
	Create          func(c *WatchFileContext, eventPath string)
	Rename          func(c *WatchFileContext, eventPath string)
	Write           func(c *WatchFileContext, eventPath string)
}

func WatchFilesForCarbonCopy(src string, dest string, ignore []string, verbose bool) {
	wf := WatchFileContext{
		Verbose:         verbose,
		SourceBase:      src,
		DestinationBase: dest,
		IgnorePaths:     ignore,
		Rename:          deleteDestination,
		Removed:         deleteDestination,
		Create:          copySourceToDestination,
		Write:           copySourceToDestination,
		Chmod:           ignoreDestination,
	}

	wf.Watch()

}

func ignoreDestination(c *WatchFileContext, eventPath string) {
}

func deleteDestination(c *WatchFileContext, eventPath string) {
	relPath, err := filepath.Rel(c.SourceBase, eventPath)
	if err != nil {
		fmt.Printf("Error calculating path for copy of %v: %v\n", eventPath, err.Error())
	}

	// get dest
	dest := filepath.Join(c.DestinationBase, relPath)

	err = os.RemoveAll(dest)
	if err != nil {
		fmt.Printf("Error deleting %v: %v\n", eventPath, err.Error())
	} else {
		fmt.Printf("Removed %v\n", eventPath)
	}
}

func copySourceToDestination(c *WatchFileContext, eventPath string) {
	relPath, err := filepath.Rel(c.SourceBase, eventPath)
	if err != nil {
		fmt.Printf("Error calculating path for copy of %v: %v\n", eventPath, err.Error())
	}

	// get dest
	dest := filepath.Join(c.DestinationBase, relPath)
	err = Copy(eventPath, dest, true, c.Verbose)
	if err != nil {
		fmt.Printf("Error copying %v: %v\n", eventPath, err.Error())
	}
}

func (c *WatchFileContext) Watch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if filepath.Ext(event.Name) != ".DS_Store" {

					if event.Op&fsnotify.Write == fsnotify.Write {
						c.Write(c, event.Name)
					} else if event.Op&fsnotify.Remove == fsnotify.Remove {
						c.Removed(c, event.Name)
					} else if event.Op&fsnotify.Create == fsnotify.Create {
						c.Create(c, event.Name)
					} else if event.Op&fsnotify.Rename == fsnotify.Rename {
						c.Rename(c, event.Name)
					} else if event.Op&fsnotify.Chmod == fsnotify.Chmod {
						c.Chmod(c, event.Name)
					}
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	// get all paths under path
	err = filepath.Walk(c.SourceBase, func(path string, info os.FileInfo, err error) error {

		// check ignore
		for _, ignorePath := range c.IgnorePaths {
			cleanIgnorePath := filepath.Clean(ignorePath)
			cleanPath := filepath.Clean(path)
			//fmt.Printf("Compare IP: %v, %v\n", cleanIgnorePath, cleanPath)
			// ignore files
			if cleanIgnorePath == cleanPath {
				if c.Verbose {
					fmt.Printf("ignoring: %v\n", cleanIgnorePath)
				}
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			// ignore directories
		}

		if err != nil {
			return err
		}

		if info.IsDir() {
			err = watcher.Add(path)
			if err != nil {
				log.Fatal(err)
			} else {
				//fmt.Printf("Watching %v\n", path)
			}
		}

		return nil
	})

	<-done
}
