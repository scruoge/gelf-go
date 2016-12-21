package main

import (
	"bytes"
	"fmt"
	"net"
)

func main() {
	sAddr, err := net.ResolveUDPAddr("udp", ":12201")
	if err != nil {
		fmt.Println("Error: ", err)
	}
	sConn, err := net.ListenUDP("udp", sAddr)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	defer sConn.Close()

	buf := make([]byte, 8192)

	// for background
	var c chan []byte = make(chan []byte, 2)
	var nn chan []byte = make(chan []byte)
	var ac chan Parsed = make(chan Parsed, 2)

	go printer(c, nn)
	go accum(c, ac)
	i := 0
	for {
		i++
		n, err := sConn.Read(buf)
		// process msg
		go worker(c, ac, nn, i)
		nn <- buf[0:n]
		// process result
		data := <-nn

		if string(data) != "" {
			fmt.Printf("%s\n", data)
		}

		if err != nil {
			fmt.Println("Error: ", err)
		}
	}

}

func worker(c chan []byte, ac chan Parsed, nn chan []byte, numer int) {
	var chunkByte []byte = []byte{0xef}
	msg := <-nn
	result := check(msg)
	if bytes.Equal(result, chunkByte) {
		ac <- extract(msg)
	} else {
		c <- result
	}
}

func accum(c chan []byte, ac chan Parsed) {
	buffer := map[string][]Parsed{}
	for {
		chunk := <-ac
		buffer[chunk.id] = append(buffer[chunk.id], chunk)

		if chunk.count == len(buffer[chunk.id]) {
			var msg string
			for _, item := range buffer[chunk.id] {
				msg += item.data
			}
			delete(buffer, chunk.id)
			c <- check([]byte(msg))
		} else {
			c <- []byte{}
		}
	}
}
func printer(c chan []byte, nn chan []byte) {
	for {
		msg := <-c
		nn <- msg
	}
}

type Parsed struct {
	id     string
	data   string
	number int
	count  int
}

func extract(chunked []byte) Parsed {
	var chunk = Parsed{}
	chunk.id = string(chunked[2:10])
	chunk.number = int(chunked[10:11][0])
	chunk.count = int(chunked[11:12][0])
	chunk.data = string(chunked[12:])
	return chunk
}
