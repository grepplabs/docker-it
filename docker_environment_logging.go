package dockerit

import (
	"fmt"
	"log"
	"os"
)

type logWriter struct{ *log.Logger }

// implements io.Writer interface
func (w logWriter) Write(b []byte) (int, error) {
	w.Printf("%s", b)
	return len(b), nil
}

func stdoutWriter(prefix string) *logWriter {
	if prefix != "" {
		return &logWriter{log.New(os.Stdout, fmt.Sprintf("%s: ", prefix), 0)}
	}
	return &logWriter{log.New(os.Stdout, "", 0)}
}

type logger struct {
	Info  *log.Logger
	Error *log.Logger
}

func newLogger() *logger {
	infoLogger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	errorLogger := log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime)
	return &logger{Info: infoLogger, Error: errorLogger}
}
