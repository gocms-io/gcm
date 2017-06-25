package plugin

import (
	"fmt"
	"github.com/gocms-io/gcm/config"
	"github.com/gocms-io/gcm/utility"
	"github.com/urfave/cli"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	// build go binary
	contentPath := filepath.Join(destDir, config.CONTENT_DIR, config.PLUGINS_DIR)
	pluginPath := filepath.Join(contentPath, pluginName)
	pluginBinaryPath := filepath.Join(pluginPath, binaryName)
	goBuild := exec.Command("go", "build", "-o", pluginBinaryPath, filepath.Join(srcDir, entryPoint))
	if c.GlobalBool(config.FLAG_VERBOSE) {
		goBuild.Stdout = os.Stdout
	}
	goBuild.Stderr = os.Stderr

	err = goBuild.Run()
	if err != nil {
		fmt.Printf("Error running 'go build -o %v %v': %v\n", pluginBinaryPath, filepath.Join(srcDir, entryPoint), err.Error())
		return nil
	}
	// set permissions to run
	err = os.Chmod(pluginBinaryPath, os.FileMode(0755))
	if err != nil {
		fmt.Printf("Error setting plugin to executable: %v\n", err.Error())
	}

	if c.StringSlice(flag_dir_file_to_copy) != nil {
		filesToCopy = append(filesToCopy, c.StringSlice(flag_dir_file_to_copy)...)
	}

	done := make(chan bool)

	copyPluginFiles(filesToCopy, pluginPath, srcDir, c.GlobalBool(config.FLAG_VERBOSE))

	if watch {
		wfc := utility.WatchFileContext{
			Verbose:     c.GlobalBool(config.FLAG_VERBOSE),
			SourceBase:  srcDir,
			IgnorePaths: []string{"vendor", ".git", "docs"},
			Chmod:       buildCopyAndRun,
			Removed:     buildCopyAndRun,
			Create:      buildCopyAndRun,
			Rename:      buildCopyAndRun,
			Write:       buildCopyAndRun,
		}

		wfc.Watch()
	}

	if runGoCMS {

		fmt.Printf("Running GoCMS\n")
		utility.StartGoCMS(destDir, goCMSDevMode)
		<-done
	}

	return nil
}

func buildCopyAndRun(c *utility.WatchFileContext, eventPath string) {
	fmt.Printf("buildCopyAndRun called by file: %v\n", eventPath)
}

func copyPluginFiles(filesToCopy []string, pluginPath string, srcDir string, verbose bool) {
	// copy files to plugin
	for _, file := range filesToCopy {

		// if file doesn't exist
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Printf("Can't copy file/directory:'%v'... Doesn't exist!\n", file)
			break
		}

		destFile := file

		// strip leading path if it isn't just a file
		if filepath.Base(file) != file {
			destFile = strings.Replace(file, srcDir, "", 1)
		}
		destFilePath := filepath.Join(pluginPath, destFile)
		err := utility.Copy(file, destFilePath, true, verbose)
		if err != nil {
			fmt.Printf("Error copying %v: %v\n", file, err.Error())
		}
	}
}
