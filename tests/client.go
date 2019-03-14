package main

import (
	"gopkg.in/Graylog2/go-gelf.v2/gelf"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

type Message struct {
	short      string
	file       string
	line       int
	level      int
	extra      map[string]interface{}
	stacktrace string
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
func main() {
	gelfWriter, err := gelf.NewUDPWriter(":12201")
	if err != nil {
		log.Fatal("Error: ", err)
	}
	// log to both stderr and graylog2
	log.SetOutput(io.MultiWriter(os.Stderr, gelfWriter))
	//log.Printf("logging to stderr & graylog2@'%s'", ":12201")
	i := 0
	for j := 0; j < 100000; j++ {
		file, line := getFileAndLine()
		replacer := strings.NewReplacer("\n", "; ", "\t", "")
		stacktrace := replacer.Replace(string(debug.Stack()))
		err := gelfWriter.WriteMessage(createGelfMessage(&Message{
			short:      "hello graylog client" + RandStringRunes(7111),
			file:       file,
			line:       line,
			stacktrace: stacktrace,
		}))
		if err != nil {
			log.Println("Error: ", err)
		}
		i++
		log.Println(i)
		//os.Exit(0)
	}
}
func getFileAndLine() (string, int) {
	_, file, line, _ := runtime.Caller(4)
	isExportCall := strings.Contains(file, "export.go")
	isChainedCall := strings.Contains(file, "chainBuilder.go")

	if isExportCall || isChainedCall {
		_, file, line, _ = runtime.Caller(5)
	}

	return filepath.Base(file), line
}

var host, _ = os.Hostname()
var env = os.Getenv("ENV_ID")
var region = os.Getenv("REGION_NAME")

func createGelfMessage(m *Message) *gelf.Message {
	msg := &gelf.Message{
		Version:  "1.1",
		Host:     host,
		TimeUnix: float64(time.Now().Unix()),
		Level:    int32(m.level),
		Facility: "Facility",
	}

	if m.short != "" {
		msg.Short = m.short
	}

	msg.Extra = map[string]interface{}{
		"_env":        env,
		"_region":     region,
		"_file":       m.file,
		"_line":       m.line,
		"_stacktrace": m.stacktrace,
	}

	if len(m.extra) > 0 {
		for k, v := range m.extra {
			msg.Extra["_"+k] = v
		}
	}

	return msg
}
