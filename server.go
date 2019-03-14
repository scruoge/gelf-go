package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	go monitor()
	go printer()
	go accumulate()
	go consume()
	go worker()
	go server()
	//Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
}

var counter uint64
var res = make(chan int)

func monitor() {
	for {
		select {
		case <-time.After(time.Millisecond * 1000):
			final := atomic.SwapUint64(&counter, 0)
			fmt.Println("msg/sec: ", final)
		}
	}
}

var queue [][]byte
var mutex = sync.Mutex{}
var resultChannel = make(chan []byte)
var sourceChannel = make(chan [][]byte)
var queueChannel = make(chan []byte, 100)
var accumulatorChannel = make(chan Parsed)

func server() {
	//sConn, err := reuseport.ListenPacket("udp", ":12201")
	sAddr, err := net.ResolveUDPAddr("udp", ":12201")
	if err != nil {
		log.Fatal("Error: ", err)
	}
	sConn, err := net.ListenUDP("udp", sAddr)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	if err := sConn.SetReadBuffer(26214400); err != nil {
		log.Fatal("Error: ", err)
	}
	defer sConn.Close()
	quit := make(chan struct{})
	for i := 0; i < runtime.NumCPU(); i++ {
		go listen(sConn, quit)
	}
	<-quit // hang until an error
}

var bufferPool = sync.Pool{
	New: func() interface{} { return make([]byte, 1500) },
}

func listen(connection *net.UDPConn, quit chan struct{}) {
	rlen, _, err := 0, new(net.UDPAddr), error(nil)
	for err == nil {
		msg := bufferPool.Get().([]byte)
		rlen, _, err = connection.ReadFrom(msg[0:])
		atomic.AddUint64(&counter, 1)
		go queuePut(rlen, msg)
	}
	fmt.Println("listener failed - ", err)
	quit <- struct{}{}
}
func queuePut(rlen int, buf []byte) {
	var msg = make([]byte, rlen)
	copy(msg, buf[:rlen])
	bufferPool.Put(buf)
	queueChannel <- msg
	for {
		select {
		case msg := <-queueChannel:
			mutex.Lock()
			queue = append(queue, msg)
			mutex.Unlock()
		}
	}
}

func consume() {
	for {
		select {
		case <-time.After(time.Millisecond * 10):
			mutex.Lock()
			if len(queue) != 0 {
				queueCopy := queue[:len(queue)]
				curQueueLength := len(queueCopy)
				//res <- len(queue)
				queue = queue[curQueueLength:]
				sourceChannel <- queueCopy
			}
			mutex.Unlock()
		}
	}
}
func worker() {
	var chunkByte = []byte{0xef}
	for {
		select {
		case msgSlice := <-sourceChannel:
			for _, msg := range msgSlice {
				result := detectMessageType(msg)
				if bytes.Equal(result, chunkByte) {
					accumulatorChannel <- extract(msg)
				} else {
					resultChannel <- result
				}
			}
		}
	}
}

func accumulate() {
	cacheChunk := map[string][]Parsed{}
	for {
		select {
		case chunk := <-accumulatorChannel:
			id := chunk.id
			cacheChunk[id] = append(cacheChunk[id], chunk)
			if chunk.count == len(cacheChunk[id]) {
				buffer := bytes.NewBuffer([]byte{})
				sort.Slice(cacheChunk[id], func(i, j int) bool {
					return cacheChunk[id][i].number < cacheChunk[id][j].number
				})
				for _, item := range cacheChunk[id] {
					if _, err := buffer.Write(item.data); err != nil {
						log.Println("Error: ", err)
					}
				}
				delete(cacheChunk, id)
				resultChannel <- detectMessageType(buffer.Bytes())
			} else {
				resultChannel <- []byte{}
			}
		}
	}
}

func printer() {
	for {
		select {
		case msg, ok := <-resultChannel:
			if !ok {
				resultChannel = nil
				return
			}
			if string(msg) != "" {
				log.Printf("%s\n", msg)
			}
		}
	}
}

type Parsed struct {
	id     string
	data   []byte
	number int
	count  int
}

func extract(chunked []byte) Parsed {
	chunk := Parsed{
		id:     fmt.Sprintf("%x", chunked[2:10]),
		number: int(chunked[10:11][0]),
		count:  int(chunked[11:12][0]),
	}
	chunk.data = make([]byte, len(chunked[12:]))
	copy(chunk.data, chunked[12:])
	return chunk
}
