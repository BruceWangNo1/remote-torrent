package rt

import (
	"encoding/json"
	"errors"
	"github.com/cheggaaa/pb"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	ipAndPort, magnet     string
	clientSideTorrentInfo torrentInfo
	barStartSignal        chan struct{}
	bar                   *pb.ProgressBar
)

func init() {
	clientSideTorrentInfo = torrentInfo{}
	barStartSignal = make(chan struct{}, 1)
	bar = pb.New64(1)
}

// Client is the remote torrent main client function
func Client(args []string) {
	go exitSignalHandlers()

	username = strings.Split(args[0], ":")[0]
	password = strings.Split(args[0], ":")[1]
	if username == "" || password == "" {
		errorLog.Println("username and password not specified")
		os.Exit(0)
	} else if len(password) < 7 {
		warningLog.Println("password too short")
	}

	ipAndPort = args[1]
	if ipAndPort == "" {
		errorLog.Println("IP address and port not specified")
		os.Exit(0)
	}

	magnet = args[2]
	if magnet == "" {
		errorLog.Println("magnet link not specified")
		os.Exit(0)
	}

	err := sendMagnetLink()
	if err != nil {
		return
	}

	err = statusCheckFromClient()
	if err != nil {
		return
	}

	err = downloadFilesRequest()
	if err != nil {
		return
	}

	err = removeFilesOnServer()
	if err != nil {
		return
	}

	return
}

func downloadFiles(s []string) {
	for _, name := range s {
		if name == ".torrent.bolt.db" {
			continue
		}

		for {
			//infoLog.Println(name)
			f, err := createFile(name)
			if err != nil {
				errorLog.Fatalf("Unable to open %s file: %s", name, err.Error())
			}

			client := &http.Client{}
			URL := "http://" + ipAndPort + "/lalaland/" + url.PathEscape(name)
			req, err := http.NewRequest("GET", URL, nil)
			if err != nil {
				f.Close()
				errorLog.Println(err)
				infoLog.Println("retrying...")
				continue
			}
			req.SetBasicAuth(username, password)
			resp, err := client.Do(req)
			if err != nil {
				resp.Body.Close()
				f.Close()
				errorLog.Println(err)
				infoLog.Println("retrying...")
				continue
			}

			if resp.StatusCode == http.StatusUnauthorized {
				errorLog.Println("Authentication failed")
				return
			}

			progressBar := pb.New64(resp.ContentLength)
			progressBar.SetUnits(pb.U_BYTES)
			progressBar.ShowTimeLeft = true
			progressBar.ShowSpeed = true
			//	progressBar.RefreshRate = time.Millisecond * 1
			infoLog.Printf("Retrieving file: %s\n", name)
			progressBar.Start()
			out := io.MultiWriter(f, progressBar)

			_, err = io.Copy(out, resp.Body)
			if err != nil {
				progressBar.Finish()
				resp.Body.Close()
				f.Close()
				errorLog.Printf("Write to file %s failed. Retrying...\n", name)
			}

			progressBar.Finish()
			resp.Body.Close()
			f.Close()
			break
		}
	}
}
func sendMagnetLink() error {
	for {
		client := &http.Client{}
		URL := "http://" + ipAndPort + "/magnet"
		v := url.Values{}
		v.Set("magnet", magnet)
		req, err := http.NewRequest("POST", URL, strings.NewReader(v.Encode()))
		if err != nil {
			errorLog.Println(err)
			infoLog.Println("retrying...")
			continue
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth(username, password)
		resp, err := client.Do(req)
		if err != nil {
			errorLog.Println(err)
			infoLog.Println("retrying...")
			continue
		}

		if resp.StatusCode == http.StatusUnauthorized {
			errorLog.Println("Authentication failed")
			return errors.New("authentication failed")
		} else if resp.StatusCode != http.StatusOK {
			infoLog.Println("Server may not be up yet. Retrying...")
			continue
		}

		infoLog.Println("Download scheduled")
		break
	}
	return nil
}

func statusCheckFromClient() error {
	// check torrent download status
	ticker := time.NewTicker(time.Second * 3)
	go progressBar()
	for range ticker.C {
		client := &http.Client{}
		URL := "http://" + ipAndPort + "/status"
		req, err := http.NewRequest("GET", URL, nil)
		if err != nil {
			errorLog.Println(err)
			infoLog.Println("retrying...")
			continue
		}
		req.SetBasicAuth(username, password)
		resp, err := client.Do(req)
		if err != nil {
			errorLog.Println(err)
			infoLog.Println("retrying...")
			continue
		}

		if resp.StatusCode == http.StatusUnauthorized {
			errorLog.Println("Authentication failed")
			return errors.New("authentication failed")
		} else if resp.StatusCode == http.StatusOK {
			bar.Finish()
			infoLog.Println("Download completed on server")
			break
		} else if resp.StatusCode == http.StatusBadRequest {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				errorLog.Printf("torrent error and ioutil.ReadAll error: %s\n", err.Error())
			}
			resp.Body.Close()
			errorLog.Printf("torrent error: %s\n", string(body))
			return errors.New("torrent error: " + string(body))
		}

		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&clientSideTorrentInfo)
		resp.Body.Close()
		if err != nil {
			errorLog.Printf("error decoding json: %v\n", err)
			continue
		}
		barStartSignal <- struct{}{}
	}
	ticker.Stop()
	return nil
}

func downloadFilesRequest() error {
	// download files from server's directory
	for {
		client := &http.Client{}
		URL := "http://" + ipAndPort + "/filenames"
		req, err := http.NewRequest("GET", URL, nil)
		if err != nil {
			errorLog.Println(err)
			infoLog.Println("retrying...")
			continue
		}
		req.SetBasicAuth(username, password)
		resp, err := client.Do(req)
		if err != nil {
			errorLog.Println(err)
			infoLog.Println("retrying...")
			continue
		}

		if resp.StatusCode == http.StatusUnauthorized {
			errorLog.Println("Authentication failed")
			return errors.New("authentication failed")
		} else if resp.StatusCode != http.StatusOK {
			errorLog.Println("statuscode not ok")
			continue
		}

		decoder := json.NewDecoder(resp.Body)
		var s []string
		err = decoder.Decode(&s)
		if err != nil {
			errorLog.Printf("error decoding json: %v\n", err)
			continue
		}
		//infoLog.Println(s)
		downloadFiles(s)

		break
	}
	return nil
}

func removeFilesOnServer() error {
	// remove torrent on server remotely
	for {
		client := &http.Client{}
		URL := "http://" + ipAndPort + "/remove"
		req, err := http.NewRequest("GET", URL, nil)
		if err != nil {
			errorLog.Println(err)
			infoLog.Println("retrying...")
			continue
		}
		req.SetBasicAuth(username, password)
		resp, err := client.Do(req)
		if err != nil {
			errorLog.Println(err)
			infoLog.Println("retrying...")
			continue
		}

		if resp.StatusCode == http.StatusUnauthorized {
			errorLog.Println("Authentication failed")
			return errors.New("authentication failed")
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			errorLog.Println(err)
			infoLog.Println("retrying...")
			continue
		}

		infoLog.Println(string(body))

		if strings.Contains(string(body), "Files removed on server") {
			//infoLog.Println("Torrent removed on server side")
			break
		}
	}
	return nil
}

func progressBar() {
	<-barStartSignal
	bar = pb.New64(clientSideTorrentInfo.TL)
	infoLog.Println("Torrent download progress on server:")
	bar.Start()
	for {
		bar.Total = clientSideTorrentInfo.TL
		bar.Set64(clientSideTorrentInfo.BC)
		<-barStartSignal
	}
}
