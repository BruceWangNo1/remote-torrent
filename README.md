[![Go Report Card](https://goreportcard.com/badge/github.com/BruceWangNo1/remote-torrent)](https://goreportcard.com/report/github.com/BruceWangNo1/remote-torrent)
[![GoDoc](https://godoc.org/github.com/brucewangno1/remote-torrent?status.svg)](https://godoc.org/github.com/brucewangno1/remote-torrent)
# remote-torrent

Download Torrent Remotely and Retrieve Files Over HTTP at Full Speed without ISP Torrent Limitation.
This repository is an extension to [anacrolix/torrent](https://github.com/anacrolix/torrent) project to download torrent remotely to your server like your VPS and retrieve the downloaded files to your local machine over HTTP. As we know, some ISPs implement torrent limitation and also sometimes downloading torrent locally does not work well. So this is why I started this project.

## Features
1. Once the server is set up, the user does not have to interact with the server manually. When the torrent download is finished, the client will automatically retrieve the downloads and remove intermedia downloads on the server side.

2. If the user does not wish to continue at any moment, Ctrl + C will be captured by client which immediately sends request to clean up the server (shutdown the current torrent task and possibly remove downloaded contents) and then shut down itself gracefully.
3. Client http partial request implemented. For example, if the user already has a partial file from the previous unfinished download process, the new download process picks up where it left off instead of restarting the download all over again.

## Notice
Before installation, I think it is better for you to know that there is great web implementation [jpillora/cloud-torrent](https://github.com/jpillora/cloud-torrent) providing similar functionalities while offering great visual aids. You may still want to use this project if you want a simpler approach with just a single command on client side that takes care of everything for you.

## Installation

Install the library package with `go get github.com/brucewangno1/remote-torrent/rt`.

## Enviroment Setup

Create a directory `/root/media` on your server to save files downloaded. Do not store anything else in that directory because after every client torrent download request all files in that directory will be purged.

## Command Usage

- Server side: run the below command in `/root/media` directory
`rt server yourPortNumber username:password`  
As most user would like to keep the process running after logging out ssh session, the following command will meets their needs:
`nohup rt server yourPortNumber username:password > /root/rtserver.log &`

- Client side: `rt client username:password yourServerIP:yourServerPortNumber "desiredMagnetLink"`
You may find downloaded files in the current directory after `rt client` finishes.

## Example
`rt server 8899 ryan:see_a_penny`

`rt client ryan:see_a_penny 192.168.100.100:8899 "magnet:?xt=urn:btih:194257a7bf4eaea978f4b5b7fbd3b4efcdd99e43&dn=ubuntu-18.04.3-live-server-amd64.iso"`

## To-do List



## Contributing

1. Fork it!
2. Create your feature branch: `git checkout -b my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin my-new-feature`
5. Submit a pull request :D

If you have any questions about this project, feel free to create an issue.
