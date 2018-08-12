package remote_torrent

import (
	"net/http"
	"strings"
	"io/ioutil"
	"time"
	"os"
	"encoding/json"
	"net/url"
	"io"
	"github.com/cheggaaa/pb"
)

var (
	ipAndPort, magnet string
	clientSideTorrentInfo torrentInfo
	barStartSignal chan struct{}
	bar *pb.ProgressBar
)

func init() {
	clientSideTorrentInfo = torrentInfo{}
	barStartSignal = make(chan struct{}, 1)
	bar = pb.New64(1)
}

func RTClient(args []string)  {
	go exitSignalHandlers()

	username = strings.Split(args[0], ":")[0]
	password = strings.Split(args[0], ":")[1]
	if username == "" || password == "" {
		Error.Println("username and password not specified")
		os.Exit(0)
	} else if len(password) < 7 {
		Warning.Println("password too short")
	}

	ipAndPort = args[1]
	if ipAndPort == "" {
		Error.Println("IP address and port not specified")
		os.Exit(0)
	}

	magnet = args[2]
	if magnet == "" {
		Error.Println("magnet link not specified")
		os.Exit(0)
	}

	for {
		client := &http.Client{}
		URL := "http://" + ipAndPort + "/magnet"
		v := url.Values{}
		v.Set("magnet", magnet)
		req, err := http.NewRequest("POST", URL, strings.NewReader(v.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth(username, password)
		resp, err := client.Do(req)
		if err != nil{
			Error.Println(err)
			Info.Println("retrying...")
			continue
		}

		if resp.StatusCode == http.StatusUnauthorized {
			Error.Println("Authentication failed")
			return
		} else if resp.StatusCode != http.StatusOK{
			Info.Println("Server may not be up yet. Retrying...")
			continue
		}

		Info.Println("Download scheduled")
		break
	}

	// check torrent download status
	ticker := time.NewTicker(time.Second * 3)
	go progressBar()
	for range ticker.C {
		client := &http.Client{}
		URL := "http://" + ipAndPort + "/status"
		req, err := http.NewRequest("GET", URL, nil)
		req.SetBasicAuth(username, password)
		resp, err := client.Do(req)
		if err != nil{
			Error.Println(err)
			Info.Println("retrying...")
			continue
		}

		if resp.StatusCode == http.StatusUnauthorized {
			Error.Println("Authentication failed")
			return
		} else if resp.StatusCode == http.StatusOK {
			bar.Finish()
			Info.Println("Download completed on server")
			break
		} else if resp.StatusCode == http.StatusBadRequest {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				Error.Printf("torrent error and ioutil.ReadAll error: %s\n", err.Error())
			}
			resp.Body.Close()
			Error.Printf("torrent error: %s\n", string(body))
			return
		}

		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&clientSideTorrentInfo)
		resp.Body.Close()
		if err != nil {
			Error.Printf("error decoding json: %v\n", err)
			continue
		}
		barStartSignal <- struct{}{}
	}
	ticker.Stop()


	// download files from server's directory
	for {
		client := &http.Client{}
		URL := "http://" + ipAndPort + "/filenames"
		req, err := http.NewRequest("GET", URL, nil)
		req.SetBasicAuth(username, password)
		resp, err := client.Do(req)
		if err != nil{
			Error.Println(err)
			Info.Println("retrying...")
			continue
		}

		if resp.StatusCode == http.StatusUnauthorized {
			Error.Println("Authentication failed")
			return
		} else if resp.StatusCode != http.StatusOK {
			Error.Println("statuscode not ok")
			continue
		}

		decoder := json.NewDecoder(resp.Body)
		var s []string
		err = decoder.Decode(&s)
		if err != nil {
			Error.Printf("error decoding json: %v\n", err)
			continue
		}
		//Info.Println(s)
		downloadFiles(s)

		break
	}

	// remove torrent on server remotely
	for {
		client := &http.Client{}
		URL := "http://" + ipAndPort + "/remove"
		req, err := http.NewRequest("GET", URL, nil)
		req.SetBasicAuth(username, password)
		resp, err := client.Do(req)
		if err != nil {
			Error.Println(err)
			Info.Println("retrying...")
			continue
		}

		if resp.StatusCode == http.StatusUnauthorized {
			Error.Println("Authentication failed")
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			Error.Println(err)
			Info.Println("retrying...")
			continue
		}

		Info.Println(string(body))

		if strings.Contains(string(body), "Files removed on server") {
			//Info.Println("Torrent removed on server side")
			break
		}
	}

	return
}

func downloadFiles(s []string) {
	for _, name := range s {
		if name == ".torrent.bolt.db" {
			continue
		}

		for {
			//Info.Println(name)
			f, err := createFile(name)
			if err != nil {
				Error.Fatalf("Unable to open %s file: %s", name, err.Error())
			}

			client := &http.Client{}
			URL := "http://" + ipAndPort + "/lalaland/" + url.PathEscape(name)
			req, err := http.NewRequest("GET", URL, nil)
			req.SetBasicAuth(username, password)
			resp, err := client.Do(req)
			if err != nil{
				resp.Body.Close()
				f.Close()
				Error.Println(err)
				Info.Println("retrying...")
				continue
			}

			if resp.StatusCode == http.StatusUnauthorized {
				Error.Println("Authentication failed")
				return
			}


			progressBar := pb.New64(resp.ContentLength)
			progressBar.SetUnits(pb.U_BYTES)
			progressBar.ShowTimeLeft = true
			progressBar.ShowSpeed = true
			//	progressBar.RefreshRate = time.Millisecond * 1
			Info.Printf("Retrieving file: %s\n", name)
			progressBar.Start()
			out := io.MultiWriter(f, progressBar)

			_, err = io.Copy(out, resp.Body)
			if err != nil {
				progressBar.Finish()
				resp.Body.Close()
				f.Close()
				Error.Printf("Write to file %s failed. Retrying...\n", name)
			}

			progressBar.Finish()
			resp.Body.Close()
			f.Close()
			break
		}
	}
}

func progressBar() {
	<-barStartSignal
	bar = pb.New64(clientSideTorrentInfo.TL)
	Info.Println("Torrent download progress on server:")
	bar.Start()
	for {
		bar.Total = clientSideTorrentInfo.TL
		bar.Set64(clientSideTorrentInfo.BC)
		<-barStartSignal
	}
}