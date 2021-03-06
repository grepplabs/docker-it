package dockerit

import (
	"fmt"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
)

const (
	testImage = "busybox"
)

func TestDockerCommands(t *testing.T) {
	a := assert.New(t)

	dc, err := newDockerClient()
	a.Nil(err)

	_, err = dc.GetImageByName(testImage)
	a.Nil(err)

	sum, err := dc.GetImageByName("this_image_does_not_exist")
	a.Nil(sum)
	a.Nil(err)

	err = dc.PullImage(testImage)
	a.Nil(err)

	err = dc.PullImage("this_image_does_not_exist")

	exposedPorts := make(nat.PortSet)
	port, err := nat.NewPort("tcp", strconv.Itoa(4771))
	a.Nil(err)

	exposedPorts[port] = struct{}{}

	portSpecs := make([]string, 0)
	portSpec := fmt.Sprintf("%s:%d:%d/%s", "127.0.0.1", 4711, 4712, "tcp")
	portSpecs = append(portSpecs, portSpec)

	salt := uuid.New().String()
	salt = salt[len(salt)-12:]
	containerName := fmt.Sprintf("test-busybox-%s", salt)

	env := []string{}
	cmd := []string{}
	binds := []string{}
	dnsServer := ""
	containerID, err := dc.CreateContainer(containerName, testImage, env, portSpecs, cmd, binds, dnsServer)
	a.Nil(err)

	container, err := dc.GetContainerByID(containerID)
	a.Nil(err)
	a.NotNil(container)

	container, err = dc.GetContainerByID("unknown-id")
	a.Nil(err)
	a.Nil(container)

	err = dc.StartContainer(containerID)
	a.Nil(err)

	err = dc.StartContainer(containerID)
	a.Nil(err)

	reader, err := dc.ContainerLogs(containerID, false)
	a.Nil(err)
	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, reader)

	err = dc.StopContainer(containerID)
	a.Nil(err)

	err = dc.StopContainer(containerID)
	a.Nil(err)

	reader, err = dc.ContainerLogs(containerID, false)
	a.Nil(err)
	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, reader)

	err = dc.RemoveContainer(containerID)
	a.Nil(err)

	_, err = dc.ContainerLogs(containerID, false)
	a.NotNil(err)

	err = dc.RemoveImageByName(testImage)
	if err != nil {
		// do not check error force remove is not used and the images can be used be another container
		fmt.Println("WARNING: Remove by name error: ", err)
	}

	// close the client
	dc.Close()
}
