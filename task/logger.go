package task

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

const (
	logFileSize = 5 * 1024 * 1024
	jsonExt     = "json"
	logExt      = "log"
)

var (
	Log     *log.Logger
	logFile *os.File
	//logJsonFile *os.File
)

type writer struct {
	Writers    []io.Writer
	TimeFormat string
}

func (w writer) Write(b []byte) (n int, err error) {
	for i := 0; i < len(w.Writers); i++ {
		n, err = w.Writers[i].Write(append([]byte(time.Now().Format(w.TimeFormat)), b...))

		if err != nil {
			break
		}
	}

	return

	//return w.Writer.Write(append([]byte(time.Now().Format(w.TimeFormat)), b...))
}

// Marshalling state into json file
func (st *State) logJson() (err error) {
	st.Time = time.Now().Format(time.RFC3339)
	var b []byte
	b, err = json.Marshal(st)
	if err != nil {
		return
	}

	jsonLogFileName := joinPaths(logDir, strings.ToLower(dtsAppName)+"."+jsonExt)
	err = writeJsonLog(jsonLogFileName, b)

	return
}

// Write []byte to file and close it
func writeJsonLog(file string, b []byte) error {
	jsonLogFile, err := openFile(file)
	defer jsonLogFile.Close()
	if err != nil {
		return err
	}

	b = append(b, '\n')
	_, err = jsonLogFile.Write(b)
	if err != nil {
		return err
	}

	return nil
}

// Create logger for
func setupLogger() {
	fPath := joinPaths(logDir, strings.ToLower(dtsAppName)+"."+logExt)
	var err error
	logFile, err = openFile(fPath)
	if err != nil {
		log.Fatal(err)
	}

	fi, err := logFile.Stat()
	if err != nil {
		log.Fatal(err)
	}

	if fi.Size() > logFileSize {
		var dst string
		dst, err = rotate(logFile.Name())
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("log rotated to dst=%s\n", dst)
	}

	writers := io.MultiWriter(&writer{
		Writers:    []io.Writer{logFile, os.Stderr},
		TimeFormat: time.RFC3339 + " ",
	})

	Log = log.New(writers, "", 0)
}
