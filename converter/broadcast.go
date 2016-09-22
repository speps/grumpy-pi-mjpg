package converter

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"os"
	"sync"
)

var boundary = flag.String("boundary", "--BOUNDARY", "boundary marker")
var buffersize = flag.Int("buffersize", 4096, "buffer size")
var data = &threadSafeSlice{
	workers: make([]*worker, 0, 1),
}

type worker struct {
	source chan []byte
	first  bool
	done   bool
}

type threadSafeSlice struct {
	sync.Mutex
	workers []*worker
}

func (s *threadSafeSlice) Len() int {
	s.Lock()
	defer s.Unlock()

	return len(s.workers)
}

func (s *threadSafeSlice) Push(w *worker) {
	s.Lock()
	defer s.Unlock()

	s.workers = append(s.workers, w)
}

func (s *threadSafeSlice) Iter(routine func(*worker) bool) {
	s.Lock()
	defer s.Unlock()

	for i := len(s.workers) - 1; i >= 0; i-- {
		remove := routine(s.workers[i])
		if remove {
			s.workers[i] = nil
			s.workers = append(s.workers[:i], s.workers[i+1:]...)
		}
	}
}

func broadcaster(ch chan []byte) {
	for {
		msg := <-ch
		data.Iter(func(w *worker) bool {
			if w.done {
				fmt.Fprintf(os.Stderr, "done %p\n", w)
				close(w.source)
				return true
			} else {
				w.source <- msg
				return false
			}
		})
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
			fmt.Fprintf(os.Stderr, "gen read err %v\n", err)
			break
		}
		ProcessData(readbuffer, n, func(image []byte) {
			// header
			fmt.Fprintf(writebuffer, "%s\r\n", *boundary)
			writebuffer.Write([]byte("Content-Type: image/jpeg\r\n"))
			fmt.Fprintf(writebuffer, "Content-Length: %d\r\n", len(image))
			writebuffer.Write([]byte("\r\n"))
			// image
			writebuffer.Write(image)
			// make a copy to send over channel
			cp := make([]byte, writebuffer.Len())
			copy(cp, writebuffer.Bytes())
			writebuffer.Reset()
			// send!
			ch <- cp
		})
	}
}

func Broadcast() {
	c := make(chan []byte)
	go broadcaster(c)
	go generator(c)
	// go testgen(c)
}

func Len() int {
	return data.Len()
}

func StreamTo(w io.Writer, closed <-chan bool) {
	wk := &worker{
		source: make(chan []byte),
		first:  true,
	}
	fmt.Fprintf(os.Stderr, "created %p\n", wk)
	data.Push(wk)
loop:
	for {
		select {
		case s, ok := <-wk.source:
			if !ok {
				break loop
			}
			if !wk.first {
				w.Write([]byte("\r\n"))
			} else {
				wk.first = false
			}
			w.Write(s)
		case <-closed:
			wk.done = true
		}
	}
}
