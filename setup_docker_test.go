package dockerit

import (
	"os"
	"testing"
)

func init() {
	// ensure docker API version
	SetDefaultDockerAPIVersion()
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
