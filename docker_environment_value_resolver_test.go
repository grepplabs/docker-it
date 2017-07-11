package dockerit

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewDockerComponentValueResolver(t *testing.T) {
	a := assert.New(t)

	environmentContext, err := newDockerEnvironmentContext()
	a.Nil(err)
	resolver := newDockerComponentValueResolver("127.0.0.2", environmentContext)

	a.Equal(environmentContext, resolver.context)
	a.Equal("127.0.0.2", resolver.ip)
}

func TestEnvironmentContextVariablesNoContainers(t *testing.T) {
	a := assert.New(t)

	environmentContext, err := newDockerEnvironmentContext()
	a.Nil(err)

	resolver := &dockerEnvironmentValueResolver{context: environmentContext}
	resolveContext, err := resolver.getEnvironmentContextVariables()
	a.Nil(err)
	a.Empty(resolveContext)
}

func TestEnvironmentContextVariablesNoBinding(t *testing.T) {
	a := assert.New(t)

	environmentContext, err := newDockerEnvironmentContext()
	a.Nil(err)

	container, err := environmentContext.addContainer(DockerComponent{Name: "REDIS", Image: "redis:latest"})
	a.Nil(err)
	container.portBindings = make([]Port, 0)

	resolver := &dockerEnvironmentValueResolver{ip: "127.0.0.1", context: environmentContext}
	resolveContext, err := resolver.getEnvironmentContextVariables()
	a.Nil(err)
	a.Equal(resolveContext, map[string]interface{}{
		"REDIS.Host": "127.0.0.1", "redis.Host": "127.0.0.1",
	})
}

func TestEnvironmentContextVariablesNoBindings(t *testing.T) {
	a := assert.New(t)

	environmentContext, err := newDockerEnvironmentContext()
	a.Nil(err)

	container1, err := environmentContext.addContainer(DockerComponent{Name: "REDIS", Image: "redis:latest"})
	a.Nil(err)
	container1.portBindings = make([]Port, 0)
	container2, err := environmentContext.addContainer(DockerComponent{Name: "kafka", Image: "kafka:latest"})
	a.Nil(err)
	container2.portBindings = make([]Port, 0)

	resolver := &dockerEnvironmentValueResolver{ip: "127.0.0.1", context: environmentContext}
	resolveContext, err := resolver.getEnvironmentContextVariables()
	a.Nil(err)
	a.Equal(resolveContext, map[string]interface{}{
		"REDIS.Host": "127.0.0.1", "redis.Host": "127.0.0.1", "kafka.Host": "127.0.0.1",
	})
}

func TestEnvironmentContextVariablesPortBindingsWithDefaultName(t *testing.T) {
	a := assert.New(t)

	environmentContext, err := newDockerEnvironmentContext()
	a.Nil(err)

	container, err := environmentContext.addContainer(DockerComponent{Name: "REDIS", Image: "redis:latest"})
	a.Nil(err)
	container.portBindings = []Port{
		{ContainerPort: 8080, HostPort: 8081},
	}
	container.DockerComponent.ExposedPorts = []Port{
		{ContainerPort: 8080, HostPort: 8081},
	}
	resolver := &dockerEnvironmentValueResolver{ip: "127.0.0.1", context: environmentContext}
	resolveContext, err := resolver.getEnvironmentContextVariables()
	a.Nil(err)
	a.Equal(resolveContext, map[string]interface{}{
		"REDIS.Host": "127.0.0.1",
		"redis.Host": "127.0.0.1",

		"REDIS.ContainerPort": "8080",
		"redis.ContainerPort": "8080",
		"REDIS.TargetPort":    "8080",
		"redis.TargetPort":    "8080",
		"REDIS.HostPort":      "8081",
		"redis.HostPort":      "8081",
		"REDIS.Port":          "8081",
		"redis.Port":          "8081",
	})
}

