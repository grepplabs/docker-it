package test_examples

import (
	dit "github.com/cloud-42/docker-it"
	"github.com/cloud-42/docker-it/wait/http"
	"github.com/cloud-42/docker-it/wait/redis"
	"os"
	"testing"
)

var dockerEnvironment *dit.DockerEnvironment

func init() {
	dockerEnvironment = newDockerEnvironment()
}

func TestMain(m *testing.M) {
	if err := dockerEnvironment.StartParallel("it-redis", "it-wiremock"); err != nil {
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
			AfterStart: &redis.RedisWait{},
		},
		dit.DockerComponent{
			Name:       "it-wiremock",
			Image:      "rodolpheche/wiremock",
			FollowLogs: true,
			ExposedPorts: []dit.Port{
				{
					ContainerPort: 8080,
				},
			},
			AfterStart: &http.HttpWait{UrlTemplate: `http://{{ value . "it-wiremock.Host"}}:{{ value . "it-wiremock.Port"}}/__admin`},
		},
	)
	if err != nil {
		panic(err)
	}
	return env
}
