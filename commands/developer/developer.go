package developer

import (
	"github.com/gocms-io/gcm/commands/developer/theme"
	"github.com/urfave/cli"
	"github.com/gocms-io/gcm/commands/developer/plugin"
)

var CMD_DEVELOPER = cli.Command{
	Name:  "developer",
	Usage: "Development tools for gocms",
	Subcommands: []cli.Command{
		theme.CMD_THEME,
		plugin.CMD_PLUGIN,
	},
}

func cmd_developer(c *cli.Context) error {

	return nil
}
