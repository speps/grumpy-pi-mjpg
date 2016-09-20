# grumpy-pi-mjpg [![Build status](https://ci.appveyor.com/api/projects/status/660s288u7re3881f?svg=true)](https://ci.appveyor.com/project/speps/grumpy-pi-mjpg)

### Simple HTTP server to provide streams of Raspberry Pi camera using MJPEG codec

This server has been designed to be the simplest possible to start streaming video
from your Raspberry Pi camera module. It uses the MJPEG codec so doesn't have sound
but supports all options of `raspivid`. The server reads the data using `stdin` which
means you need to use the `-o -` option and pipe it to the server.

The program `raspivid` can output MJPEG but it doesn't conform to what a browser
expects in a webpage. Instead, it outputs a JPEG image back to back. This is easy to
split and prepare each image to be a proper MJPEG which includes the right HTTP headers.

### How to download

Type this in your terminal :

    > wget https://dl.bintray.com/speps/grumpy-pi-mjpg/mjpg-server

### How to use

Type (those options are the minimum required) :

    raspivid -cd MJPEG -t 0 -o - | mjpg-server

Of course, `raspivid` can take any options like so :

    raspivid -cd MJPEG -w 640 -h 360 -fps 10 -t 0 -n -o - | mjpg-server

### How to compile

**NOTE**: prefer binary releases, see how to download above

You need to install Golang to compile : https://golang.org/doc/install

There are 2 options :

* Downloading the `linux-armv6l` version and following the instructions
* Download for another operating system and cross-compiling
    * For this, set `GOOS=linux` and `GOARCH=arm`

Once this is done, run this :

    > go build grumpy-pi-mjpg/mjpg-server.go

This generates the `mjpg-server` executable.
