package utility

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"path/filepath"
)

type WatchFileContext struct {
	SourceBase      string
	DestinationBase string
	Chmod           func(c *WatchFileContext, eventPath string)
	Removed         func(c *WatchFileContext, eventPath string)
	Create          func(c *WatchFileContext, eventPath string)
	Rename          func(c *WatchFileContext, eventPath string)
	Write           func(c *WatchFileContext, eventPath string)
}

func WatchFilesForCarbonCopy(src string, dest string) {
	wf := WatchFileContext{
		SourceBase:      src,
		DestinationBase: dest,
		Rename:          deleteDestination,
		Removed:         deleteDestination,
		Create:          copySourceToDestination,
		Write:           copySourceToDestination,
		Chmod:           ignoreDestination,
	}

	watch(&wf)

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
	err = Copy(eventPath, dest, true)
	if err != nil {
		fmt.Printf("Error copying %v: %v\n", eventPath, err.Error())
	} else {
		fmt.Printf("Copied %v\n", eventPath)
	}
}

func watch(c *WatchFileContext) {
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
