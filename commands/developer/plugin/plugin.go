package plugin

import (
	"fmt"
	"github.com/gocms-io/gcm/config"
	"github.com/gocms-io/gcm/models"
	"github.com/gocms-io/gcm/utility"
	"github.com/gocms-io/gocms/utility/errors"
	"github.com/urfave/cli"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const flag_hard = "delete"
const flag_hard_short = "d"
const flag_watch = "watch"
const flag_watch_short = "w"
const flag_entry = "entry"
const flag_entry_short = "e"
const flag_dir_file_to_copy = "copy"
const flag_dir_file_to_copy_short = "c"
const flag_run_gocms = "run"
const flag_run_gocms_short = "r"
const flag_gocms_dev_mode = "gocms"
const flag_gocms_dev_mode_short = "g"
const flag_ignore_files = "ignore"
const flag_ignore_files_short = "i"

type pluginContext struct {
	hardCopy                   bool
	watch                      bool
	buildEntry                 string
	filesToCopy                []string
	run                        bool
	devMode                    bool
	pluginPath                 string
	srcDir                     string
	destDir                    string
	verbose                    bool
	manifest                   *models.PluginManifest
	goCMSDoneChan              chan bool
	watcherDoneChan            chan bool
	systemDoneChan             chan bool
	ignorePath                 []string
	goBuildExec                *exec.Cmd
	watcherFileContext         *utility.WatchFileContext
	skipRunDueToFailedComplile bool
}

var CMD_PLUGIN = cli.Command{
	Name:      "plugin",
	Usage:     "copy plugin files from development directory into the gocms plugin directory",
	ArgsUsage: "<source> <gocms installation>",
	Action:    cmd_copy_plugin,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  flag_hard + ", " + flag_hard_short,
			Usage: "Delete the existing destination and replace with the contents of the source.",
		},
		cli.BoolFlag{
			Name:  flag_watch + ", " + flag_watch_short,
			Usage: "Watch for file changes in source and copy to destination on change.",
		},
		cli.StringFlag{
			Name:  flag_entry + ", " + flag_entry_short,
			Usage: "Build the plugin using the following entry point. Defaults to 'main.go'.",
		},
		cli.StringSliceFlag{
			Name:  flag_dir_file_to_copy + ", " + flag_dir_file_to_copy_short,
			Usage: "Directory or file to copy with plugin. Accepts multiple instances of the flag.",
		},
		cli.BoolFlag{
			Name:  flag_run_gocms + ", " + flag_run_gocms_short,
			Usage: "Run gocms after plugin is compiled and copied.",
		},
		cli.BoolFlag{
			Name:  flag_gocms_dev_mode + ", " + flag_gocms_dev_mode_short,
			Usage: "This option is intended for use only during gocms development. It is used to compile and run gocms from the specified destination directory.",
		},
		cli.StringSliceFlag{
			Name:  flag_ignore_files + ", " + flag_ignore_files_short,
			Usage: "Files to ignore while watching. Multiple ignore flags can be given to ignore multiple files. Ignore files are regex capable. ex: .git*",
		},
	},
}

func cmd_copy_plugin(c *cli.Context) error {

	// get command context from cli
	pctx, err := buildContextFromFlags(c)
	if err != nil {
		return err
	}

	// build binary
	err = pctx.getBinaryBuildCommand()
	if err != nil {
		return err
	}

	fmt.Printf("Starting Build and Copy - %v\n", time.Now().Format("03:04:05"))

	// to try run go generate or fail nice and continue
	err = pctx.goGenerate()
	if err != nil {
		fmt.Println("Error running go generate. Continue anyway.")
	}

	// build binary
	err = pctx.runBinaryBuildCommand()
	if err != nil {
		return err
	}

	// copy files
	pctx.copyPluginFiles()

	fmt.Printf("Build and Copy Complete - %v\n", time.Now().Format("03:04:05"))

	if pctx.run || pctx.watch {
		pctx.systemDoneChan = make(chan bool)
		if pctx.run {

			// setup gracful close to prevent port leaks
			c := make(chan os.Signal, 2)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-c
				fmt.Printf("Quiting...\n")
				close(pctx.goCMSDoneChan)
				time.Sleep(time.Second * 2)
				close(pctx.systemDoneChan)
			}()

			pctx.goCMSDoneChan = make(chan bool)
			pctx.runGoCMS()
		}

		// if watch create watcher context and run
		if pctx.watch {

			pctx.watcherDoneChan = make(chan bool)

			pctx.startFileWatcher()
		}

		// listen
		<-pctx.systemDoneChan
	}

	return nil
}

func (pctx *pluginContext) startFileWatcher() {

	pctx.watcherFileContext = &utility.WatchFileContext{
		Verbose:          pctx.verbose,
		SourceBase:       pctx.srcDir,
		IgnorePaths:      pctx.ignorePath,
		DoneChan:         pctx.watcherDoneChan,
		ChangeTimeoutMap: make(map[string]time.Time),
		Chmod:            utility.IgnoreDestination,
		Removed:          pctx.onFileChangeHandler,
		Create:           pctx.onFileChangeHandler,
		Rename:           utility.IgnoreDestination,
		Write:            pctx.onFileChangeHandler,
	}

	if pctx.devMode {
		fmt.Printf("Dev mode enabled. Waiting 5 seconds before watching files for change.\n")
		time.Sleep(time.Second * 5)
	}
	go pctx.watcherFileContext.Watch()
}

