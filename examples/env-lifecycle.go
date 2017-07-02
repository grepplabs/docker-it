package main

import (
	dit "github.com/cloud-42/docker-it"
	"time"
)

func main() {
	env, err := dit.NewDockerEnvironment(
		dit.DockerComponent{
			Name:  "redis-1",
			Image: "redis",
			ExposedPorts: []dit.Port{
				{
					ContainerPort: 6379,
				},
			},
			EnvironmentVariables: map[string]string {
				"my-key": "my-value",
			},
		},
	)

	if err != nil {
		panic(err)
	}

	if err := env.Start("redis-1"); err != nil {
		panic(err)
	}

	if err := env.Stop("redis-1"); err != nil {
		panic(err)
	}

	time.Sleep(2 * time.Second)

	if err := env.Destroy("redis-1"); err != nil {
		panic(err)
	}
}
