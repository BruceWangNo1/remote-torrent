package remote_torrent

import (
	"os/exec"
	"log"
	"bytes"
	"fmt"
)

var out bytes.Buffer

func download(magnet string) error  {
	info.Printf("magnet link: %s download scheduled\n", magnet)

	_, err := exec.LookPath("torrent")
	if err != nil {
		log.Fatal("Please install anacrolix/torrent first.\n")
		return err
	}

	go func() {
		cmd := exec.Command("torrent", magnet)
		cmd.Dir = "/root/media/"
		cmd.Stdout = &out
		err := cmd.Start()
		if err != nil {
			log.Fatal(err)
			return
		}
		//downloadFinished = make(chan string)
		//downloadFinished <- "finished"
		fmt.Println("Torrent Download Finished Successfully.")
	}()

	return nil
}