package dockerit

import (
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestLogWriterWithPrefix(t *testing.T) {

	osStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	writer := stdoutWriter("my-container")
	reader := strings.NewReader("my content")
	io.Copy(writer, reader)

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = osStdout

	a := assert.New(t)
	a.Equal("my-container: my content\n", string(out))
}

func TestLogWriterWithoutPrefix(t *testing.T) {

	osStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	writer := stdoutWriter("")
	reader := strings.NewReader("my content")
	io.Copy(writer, reader)

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = osStdout

	a := assert.New(t)
	a.Equal("my content\n", string(out))
}

func TestLogger(t *testing.T) {

	osStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := newLogger()
	logger.Info.Println("info logger")
	logger.Error.Println("error logger")

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = osStdout

	a := assert.New(t)
	a.Contains(string(out), "INFO: ")
	a.Contains(string(out), "ERROR: ")

}
