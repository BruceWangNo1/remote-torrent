package remote_torrent

import (
	"os"
	"strings"
	"log"
)

func createFile(name string) (f *os.File, err error) {
	err = checkDir(name)
	if err != nil {
		log.Fatalf("error creating %s directory: %s", name, err.Error())
	}
	f, err = os.OpenFile(name, os.O_CREATE | os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Unable to open %s file: %s", name, err.Error())
	}

	return
}

func checkDir(name string) error {
	index := strings.LastIndex(name, "/")
	if index == -1 {
		return nil
	}

	dir := name[0:index]
	err := os.MkdirAll(dir, 0731)
	if err != nil {
		return err
	}
	return nil
}