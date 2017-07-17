package test_examples

import (
	dit "github.com/cloud-42/docker-it"
	"github.com/cloud-42/docker-it/wait"
	"github.com/cloud-42/docker-it/wait/elastic"
	"github.com/cloud-42/docker-it/wait/http"
	"github.com/cloud-42/docker-it/wait/kafka"
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
		"it-kafka",
		"it-es",
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
			FollowLogs: false,
			ExposedPorts: []dit.Port{
				{
					ContainerPort: 6379,
				},
			},
			AfterStart: redis.NewRedisWait(redis.Options{}),
		},
		dit.DockerComponent{
			Name:       "it-http",
			Image:      "rodolpheche/wiremock",
			ForcePull:  true,
			FollowLogs: false,
			ExposedPorts: []dit.Port{
				{
					ContainerPort: 8080,
				},
			},
			AfterStart: http.NewHttpWait(
				`http://{{ value . "it-http.Host"}}:{{ value . "it-http.Port"}}/__admin`,
				http.Options{},
			),
		},
		dit.DockerComponent{
			Name:       "it-postgres",
			Image:      "postgres:9.6",
			ForcePull:  true,
			FollowLogs: false,
			ExposedPorts: []dit.Port{
				{
					ContainerPort: 5432,
				},
			},
			AfterStart: postgres.NewPostgresWait(
				`postgres://postgres:postgres@{{ value . "it-postgres.Host"}}:{{ value . "it-postgres.Port"}}/postgres?sslmode=disable`,
				postgres.Options{}),
		},
		dit.DockerComponent{
			Name:       "it-mysql",
			Image:      "mysql:8.0",
			ForcePull:  true,
			FollowLogs: false,
			ExposedPorts: []dit.Port{
				{
					ContainerPort: 3306,
				},
			},
			EnvironmentVariables: map[string]string{
				"MYSQL_ROOT_PASSWORD": "mypassword",
			},
			AfterStart: mysql.NewMySQLWait(
				`root:mypassword@tcp({{ value . "it-mysql.Host"}}:{{ value . "it-mysql.Port"}})/`,
				mysql.Options{
					WaitOptions: wait.Options{AtMost: 60 * time.Second},
				},
			),
		},
		// see https://github.com/spotify/docker-kafka/pull/70
		dit.DockerComponent{
			Name:       "it-kafka",
			Image:      "spotify/kafka",
			ForcePull:  true,
			FollowLogs: false,
			ExposedPorts: []dit.Port{
				{
					ContainerPort: 9092,
				},
				{
					Name:          "zookeeper",
					ContainerPort: 2181,
				},
			},
			EnvironmentVariables: map[string]string{
				"ADVERTISED_HOST": `{{ value . "it-kafka.Host"}}`,
				"ADVERTISED_PORT": `{{ value . "it-kafka.Port"}}`,
			},
			AfterStart: kafka.NewKafkaWait(
				`{{ value . "it-kafka.Host"}}:{{ value . "it-kafka.Port"}}`,
				kafka.Options{},
			),
		},
		dit.DockerComponent{
			Name:       "it-es",
			Image:      "docker.elastic.co/elasticsearch/elasticsearch:5.5.0",
			ForcePull:  true,
			FollowLogs: false,
			ExposedPorts: []dit.Port{
				{
					ContainerPort: 9200,
				},
			},
			EnvironmentVariables: map[string]string{
				"http.host":      "0.0.0.0",
				"transport.host": "127.0.0.1",
			},
			AfterStart: elastic.NewElasticWait(
				`http://{{ value . "it-es.Host"}}:{{ value . "it-es.Port"}}/`,
				elastic.Options{
					WaitOptions: wait.Options{AtMost: 60 * time.Second},
					Username: "elastic",
					Password: "changeme",
				},
			),
		},
	)
	if err != nil {
		panic(err)
	}
	return env
}
