package dockerit

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"net"
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
		Image:      "redis",
		ExposedPorts: []Port{
			{
				ContainerPort: 6379,
			},
		},
	})
	a.EqualError(err, "DockerComponent Name and Image must not be empty")

	_, err = context.addContainer(DockerComponent{
		Name:       "it-redis",
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
		Name:       "it-redis",
		Image:      "redis",
	})
	a.Nil(err)

	_, err = context.addContainer(DockerComponent{
		Name:       "it-redis",
		Image:      "redis:3.2",
	})
	a.EqualError(err, "DockerComponent [it-redis] is configured twice")

	_, err = context.addContainer(DockerComponent{
		Name:       "IT-REDIS",
		Image:      "redis:3.2",
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
		ExposedPorts: []Port{
			{
				ContainerPort: 6379,
			},
		},
	})
	a.Nil(err)
	a.NotNil(container)

	//TODO: Implement

}