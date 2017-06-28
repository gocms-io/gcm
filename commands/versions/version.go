package versions

import (
	"fmt"
	"github.com/urfave/cli"
	"io"
	"log"
	"net/http"
	"os"
)

var CMD_VERSIONS = cli.Command{
	Name:   "versions",
	Usage:  "List all available gocms versions.",
	Action: cmd_versions,
}

func cmd_versions(c *cli.Context) error {

	response, err := http.Get("http://release.gocms.io/alpha-release/versions.txt")
	if err != nil {
		log.Fatal(err)
	} else {
		defer response.Body.Close()
		_, err := io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("")

	return nil
}
