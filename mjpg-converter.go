package main

import (
	"os"
	"bytes"
	"flag"
	"log"
	"runtime/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var boundary = flag.String("boundary", "--BOUNDARY", "boundary marker")
var magic = []byte { 0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 0x4a, 0x46, 0x49, 0x46, 0x00 }
var work = make([]byte, len(magic))
var indices = make([]int, 0, 4)
var total = 0

func processBuffer(data []byte, n int) {
	indices = indices[0:0]
	for i := 0; i < n; i++ {
		work = work[1:]
		work = append(work, data[i])
		if bytes.Compare(work, magic) == 0 {
			start := i-len(magic)+1
			if start < 0 {
				start = 0
			}
			indices = append(indices, start)
		}
	}
	total += len(indices)
	if len(indices) > 0 {
		for i := 0; i < len(indices); i++ {
			start := indices[i]
			end := start
			if i == len(indices)-1 {
				end = n
			} else {
				end = indices[i+1]
			}
			if i == 0 && start != 0 {
				os.Stdout.Write(data[:start])
			}
			os.Stdout.Write([]byte(*boundary))
			os.Stdout.Write(data[start:end])
		}
	} else {
		os.Stdout.Write(data);
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

	data := make([]byte, 4096)
	n, _ := os.Stdin.Read(data)
	for n > 0 {
		processBuffer(data, n)
		n, _ = os.Stdin.Read(data)
	}
}