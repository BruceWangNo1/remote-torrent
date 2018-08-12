package remote_torrent

import (
	"net/http"
	"fmt"
	"strings"
	"os"
	"log"
	"time"
	"path/filepath"
	"encoding/json"
)
var (
	port string
	username string
	password string
	mediaDir string
	downloadFinished bool

	downloadInProgress chan struct{}
	clientCleanupSignal chan struct{}
	clientCleanupFinished chan struct{}

	Info *log.Logger
	Warning *log.Logger
	Error *log.Logger
)

func init() {
	Info = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	downloadFinished = false
	downloadInProgress = make(chan struct{}, 1)
	clientCleanupSignal = make(chan struct{}, 1)
	clientCleanupFinished = make(chan struct{}, 1)
	mediaDir = "/root/media"
}

func RTServer(args []string) {
	port = args[0]
	if port == "" {
		port = "6789"
		Info.Println("port number is not specified. set to default 6789")
	}

	username = strings.Split(args[1], ":")[0]
	password = strings.Split(args[1], ":")[1]
	if username == "" || password == "" {
		Error.Println("username and password not specified")
		os.Exit(0)
	} else if len(password) < 7 {
		Warning.Println("password too short")
	}

	// Simple static webserver
	mux := http.NewServeMux()
	mux.HandleFunc("/lalaland/", authenticationByBasicAuth(http.StripPrefix("/lalaland/", http.FileServer(http.Dir(mediaDir))).ServeHTTP))
	mux.HandleFunc("/magnet", authenticationByBasicAuth(torrentDownloadAssignment))
	mux.HandleFunc("/status", authenticationByBasicAuth(statusCheck))
	mux.HandleFunc("/remove", authenticationByBasicAuth(removeTorrent))
	mux.HandleFunc("/server-cleanup", authenticationByBasicAuth(serverCleanup))
	mux.HandleFunc("/filenames", authenticationByBasicAuth(filenames))
	http.ListenAndServe(":"+port, mux)
}

func authenticationByBasicAuth(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		// log.Printf("BasicAuth: %v:%v\n", user, pass)
		if !ok || !check(user, pass) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Username and Password, Please"`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised"))
			return
		}

		fn(w, r)
	}
}

func check(u, p string) bool {
	if u == username && p == password {
		return true
	} else {
		return false
	}
}

func torrentDownloadAssignment(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		Error.Println("ParseForm problem occurred")
	}
	magnet := r.PostFormValue("magnet")

	go download(magnet)
	torrentErrorForHTTPHandler = nil

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Torrent Download Scheduled")
}

func statusCheck(w http.ResponseWriter, r *http.Request) {
	if downloadFinished == true {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Torrent Download Finished")
		return
	} else if torrentErrorForHTTPHandler != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, torrentErrorForHTTPHandler.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(serverSideTorrentInfo)

	//select {
	//case <-downloadFinished:
	//	Info.Println("Torrent Download Finished")
	//	fmt.Fprintf(w, "Torrent Download Finished")
	//	return
	//case <-time.After(time.Second):
	//	Info.Println("Torrent Download Ongoing")
	//	fmt.Fprintf(w, "Torrent Download Ongoing")
	//	return
	//}
}

func removeTorrent(w http.ResponseWriter, r *http.Request) {
	err := RemoveContents(mediaDir)
	if err != nil {
		Error.Println(err)
		fmt.Fprintf(w, "remove contents failed")
	} else {
		fmt.Fprintf(w, "Files removed on server")
	}
}

func serverCleanup(w http.ResponseWriter, r *http.Request) {
	select {
	case <-downloadInProgress:
		clientCleanupSignal <- struct{}{}
		//Info.Println("Cleanup signal delivered")
		select {
		case <-clientCleanupFinished:
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Cleanup signal delivered")
		case <-time.After(2 * time.Second):
			w.WriteHeader(http.StatusGatewayTimeout)
			fmt.Fprintf(w, "Cleanup process timeout")
		}
	case <-time.After(2 * time.Second):
		//Info.Println("No need to clean up")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "No need to clean up")
	}
}

func filenames(w http.ResponseWriter, r *http.Request) {
	var s []string

	err := filepath.Walk(mediaDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			path = strings.TrimPrefix(path, mediaDir + "/")
			s = append(s, path)
		}

		return nil
	})

	if err != nil {
		Error.Fatalf("filepath.Walk() returned %v\n", err)
	}

	if err != nil {
		Error.Fatalln(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(s)
}