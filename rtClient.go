package remote_torrent

import (
	"net/http"
	"net/url"
	"strings"
	"io/ioutil"
	"time"
	"os/exec"
	"os"
)

func RTClient(args []string)  {
	username := strings.Split(args[0], ":")[0]
	password := strings.Split(args[0], ":")[1]
	if username == "" || password == "" {
		Error.Println("username and password not specified")
		os.Exit(0)
	} else if len(password) < 7 {
		Warning.Println("password too short")
	}

	ipAndPort := args[1]
	if ipAndPort == "" {
		Error.Println("IP address and port not specified")
		os.Exit(0)
	}

	magnet := args[2]
	if magnet == "" {
		Error.Println("magnet link not specified")
		os.Exit(0)
	}

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
			Error.Println(err)
			Info.Println("retrying...")
			continue
		} else {
			Info.Println("Torrent download scheduling over http succeeds")
		}

		if schedulingResp.StatusCode == http.StatusUnauthorized {
			Error.Println("Authentication failed")
			return
		}

		body, err := ioutil.ReadAll(schedulingResp.Body)
		schedulingResp.Body.Close()

		if err != nil {
			Error.Println(err)
			Info.Println("retrying...")
			continue
		}

		Info.Println(string(body))
		break
	}

	// status check
	ticker := time.NewTicker(time.Second * 3)

	for range ticker.C {
		statusCheckResp, statusCheckErr := http.PostForm("http://" + ipAndPort + "/status", url.Values{"username": {username}, "password": {password}})
		if statusCheckErr != nil {
			Error.Println(statusCheckErr)
		}

		if statusCheckResp.StatusCode == http.StatusUnauthorized {
			Error.Println("Authentication failed.")
			return
		}

		body, err := ioutil.ReadAll(statusCheckResp.Body)
		if err != nil {
			Error.Println(err)
		}

		Info.Printf("Status check: %s\n", string(body))

		statusCheckResp.Body.Close()

		if strings.Contains(string(body), "Torrent Download Finished") {
			Info.Println("Torrent download completed 100% on server side")
			break
		}
	}
	ticker.Stop()

	// download files from server's directory
	for {
		_, err := exec.LookPath("wget")
		if err != nil {
			Error.Println("Please install wget on your local machine first")
			return
		}
		cmd := exec.Command("wget", "-r", "--post-data", "username=" + username + "&password=" + password, "http://" + ipAndPort + "/files/")
		cmd.Stderr = os.Stdout
		err = cmd.Run()
		if err != nil {
			Error.Printf("Error starting wget: %v\n", err)
		}

		//go func() {
		//	finished <- cmd.Wait()
		//}()


		Info.Println("Torrent successfully retrieved to client")
		break
	}

	// remove torrent on server remotely
	for {
		removeTorrentResponse, removeTorrentErr := http.PostForm("http://" + ipAndPort + "/remove", url.Values{"username": {username}, "password": {password}})
		if removeTorrentErr != nil {
			Error.Println(removeTorrentErr)
			Info.Println("retrying...")
			continue
		}

		if removeTorrentResponse.StatusCode == http.StatusUnauthorized {
			Error.Println("Authentication failed")
			return
		}

		body, err := ioutil.ReadAll(removeTorrentResponse.Body)
		removeTorrentResponse.Body.Close()

		if err != nil {
			Error.Println(err)
			continue
		}
		Info.Println(string(body))
		break
	}

	return
}