package converter

import (
	"os"
	"flag"
	"fmt"
	"io"
	"sync"
	"image"
	"image/color"
	"image/jpeg"
	"bytes"
)

var boundary = flag.String("boundary", "--BOUNDARY", "boundary marker")
var buffersize = flag.Int("buffersize", 4096, "buffer size")
var data = &threadSafeSlice{
	workers: make([]*worker, 0, 1),
}

type worker struct {
	source chan []byte
}

type threadSafeSlice struct {
	sync.Mutex
	workers []*worker
}

func (s *threadSafeSlice) Push(w *worker) {
	s.Lock()
	defer s.Unlock()

	s.workers = append(s.workers, w)
}

func (s *threadSafeSlice) Iter(routine func(*worker)) {
	s.Lock()
	defer s.Unlock()

	for _, worker := range s.workers {
		routine(worker)
	}
}

func broadcaster(ch chan []byte) {
	for {
		msg := <-ch
		data.Iter(func(w *worker) { w.source <- msg })
	}
}

func testgen(ch chan []byte) {
	buffer := new(bytes.Buffer)
	imgbuffer := new(bytes.Buffer)
	m := image.NewRGBA(image.Rect(0, 0, 256, 256))
	index := 0
	for {
		// generate
		x := index % m.Bounds().Max.X
		y := index / m.Bounds().Max.X
		m.Set(x, y, color.RGBA{255, 0, 255, 255})
		jpeg.Encode(imgbuffer, m, nil)

		// output
		if index > 0 {
			buffer.Write([]byte("\r\n"))
		}
		fmt.Fprintf(buffer, "%s\r\n", *boundary)
		fmt.Fprintf(buffer, "Content-Type: image/jpeg\r\n")
		fmt.Fprintf(buffer, "Content-Length: %d\r\n", imgbuffer.Len())
		buffer.Write([]byte("\r\n"))
		imgbuffer.WriteTo(buffer)
		cp := make([]byte, buffer.Len())
		copy(cp, buffer.Bytes())
		ch <- cp
		buffer.Reset()
		index++
	}
}

func generator(ch chan []byte) {
	readbuffer := make([]byte, *buffersize)
	writebuffer := new(bytes.Buffer)
	for {
		n, err := os.Stdin.Read(readbuffer)
		if err != nil {
			break
		}
		ProcessData(readbuffer, n, func(image []byte) {
			writebuffer.Write([]byte("\r\n"))
			writebuffer.Write([]byte("--BOUNDARY\r\n"))
			writebuffer.Write([]byte("Content-Type: image/jpeg\r\n"))
			fmt.Fprintf(writebuffer, "Content-Length: %d\r\n", len(image))
			writebuffer.Write([]byte("\r\n"))
			writebuffer.Write(image)
			cp := make([]byte, writebuffer.Len())
			copy(cp, writebuffer.Bytes())
			ch <- cp
			writebuffer.Reset()
		})
	}
}

func Broadcast() {
	c := make(chan []byte)
	go broadcaster(c)
	go generator(c)
	// go testgen(c)
}

func StreamTo(w io.Writer) {
	wk := &worker{
		source: make(chan []byte),
	}
	data.Push(wk)
	for {
		w.Write(<-wk.source)
	}
}
