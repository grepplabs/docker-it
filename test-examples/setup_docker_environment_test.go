package test_examples

import (
	dit "github.com/cloud-42/docker-it"
	"github.com/cloud-42/docker-it/wait"
	"github.com/cloud-42/docker-it/wait/http"
	"github.com/cloud-42/docker-it/wait/mysql"
	"github.com/cloud-42/docker-it/wait/postgres"
	"github.com/cloud-42/docker-it/wait/redis"
	"os"
	"testing"
	"time"
)

var dockerEnvironment *dit.DockerEnvironment

func init() {
	dockerEnvironment = newDockerEnvironment()
}

func TestMain(m *testing.M) {
	components := []string{
		"it-redis",
		"it-http",
		"it-postgres",
		"it-mysql",
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
			Name:       "it-http",
			Image:      "rodolpheche/wiremock",
			ForcePull:  true,
			FollowLogs: true,
			ExposedPorts: []dit.Port{
				{
					ContainerPort: 8080,
				},
			},
			AfterStart: &http.HttpWait{
				UrlTemplate: `http://{{ value . "it-http.Host"}}:{{ value . "it-http.Port"}}/__admin`},
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
			AfterStart: &postgres.PostgresWait{
				DatabaseUrl: `postgres://postgres:postgres@{{ value . "it-postgres.Host"}}:{{ value . "it-postgres.Port"}}/postgres?sslmode=disable`},
		},
		dit.DockerComponent{
			Name:       "it-mysql",
			Image:      "mysql:8.0",
			ForcePull:  true,
			FollowLogs: true,
			ExposedPorts: []dit.Port{
				{
					ContainerPort: 3306,
				},
			},
			EnvironmentVariables: map[string]string{
				"MYSQL_ROOT_PASSWORD": "mypassword",
			},
			AfterStart: &mysql.MySQLWait{
				DatabaseUrl: `root:mypassword@tcp({{ value . "it-mysql.Host"}}:{{ value . "it-mysql.Port"}})/`,
				Wait:        wait.Wait{AtMost: 60 * time.Second},
			},
		},
	)
	if err != nil {
		panic(err)
	}
	return env
}
