package dockerit

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestGetNormalizedExposedPortsWhenNoPortsAreExposed(t *testing.T) {
	a := assert.New(t)

	environmentContext, err := NewDockerEnvironmentContext()
	a.Nil(err)

	binder := NewDockerEnvironmentPortBinding("0.0.0.0", environmentContext)

	_, err = environmentContext.addContainer(DockerComponent{Name: "redis", Image: "redis:latest"})
	a.Nil(err)

	mapping, err := binder.getNormalizedExposedPorts()
	a.Nil(err)
	a.Equal(map[string][]Port{"redis": []Port{}}, mapping)
}

func TestExposedPortContainerPortMustBeProvided(t *testing.T) {
	a := assert.New(t)

	environmentContext, err := NewDockerEnvironmentContext()
	a.Nil(err)

	binder := NewDockerEnvironmentPortBinding("0.0.0.0", environmentContext)

	container, err := environmentContext.addContainer(DockerComponent{Name: "redis", Image: "redis:latest"})
	a.Nil(err)
	container.ExposedPorts = []Port{
		{
			Name: "Container_port_is_not_provided",
		},
	}

	_, err = binder.getNormalizedExposedPorts()
	a.True(err != nil)
	a.Equal(`DockerComponent 'redis' ContainerPort '0' is invalid`, err.Error())
}

func TestGetNormalizedExposedPorts(t *testing.T) {
	a := assert.New(t)

	environmentContext, err := NewDockerEnvironmentContext()
	a.Nil(err)

	binder := NewDockerEnvironmentPortBinding("0.0.0.0", environmentContext)

	container, err := environmentContext.addContainer(DockerComponent{Name: "myapp", Image: "myapp:latest"})
	a.Nil(err)
	container.ExposedPorts = []Port{
		{
			ContainerPort: 8080,
		},
		{
			Name:          "port-1",
			ContainerPort: 8081,
		},
		{
			Name:          "port-2",
			ContainerPort: 8082,
			HostPort:      38082,
		},
		{
			Name:          "PORT-3",
			ContainerPort: 8083,
			HostPort:      38083,
		},
	}

	mapping, err := binder.getNormalizedExposedPorts()
	a.Nil(err)
	a.Equal(map[string][]Port{"myapp": []Port{
		{
			Name:          "myapp",
			ContainerPort: 8080,
			HostPort:      0,
		},
		{
			Name:          "port-1",
			ContainerPort: 8081,
			HostPort:      0,
		},
		{
			Name:          "port-2",
			ContainerPort: 8082,
			HostPort:      38082,
		},
		{
			Name:          "port-3",
			ContainerPort: 8083,
			HostPort:      38083,
		},
	}}, mapping)

}

func TestGetNormalizedExposedPortsFailsWhenPortNameUsedTwice(t *testing.T) {
	a := assert.New(t)

	environmentContext, err := NewDockerEnvironmentContext()
	a.Nil(err)

	binder := NewDockerEnvironmentPortBinding("0.0.0.0", environmentContext)

	container, err := environmentContext.addContainer(DockerComponent{Name: "myapp", Image: "myapp:latest"})
	a.Nil(err)
	container.ExposedPorts = []Port{
		{
			Name:          "port-3",
			ContainerPort: 8082,
			HostPort:      38082,
		},
		{
			Name:          "PORT-3",
			ContainerPort: 8083,
			HostPort:      38083,
		},
	}

	_, err = binder.getNormalizedExposedPorts()
	a.NotNil(err)
	a.Equal(`DockerComponent 'myapp' port name 'port-3' is configured twice`, err.Error())
}

func TestPortBindings(t *testing.T) {
	a := assert.New(t)
	componentPorts := map[string][]Port{"myapp": []Port{
		{
			Name:          "myapp",
			ContainerPort: 8080,
			HostPort:      0,
		},
		{
			Name:          "port-1",
			ContainerPort: 8081,
			HostPort:      0,
		},
		{
			Name:          "port-2",
			ContainerPort: 8082,
			HostPort:      38082,
		},
		{
			Name:          "port-3",
			ContainerPort: 8083,
			HostPort:      38083,
		},
	}}
	portBindings, err := getPortBindings("0.0.0.0", componentPorts)
	a.Nil(err)
	a.Equal(1, len(portBindings))
	ports := portBindings["myapp"]
	a.Equal(4, len(ports))

	port := ports[0]
	a.Equal("myapp", port.Name)
	a.Equal(8080, port.ContainerPort)
	a.True(port.HostPort > 0)

	port = ports[1]
	a.Equal("port-1", port.Name)
	a.Equal(8081, port.ContainerPort)
	a.True(port.HostPort > 0)

	port = ports[2]
	a.Equal("port-2", port.Name)
	a.Equal(8082, port.ContainerPort)
	a.Equal(38082, port.HostPort)

	port = ports[3]
	a.Equal("port-3", port.Name)
	a.Equal(8083, port.ContainerPort)
	a.Equal(38083, port.HostPort)
}

func TestPortBindingsFailsWhenHostPortConfiguredTwice(t *testing.T) {
	a := assert.New(t)
	componentPorts := map[string][]Port{"myapp": []Port{
		{
			Name:          "port-1",
			ContainerPort: 8082,
			HostPort:      38084,
		},
		{
			Name:          "port-2",
			ContainerPort: 8083,
			HostPort:      38084,
		},
	}}
	_, err := getPortBindings("0.0.0.0", componentPorts)
	a.NotNil(err)
	a.Contains(strings.ToLower(err.Error()), "address already in use")
}

func TestConfigurePortBinding(t *testing.T) {
	a := assert.New(t)

	environmentContext, err := NewDockerEnvironmentContext()
	a.Nil(err)

	container1, err := environmentContext.addContainer(DockerComponent{Name: "redis", Image: "redis:latest"})
	a.Nil(err)
	container1.ExposedPorts = []Port{
		{ContainerPort: 6379},
	}

	container2, err := environmentContext.addContainer(DockerComponent{Name: "kafka", Image: "kafka:latest"})
	a.Nil(err)
	container2.ExposedPorts = []Port{
		{ContainerPort: 9094},
	}

	binder := NewDockerEnvironmentPortBinding("0.0.0.0", environmentContext)
	err = binder.configurePortBindings()

	binding1 := container1.portBindings
	a.Equal(1, len(binding1))
	a.Equal("redis", binding1[0].Name)
	a.Equal(6379, binding1[0].ContainerPort)
	a.True(binding1[0].HostPort > 0)

	binding2 := container2.portBindings
	a.Equal(1, len(binding2))
	a.Equal("kafka", binding2[0].Name)
	a.Equal(9094, binding2[0].ContainerPort)
	a.True(binding2[0].HostPort > 0)
}