func TestEnvironmentContextVariablesPortBindingsWithNamedPort(t *testing.T) {
	a := assert.New(t)

	environmentContext, err := newDockerEnvironmentContext()
	a.Nil(err)

	container, err := environmentContext.addContainer(DockerComponent{Name: "REDIS", Image: "redis:latest"})
	a.Nil(err)
	container.portBindings = []Port{
		{Name: "my-port", ContainerPort: 8080, HostPort: 8081},
	}
	container.DockerComponent.ExposedPorts = []Port{
		{Name: "MY-PORT", ContainerPort: 8080, HostPort: 8081},
	}

	resolver := &dockerEnvironmentValueResolver{ip: "127.0.0.1", context: environmentContext}
	resolveContext, err := resolver.getEnvironmentContextVariables()
	a.Nil(err)
	a.Equal(resolveContext, map[string]interface{}{
		"REDIS.Host": "127.0.0.1",
		"redis.Host": "127.0.0.1",

		"REDIS.MY-PORT.ContainerPort": "8080",
		"redis.MY-PORT.ContainerPort": "8080",
		"REDIS.MY-PORT.TargetPort":    "8080",
		"redis.MY-PORT.TargetPort":    "8080",
		"REDIS.MY-PORT.HostPort":      "8081",
		"redis.MY-PORT.HostPort":      "8081",
		"REDIS.MY-PORT.Port":          "8081",
		"redis.MY-PORT.Port":          "8081",

		"REDIS.my-port.ContainerPort": "8080",
		"redis.my-port.ContainerPort": "8080",
		"REDIS.my-port.TargetPort":    "8080",
		"redis.my-port.TargetPort":    "8080",
		"REDIS.my-port.HostPort":      "8081",
		"redis.my-port.HostPort":      "8081",
		"REDIS.my-port.Port":          "8081",
		"redis.my-port.Port":          "8081",
	})

}

func TestResolveUsingContextVariables(t *testing.T) {
	a := assert.New(t)

	environmentContext, err := newDockerEnvironmentContext()
	a.Nil(err)

	container, err := environmentContext.addContainer(DockerComponent{Name: "redis", Image: "redis:latest"})
	a.Nil(err)

	container.portBindings = []Port{
		{ContainerPort: 6379, HostPort: 32401},
		{Name: "sentinel", ContainerPort: 26379, HostPort: 32402},
	}
	container.DockerComponent.ExposedPorts = []Port{
		{ContainerPort: 6379, HostPort: 32401},
		{Name: "sentinel", ContainerPort: 26379, HostPort: 32402},
	}

	resolver := &dockerEnvironmentValueResolver{ip: "192.168.178.44", context: environmentContext}

	value, err := resolver.resolve(`redis://{{ value . "redis.Host"}}:{{ value . "redis.Port"}}`)
	a.Nil(err)
	a.Equal(`redis://192.168.178.44:32401`, value)

	value, err = resolver.resolve(`redis://{{ value . "redis.Host"}}:{{ value . "redis.ContainerPort"}}`)
	a.Nil(err)
	a.Equal(`redis://192.168.178.44:6379`, value)

	value, err = resolver.resolve(`redis://{{ value . "redis.Host"}}:{{ value . "redis.sentinel.Port"}}`)
	a.Nil(err)
	a.Equal(`redis://192.168.178.44:32402`, value)

	value, err = resolver.resolve(`redis://{{ value . "redis.Host"}}:{{ value . "redis.sentinel.ContainerPort"}}`)
	a.Nil(err)
	a.Equal(`redis://192.168.178.44:26379`, value)
}

func TestResolveUsingSystemVariables(t *testing.T) {
	a := assert.New(t)

	environmentContext, err := newDockerEnvironmentContext()
	a.Nil(err)

	container, err := environmentContext.addContainer(DockerComponent{Name: "redis", Image: "redis:latest"})
	a.Nil(err)

	container.portBindings = []Port{
		{ContainerPort: 6379, HostPort: 32401},
	}
	container.DockerComponent.ExposedPorts = []Port{
		{ContainerPort: 6379, HostPort: 32401},
	}
	os.Setenv("my-os-env-variable", "4711")

	resolver := &dockerEnvironmentValueResolver{ip: "192.168.178.44", context: environmentContext}

	value, err := resolver.resolve(`redis://{{ value . "redis.Host"}}:{{ value . "redis.Port"}}`)
	a.Nil(err)
	a.Equal(`redis://192.168.178.44:32401`, value)

	value, err = resolver.resolve(`redis://{{ value . "redis.Host"}}:{{ value . "redis.ContainerPort"}}`)
	a.Nil(err)
	a.Equal(`redis://192.168.178.44:6379`, value)

	value, err = resolver.resolve(`redis://{{ value . "redis.Host"}}:{{ value . "my-os-env-variable"}}`)
	a.Nil(err)
	a.Equal(`redis://192.168.178.44:4711`, value)

}

