# remote-torrent

Download Torrent Remotely and Retrieve Files Over HTTP at Full Speed without ISP Torrent Limitation.

This repository is an extension to [anacrolix/torrent](https://github.com/anacrolix/torrent) project to download torrent remotely to your server like your VPS and retrieve the downloaded files to your local machine over HTTP. As we know, some ISPs implement torrent limitation and also sometimes downloading torrent locally does not work well. So this is why I start this project.

## Installation

Install the library package with `go get github.com/brucewangno1/remote-torrent/rt`.

## Development Enviroment

Ubuntu on server side and MacOS on client side.

## Requirement

Install [anacrolix/torrent](https://github.com/anacrolix/torrent) first on your server and wget on your client via brew.

## Command Usage

Server side: `rt server yourPortNumber username:password`
Client side: `rt client username:password yourServerIP:yourServerPortNumber magnetLink`

## Contribution

Looking forward to your contribution.
