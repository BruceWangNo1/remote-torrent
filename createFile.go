package rt

import (
	"log"
	"os"
	"strconv"
	"strings"
)

func getFile(name string) (f *os.File, byteRange string) {
	err := checkDir(name)
	if err != nil {
		log.Fatalf("error creating %s directory: %s", name, err.Error())
	}

	if stat, err := os.Stat(name); err == nil {
		byteRange = strconv.FormatInt(stat.Size(), 10) + "-"
		f, err = os.OpenFile(name, os.O_APPEND|os.O_WRONLY, 0611)
		if err != nil {
			log.Fatalf("Unable to open %s file: %s", name, err.Error())
		}
		return
	} else if os.IsNotExist(err) {
		f, err = os.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0611)
		if err != nil {
			log.Fatalf("Unable to open %s file: %s", name, err.Error())
		}
		return
	}

	log.Fatalf("Unable to get the stat of file %s\n", err.Error())
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
