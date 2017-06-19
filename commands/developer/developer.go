package developer

import (
	"github.com/gocms-io/gcm/commands/developer/theme"
	"github.com/urfave/cli"
)

var CMD_DEVELOPER = cli.Command{
	Name:  "developer",
	Usage: "Development tools for gocms",
	Subcommands: []cli.Command{
		theme.CMD_THEME,
	},
}

func cmd_developer(c *cli.Context) error {

	return nil
}
