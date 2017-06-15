package install

import (
	"fmt"
	"github.com/gocms-io/gcm/config"
	"github.com/gocms-io/gcm/utility"
	"github.com/urfave/cli"
	"os"
	"path"
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

	// clean destination file path
	// todo make os agnostic
	downloadLocation := fmt.Sprintf("%v/%v", path.Clean(c.Args().First()), config.BINARY_NAME)

	urlLocation := fmt.Sprintf("%v://%v.%v/%v/%v/%v/%v", config.BINARY_PROTOCOL, config.BINARY_HOST, config.BINARY_DOMAIN, config.BINARY_DEFAULT_RELEASE, config.BINARY_DEFAULT_VERSION, config.BINARY_OS_PATH, config.BINARY_NAME)
	fmt.Println("Downloading: ", urlLocation)
	err := utility.DownloadFile(downloadLocation, urlLocation)

	if err != nil {
		fmt.Printf("Error downloading GoCMS package: %v\n", err.Error())
		os.Remove(downloadLocation)
		return nil
	}

	fmt.Println("GoCMS Installed Successfully")

	return nil
}
