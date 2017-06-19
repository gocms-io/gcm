package utility

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type CopyContext struct {
	Source      string
	Destination string
}

func Copy(source string, dest string, hardCopy bool) error {

	// check source
	srcInfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	// if source is a directory start copy
	if srcInfo.IsDir() {
		return copyDir(source, dest, hardCopy)
	}

	// get path to make if needed
	err = os.MkdirAll(filepath.Dir(dest), os.ModePerm)
	if err != nil {
		return err
	}
	// if we are trying to copy a file into a directory - fix it
	if dest[len(dest)-1:] == fmt.Sprintf("%c", os.PathSeparator) {
		dest = filepath.Join(dest, srcInfo.Name())
	}
	return copyFile(source, dest)
}

func copyDir(source string, dest string, hardCopy bool) error {

	cc := CopyContext{
		Source:      source,
		Destination: dest,
	}

	// if we are trying to copy a directory into a file - fix it
	if dest[len(dest)-1:] != fmt.Sprintf("%c", os.PathSeparator) {
		dest = dest + fmt.Sprintf("%c", os.PathSeparator)
		cc.Destination = dest
	}

	// if we are doing a hard copy we must delete the dest contents first
	if hardCopy {
		_ = os.RemoveAll(cc.Destination)
	}

	// create if doesn't exist
	err := os.MkdirAll(cc.Destination, os.ModePerm)
	if err != nil {
		return err
	}

	err = filepath.Walk(cc.Source, cc.copyDirWalk)
	if err != nil {
		return err
	}

	return nil
}

func (cc *CopyContext) copyDirWalk(src string, info os.FileInfo, err error) error {
	dst := strings.Replace(src, cc.Source, cc.Destination, 1)

	if err != nil {
		return err
	}

	// if dir that doesn't exist create it
	if info.IsDir() {
		err = os.MkdirAll(dst, os.ModePerm)
		if err != nil {
			return err
		}
		return nil
	}

	// otherwise copy file
	return copyFile(src, dst)

}

func copyFile(src, dst string) error {

	sfi, err := os.Stat(src)
	if err != nil {
		return err
	}

	// copy file contents
	err = copyFileContents(src, dst)

	if err != nil {
		return err
	}

	// try to set same perms
	err = os.Chmod(dst, sfi.Mode().Perm())
	if err != nil {
		return err
	}

	fmt.Printf("Copied %v\n", src)
	return nil
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}
