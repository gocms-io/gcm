package utility

import (
	"github.com/cavaliercoder/grab"
	"fmt"
	"time"
)

func DownloadFile(filepath string, downloadUrl string) (err error) {

	client := grab.NewClient()
	req, _ := grab.NewRequest(filepath, downloadUrl)

	resp := client.Do(req)


	// start Progress loop
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			fmt.Printf("	transferred %v /%v bytes (%.2f%%)\n",
				resp.BytesComplete(),
				resp.Size,
				100*resp.Progress())

		case <-resp.Done:
			break Loop
		}
	}

	// check errors
	if err := resp.Err(); err != nil {
		return err
	}

	return nil
}