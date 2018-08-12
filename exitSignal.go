package remote_torrent

import (
	"os"
	"os/signal"
	"syscall"
	"net/http"
)

var Mode string

func exitSignalHandlers() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	Info.Printf("close signal received: %+v\n", <-c)

	if Mode == "server" {
		os.Exit(1)
	} else {
		counter := 0
		for {
			if counter > 2 {
				break
			}
			client := &http.Client{}
			URL := "http://" + ipAndPort + "/server-cleanup"
			req, err := http.NewRequest("GET", URL, nil)
			req.SetBasicAuth(username, password)
			resp, err := client.Do(req)
			if err != nil {
				Error.Println(err)
				Info.Println("retrying...")
				counter = counter + 1
				continue
			}

			if resp.StatusCode == http.StatusUnauthorized {
				Error.Println("Authentication failed")
				return
			} else if resp.StatusCode == http.StatusOK {
				Info.Println("Server cleanup successful")
				break
			} else {
				counter = counter + 1
				Error.Printf("status code: %v\n", resp.StatusCode)
				continue
			}
		}
		os.Exit(1)
	}
}
