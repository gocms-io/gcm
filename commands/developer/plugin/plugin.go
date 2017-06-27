package plugin

import (
	"fmt"
	"github.com/gocms-io/gcm/config"
	"github.com/gocms-io/gcm/utility"
	"github.com/gocms-io/gocms/utility/errors"
	"github.com/urfave/cli"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const plugin_name = "name"
const plugin_name_short = "n"
const flag_hard = "delete"
const flag_hard_short = "d"
const flag_watch = "watch"
const flag_watch_short = "w"
const flag_binary = "binary"
const flag_binary_short = "b"
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
	pluginName         string
	hardCopy           bool
	watch              bool
	buildEntry         string
	binaryName         string
	filesToCopy        []string
	run                bool
	devMode            bool
	pluginPath         string
	srcDir             string
	destDir            string
	verbose            bool
	doneChan           chan bool
	ignorePath         []string
	goBuildExec        *exec.Cmd
	watcherFileContext *utility.WatchFileContext
}

var CMD_PLUGIN = cli.Command{
	Name:      "plugin",
	Usage:     "copy plugin files from development directory into the gocms plugin directory",
	ArgsUsage: "<source> <gocms installation>",
	Action:    cmd_copy_plugin,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  plugin_name + ", " + plugin_name_short,
			Usage: "Name of the plugin. *Required",
		},
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
		cli.StringFlag{
			Name:  flag_binary + ", " + flag_binary_short,
			Usage: "Build the plugin using the following name for the output. Defaults to -n <plugin name>.",
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
		return nil
	}

	// build binary
	err = pctx.getBinaryBuildCommand()
	if err != nil {
		return nil
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
		return nil
	}

	// copy files
	pctx.copyPluginFiles()

	fmt.Printf("Build and Copy Complete - %v\n", time.Now().Format("03:04:05"))

	if pctx.run || pctx.watch {
		pctx.doneChan = make(chan bool)

		if pctx.run {
			go pctx.runGoCMS()
		}

		// if watch create watcher context and run
		if pctx.watch {
			pctx.watcherFileContext = &utility.WatchFileContext{
				Verbose:          pctx.verbose,
				SourceBase:       pctx.srcDir,
				IgnorePaths:      pctx.ignorePath,
				DoneChan:         pctx.doneChan,
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
			if !pctx.run {
				fmt.Print("Waiting for changes:\n")
			}
		}

		<-pctx.doneChan
	}

	return nil
}

func (pctx *pluginContext) onFileChangeHandler(c *utility.WatchFileContext, eventPath string) {

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
	if currentTime.Sub(fileChangedLastTime) < (time.Second * 1) {
		return
	}

	// set file change time
	c.ChangeTimeoutMap[eventPath] = currentTime

	fmt.Printf("Changes Dectected. Rebuilding Plugin - %v\n", time.Now().Format("03:04:05"))

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
	}

	err = pctx.copyPluginFiles()
	if err != nil {
		fmt.Printf("Error copying files: %v\n", err.Error())
	}

	// if we are suppose to run the new binary within gocms
	if pctx.run {
		fmt.Println("rerun reached")
		// close chanel so we can start a new one
		close(pctx.doneChan)
		pctx.doneChan = make(chan bool)
		go pctx.runGoCMS()
	} else {
		fmt.Printf("Rebuild & Copy Complete - %v\n", time.Now().Format("03:04:05"))
		fmt.Print("Waiting for changes:\n")
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
		if filepath.Base(file) != file {
			destFile = strings.Replace(file, pctx.srcDir, "", 1)
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
	fullBinPath := filepath.Join(pctx.pluginPath, pctx.binaryName)
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
	utility.StartGoCMS(pctx.destDir, pctx.devMode)
}

func (pctx *pluginContext) getBinaryBuildCommand() error {
	// build go binary exce
	pctx.pluginPath = filepath.Join(pctx.destDir, config.CONTENT_DIR, config.PLUGINS_DIR, pctx.pluginName)
	pctx.goBuildExec = exec.Command("go", "build", "-o", filepath.Join(pctx.pluginPath, pctx.binaryName), filepath.Join(pctx.srcDir, pctx.buildEntry))
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

	// verify there is a source and destination
	if !c.Args().Present() {
		errStr := "A source and destination directory must be specified."
		fmt.Println(errStr)
		return nil, errors.New(errStr)
	}

	// verify that a plugin name is given
	if c.String(plugin_name) == "" {
		errStr := "A plugin name must be specified with the --name or -n flag."
		fmt.Println(errStr)
		return nil, errors.New(errStr)
	}

	pctx := pluginContext{
		buildEntry: "main.go",
		pluginName: c.String(plugin_name),
		binaryName: c.String(plugin_name),
		devMode:    false,
		run:        false,
		watch:      false,
		srcDir:     c.Args().Get(0),
		destDir:    c.Args().Get(1),
	}

	// binary
	if c.String(flag_binary) != "" {
		pctx.binaryName = c.String(flag_binary)
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

	// verify src and dest exist
	if pctx.srcDir == "" || pctx.destDir == "" {
		errStr := "A source and destination directory must be specified."
		fmt.Println(errStr)
		return nil, errors.New(errStr)
	}

	// add default files
	pctx.filesToCopy = append(pctx.filesToCopy, filepath.Join(pctx.srcDir, config.PLUGIN_MANIFEST))
	pctx.filesToCopy = append(pctx.filesToCopy, filepath.Join(pctx.srcDir, config.PLUGIN_DOCS))

	// add additional files
	if c.StringSlice(flag_dir_file_to_copy) != nil {
		pctx.filesToCopy = append(pctx.filesToCopy, c.StringSlice(flag_dir_file_to_copy)...)
	}

	// add default ignore files
	pctx.ignorePath = append(pctx.ignorePath, []string{"vendor", ".git", "docs", ".idea", "___*", "node_modules"}...)

	// add files to ignore
	if c.StringSlice(flag_ignore_files) != nil {
		pctx.ignorePath = append(pctx.ignorePath, c.StringSlice(flag_ignore_files)...)
	}

	return &pctx, nil
}
