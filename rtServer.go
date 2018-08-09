package remote_torrent

import (
	"net/http"
	"fmt"
	"strings"
	"os/exec"
	"os"
	"log"
	"time"
)
var (
	port string
	username string
	password string
	downloadFinished chan string

	Info *log.Logger
	Warning *log.Logger
	Error *log.Logger
)

func init() {
	Info = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
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
	mux.HandleFunc("/files/", simpleAuthentication(http.FileServer(http.Dir("/root/media/"))))
	mux.HandleFunc("/magnet", torrentDownloadAssignment)
	mux.HandleFunc("/status", statusCheck)
	mux.HandleFunc("/remove", removeTorrent)
	http.ListenAndServe(":"+port, mux)
}


func simpleAuthentication(fn http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//user, pass, _ := r.BasicAuth()
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Unauthorized.", 401)
			return
		}
		username := r.PostFormValue("username")
		password := r.PostFormValue("password")

		if !check(username, password) {
			http.Error(w, "Unauthorized.", 401)
			return
		}
		fn.ServeHTTP(w, r)
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
	if check(r.PostFormValue("username"), r.PostFormValue("password")) {
		magnet := r.PostFormValue("magnet")
		go download(magnet)
		fmt.Fprintf(w, "Torrent Download Scheduled")
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Authentication Failed.")
	}
}

func statusCheck(w http.ResponseWriter, r *http.Request) {
	if check(r.PostFormValue("username"), r.PostFormValue("password")) {
		select {
		case <-downloadFinished:
			Info.Println("Torrent Download Finished")
			fmt.Fprintf(w, "Torrent Download Finished")
			return
		case <-time.After(time.Second):
			Info.Println("Torrent Download Ongoing")
			fmt.Fprintf(w, "Torrent Download Ongoing")
			return
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Authentication Failed.")
	}
}

func removeTorrent(w http.ResponseWriter, r *http.Request) {
	if check(r.PostFormValue("username"), r.PostFormValue("password")) {
		//cmd := exec.Command("bash")
		cmd := exec.Command("bash", "-c", "rm -rf /root/media/*")
		//cmd.Stdin = strings.NewReader("rm -rf /root/media/*")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
		err := cmd.Run()
		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(w, "'rm' command failed.")
		}

		fmt.Fprintf(w, "Torrent removed on server side.")
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Authentication Failed.")
	}
}