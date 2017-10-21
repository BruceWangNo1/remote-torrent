package remote_torrent

import (
	"net/http"
	"net/url"
	"strings"
	"fmt"
	"io/ioutil"
	"time"
	"os/exec"
	"log"
	"os"
)

func RTClient(args []string)  {
	username := strings.Split(args[0], ":")[0]
	password := strings.Split(args[0], ":")[1]
	ipAndPort := args[1]
	magnet := args[2]

	//finished := make(chan error, 2)
	//
	//c := make(chan os.Signal, 2)
	//signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	//
	//go func() {
	//	select {
	//	case <-c:
	//		if err := cmd.Process.Kill(); err != nil {
	//			log.Fatal("failed to kill: ", err)
	//		}
	//		os.Exit(1)
	//		return
	//	case err := <-finished:
	//		if err != nil {
	//			fmt.Println(err)
	//			return
	//		}
	//	}
	//}()

	// schedule torrent download
	for {
		schedulingResp, err := http.PostForm("http://" + ipAndPort + "/magnet", url.Values{"username": {username}, "password": {password}, "magnet": {magnet}})
		if err != nil {
			fmt.Println(err)
			fmt.Println("retrying...")
			continue
		} else {
			fmt.Println("Torrent download scheduling over http succeeds.")
		}

		if schedulingResp.StatusCode == http.StatusUnauthorized {
			log.Fatal("Authentication failed.")
			return
		}

		body, err := ioutil.ReadAll(schedulingResp.Body)
		schedulingResp.Body.Close()

		if err != nil {
			fmt.Println(err)
			fmt.Println("retrying...")
			continue
		}

		fmt.Println(string(body))
		break
	}

	// status check
	ticker := time.NewTicker(time.Second * 3)

	for range ticker.C {
		statusCheckResp, statusCheckErr := http.PostForm("http://" + ipAndPort + "/status", url.Values{"username": {username}, "password": {password}})
		if statusCheckErr != nil {
			fmt.Println(statusCheckErr)
		}

		if statusCheckResp.StatusCode == http.StatusUnauthorized {
			log.Fatal("Authentication failed.")
			return
		}

		body, err := ioutil.ReadAll(statusCheckResp.Body)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("\rStatus check: %s.", string(body))

		statusCheckResp.Body.Close()

		if strings.Contains(string(body), "Torrent Download Finished.") {
			fmt.Println("\nTorrent download completed 100% on server side.")
			break
		}
	}
	ticker.Stop()
	close(ticker)

	// download files from server's directory
	for {
		_, err := exec.LookPath("wget")
		if err != nil {
			log.Fatal("Please install wget on your local machine first.\n")
			return
		}
		cmd := exec.Command("wget", "-r", "--post-data", "username=" + username + "&password=" + password, "http://" + ipAndPort + "/")
		cmd.Stderr = os.Stdout
		err = cmd.Run()
		if err != nil {
			log.Fatal("Error starting wget: %v", err)
		}

		//go func() {
		//	finished <- cmd.Wait()
		//}()


		fmt.Println("Torrent successfully retrieved to client.")
		break
	}

	// remove torrent on server remotely
	for {
		removeTorrentResponse, removeTorrentErr := http.PostForm("http://" + ipAndPort + "/remove", url.Values{"username": {username}, "password": {password}})
		if removeTorrentErr != nil {
			fmt.Println(removeTorrentErr)
			fmt.Println("retrying...")
			continue
		}

		if removeTorrentResponse.StatusCode == http.StatusUnauthorized {
			log.Fatal("Authentication failed.")
			return
		}

		body, err := ioutil.ReadAll(removeTorrentResponse.Body)
		removeTorrentResponse.Body.Close()

		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(string(body))
		break
	}

	return
}