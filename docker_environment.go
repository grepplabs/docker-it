package dockerit

import (
	"errors"
	"sync"
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
	// new context
	context, err := NewDockerEnvironmentContext()
	if err != nil {
		return nil, err
	}
	for _, component := range components {
		if _, err := context.addContainer(component); err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	// we could use 0.0.0.0
	portBinding := NewDockerEnvironmentPortBinding(context.externalIP, context)
	if err := portBinding.configurePortBindings(); err != nil {
		return nil, err
	}

	valueResolver := NewDockerComponentValueResolver(context.externalIP, context)
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
	if (len(names)) == 0 {
		return errors.New("No component was provided")
	}

	r.context.logger.Info.Println("Starting components in parallel", names)

	var wg sync.WaitGroup
	errorChannel := make(chan error, len(names))
	doneChannel := make(chan struct{}, 1)

	for _, name := range names {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			err := r.Start(name)
			if err != nil {
				r.context.logger.Error.Println("Component start error", err)
				errorChannel <- err
			}
		}(name)
	}
	go func() {
		defer func() {
			doneChannel <- struct{}{}
		}()
		wg.Wait()
	}()

	select {
	case err := <-errorChannel:
		return err
	case <-doneChannel:
	}
	r.context.logger.Info.Println("All components started")
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
