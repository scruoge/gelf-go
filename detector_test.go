package main

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"testing"
	"time"

	"github.com/duythinht/gelf"
	"github.com/duythinht/gelf/chunk"
)

type Test struct {
	in  []byte
	out string
}

var tests = []Test{
	{makeGzipMsg("test gzip"), "test gzip"},
	{makeZlibMsg("test zlib"), "test zlib"},
	{[]byte{0x00}, ""},
}

func TestCheck(t *testing.T) {
	for _, test := range tests {
		var out = string(detectMessageType(test.in))
		if out != test.out {
			t.Error("unexpexted: " + out)
		}
	}
}

func TestError(t *testing.T) {
	expected := []byte{0x1f, 0x8b}
	detectMessageType(expected)
}

func BenchmarkChunked(b *testing.B) {
	expected := []byte{0xef}
	var buffers = makeGelfChunked("some msg")
	for n := 0; n < b.N; n++ {
		for _, chunked := range buffers {
			out := detectMessageType(chunked)
			if bytes.Compare(out, expected) != 0 {
				b.Error("unexpexted: " + string(out))
			}
		}
	}
}

func makeGelfChunked(input string) [][]byte {
	message := gelf.Create(input).
		SetTimestamp(time.Now().Unix()).
		SetFullMessage("This is full message").
		SetLevel(3).
		SetHost("chat Server").
		ToJSON()
	ZippedMessage := chunk.ZipMessage(message)
	var MaxChunkSize = 50
	var buffers [][]byte
	//fmt.Println(len(ZippedMessage))
	if len(ZippedMessage) > MaxChunkSize {
		buffers = chunk.GetGelfChunks(ZippedMessage, MaxChunkSize)
	}
	return buffers
}

func makeGzipMsg(input string) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write([]byte(input))
	w.Close()
	return b.Bytes()
}

func makeZlibMsg(input string) []byte {
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write([]byte(input))
	w.Close()
	return b.Bytes()
}

func benchmarkCheck(i int, b *testing.B) {
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		var out = string(detectMessageType(tests[i].in))
		if out != tests[i].out {
			b.Error("unexpexted: " + out)
		}
	}
}

func BenchmarkCase0(b *testing.B)  {
	benchmarkCheck(0, b )
}
func BenchmarkCase1(b *testing.B)  {
	benchmarkCheck(1, b )
}
func BenchmarkCase2(b *testing.B)  {
	benchmarkCheck(2, b )
}