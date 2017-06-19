package update

import (
	"fmt"
	"github.com/gocms-io/gcm/commands/install"
	"github.com/gocms-io/gcm/config"
	"github.com/gocms-io/gcm/utility"
	"github.com/urfave/cli"
	"os"
	"path/filepath"
)

var CMD_UPDATE = cli.Command{
	Name:      "update",
	Usage:     "Update gocms - The .env file will be preserved",
	ArgsUsage: "<directory>",
	Action:    cmd_update,
}

func cmd_update(c *cli.Context) error {

	if !c.Args().Present() {
		fmt.Println("An install directory must be specified.")
		return nil
	}

	fph := utility.NewFilePathHelper(c.Args().First())

	// verify that this is a gocms install
	if _, err := os.Stat(fph.AddWorkingDirPath(config.BINARY_FILE)); os.IsNotExist(err) {
		fmt.Println("The provided directory doesn't appear to be an active GoCMS installation.")
		return nil
	}

	// create backup dir
	err := os.Mkdir(fph.AddBackupDirPath(""), os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating backup directory: %v\n", err.Error())
		return nil
	}

	// move current files to backup dir
	fmt.Print("Backing up current installation...")
	// gocms
	_ = fph.WorkingToBackup(config.BINARY_FILE)
	// content
	_ = fph.WorkingToBackup(config.CONTENT_DIR)
	// .env
	_ = fph.WorkingToBackup(config.ENV_FILE)
	fmt.Print("complete!\n")

	// do basic install and rollback on error
	err = install.BasicInstall(fph.AddStagingDirPath(""))
	if err != nil {
		// roll back update
		fmt.Print("Rolling back changes...")
		// gocms
		_ = fph.BackupToWorking(config.BINARY_FILE)
		// content
		_ = fph.BackupToWorking(config.CONTENT_DIR)
		// .env
		_ = fph.BackupToWorking(config.ENV_FILE)
		fmt.Print("complete!\n")
		return nil
	}

	// merge backup and staging into installation dir
	fmt.Print("Applying update...")
	// gocms
	_ = fph.StagingToWorking(config.BINARY_FILE)
	// .env
	_ = fph.BackupToWorking(config.ENV_FILE)
	// make content
	_ = os.Mkdir(fph.AddWorkingDirPath(config.CONTENT_DIR), os.ModePerm)
	// content/docs
	_ = fph.StagingToWorking(filepath.Join(config.CONTENT_DIR, config.DOCS_DIR))
	// content/gocms (admin)
	_ = fph.StagingToWorking(filepath.Join(config.CONTENT_DIR, config.GOCMS_ADMIN_DIR))
	// content/templates
	_ = fph.StagingToWorking(filepath.Join(config.CONTENT_DIR, config.TEMPLATES_DIR))
	// plugins
	_ = fph.BackupToWorking(filepath.Join(config.CONTENT_DIR, config.PLUGINS_DIR))
	// themes
	_ = fph.BackupToWorking(filepath.Join(config.CONTENT_DIR, config.THEMES_DIR))
	// remove default theme and add new

	_ = os.Remove(fph.AddWorkingDirPath(filepath.Join(config.CONTENT_DIR, config.THEMES_DIR, config.THEMES_DEFAULT_DIR)))
	_ = fph.StagingToWorking(filepath.Join(config.CONTENT_DIR, config.THEMES_DIR, config.THEMES_DEFAULT_DIR))
	fmt.Print("complete!\n")

	fmt.Println("GoCMS Installed Updated!")

	return nil
}
