package dockerit

import (
	"errors"
)

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

	context := NewDockerEnvironmentContext()
	for _, component := range components {
		if err := context.addContainer(component); err != nil {
			return nil, err
		}
	}
	lifecycleHandler, err := NewDockerLifecycleHandler(context)
	if err != nil {
		return nil, err
	}
	return &DockerEnvironment{context: context, lifecycleHandler: lifecycleHandler}, nil
}

type dockerContainerCommand interface {
	exec(*DockerContainer) error
}

func (r *DockerEnvironment) Start(names ...string) error {
	return r.forEach(r.lifecycleHandler.Start, names...)
}

func (r *DockerEnvironment) StartParallel(names ...string) error {
	//TODO: implement
	return nil
}

func (r *DockerEnvironment) Stop(names ...string) error {
	return r.forEach(r.lifecycleHandler.Stop, names...)
}

func (r *DockerEnvironment) Destroy(names ...string) error {
	return r.forEach(r.lifecycleHandler.Destroy, names...)
}

func (r *DockerEnvironment) forEach(f func(*DockerContainer) error, names ...string) error {
	for _, name := range names {
		if container, err := r.context.getContainer(name); err != nil {
			return err
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
