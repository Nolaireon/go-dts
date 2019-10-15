package task

import (
	"encoding/json"
	"errors"
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
	Log           *log.Logger
	logFile       *os.File
	ErrShortWrite = errors.New("short write")
)

type writer struct {
	Writers    []io.Writer
	TimeFormat string
}

// Custom realization of writer interface aimed on writing same data to every writer in the given slice
func (w writer) Write(b []byte) (n int, err error) {
	for i := 0; i < len(w.Writers); i++ {
		n, err = w.Writers[i].Write(append([]byte(time.Now().Format(w.TimeFormat)), b...))

		if err != nil {
			break
		}
	}

	return
}

// Marshalling state into json file
func (st *State) logJson() (err error) {
	st.Time = time.Now().Format(time.RFC3339)
	var b []byte
	b, err = json.Marshal(st)
	if err != nil {
		return
	}

	jsonLogFileName := joinPaths("logs", strings.ToLower(dtsAppName)+"."+jsonExt)
	err = writeJsonLog(jsonLogFileName, b)

	return
}

// writeJsonLog
func writeJsonLog(file string, b []byte) error {
	jsonLogFile, err := openLogFile(file)
	if err != nil {
		return err
	}

	b = append(b, '\n')
	n, err := jsonLogFile.Write(b)
	if err == nil && n < len(b) {
		err = ErrShortWrite
	}

	if err1 := jsonLogFile.Close(); err == nil {
		err = err1
	}

	return err
}

// Create and setting up logger responsible to write plain log in two writers: stderr and file.
// There is no need to write additional functions to manually close logger writers on app termination events,
// log package will take care of correct closure for every writer that was passed to logger, including any emergency exits
func setupLogger() {
	dtsDir, err := getExecutablePath()
	if err != nil {
		log.Fatal(err)
	}

	err = os.Chdir(dtsDir)
	if err != nil {
		log.Fatal(err)
	}

	fPath := joinPaths("logs", strings.ToLower(dtsAppName)+"."+logExt)
	logFile, err = openLogFile(fPath)
	if err != nil {
		log.Fatal(err)
	}

	fi, err := logFile.Stat()
	if err != nil {
		log.Fatal(err)
	}

	if fi.Size() > logFileSize {
		if err = logFile.Close(); err != nil {
			log.Fatal(err)
		}

		var dst string
		dst, err = rotate(logFile.Name())
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("log rotated to dst=%s\n", dst)

		logFile, err = openLogFile(fPath)
		if err != nil {
			log.Fatal(err)
		}
	}

	writers := io.MultiWriter(&writer{
		Writers:    []io.Writer{logFile, os.Stderr},
		TimeFormat: time.RFC3339 + " ",
	})

	Log = log.New(writers, "", 0)
}
