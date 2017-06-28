package utility

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type CopyContext struct {
	Source      string
	Destination string
	Verbose     bool
	Ignore      []string
}

func Copy(source string, dest string, hardCopy bool, verbose bool, ignore ...string) error {
	// clean
	source = filepath.Clean(source)
	dest = filepath.Clean(dest)

	if verbose {
		fmt.Printf("Starting to Copy: %v to %v...", source, dest)
	}

	// check source
	srcInfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	// if source is a directory start copy
	if srcInfo.IsDir() {

		err = copyDir(source, dest, hardCopy, verbose, ignore...)
		if err != nil {
			fmt.Printf(" failed!\nerror: %v\n", err.Error())
			return err
		}

		if verbose {
			fmt.Printf(" done.\n")
		}
		return nil
	}

	// get path to make if needed
	err = os.MkdirAll(filepath.Dir(dest), os.ModePerm)
	if err != nil {
		fmt.Printf(" failed!\nerror: %v\n", err.Error())
		return err
	}

	err = copyFile(source, dest, verbose)
	if err != nil {
		fmt.Printf(" failed!\nerror: %v\n", err.Error())
		return err
	}

	if verbose {
		fmt.Printf(" done.\n")
	}

	return nil
}

func copyDir(source string, dest string, hardCopy bool, verbose bool, ignore ...string) error {

	cc := CopyContext{
		Source:      source,
		Destination: dest,
		Verbose:     verbose,
		Ignore:      ignore,
	}

	// if we are trying to copy a directory into a file - fix it
	if cc.Destination[len(cc.Destination[len(cc.Destination[len(dest)-1:])-1:])-1:] != fmt.Sprintf("%c", os.PathSeparator) {
		cc.Destination = cc.Destination + fmt.Sprintf("%c", os.PathSeparator)
		cc.Destination = filepath.Clean(cc.Destination)
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

	for _, ignoreFile := range cc.Ignore {

		ignoreRegex, err := regexp.Compile(ignoreFile)
		if err != nil {
			fmt.Printf("Can't compile ignore regex %v: %v\n", ignoreFile, err.Error())
			return nil
		}
		if ignoreRegex.MatchString(src) {
			if cc.Verbose {
				fmt.Printf("ignoring regex %v and file: %v\n", ignoreRegex, src)
			}
			return nil
		}
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
	return copyFile(src, dst, cc.Verbose)

}

func copyFile(src, dst string, verbose bool) error {

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

	if verbose {
		fmt.Printf("Copied %v to %v\n", src, dst)
	}
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