func TestResolveContextVariableOutweighsSystemVariable(t *testing.T) {
	a := assert.New(t)

	environmentContext, err := newDockerEnvironmentContext()
	a.Nil(err)

	container, err := environmentContext.addContainer(DockerComponent{Name: "redis", Image: "redis:latest"})
	a.Nil(err)

	container.portBindings = []Port{
		{ContainerPort: 6379, HostPort: 32401},
	}
	os.Setenv("redis.HostPort", "32402")
	os.Setenv("redis.HostPort2", "32403")

	resolver := &dockerEnvironmentValueResolver{ip: "192.168.178.44", context: environmentContext}

	value, err := resolver.resolve(`redis://{{ value . "redis.Host"}}:{{ value . "redis.HostPort"}}`)
	a.Nil(err)
	a.Equal(`redis://192.168.178.44:32401`, value)

	value, err = resolver.resolve(`redis://{{ value . "redis.Host"}}:{{ value . "redis.HostPort2"}}`)
	a.Nil(err)
	a.Equal(`redis://192.168.178.44:32403`, value)
}

func TestResolveReturnsErrorWhenVariableIsNotFound(t *testing.T) {
	a := assert.New(t)

	environmentContext, err := newDockerEnvironmentContext()
	a.Nil(err)

	container, err := environmentContext.addContainer(DockerComponent{Name: "redis", Image: "redis:latest"})
	a.Nil(err)

	container.portBindings = []Port{
		{ContainerPort: 6379, HostPort: 32401},
	}

	resolver := &dockerEnvironmentValueResolver{ip: "192.168.178.44", context: environmentContext}

	_, err = resolver.resolve(`redis://{{ value . "redis.Host"}}:{{ value . "redis.HostBad"}}`)
	a.True(err != nil)
	a.Contains(err.Error(), "Unknown key 'redis.HostBad'")
}

func TestConfigureContainersEnv(t *testing.T) {
	a := assert.New(t)

	environmentContext, err := newDockerEnvironmentContext()
	a.Nil(err)

	container1, err := environmentContext.addContainer(DockerComponent{Name: "redis", Image: "redis:latest"})
	a.Nil(err)
	container1.portBindings = []Port{
		{ContainerPort: 6379, HostPort: 32401},
	}
	container1.EnvironmentVariables = map[string]string{
		"REDIS_TARGET": `redis://{{ value . "redis.Host"}}:{{ value . "redis.HostPort"}}`,
		"KAFKA_TARGET": `kafka-other://{{ value . "kafka.Host"}}:{{ value . "kafka.HostPort"}}`,
	}

	container2, err := environmentContext.addContainer(DockerComponent{Name: "kafka", Image: "kafka:latest"})
	a.Nil(err)
	container2.portBindings = []Port{
		{ContainerPort: 9094, HostPort: 32402},
	}
	container2.EnvironmentVariables = map[string]string{
		"KAFKA_TARGET": `kafka://{{ value . "kafka.Host"}}:{{ value . "kafka.HostPort"}}`,
		"REDIS_TARGET": `redis-other://{{ value . "redis.Host"}}:{{ value . "redis.HostPort"}}`,
	}

	resolver := &dockerEnvironmentValueResolver{ip: "127.0.0.1", context: environmentContext}
	err = resolver.configureContainersEnv()
	a.Nil(err)

	// cross container resolution
	a.Equal(`redis://127.0.0.1:32401`, container1.env["REDIS_TARGET"])
	a.Equal(`kafka-other://127.0.0.1:32402`, container1.env["KAFKA_TARGET"])

	a.Equal(`kafka://127.0.0.1:32402`, container2.env["KAFKA_TARGET"])
	a.Equal(`redis-other://127.0.0.1:32401`, container2.env["REDIS_TARGET"])
}

func TestRequireContainerPortBindings(t *testing.T) {
	a := assert.New(t)

	environmentContext, err := newDockerEnvironmentContext()
	a.Nil(err)

	_, err = environmentContext.addContainer(DockerComponent{Name: "REDIS", Image: "redis:latest"})
	a.Nil(err)

	resolver := &dockerEnvironmentValueResolver{ip: "127.0.0.1", context: environmentContext}
	_, err = resolver.getEnvironmentContextVariables()
	a.True(err != nil)
	a.Equal(`portBindings for 'redis' is not defined`, err.Error())
}
