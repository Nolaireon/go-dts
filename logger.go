package main

import (
	"io"
	"log"
	"os"
	"time"
)

type writer struct {
	io.Writer
	timeFormat string
}

func (w writer) Write(b []byte) (n int, err error) {
	return w.Writer.Write(append([]byte(time.Now().Format(w.timeFormat)), b...))
}

var Logger = log.New(&writer{os.Stdout, time.RFC3339 + " "}, "", 0)
