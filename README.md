# docker-it - integration testing with Docker
golang library for integration testing with Docker

[![Build Status](https://travis-ci.org/grepplabs/docker-it.svg?branch=master)](https://travis-ci.org/grepplabs/docker-it)
[![Go Report Card](https://goreportcard.com/badge/github.com/grepplabs/docker-it)](https://goreportcard.com/report/github.com/grepplabs/docker-it)
[![Coverage Status](https://coveralls.io/repos/github/grepplabs/docker-it/badge.svg?branch=master)](https://coveralls.io/github/grepplabs/docker-it?branch=master)
[![GoDoc](https://godoc.org/github.com/grepplabs/docker-it?status.svg)](https://godoc.org/github.com/grepplabs/docker-it)

This utility library allows you to create a test environment based on docker containers:

* Dynamic host port binding - multiple test environments can be run simultaneously e.g. multi-branch CI pipeline
* Resolve values of named port between defined components
* Containers can be started in parallel
* Full control of the container lifecycle - you can stop and restart a container to test connectivity problems
* Follow container log output
* Define a wait for container application startup before your tests start
* Bind mounts
* Use DOCKER_API_VERSION environment variable to set API version
 
Prerequisites
========
[Go 1.9 or higher](https://golang.org/doc/install)  
[Docker](https://docs.docker.com/engine/installation/linux/docker-ce/ubuntu/)

Building
========

```bash
$ export GOPATH=$(pwd)    # first set GOPATH if not done already
$ go get -d github.com/grepplabs/docker-it
$ go get -u github.com/golang/dep/cmd/dep
$ cd $GOPATH/src/github.com/grepplabs/docker-it
$ dep ensure
$ go build .
$ go test -v ./
$ go test -v ./test-examples/...
```

Example usage
========
Define your test environment

```go
package mytestwithdocker

import (
	dit "github.com/grepplabs/docker-it"
	"github.com/grepplabs/docker-it/wait"
	"github.com/grepplabs/docker-it/wait/elastic"
	"github.com/grepplabs/docker-it/wait/redis"
	"github.com/grepplabs/docker-it/wait/http"
	"testing"
	"time"
)

func TestWithDocker(t *testing.T) {
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
			FollowLogs: false,
			ExposedPorts: []dit.Port{
				{
					ContainerPort: 8200,
				},
			},
			EnvironmentVariables: map[string]string{
				"VAULT_ADDR": "http://127.0.0.1:8200",
			},
			Cmd: []string{
				"server", "-dev", "-config=/etc/vault/vault_config.hcl",
			},
			Binds: []string{
				"/tmp/vault_config.hcl", "/etc/vault",
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
	if err != env.StartParallel("it-redis", "it-es", "it-vault") {
		panic(err)
	}

	// your tests

	env.Shutdown()

}
```

Command `env.StartParallel("it-redis", "it-es")` will start redis and elastic search in parallel.
With `AfterStart` option you can define wait condition.

```
INFO: 2017/07/30 23:48:35 Using IP 192.168.178.20
INFO: 2017/07/30 23:48:35 Starting components in parallel [it-redis it-es]
INFO: 2017/07/30 23:48:35 Start component it-redis
INFO: 2017/07/30 23:48:35 Start component it-es
INFO: 2017/07/30 23:48:35 Pulling image redis
INFO: 2017/07/30 23:48:35 Pulling image docker.elastic.co/elasticsearch/elasticsearch:5.5.0
INFO: 2017/07/30 23:48:37 Creating container for it-redis name it-redis-7bf7791d55ba env [] portSpecs [192.168.178.20:33139:6379/tcp]
INFO: 2017/07/30 23:48:37 Created new container c213214253e2 for it-redis
INFO: 2017/07/30 23:48:37 Starting container c213214253e2 for it-redis
INFO: 2017/07/30 23:48:37 Creating container for it-es name it-es-7bf7791d55ba env [http.host=0.0.0.0 transport.host=127.0.0.1] portSpecs [192.168.178.20:42705:9200/tcp]
INFO: 2017/07/30 23:48:37 Created new container 0fae9d0ac54a for it-es
INFO: 2017/07/30 23:48:37 Starting container 0fae9d0ac54a for it-es
INFO: 2017/07/30 23:48:37 Start follow logs c213214253e2
WAIT FOR it-redis: 2017/07/30 23:48:37 Waiting for redis 192.168.178.20:33139
it-redis: 1:C 30 Jul 21:48:37.440 # oO0OoO0OoO0Oo Redis is starting oO0OoO0OoO0Oo
it-redis: 1:C 30 Jul 21:48:37.440 # Redis version=4.0.1, bits=64, commit=00000000, modified=0, pid=1, just started
it-redis: 1:C 30 Jul 21:48:37.440 # Warning: no config file specified, using the default config. In order to specify a config file use redis-server /path/to/redis.conf
it-redis: 1:M 30 Jul 21:48:37.440 * Running mode=standalone, port=6379.
it-redis: 1:M 30 Jul 21:48:37.441 # WARNING: The TCP backlog setting of 511 cannot be enforced because /proc/sys/net/core/somaxconn is set to the lower value of 128.
it-redis: 1:M 30 Jul 21:48:37.441 # Server initialized
it-redis: 1:M 30 Jul 21:48:37.441 # WARNING overcommit_memory is set to 0! Background save may fail under low memory condition. To fix this issue add 'vm.overcommit_memory = 1' to /etc/sysctl.conf and then reboot or run the command 'sysctl vm.overcommit_memory=1' for this to take effect.
it-redis: 1:M 30 Jul 21:48:37.441 # WARNING you have Transparent Huge Pages (THP) support enabled in your kernel. This will create latency and memory usage issues with Redis. To fix this issue run the command 'echo never > /sys/kernel/mm/transparent_hugepage/enabled' as root, and add it to your /etc/rc.local in order to retain the setting after a reboot. Redis must be restarted after THP is disabled.
it-redis: 1:M 30 Jul 21:48:37.441 * Ready to accept connections
WAIT FOR it-redis: 2017/07/30 23:48:37 Component is up after 201.235µs
WAIT FOR it-es: 2017/07/30 23:48:37 Waiting for elastic http://192.168.178.20:42705/
WAIT FOR it-es: 2017/07/30 23:48:45 Component is up after 8.166139275s
INFO: 2017/07/30 23:48:45 All components started
INFO: 2017/07/30 23:48:45 Destroy component it-redis container c213214253e2
INFO: 2017/07/30 23:48:45 Received stop follow logs c213214253e2
INFO: 2017/07/30 23:48:45 Stop component it-redis
it-redis: 1:signal-handler (1501451325) Received SIGTERM scheduling shutdown...
it-redis: 1:M 30 Jul 21:48:45.855 # User requested shutdown...
it-redis: 1:M 30 Jul 21:48:45.855 * Saving the final RDB snapshot before exiting.
it-redis: 1:M 30 Jul 21:48:45.857 * DB saved on disk
it-redis: 1:M 30 Jul 21:48:45.857 # Redis is now ready to exit, bye bye...
INFO: 2017/07/30 23:48:46 Remove container c213214253e2
INFO: 2017/07/30 23:48:46 Destroy component it-es container 0fae9d0ac54a
INFO: 2017/07/30 23:48:46 Stop component it-es
INFO: 2017/07/30 23:48:47 Remove container 0fae9d0ac54a
INFO: 2017/07/30 23:48:47 Closing docker lifecycle handler
```

As `it-redis` component sets `FollowLogs` option, the logs from its container are logged with the component name as prefix

```
it-redis: 1:M 30 Jul 21:48:37.440 * Running mode=standalone, port=6379.
```

When the tests are finished shutdown the test environment with `env.Shutdown()` and the test containers will be removed

```
INFO: 2017/07/30 23:48:45 Destroy component it-redis container c213214253e2
INFO: 2017/07/30 23:48:46 Remove container c213214253e2
INFO: 2017/07/30 23:48:46 Destroy component it-es container 0fae9d0ac54a
INFO: 2017/07/30 23:48:47 Remove container 0fae9d0ac54a
```


Using TestMain
========

You can use `func TestMain(m *testing.M)` to start your environment before and shutdown it after testing.

```go
package testexamples

import (
	dit "github.com/grepplabs/docker-it"
	"os"
	"testing"
)

var dockerEnvironment *dit.DockerEnvironment

func init() {
	dockerEnvironment = newDockerEnvironment()
}

func TestMain(m *testing.M) {
	components := []string{
		"it-my-app",
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
			Name:       "it-my-app",
			Image:      "my-app:latest",
			RemoveImageAfterDestroy: true,
		},
	)
	if err != nil {
		panic(err)
	}
	return env
}
```

Test examples
========
Run [test-examples](https://github.com/grepplabs/docker-it/tree/master/test-examples)

* HTTP
* Elasticsearch
* MySQL
* Postgres
* Kafka
* Redis
* Vault

```bash
go test -v ./test-examples/...
```

