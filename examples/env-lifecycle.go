package main

import (
	dit "github.com/cloud-42/docker-it"
	"time"
	"fmt"
)

func main() {
	env, err := dit.NewDockerEnvironment(
		dit.DockerComponent{
			Name:  "redis-1",
			Image: "redis",
			FollowLogs: true,
			ExposedPorts: []dit.Port{
				{
					Name:          "redis-1",
					ContainerPort: 6379,
				},
				{
					Name:          "sentinel",
					ContainerPort: 6380,
				},
			},
			EnvironmentVariables: map[string]string{
				"my-key":            "my-value",
				"my-redis-port1":    `http://localhost:{{ value . "redis-1.redis-1.HostPort"}}`,
				"my-redis-port2":    `http://localhost:{{ value . "redis-1.HostPort"}}`,
				"my-sentinel-port1": `http://localhost:{{ value . "redis-1.sentinel.HostPort"}}`,
			},
		},
		dit.DockerComponent{
			Name:  "REDIS-2",
			Image: "redis",
			ExposedPorts: []dit.Port{
				{
					ContainerPort: 6379,
				},
				{
					Name:          "SENTINEL",
					ContainerPort: 6380,
				},
			},
			EnvironmentVariables: map[string]string{
				"my-key":              "my-value",
				"my-redis-port":       `http://localhost:{{ value . "redis-1.HostPort"}}`,
				"my-sentinel-port":    `http://localhost:{{ value . "REDIS-2.SENTINEL.HostPort"}}`,
				"other-sentinel-port": `http://localhost:{{ value . "redis-1.sentinel.HostPort"}}`,
			},
		},
	)

	if err != nil {
		panic(err)
	}

	fmt.Println("Start container")
	if err := env.Start("redis-1"); err != nil {
		panic(err)
	}

	time.Sleep(10 * time.Second)

	fmt.Println("Stop container")
	if err := env.Stop("redis-1"); err != nil {
		panic(err)
	}

	time.Sleep(1 * time.Second)

	fmt.Println("Destroy container")
	if err := env.Destroy("redis-1"); err != nil {
		panic(err)
	}
}
