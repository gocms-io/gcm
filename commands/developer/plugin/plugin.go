package plugin

import (
	"github.com/urfave/cli"
	"github.com/gocms-io/gcm/commands/developer/plugin/plugin_copy"
	"github.com/gocms-io/gcm/commands/developer/plugin/manifest"
)

var CMD_PLUGIN = cli.Command{
	Name:  "plugin",
	Usage: "developer plugin tools for gocms",
	Subcommands: []cli.Command{
		plugin_copy.CMD_PLUGIN_COPY,
		plugin_manifest.CMD_PLUGIN_MANIFEST,
	},
}

func cmd_plugin(c *cli.Context) error {

	return nil
}
