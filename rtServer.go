package remote_torrent

import (
	"net/http"
	"fmt"
	"strings"
	"os/exec"
	"os"
)
var (
	port string
	username string
	password string
	//downloadFinished chan string
)

func RTServer(args []string) {
	port = args[0]
	username = strings.Split(args[1], ":")[0]
	password = strings.Split(args[1], ":")[1]

	// Simple static webserver
	mux := http.NewServeMux()
	mux.HandleFunc("/", simpleAuthentication(http.FileServer(http.Dir("/root/media/"))))
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
		fmt.Fprintf(w, "Torrent Download Scheduled.")
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Authentication Failed.")
	}
}

func statusCheck(w http.ResponseWriter, r *http.Request) {
	if check(r.PostFormValue("username"), r.PostFormValue("password")) {
		downloadLog := out.String()
		if strings.Contains(downloadLog, "utpSocketUtpPacketsReceived") {
			fmt.Println("Torrent Download Finished.")
			fmt.Fprintf(w, "Torrent Download Finished.")
			return
		}
		info := strings.Split(downloadLog, "% downloading")
		infoLength := len(info)
		if infoLength < 2 {
			fmt.Fprintf(w, "not started yet")
		} else {
			mostUpToDateStatus := info[infoLength - 2]
			mostUpToDateStatusDownloaded := strings.Split(strings.Split(info[infoLength - 1], "(")[1], "/")[0]
			mostUpToDateStatusOverallData := strings.Split(strings.Split(info[infoLength - 1], "/")[1], ")")[0]
			length := len(mostUpToDateStatus)
			statusPercentage := mostUpToDateStatus[length-3:]

			fmt.Println(mostUpToDateStatusDownloaded)
			fmt.Fprintf(w, statusPercentage + "%% " + mostUpToDateStatusDownloaded + "/"+ mostUpToDateStatusOverallData)
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