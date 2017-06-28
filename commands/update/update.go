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
	Usage:     "Update gocms - The .env file will be preserved. Updates apply to the current working directory.",
	ArgsUsage: "<directory>",
	Action:    cmd_update,
}

type updatePluginContext struct {
	backupDir  string
	stagingDir string
	installDir string
	verbose    bool
}

func cmd_update(c *cli.Context) error {

	//if !c.Args().Present() {
	//	fmt.Println("An install directory must be specified.")
	//	return nil
	//}

	installDir, _ := filepath.Abs(".")

	// verify this is a gocms dir
	if _, err := os.Stat(filepath.Join(installDir, config.BINARY_FILE)); os.IsNotExist(err) {
		fmt.Println("The provided directory doesn't appear to be an active GoCMS installation.")
		return nil
	}

	uctx := updatePluginContext{
		backupDir:  filepath.Join(installDir, config.BACKUP_DIR),
		stagingDir: filepath.Join(installDir, config.STAGING_DIR),
		installDir: installDir,
		verbose:    c.GlobalBool(config.FLAG_VERBOSE),
	}

	// copy current install to backup
	err := utility.Copy(uctx.installDir, uctx.backupDir, true, uctx.verbose, "\\.bk", ".bk.*")
	if err != nil {
		return nil
	}

	// create staging dir
	err = os.Mkdir(uctx.stagingDir, os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating staging directory %v: %v\n", uctx.stagingDir, err.Error())
		return nil
	}

	// do basic install and rollback on error
	err = install.BasicInstall(uctx.stagingDir)
	if err != nil {
		// roll back update
		fmt.Print("Rolling back changes...")
		os.RemoveAll(uctx.backupDir)
		os.RemoveAll(uctx.installDir)
		return nil
	}

	// merge backup and staging into installation dir
	fmt.Print("Applying update to staging...")

	// .env
	err = os.Rename(filepath.Join(uctx.backupDir, config.ENV_FILE), filepath.Join(uctx.stagingDir, config.ENV_FILE))
	if err != nil {
		fmt.Printf("Error applying .env file: %v\n", err.Error())
		uctx.rollback()
		return nil
	}

	// plugins
	err = utility.ForceRename(filepath.Join(uctx.backupDir, config.CONTENT_DIR, config.PLUGINS_DIR), filepath.Join(uctx.stagingDir, config.CONTENT_DIR, config.PLUGINS_DIR))
	if err != nil {
		fmt.Printf("Error applying plugins file: %v\n", err.Error())
		uctx.rollback()
		return nil
	}

	// docs
	// remove these default dir first
	_ = os.RemoveAll(filepath.Join(uctx.backupDir, config.CONTENT_DIR, config.THEMES_DIR, config.THEMES_DEFAULT_DIR))
	err = utility.ForceRename(filepath.Join(uctx.backupDir, config.CONTENT_DIR, config.THEMES_DIR), filepath.Join(uctx.stagingDir, config.CONTENT_DIR, config.THEMES_DIR))
	if err != nil {
		fmt.Printf("Error applying themes file: %v\n", err.Error())
		uctx.rollback()
		return nil
	}

	// move everything into production
	fmt.Printf("Moving staging into production\n")
	err = utility.Copy(uctx.stagingDir, uctx.installDir, false, uctx.verbose)
	if err != nil {
		fmt.Printf("Erorr moving staging into production: %v\n", err.Error())
		uctx.rollback()
		return nil
	}

	// clean up
	fmt.Println("Cleaning up temp files")
	err = os.RemoveAll(uctx.backupDir)
	if err != nil {
		fmt.Printf("Error removing backup: %v\n", err.Error())
	}
	err = os.RemoveAll(uctx.stagingDir)
	if err != nil {
		fmt.Printf("Error removing staging: %v\n", err.Error())
	}

	fmt.Println("GoCMS Installed Updated!\n")

	return nil
}

func (uctx *updatePluginContext) rollback() {
	fmt.Print("Rolling back changes...")
	err := utility.Copy(uctx.backupDir, uctx.installDir, false, uctx.verbose)
	if err != nil {
		fmt.Printf("Error moving backup into production: %v\n", err.Error())
	}
	err = os.RemoveAll(uctx.backupDir)
	if err != nil {
		fmt.Printf("Error removing backup: %v\n", err.Error())
	}
	err = os.RemoveAll(uctx.stagingDir)
	if err != nil {
		fmt.Printf("Error removing staging: %v\n", err.Error())
	}
	fmt.Print("complete!\n")
}
