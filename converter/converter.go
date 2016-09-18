package converter

import (
	"bytes"
	"fmt"
	"os"
)

var magic = []byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 0x4a, 0x46, 0x49, 0x46, 0x00}
var work = make([]byte, len(magic))
var indices = make([]int, 0, 4)
var total = 0
var buffer = new(bytes.Buffer)

func debug(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args)
}

type ImageCallback func(image []byte)

func ProcessData(data []byte, n int, callback ImageCallback) {
	indices = indices[0:0]
	for i := 0; i < n; i++ {
		work = work[1:]
		work = append(work, data[i])
		if bytes.Compare(work, magic) == 0 {
			start := i - len(magic) + 1
			if start < 0 {
				start = 0
			}
			indices = append(indices, start)
		}
	}
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
				buffer.Write(data[:start])
			}
			callback(buffer.Bytes())
			buffer.Reset()
			buffer.Write(data[start:end])
			total++
		}
	} else {
		buffer.Write(data[:n])
	}
}