func (pctx *pluginContext) onFileChangeHandler(c *utility.WatchFileContext, eventPath string) {

	// ignore changes to "."
	if eventPath == "." || eventPath == "/" || eventPath == "./" || eventPath == "" {
		return
	}

	// ignore paths as specified
	for _, ignorePath := range c.IgnorePaths {
		ignorePathRegex, _ := regexp.Compile(ignorePath)
		if ignorePathRegex.MatchString(filepath.Clean(eventPath)) {
			return
		}
	}

	// check if file change is to soon and skip if needed
	currentTime := time.Now()
	fileChangedLastTime := c.ChangeTimeoutMap[eventPath]
	if currentTime.Sub(fileChangedLastTime) < (time.Second * 5) {
		return
	}

	// set file change time
	c.ChangeTimeoutMap[eventPath] = currentTime

	fmt.Printf("Changes Dectected in '%v'\n", eventPath)

	if pctx.run {
		if !pctx.skipRunDueToFailedComplile {
			fmt.Printf("Stopping GoCMS\n")
			close(pctx.goCMSDoneChan)
		}
	}

	fmt.Printf("Start Rebuild & Copy - %v\n", time.Now().Format("03:04:05"))

	// get binary build command
	err := pctx.getBinaryBuildCommand()
	if err != nil {
		fmt.Printf("Error getting new binary build command: %v\n", err.Error())
	}

	// run go generate
	err = pctx.goGenerate()
	if err != nil {
		fmt.Printf("Error running go generate: %v\n", err.Error())
	}

	// run binary build command
	err = pctx.runBinaryBuildCommand()
	if err != nil {
		fmt.Printf("Error running new binary build command: %v\n", err.Error())
		pctx.skipRunDueToFailedComplile = true
		return
	}

	err = pctx.copyPluginFiles()
	if err != nil {
		fmt.Printf("Error copying files: %v\n", err.Error())
	}

	// if we are suppose to run the new binary within gocms
	pctx.skipRunDueToFailedComplile = false
	if pctx.run {
		newGoCMSDonChan := make(chan bool)
		pctx.goCMSDoneChan = newGoCMSDonChan
		pctx.runGoCMS()
	} else {
		fmt.Printf("Rebuild & Copy Complete - %v\n", time.Now().Format("03:04:05"))
	}

}

func (pctx *pluginContext) copyPluginFiles() error {
	// copy files to plugin
	for _, file := range pctx.filesToCopy {

		// if file doesn't exist
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Printf("Can't copy file/directory:'%v'... Doesn't exist!\n", file)
			return err
		}

		destFile := file

		// strip leading path if it isn't just a file
		if filepath.Base(file) != file && pctx.srcDir != "." {
			destFile = strings.Replace(file, pctx.srcDir, "", 1)
			if pctx.verbose {
				fmt.Printf("compaired %v and replaced %v, with %v\n", pctx.srcDir, file, destFile)
			}
		}
		destFilePath := filepath.Join(pctx.pluginPath, destFile)
		err := utility.Copy(file, destFilePath, true, pctx.verbose)
		if err != nil {
			fmt.Printf("Error copying %v: %v\n", file, err.Error())
			return err
		}
	}

	return nil
}

func (pctx *pluginContext) runBinaryBuildCommand() error {
	fullBinPath := filepath.Join(pctx.pluginPath, pctx.manifest.Services.Bin)
	if runtime.GOOS == "windows" {
		fullBinPath = fmt.Sprintf("%v.exe", fullBinPath)
		fmt.Printf("Adding .exe for windows: %v\n", fullBinPath)
	}

	err := pctx.goBuildExec.Run()
	if err != nil {
		fmt.Printf("Error running 'go build -o %v %v': %v\n", fullBinPath, filepath.Join(pctx.srcDir, pctx.buildEntry), err.Error())
		return err
	}
	// set permissions to run
	err = os.Chmod(fullBinPath, os.FileMode(0755))
	if err != nil {
		fmt.Printf("Error setting plugin to executable: %v\n", err.Error())
		return err
	}
	return nil
}

func (pctx *pluginContext) runGoCMS() {
	fmt.Printf("Running GoCMS\n")
	go utility.StartGoCMS(pctx.destDir, pctx.devMode, pctx.goCMSDoneChan)
}

func (pctx *pluginContext) getBinaryBuildCommand() error {
	// build go binary exce
	pctx.pluginPath = filepath.Join(pctx.destDir, config.CONTENT_DIR, config.PLUGINS_DIR, pctx.manifest.Id)
	pctx.goBuildExec = exec.Command("go", "build", "-o", filepath.Join(pctx.pluginPath, pctx.manifest.Services.Bin), filepath.Join(pctx.srcDir, pctx.buildEntry))
	//if pctx.verbose {
	pctx.goBuildExec.Stdout = os.Stdout
	//}
	pctx.goBuildExec.Stderr = os.Stderr

	return nil
}

