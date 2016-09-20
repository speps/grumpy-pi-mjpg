package main

import "./converter"

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"runtime/pprof"
)

const defaultTemplate = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>{{.Title}}</title>
	</head>
	<body>
		<h1>Live</h1>
		<p>{{.Len}} viewers when page loaded</p>
		<img src="{{.Stream}}"/>
	</body>
</html>`

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var port = flag.String("port", ":8081", "HTTP listen port")

func writeStreamOutput(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "multipart/x-mixed-replace;boundary=--BOUNDARY")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	if c, ok := w.(http.CloseNotifier); ok {
		converter.StreamTo(w, c.CloseNotify())
	}
}

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}
	t, err := template.New("webpage").Parse(defaultTemplate)
	check(err)

	data := struct {
		Title  string
		Len int
		Stream string
	}{
		Title:  "MJPG Server",
		Stream: "/stream",
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data.Len = converter.Len()
		err = t.Execute(w, data)
		check(err)
	})

	http.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		writeStreamOutput(w)
	})

	converter.Broadcast()
	log.Fatal(http.ListenAndServe(*port, nil))
}
