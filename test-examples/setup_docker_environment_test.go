package test_examples

import (
	dit "github.com/cloud-42/docker-it"
	"github.com/cloud-42/docker-it/wait/http"
	"github.com/cloud-42/docker-it/wait/postgres"
	"github.com/cloud-42/docker-it/wait/redis"
	"os"
	"testing"
)

var dockerEnvironment *dit.DockerEnvironment

func init() {
	dockerEnvironment = newDockerEnvironment()
}

func TestMain(m *testing.M) {
	components := []string{
		"it-redis",
		"it-wiremock",
		"it-postgres",
	}

	if err := dockerEnvironment.StartParallel(components...); err != nil {
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
			ForcePull:  true,
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
			ForcePull:  true,
			FollowLogs: true,
			ExposedPorts: []dit.Port{
				{
					ContainerPort: 8080,
				},
			},
			AfterStart: &http.HttpWait{UrlTemplate: `http://{{ value . "it-wiremock.Host"}}:{{ value . "it-wiremock.Port"}}/__admin`},
		},
		dit.DockerComponent{
			Name:       "it-postgres",
			Image:      "postgres:9.6",
			ForcePull:  true,
			FollowLogs: true,
			ExposedPorts: []dit.Port{
				{
					ContainerPort: 5432,
				},
			},
			AfterStart: &postgres.PostgresWait{DatabaseUrl: `postgres://postgres:postgres@{{ value . "it-postgres.Host"}}:{{ value . "it-postgres.Port"}}/postgres?sslmode=disable`},
		},
	)
	if err != nil {
		panic(err)
	}
	return env
}
