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

type DockerEnvironment struct {
	context          *DockerEnvironmentContext
	lifecycleHandler *DockerLifecycleHandler
}

func NewDockerEnvironment(components ...DockerComponent) (*DockerEnvironment, error) {
	if len(components) == 0 {
		return nil, errors.New("Component list is empty")
	}
	// TODO: all properties and port resolution are collected
	//       host, containerPort, targetPort, hostPort, port
	// TODO: env variable can be a template and is resolved
	// TODO: remember found host ports (prevent double assignemt)

	// TODO: Port struct: name and containerPort are mandatory
	// TODO: add shutdown hook

	containers, err := toContainers(components...)
	if err != nil {
		return nil, err
	}
	id := uuid.New().String()
	context := &DockerEnvironmentContext{ID: id, containers: containers}
	lifecycleHandler,err := NewDockerLifecycleHandler(context)
	if err != nil {
		return nil, err
	}
	return &DockerEnvironment{context: context, lifecycleHandler: lifecycleHandler}, nil
}
func toContainers(components ...DockerComponent) (map[string]*DockerContainer, error) {
	containers := make(map[string]*DockerContainer)
	for _, component := range components {
		if component.Name == "" || component.Image == "" {
			return nil, errors.New("DockerComponent Name and Image must not be empty")
		}
		name := containerName(component.Name)
		container := NewDockerContainer(component)
		container.DockerComponent.Name = strings.ToLower(component.Name)
		if _, exits := containers[name]; exits {
			return nil, fmt.Errorf("DockerComponent [%s] is configured twice", name)
		}
		containers[name] = container
	}
	return containers, nil
}
func containerName(componentName string) string {
	return strings.ToLower(componentName)
}


type dockerContainerCommand interface {
    exec (*DockerContainer) error
}

func (r *DockerEnvironment) Start(names ...string) error {
	return r.forEach(r.lifecycleHandler.Start, names ...)
}

func (r *DockerEnvironment) StartParallel(names ...string) error {
	//TODO: implement
	return nil
}

func (r *DockerEnvironment) Stop(names ...string) error {
	return r.forEach(r.lifecycleHandler.Stop, names ...)
}

func (r *DockerEnvironment) Destroy(names ...string) error {
	return r.forEach(r.lifecycleHandler.Destroy, names ...)
}

func (r *DockerEnvironment) forEach(f func(*DockerContainer) error, names ...string) error {
	for _, name := range names {
		if container, exits := r.context.containers[containerName(name)]; !exits {
			return fmt.Errorf("DockerComponent [%s] is not configured", name)
		} else {
			if err := f(container); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *DockerEnvironment) Close() {
	r.lifecycleHandler.Close()
}

// resolve template with registered variables e.g. ports
func (r *DockerEnvironment) Resolve(template string) (string, error) {
	//TODO: implement
	return "", nil
}

//TODO: find port bind on all interfaces (or variable)
//TODO: take getPublicFacingIP
