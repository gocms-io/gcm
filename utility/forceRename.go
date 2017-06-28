package utility

import (
	"fmt"
	"os"
)

func ForceRename(src string, dest string) error {

	err := os.RemoveAll(dest)
	if err != nil {
		fmt.Printf("Error force renaming %v. Can't remove: %v\n", dest, err.Error())
		return err
	}

	err = os.Rename(src, dest)
	if err != nil {
		fmt.Printf("Error force renaming %v. Can't move: %v\n", src, err.Error())
	}

	return nil
}
