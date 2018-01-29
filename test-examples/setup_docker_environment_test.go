package testexamples

import (
	"fmt"
	dit "github.com/grepplabs/docker-it"
	"github.com/grepplabs/docker-it/wait"
	"github.com/grepplabs/docker-it/wait/elastic"
	"github.com/grepplabs/docker-it/wait/http"
	"github.com/grepplabs/docker-it/wait/kafka"
	"github.com/grepplabs/docker-it/wait/mysql"
	"github.com/grepplabs/docker-it/wait/postgres"
	"github.com/grepplabs/docker-it/wait/redis"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

var dockerEnvironment, dockerEnvironment2 *dit.DockerEnvironment

func init() {
	dockerEnvironment = newDockerEnvironment()
	dockerEnvironment2 = newDockerEnvironment2()

}

func TestMain(m *testing.M) {
	components := []string{
		"it-redis",
		"it-http",
		"it-postgres",
		"it-mysql",
		"it-kafka",
		"it-es",
		"it-vault",
	}

	go handleInterrupt()

	if err := dockerEnvironment.StartParallel(components...); err != nil {
		dockerEnvironment.Shutdown()
		panic(err)
	}

	if err := dockerEnvironment2.Start("it-redis2"); err != nil {
		dockerEnvironment2.Shutdown()
		panic(err)
	}

	code := m.Run()
	dockerEnvironment.Shutdown()
	dockerEnvironment2.Shutdown()

	os.Exit(code)
}

func newDockerEnvironment() *dit.DockerEnvironment {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	_, err = os.Stat(filepath.Join(pwd, "vault_config.hcl"))
	if os.IsNotExist(err) {
		panic(err)
	}

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
			AfterStart: redis.NewRedisWait(redis.Options{}),
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
					WaitOptions: wait.Options{AtMost: 180 * time.Second},
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
				kafka.Options{
					WaitOptions: wait.Options{AtMost: 120 * time.Second},
				},
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
					Username:    "elastic",
					Password:    "changeme",
				},
			),
		},
		dit.DockerComponent{
			Name:       "it-vault",
			Image:      "vault:0.9.1",
			ForcePull:  true,
			FollowLogs: true,
			ExposedPorts: []dit.Port{
				{
					ContainerPort: 8200,
				},
			},
			Cmd: []string{
				"server", "-dev", "-config=/etc/vault/vault_config.hcl",
			},
			Binds: []string{
				fmt.Sprintf("%s:%s", pwd, "/etc/vault"),
			},
			AfterStart: http.NewHttpWait(
				`http://{{ value . "it-vault.Host"}}:{{ value . "it-vault.Port"}}/v1/sys/seal-status`,
				http.Options{},
			),
			DNSServer: "8.8.8.8",
		},
	)
	if err != nil {
		panic(err)
	}
	return env
}

func newDockerEnvironment2() *dit.DockerEnvironment {

	env, err := dit.NewDockerEnvironment(
		dit.DockerComponent{
			Name:       "it-redis2",
			Image:      "redis",
			ForcePull:  false,
			FollowLogs: false,
			ExposedPorts: []dit.Port{
				{
					ContainerPort: 6379,
				},
			},
			AfterStart: redis.NewRedisWait(redis.Options{}),
		},
	)
	if err != nil {
		panic(err)
	}
	return env
}

func handleInterrupt() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
	println("FATAL: Stop signal was received")
	dockerEnvironment.Shutdown()
	os.Exit(1)
}
