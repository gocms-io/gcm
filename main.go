package main

import (
	"github.com/gocms-io/gcm/commands/install"
	cli "github.com/urfave/cli"
	"os"
)

func main() {
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Name = "GoCMS Manager (gcm)"
	app.Usage = "Interface to manage all things GoCMS"
	app.HelpName = "gcm"
	app.Version = "0.0.1"
	app.Commands = []cli.Command{
		install.CMD_INSTALL,
	}

	app.Run(os.Args)
}
