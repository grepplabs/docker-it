package dockerit

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEnvironmentContextVariablesNoContainers(t *testing.T) {
	a := assert.New(t)

	environmentContext, err := NewDockerEnvironmentContext()
	a.Nil(err)

	resolver := &DockerEnvironmentValueResolver{context: environmentContext}
	resolveContext, err := resolver.getEnvironmentContextVariables()
	a.Nil(err)
	a.Empty(resolveContext)
}

func TestEnvironmentContextVariablesNoBindings(t *testing.T) {
	a := assert.New(t)

	environmentContext, err := NewDockerEnvironmentContext()
	a.Nil(err)

	container, err := environmentContext.addContainer(DockerComponent{Name: "REDIS", Image: "redis:latest"})
	container.portBindings = make([]Port, 0)
	a.Nil(err)

	resolver := &DockerEnvironmentValueResolver{ip: "127.0.0.1", context: environmentContext}
	resolveContext, err := resolver.getEnvironmentContextVariables()
	a.Nil(err)
	a.Equal(resolveContext, map[string]interface{}{
		"REDIS.Host": "127.0.0.1", "redis.Host": "127.0.0.1",
	})
}
