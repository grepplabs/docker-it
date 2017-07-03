package dockerit

import (
	"errors"
)

type DockerEnvironment struct {
	context          *DockerEnvironmentContext
	lifecycleHandler *DockerLifecycleHandler
	valueResolver    *DockerEnvironmentValueResolver
}

func NewDockerEnvironment(components ...DockerComponent) (*DockerEnvironment, error) {
	if len(components) == 0 {
		return nil, errors.New("Component list is empty")
	}
	// TODO: add shutdown hook

	// new context
	context := NewDockerEnvironmentContext()
	for _, component := range components {
		if err := context.addContainer(component); err != nil {
			return nil, err
		}
	}
	//TODO: public facing IP or variable
	portBinding := NewDockerEnvironmentPortBinding("0.0.0.0", context)
	if err := portBinding.configurePortBindings(); err != nil {
		return nil, err
	}

	//TODO: public facing IP or variable
	valueResolver := NewDockerComponentValueResolver("127.0.0.1", context)
	if err := valueResolver.configureContainersEnv(); err != nil {
		return nil, err
	}

	// new lifecycle handler
	lifecycleHandler, err := NewDockerLifecycleHandler(context)
	if err != nil {
		return nil, err
	}
	return &DockerEnvironment{context: context, lifecycleHandler: lifecycleHandler, valueResolver: valueResolver}, nil
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

func (r *DockerEnvironment) Resolve(template string) (string, error) {
	return r.valueResolver.resolve(template)
}

//TODO: take getPublicFacingIP
