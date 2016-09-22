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
	fmt.Fprintf(os.Stderr, format, args...)
}

type ImageCallback func(image []byte)

func ProcessData(data []byte, n int, callback ImageCallback) {
	start := 0
	for i := 0; i < n; i++ {
		work = work[1:]
		work = append(work, data[i])
		if bytes.Compare(work, magic) == 0 {
			buffer.Write(data[start:i])
			if buffer.Len() > 0 {
				end := buffer.Len() - len(magic) + 1
				image := buffer.Bytes()[:end]
				rest := buffer.Bytes()[end:]
				callback(image)
				buffer.Reset()
				buffer.Write(rest)
				start = i
			}
		}
	}
	buffer.Write(data[start:n])
	total += n
}
