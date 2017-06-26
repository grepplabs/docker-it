package dockerit

import "errors"

type DockerEnvironment struct {
	context          DockerEnvironmentContext
	lifecycleHandler DockerLifecycleHandler
}

type DockerEnvironmentContext struct {
	ID         string
	containers map[string]DockerContainer
}

func NewDockerEnvironment(components ...DockerComponent) (*DockerEnvironment, error) {
	if len(components) == 0 {
		return nil, errors.New("Component list is empty")
	}
	// TODO: all properties and port resolution are collected
	//       host, containerPort, targetPort, hostPort, port
	// TODO: env variable can be a template and is resolved
	// TODO: lowercase ???
	// TODO: remember found host ports (prevent double assignemt)

	// TODO: Port struct: name and containerPort are mandatory

	// TODO: add shutdown hook
	// TODO: containers + ID
	return &DockerEnvironment{}, nil
}

func (r *DockerEnvironment) Start(names ...string) error {
	return nil
}

func (r *DockerEnvironment) StartParallel(names ...string) error {
	return nil
}

func (r *DockerEnvironment) Stop(names ...string) error {
	return nil
}

func (r *DockerEnvironment) Destroy(names ...string) error {
	return nil
}

func (r *DockerEnvironment) Close() {
	r.lifecycleHandler.Close()
}

// resolve template with registered variables e.g. ports
func (r *DockerEnvironment) Resolve(template string) (string, error) {
	return "", nil
}

//TODO: find port bind on all interfaces (or variable)
//TODO: take getPublicFacingIP
