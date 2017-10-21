# remote-torrent

Download Torrent Remotely and Retrieve Files Over HTTP at Full Speed without ISP Torrent Limitation.
This repository is an extension to [anacrolix/torrent](https://github.com/anacrolix/torrent) project to download torrent remotely to your server like your VPS and retrieve the downloaded files to your local machine over HTTP. As we know, some ISPs implement torrent limitation and also sometimes downloading torrent locally does not work well. So this is why I start this project.

## Notice
Before installation, I think it is better for you to know that there is great web implementation [jpillora/cloud-torrent](https://github.com/jpillora/cloud-torrent) providing similar functionalities while offering great visual aids. You may still want to use this project if you want a simpler usage approach with just a single command on client side that takes care of everything for you.

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
