package rt

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// Mode represents rt running mode, either "client" or "server"
var Mode string

func exitSignalHandlers() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	infoLog.Printf("close signal received: %+v\n", <-c)

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
				errorLog.Println(err)
				infoLog.Println("retrying...")
				counter = counter + 1
				continue
			}

			if resp.StatusCode == http.StatusUnauthorized {
				errorLog.Println("Authentication failed")
				return
			} else if resp.StatusCode == http.StatusOK {
				infoLog.Println("Server cleanup successful")
				break
			} else {
				counter = counter + 1
				errorLog.Printf("status code: %v\n", resp.StatusCode)
				continue
			}
		}
		os.Exit(1)
	}
}
