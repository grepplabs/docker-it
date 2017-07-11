package test_examples

import (
	dit "github.com/cloud-42/docker-it"
	"github.com/cloud-42/docker-it/wait/redis"
	"os"
	"testing"
)

var dockerEnvironment *dit.DockerEnvironment

func init() {
	dockerEnvironment = newDockerEnvironment()
}

func TestMain(m *testing.M) {
	if err := dockerEnvironment.StartParallel("it-redis"); err != nil {
		dockerEnvironment.Shutdown()
		panic(err)
	}

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
			AfterStart: &redis.Wait{},
		},
	)
	if err != nil {
		panic(err)
	}
	return env
}
