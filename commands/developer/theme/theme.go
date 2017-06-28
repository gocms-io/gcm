package theme

import (
	"fmt"
	"github.com/gocms-io/gcm/config"
	"github.com/gocms-io/gcm/utility"
	"github.com/urfave/cli"
	"path/filepath"
)

const theme_name = "name"
const theme_name_short = "n"
const flag_hard = "delete"
const flag_hard_short = "d"
const flag_watch = "watch"
const flag_watch_short = "w"

var CMD_THEME = cli.Command{
	Name:      "theme",
	Usage:     "copy theme files from development directory into the gocms themes directory",
	ArgsUsage: "<source> <gocms installation>",
	Action:    cmd_copy_theme,
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
			Name:  theme_name + ", " + theme_name_short,
			Usage: "Name of the theme. *Required",
		},
	},
}

func cmd_copy_theme(c *cli.Context) error {

	// verify there is a source and destination
	if !c.Args().Present() {
		fmt.Println("A source and destination directory must be specified.")
		return nil
	}

	srcDir := c.Args().Get(0)
	destDir := c.Args().Get(1)

	if srcDir == "" || destDir == "" {
		fmt.Println("A source and destination directory must be specified.")
	}

	// verify that a plugin name is given
	if c.String(theme_name) == "" {
		fmt.Println("A plugin name must be specified with the --name or -n flag.")
		return nil
	}
	themeName := c.String(theme_name)
	themeDirPath := filepath.Join(destDir, config.CONTENT_DIR, config.THEMES_DIR, themeName)

	err := utility.Copy(filepath.Clean(srcDir), themeDirPath, c.Bool(flag_hard), c.GlobalBool(config.FLAG_VERBOSE))
	if err != nil {
		fmt.Printf("Error copying theme dir: %v\n", err.Error())
		return nil
	}

	if c.Bool(flag_watch) {
		fmt.Println("Watching source directory for changes...")
		utility.WatchFilesForCarbonCopy(srcDir, themeDirPath, nil, c.GlobalBool(config.FLAG_VERBOSE))
	}

	return nil
}
