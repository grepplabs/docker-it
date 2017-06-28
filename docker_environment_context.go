package dockerit

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"strings"
)

type DockerEnvironmentContext struct {
	ID         string
	containers map[string]*DockerContainer
}

func NewDockerEnvironmentContext() *DockerEnvironmentContext {
	id := uuid.New().String()
	return &DockerEnvironmentContext{ID: id, containers: make(map[string]*DockerContainer)}
}

func (r *DockerEnvironmentContext) addContainer(component DockerComponent) error {
	if component.Name == "" || component.Image == "" {
		return errors.New("DockerComponent Name and Image must not be empty")
	}
	name := r.toName(component.Name)
	container := NewDockerContainer(component)
	container.DockerComponent.Name = name
	if _, exits := r.containers[name]; exits {
		return fmt.Errorf("DockerComponent [%s] is configured twice", name)
	}
	r.containers[name] = container
	return nil
}

func (r *DockerEnvironmentContext) getContainer(name string) (*DockerContainer, error) {
	if container, exits := r.containers[r.toName(name)]; !exits {
		return nil, fmt.Errorf("DockerComponent [%s] is not configured", name)
	} else {
		return container, nil
	}
}
func (r *DockerEnvironmentContext) toName(componentName string) string {
	return strings.ToLower(componentName)
}
