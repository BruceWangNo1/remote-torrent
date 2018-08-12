package remote_torrent

import (
	"log"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/tagflag"
	"net"
	"net/http"
	"expvar"
	"github.com/anacrolix/envpprof"
	"github.com/anacrolix/torrent/metainfo"
	"os"
	"strings"
	"time"
)

var flags = struct {
	Mmap            bool           `help:"memory-map torrent data"`
	TestPeer        []*net.TCPAddr `help:"addresses of some starting peers"`
	Seed            bool           `help:"seed after download is complete"`
	Addr            *net.TCPAddr   `help:"network listen addr"`
	UploadRate      tagflag.Bytes  `help:"max piece bytes to send per second"`
	DownloadRate    tagflag.Bytes  `help:"max bytes per second down from peers"`
	Debug           bool
	PackedBlocklist string
	Stats           *bool
	tagflag.StartPos
	Torrent []string `arity:"+" help:"torrent file path or magnet uri"`
}{
	UploadRate:   -1,
	DownloadRate: -1,
}

type torrentInfo struct {
	TL int64 `json:"tl"`
	BC int64 `json:"bc"`
}

var	(
	serverSideTorrentInfo torrentInfo
	torrentError chan error
	torrentErrorForHTTPHandler error
)

func init() {
	serverSideTorrentInfo = torrentInfo{TL: 1, BC: 0}
	torrentError = make(chan error, 1)
}

func download(magnet string) (err error)  {
	select {
	case <-downloadInProgress:
		downloadInProgress <- struct{}{}
	default:
		downloadInProgress <- struct{}{}
	}
	downloadFinished = false

	//log.SetFlags(log.LstdFlags | log.Lshortfile)
	Info.Printf("magnet link: %s download scheduled\n", magnet)
	defer envpprof.Stop()
	flags.Torrent = []string{magnet}
	flags.Seed = false
	flags.Debug = false
	clientConfig := torrent.NewDefaultClientConfig()
	clientConfig.Debug = flags.Debug
	clientConfig.Seed = flags.Seed

	client, err := torrent.NewClient(clientConfig)
	if err != nil {
		Error.Printf("error creating client: %s\n", err)
		return err
	}
	defer client.Close()
	go clientSignalHandler(client)
	go torrentErrorHandler(client)

	// Write status on the root path on the default HTTP muxer. This will be
	// bound to localhost somewhere if GOPPROF is set, thanks to the envpprof
	// import.
	//http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
	//	client.WriteStatus(w)
	//})

	addTorrents(client)
	if client.WaitAll() {
		downloadFinished = true
		Info.Println("downloaded ALL the torrents")
	} else {
		Info.Println("y u no complete torrents?!")
		return
	}
	if flags.Seed {
		outputStats(client)
		select {}
	}
	outputStats(client)

	client.Close()

	<-downloadInProgress

	return nil
}

func addTorrents(client *torrent.Client) {
	for _, arg := range flags.Torrent {
		t := func() *torrent.Torrent {
			if strings.HasPrefix(arg, "magnet:") {
				t, err := client.AddMagnet(arg)
				if err != nil {
					Error.Printf("error adding magnet: %s\n", err)
					torrentError <- err
					return nil
				}
				return t
			} else if strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://") {
				response, err := http.Get(arg)
				if err != nil {
					log.Fatalf("Error downloading torrent file: %s", err)
				}

				metaInfo, err := metainfo.Load(response.Body)
				defer response.Body.Close()
				if err != nil {
					fmt.Fprintf(os.Stderr, "error loading torrent file %q: %s\n", arg, err)
					os.Exit(1)
				}
				t, err := client.AddTorrent(metaInfo)
				if err != nil {
					log.Fatal(err)
				}
				return t
			} else if strings.HasPrefix(arg, "infohash:") {
				t, _ := client.AddTorrentInfoHash(metainfo.NewHashFromHex(strings.TrimPrefix(arg, "infohash:")))
				return t
			} else {
				metaInfo, err := metainfo.LoadFromFile(arg)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error loading torrent file %q: %s\n", arg, err)
					os.Exit(1)
				}
				t, err := client.AddTorrent(metaInfo)
				if err != nil {
					log.Fatal(err)
				}
				return t
			}
		}()

		if t == nil {
			continue
		}

		getInfo(t)
		t.AddPeers(func() (ret []torrent.Peer) {
			for _, ta := range flags.TestPeer {
				ret = append(ret, torrent.Peer{
					IP:   ta.IP,
					Port: ta.Port,
				})
			}
			return
		}())
		go func() {
			<-t.GotInfo()
			t.DownloadAll()
		}()
	}
}

func getInfo(t *torrent.Torrent) {
	go func() {
		<-t.GotInfo()
		serverSideTorrentInfo.TL = t.Info().TotalLength()
		for {
			serverSideTorrentInfo.BC = t.BytesCompleted()
			time.Sleep(time.Second)
		}
	}()
}

func statsEnabled() bool {
	if flags.Stats == nil {
		return flags.Debug
	}
	return *flags.Stats
}

func outputStats(cl *torrent.Client) {
	if !statsEnabled() {
		return
	}
	expvar.Do(func(kv expvar.KeyValue) {
		fmt.Printf("%s: %s\n", kv.Key, kv.Value)
	})
	cl.WriteStatus(os.Stdout)
}

func clientSignalHandler(client *torrent.Client) {
	<-clientCleanupSignal
	Info.Println("close signal received\n")
	client.Close()
	RemoveContents(mediaDir)
	clientCleanupFinished <- struct{}{}
}

func torrentErrorHandler(client *torrent.Client) {
	err := <-torrentError
	Error.Printf("torrent error: %v\n", err)
	client.Close()
	torrentErrorForHTTPHandler = err
}