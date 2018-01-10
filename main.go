package main

import (
	"github.com/gocms-io/gcm/commands/developer"
	"github.com/gocms-io/gcm/commands/install"
	"github.com/gocms-io/gcm/commands/update"
	"github.com/gocms-io/gcm/commands/versions"
	"github.com/gocms-io/gcm/config"
	cli "github.com/urfave/cli"
	"os"
	"sort"
)

func main() {
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Name = "GoCMS Manager (gcm)"
	app.Usage = "Interface to manage all things GoCMS"
	app.HelpName = "gcm"
	app.Version = "0.0.1"
	app.Commands = []cli.Command{
		developer.CMD_DEVELOPER,
		install.CMD_INSTALL,
		update.CMD_UPDATE,
		versions.CMD_VERSIONS,
	}

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  config.FLAG_VERBOSE,
			Usage: "Enable verbose output.",
		},
		cli.StringFlag{
			Name:  config.FLAG_SET_VERSION,
			Usage: "Set the version to use for updates or install. Defaults to current.",
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		os.Exit(1)
	}
}
