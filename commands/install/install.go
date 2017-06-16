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
	Usage:     "Install gocms. Defaults to current directory unless one is provided.",
	ArgsUsage: "<directory>",
	Action:    cmd_install,
}

func cmd_install(c *cli.Context) error {

	if !c.Args().Present() {
		fmt.Println("An install directory must be specified.")
		return nil
	}

	// download file
	downloadPath := path.Clean(c.Args().First())
	downloadLocation := fmt.Sprintf("%v/%v", downloadPath, config.BINARY_NAME)
	downloadLocation = filepath.FromSlash(downloadLocation)
	urlLocation := fmt.Sprintf("%v://%v.%v/%v/%v/%v/%v", config.BINARY_PROTOCOL, config.BINARY_HOST, config.BINARY_DOMAIN, config.BINARY_DEFAULT_RELEASE, config.BINARY_DEFAULT_VERSION, config.BINARY_OS_PATH, config.BINARY_NAME)
	fmt.Printf("Downloading: %v...\n", urlLocation)
	err := utility.DownloadFile(downloadLocation, urlLocation)
	if err != nil {
		fmt.Printf("Error downloading GoCMS package: %v\n", err.Error())
		_ = os.Remove(downloadLocation)
		return nil
	}

	// unzip file
	fmt.Printf("Unpacking %v to %v\n", downloadLocation, downloadPath)
	err = utility.Unzip(downloadLocation, downloadPath)
	if err != nil {
		fmt.Printf("Error unpacking GoCMS package: %v\n", err.Error())
		_ = os.Remove(downloadLocation)
		return nil
	}

	// clean up zip file
	_ = os.Remove(downloadLocation)

	fmt.Println("GoCMS Installed Successfully!")

	return nil
}