func (pctx *pluginContext) goGenerate() error {
	// run go generate
	goGenerate := exec.Command("go", "generate", filepath.Join(pctx.srcDir, pctx.buildEntry))
	if pctx.verbose {
		goGenerate.Stdout = os.Stdout
	}
	goGenerate.Stderr = os.Stderr
	err := goGenerate.Run()
	if err != nil {
		errStr := fmt.Sprintf("Error running 'go generate %v': %v\n", filepath.Join(pctx.srcDir, pctx.buildEntry), err.Error())
		fmt.Println(errStr)
		return errors.New(errStr)
	}

	return nil
}

func buildContextFromFlags(c *cli.Context) (*pluginContext, error) {

	// get src dir and dest dri
	srcDir := c.Args().Get(0)
	destDir := c.Args().Get(1)

	// verify src and dest exist
	if srcDir == "" || destDir == "" {
		errStr := "A source and destination directory must be specified."
		fmt.Println(errStr)
		return nil, errors.New(errStr)
	}

	// clean dirs
	srcDir = filepath.Clean(srcDir)
	destDir = filepath.Clean(destDir)

	// parse manifest file
	manifestPath := filepath.Join(srcDir, "manifest.json")
	manifest, err := utility.ParseManifest(manifestPath)
	if err != nil {
		fmt.Printf("Error parsing manifest file %v: %v\n", manifestPath, err.Error())
		return nil, err
	}

	pctx := pluginContext{
		buildEntry: "main.go",
		manifest:   manifest,
		devMode:    false,
		run:        false,
		watch:      false,
		srcDir:     srcDir,
		destDir:    destDir,
	}

	// entry
	if c.String(flag_entry) != "" {
		pctx.buildEntry = c.String(flag_entry)
	}

	// dev mode
	if c.Bool(flag_gocms_dev_mode) {
		pctx.devMode = true
	}

	// run
	if c.Bool(flag_run_gocms) {
		pctx.run = true
	}

	// watch
	if c.Bool(flag_watch) {
		pctx.watch = true
	}

	// verbose
	if c.GlobalBool(config.FLAG_VERBOSE) {
		pctx.verbose = true
	}

	// add default files
	pctx.filesToCopy = append(pctx.filesToCopy, filepath.Join(pctx.srcDir, config.PLUGIN_MANIFEST))

	// add docs if they exist
	if pctx.manifest.Services.Docs != "" {
		pctx.filesToCopy = append(pctx.filesToCopy, filepath.Join(pctx.srcDir, pctx.manifest.Services.Docs))
	}

	// add additional files
	if c.StringSlice(flag_dir_file_to_copy) != nil {
		pctx.filesToCopy = append(pctx.filesToCopy, c.StringSlice(flag_dir_file_to_copy)...)
	}

	// add interface files as needed
	pctx.load_interface_for_plugin()

	// add default ignore files
	pctx.ignorePath = append(pctx.ignorePath, []string{"vendor", ".git", "docs", ".idea", "___*", "node_modules"}...)

	// add files to ignore
	if c.StringSlice(flag_ignore_files) != nil {
		pctx.ignorePath = append(pctx.ignorePath, c.StringSlice(flag_ignore_files)...)
	}

	return &pctx, nil
}

func (pctx *pluginContext) load_interface_for_plugin() {

	// public
	if pctx.manifest.Interface.Public != "" {
		pctx.loadFileOrUrl(pctx.manifest.Interface.Public)
	}

	// public vendor
	if pctx.manifest.Interface.PublicVendor != "" {
		pctx.loadFileOrUrl(pctx.manifest.Interface.PublicVendor)
	}

	// public style
	if pctx.manifest.Interface.PublicStyle != "" {
		pctx.loadFileOrUrl(pctx.manifest.Interface.PublicStyle)
	}

	// admin
	if pctx.manifest.Interface.Admin != "" {
		pctx.loadFileOrUrl(pctx.manifest.Interface.Admin)
	}

	// admin vendor
	if pctx.manifest.Interface.AdminVendor != "" {
		pctx.loadFileOrUrl(pctx.manifest.Interface.AdminVendor)
	}

	// admin style
	if pctx.manifest.Interface.AdminStyle != "" {
		pctx.loadFileOrUrl(pctx.manifest.Interface.AdminStyle)
	}

	// admin goes here

}

func (pctx *pluginContext) loadFileOrUrl(path string) {
	// if file rather than request
	_, err := url.ParseRequestURI(path)
	if err != nil {
		if pctx.verbose {
			fmt.Printf("public vendor interface is a file: %v. Add it for copy.\n", path)
		}
		// add file
		pctx.filesToCopy = append(pctx.filesToCopy, filepath.Join(pctx.srcDir, config.CONTENT_DIR, path))
	} else { // skip url
		if pctx.verbose {
			fmt.Printf("public vendor interface is a url: %v. Don't copy as a file.\n", path)
		}
	}
}
