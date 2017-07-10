package test_examples

import (
	dit "github.com/cloud-42/docker-it"
	"os"
	"testing"
)

var dockerEnvironment *dit.DockerEnvironment

func init() {
	dockerEnvironment = newDockerEnvironment()
}

func TestMain(m *testing.M) {
	dockerEnvironment.StartParallel("it-redis")
	code := m.Run()
	dockerEnvironment.Shutdown()
	os.Exit(code)
}

func newDockerEnvironment() *dit.DockerEnvironment {
	env, err := dit.NewDockerEnvironment(
		dit.DockerComponent{
			Name:       "it-redis",
			Image:      "redis",
			FollowLogs: true,
			ExposedPorts: []dit.Port{
				{
					ContainerPort: 6379,
				},
			},
		},
	)
	if err != nil {
		panic(err)
	}
	return env
}
