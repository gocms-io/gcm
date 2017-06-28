package install

import (
	"fmt"
	"github.com/gocms-io/gcm/config"
	"github.com/gocms-io/gcm/utility"
	"github.com/urfave/cli"
	"os"
	"path"
	"path/filepath"
)

var CMD_INSTALL = cli.Command{
	Name:      "install",
	Usage:     "Install gocms",
	ArgsUsage: "<directory>",
	Action:    cmd_install,
}

func cmd_install(c *cli.Context) error {

	if !c.Args().Present() {
		fmt.Println("An install directory must be specified.")
		return nil
	}

	versionToUse := config.BINARY_DEFAULT_VERSION
	if c.GlobalString(config.FLAG_SET_VERSION) != "" {
		versionToUse = c.GlobalString(config.FLAG_SET_VERSION)
	}

	err := BasicInstall(c.Args().First(), versionToUse)
	if err != nil {
		return nil
	}

	fmt.Println("GoCMS Installed Successfully!")

	return nil
}

func BasicInstall(installPath string, versionToUse string) error {

	// download file
	downloadPath := path.Clean(installPath)
	downloadLocation := fmt.Sprintf("%v/%v", downloadPath, config.BINARY_ARCHIVE)
	downloadLocation = filepath.FromSlash(downloadLocation)
	urlLocation := fmt.Sprintf("%v://%v.%v/%v/%v/%v/%v", config.BINARY_PROTOCOL, config.BINARY_HOST, config.BINARY_DOMAIN, config.BINARY_DEFAULT_RELEASE, versionToUse, config.BINARY_OS_PATH, config.BINARY_ARCHIVE)
	fmt.Printf("Downloading: %v...\n", urlLocation)
	err := utility.DownloadFile(downloadLocation, urlLocation)
	if err != nil {
		fmt.Printf("Error downloading GoCMS package: %v\n", err.Error())
		fmt.Printf("cleaning up files at %v\n", downloadLocation)
		_ = os.Remove(downloadLocation)
		return nil
	}

	// unzip file
	fmt.Printf("Unpacking %v to %v\n", downloadLocation, downloadPath)
	err = utility.Unzip(downloadLocation, downloadPath)
	if err != nil {
		fmt.Printf("Error unpacking GoCMS package: %v\n", err.Error())
		fmt.Printf("cleaning up files at %v\n", downloadLocation)
		_ = os.Remove(downloadLocation)
		return err
	}

	// clean up zip file
	fmt.Printf("Cleaning up files at %v\n", downloadPath)
	_ = os.Remove(downloadLocation)

	return nil
}
