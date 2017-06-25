package plugin

import (
	"fmt"
	"github.com/gocms-io/gcm/config"
	"github.com/gocms-io/gcm/utility"
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

type pluginWatcherContext struct {
	filesToCopy          []string
	pluginPath           string
	pluginBinaryPath     string
	srcDir               string
	destDir              string
	entryPoint           string
	goCMSDevelopmentMode bool
	verbose              bool
	doneChan             chan bool
	goBuildExec          *exec.Cmd
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
	},
}

func cmd_copy_plugin(c *cli.Context) error {
	entryPoint := "main.go"
	// verify there is a source and destination
	if !c.Args().Present() {
		fmt.Println("A source and destination directory must be specified.")
		return nil
	}

	// verify that a plugin name is given
	if c.String(plugin_name) == "" {
		fmt.Println("A plugin name must be specified with the --name or -n flag.")
		return nil
	}
	pluginName := c.String(plugin_name)
	binaryName := pluginName
	goCMSDevMode := false
	runGoCMS := false
	watch := false

	srcDir := c.Args().Get(0)
	destDir := c.Args().Get(1)

	if srcDir == "" || destDir == "" {
		fmt.Println("A source and destination directory must be specified.")
		return nil
	}

	// binary
	if c.String(flag_binary) != "" {
		binaryName = c.String(flag_binary)
	}

	// entry
	if c.String(flag_entry) != "" {
		entryPoint = c.String(flag_entry)
	}

	// dev mode
	if c.Bool(flag_gocms_dev_mode) {
		goCMSDevMode = true
	}

	// run
	if c.Bool(flag_run_gocms) {
		runGoCMS = true
	}

	// watch
	if c.Bool(flag_watch) {
		watch = true
	}

	var filesToCopy []string
	filesToCopy = append(filesToCopy, filepath.Join(srcDir, config.PLUGIN_MANIFEST))
	filesToCopy = append(filesToCopy, filepath.Join(srcDir, config.PLUGIN_DOCS))

	// run go generate
	goGenerate := exec.Command("go", "generate", filepath.Join(srcDir, entryPoint))
	if c.GlobalBool(config.FLAG_VERBOSE) {
		goGenerate.Stdout = os.Stdout
	}
	goGenerate.Stderr = os.Stderr
	err := goGenerate.Run()
	if err != nil {
		fmt.Printf("Error running 'go generate %v': %v\n", filepath.Join(srcDir, entryPoint), err.Error())
		return nil
	}

	// build go binary exce
	contentPath := filepath.Join(destDir, config.CONTENT_DIR, config.PLUGINS_DIR)
	pluginPath := filepath.Join(contentPath, pluginName)
	pluginBinaryPath := filepath.Join(pluginPath, binaryName)
	goBuild := exec.Command("go", "build", "-o", pluginBinaryPath, filepath.Join(srcDir, entryPoint))
	if c.GlobalBool(config.FLAG_VERBOSE) {
		goBuild.Stdout = os.Stdout
	}
	goBuild.Stderr = os.Stderr

	done := make(chan bool)

	pc := pluginWatcherContext{
		srcDir:               srcDir,
		filesToCopy:          filesToCopy,
		pluginPath:           pluginPath,
		pluginBinaryPath:     pluginBinaryPath,
		verbose:              c.GlobalBool(config.FLAG_VERBOSE),
		goCMSDevelopmentMode: goCMSDevMode,
		destDir:              destDir,
		doneChan:             done,
		goBuildExec:          goBuild,
		entryPoint:           entryPoint,
	}

	err = pc.buildPluginBinary()
	if err != nil {
		return nil
	}

	if c.StringSlice(flag_dir_file_to_copy) != nil {
		filesToCopy = append(filesToCopy, c.StringSlice(flag_dir_file_to_copy)...)
	}

	pc.copyPluginFiles()

	if watch {
		wfc := utility.WatchFileContext{
			Verbose:          c.GlobalBool(config.FLAG_VERBOSE),
			SourceBase:       srcDir,
			DoneChan:         done,
			IgnorePaths:      []string{"vendor", ".git", "docs", ".idea", "___*"},
			ChangeTimeoutMap: make(map[string]time.Time),
			Chmod:            utility.IgnoreDestination,
			Removed:          pc.buildCopyAndRun,
			Create:           pc.buildCopyAndRun,
			Rename:           utility.IgnoreDestination,
			Write:            pc.buildCopyAndRun,
		}

		go wfc.Watch()
	}

	if runGoCMS {
		//go pc.runGoCMS()
		<-done
	}

	if runGoCMS || watch {
		//<-done
	}

	return nil
}

func (pc *pluginWatcherContext) runGoCMS() {
	fmt.Printf("Running GoCMS\n")
	utility.StartGoCMS(pc.destDir, pc.goCMSDevelopmentMode)
}

func (pc *pluginWatcherContext) buildCopyAndRun(c *utility.WatchFileContext, eventPath string) {

	for _, ignorePath := range c.IgnorePaths {
		ignorePathRegex, _ := regexp.Compile(ignorePath)
		if ignorePathRegex.MatchString(filepath.Clean(eventPath)) {
			return
		}
	}

	currentTime := time.Now()
	fileChangedLastTime := c.ChangeTimeoutMap[eventPath]

	// if file change is to soon
	if currentTime.Sub(fileChangedLastTime) < (time.Second * 1) {
		return
	}
	// probably need to copy, assign, then save
	c.ChangeTimeoutMap[eventPath] = currentTime
	fmt.Printf("Changes Dectected. Rebuilding Plugin & Restarting GoCMS\n")
	close(pc.doneChan)

	// if changes are in path of copy files copy and restart
	//pc.copyPluginFiles()

}

func (pc *pluginWatcherContext) copyPluginFiles() {
	// copy files to plugin
	for _, file := range pc.filesToCopy {

		// if file doesn't exist
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Printf("Can't copy file/directory:'%v'... Doesn't exist!\n", file)
			break
		}

		destFile := file

		// strip leading path if it isn't just a file
		if filepath.Base(file) != file {
			destFile = strings.Replace(file, pc.srcDir, "", 1)
		}
		destFilePath := filepath.Join(pc.pluginPath, destFile)
		err := utility.Copy(file, destFilePath, true, pc.verbose)
		if err != nil {
			fmt.Printf("Error copying %v: %v\n", file, err.Error())
		}
	}
}

func (pc *pluginWatcherContext) buildPluginBinary() error {
	err := pc.goBuildExec.Run()
	if err != nil {
		fmt.Printf("Error running 'go build -o %v %v': %v\n", pc.pluginBinaryPath, filepath.Join(pc.srcDir, pc.entryPoint), err.Error())
		return err
	}
	// set permissions to run
	err = os.Chmod(pc.pluginBinaryPath, os.FileMode(0755))
	if err != nil {
		fmt.Printf("Error setting plugin to executable: %v\n", err.Error())
		return err
	}
	return nil
}
