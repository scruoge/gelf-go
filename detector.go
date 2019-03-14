package main

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"io"
	"log"
)

func detectMessageType(testBytes []byte) []byte {
	var readCloser io.ReadCloser
	var err error
	if testBytes[0] == 0x1f && testBytes[1] == 0x8b {
		readCloser, err = gzip.NewReader(bytes.NewBuffer(testBytes))
	} else if testBytes[0] == 0x78 && testBytes[1] == 0x9c {
		readCloser, err = zlib.NewReader(bytes.NewBuffer(testBytes))
	} else if testBytes[0] == 0x1e && testBytes[1] == 0x0f { // gelf chunk
		return []byte{0xef}
	} else {
		log.Println("Warn: MessageType unknown")
		return []byte{}
	}
	if err != nil {
		log.Println("Error: ", err)
		return []byte{}
	}
	defer readCloser.Close()
	out := &bytes.Buffer{}
	if _, err := out.ReadFrom(readCloser); err != nil {
		log.Println("Error: ", err)
		return []byte{}
	}
	return out.Bytes()
}
