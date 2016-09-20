# grumpy-pi-mjpg

## Simple HTTP server to provide streams of Raspberry Pi camera using MJPEG codec

This server has been designed to be the simplest possible to start streaming video
from your Raspberry Pi camera module. It uses the MJPEG codec so doesn't have sound
but supports all options of `raspivid`. The server reads the data using `stdin` which
means you need to use the `-o -` option and pipe it to the server.

### Usage

Type (those options are the minimum required) :

    raspivid -cd MJPEG -t 0 -o - | mjpg-server

Of course, `raspivid` can take any options like so :

    raspivid -cd MJPEG -w 640 -h 360 -fps 10 -t 0 -n -o - | mjpg-server
