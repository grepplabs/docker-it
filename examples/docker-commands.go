package main

import (
	"fmt"
	dit "github.com/cloud-42/docker-it"
	"github.com/docker/go-connections/nat"
	"strconv"
)

func main() {
	dc, err := dit.NewDockerClient()
	if err != nil {
		panic(err)
	}
	imageSummary, err := dc.GetImageByName("redis")
	if err != nil {
		panic(err)
	}
	fmt.Println(imageSummary)

	err = dc.PullImage("redis")
	if err != nil {
		panic(err)
	}

	exposedPorts := make(nat.PortSet)
	port, err := nat.NewPort("tcp", strconv.Itoa(4771))
	if err != nil {
		panic(err)
	}
	exposedPorts[port] = struct{}{}

	portSpecs := make([]string, 0)
	portSpec := fmt.Sprintf("%s:%d:%d/%s", "127.0.0.1", 4711, 4712, "tcp")
	portSpecs = append(portSpecs, portSpec)

	env := []string{}
	dc.RemoveContainer("my-redis")
	id, err := dc.CreateContainer("my-redis", "redis", env, portSpecs)
	if err != nil {
		panic(err)
	}
	fmt.Println(id)

}
