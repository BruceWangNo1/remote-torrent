package main

import (
	"os"
	"log"
	rt "github.com/brucewangno1/remote-torrent"
)

// 1) Client Mode (e.g. go run main.go client userName:password ip:port torrentFile/magnetLink)
// 2) Server Mode (e.g. go run main.go server portToListen userName:Password)
func main() {
	mode := os.Args[1]
	if mode == "client" {
		rt.RTClient(os.Args[2:])
	} else if mode == "server" {
		rt.RTServer(os.Args[2:])
	} else {
		log.Fatal("Args input is wrong. Please check usage\n")
	}
}
