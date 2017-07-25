package dockerit

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestExternalIPv4Address(t *testing.T) {
	a := assert.New(t)
	s, err := externalIP()
	a.Nil(err)
	ip := net.ParseIP(s)
	a.NotNil(ip)
	a.NotNil(ip.To4())
}

func TestNewDockerEnvironmentFailsOnMissingNameOrImage(t *testing.T) {
	a := assert.New(t)
	context, err := newDockerEnvironmentContext()
	a.Nil(err)

	_, err = context.addContainer(DockerComponent{
		Image: "redis",
		ExposedPorts: []Port{
			{
				ContainerPort: 6379,
			},
		},
	})
	a.EqualError(err, "DockerComponent Name and Image must not be empty")

	_, err = context.addContainer(DockerComponent{
		Name: "it-redis",
		ExposedPorts: []Port{
			{
				ContainerPort: 6379,
			},
		},
	})
	a.EqualError(err, "DockerComponent Name and Image must not be empty")
}

func TestNewDockerEnvironmentFailsOnDuplicateComponent(t *testing.T) {
	a := assert.New(t)
	context, err := newDockerEnvironmentContext()
	a.Nil(err)

	_, err = context.addContainer(DockerComponent{
		Name:  "it-redis",
		Image: "redis",
	})
	a.Nil(err)

	_, err = context.addContainer(DockerComponent{
		Name:  "it-redis",
		Image: "redis:3.2",
	})
	a.EqualError(err, "DockerComponent [it-redis] is configured twice")

	_, err = context.addContainer(DockerComponent{
		Name:  "IT-REDIS",
		Image: "redis:3.2",
	})
	a.EqualError(err, "DockerComponent [it-redis] is configured twice")
}

func TestNewDockerEnvironment(t *testing.T) {
	a := assert.New(t)
	context, err := newDockerEnvironmentContext()
	a.Nil(err)

	container, err := context.addContainer(DockerComponent{
		Name:       "it-redis",
		Image:      "redis",
		ForcePull:  true,
		FollowLogs: false,
	})
	a.Nil(err)
	a.NotNil(container)

	err = context.configurePortBindings()
	a.Nil(err)
	err = context.configureContainersEnv()
	a.Nil(err)

	container2, err := context.getContainer("it-redis")
	a.Nil(err)
	a.Exactly(container, container2)

	container2, err = context.getContainer("IT-REDIS")
	a.Nil(err)
	a.Exactly(container, container2)

	a.Equal(context.externalIP, context.Host())
	host, err := context.Resolve(`{{ value . "it-redis.Host"}}`)
	a.Nil(err)
	a.Equal(context.externalIP, host)
}

func TestNewDockerEnvironmentPortMapping(t *testing.T) {
	a := assert.New(t)
	context, err := newDockerEnvironmentContext()
	a.Nil(err)

	_, err = context.addContainer(DockerComponent{
		Name:       "it-redis",
		Image:      "redis",
		ForcePull:  true,
		FollowLogs: false,
		ExposedPorts: []Port{
			{
				ContainerPort: 6379,
			},
			{
				Name:          "port-2",
				ContainerPort: 6380,
			},
		},
	})
	a.Nil(err)

	err = context.configurePortBindings()
	a.Nil(err)

	a.Equal(context.externalIP, context.Host())
	host, err := context.Resolve(`{{ value . "it-redis.Host"}}`)
	a.Nil(err)
	a.Equal(context.externalIP, host)

	port, err := context.Port("it-redis", "")
	a.Nil(err)
	a.True(port > 0)

	port2, err := context.Port("it-redis", "it-redis")
	a.Nil(err)
	a.Equal(port, port2)

	port2, err = context.Port("it-redis", "port-2")
	a.Nil(err)
	a.NotEqual(port, port2)

	_, err = context.Port("it-redis", "unknown")
	a.NotNil(err)
}

func TestNewDockerEnvironmentEnvVariables(t *testing.T) {
	a := assert.New(t)
	context, err := newDockerEnvironmentContext()
	a.Nil(err)

	container, err := context.addContainer(DockerComponent{
		Name:  "it-redis",
		Image: "redis",
		EnvironmentVariables: map[string]string{
			"MYSQL_ROOT_PASSWORD": "mypassword",
		},
	})
	a.Nil(err)

	err = context.configurePortBindings()
	a.Nil(err)
	err = context.configureContainersEnv()
	a.Nil(err)

	a.Equal("mypassword", container.EnvironmentVariables["MYSQL_ROOT_PASSWORD"])

}
