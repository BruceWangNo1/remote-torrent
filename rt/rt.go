package main

import (
	rt "github.com/brucewangno1/remote-torrent"
	"log"
	"os"
)

// 1) Client Mode (e.g. go run main.go client userName:password ip:port torrentFile/magnetLink)
// 2) Server Mode (e.g. go run main.go server portToListen userName:Password)
func main() {
	rt.Mode = os.Args[1]
	if rt.Mode == "client" {
		rt.Client(os.Args[2:])
	} else if rt.Mode == "server" {
		rt.Server(os.Args[2:])
	} else {
		log.Fatal("Args input is wrong. Please check usage\n")
	}
}
