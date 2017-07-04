package dockerit

import (
	"errors"
	"net"
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
	context := NewDockerEnvironmentContext()
	for _, component := range components {
		if err := context.addContainer(component); err != nil {
			return nil, err
		}
	}
	ip, err := externalIP()
	if err != nil {
		return nil, err
	}
	// we could use 0.0.0.0
	portBinding := NewDockerEnvironmentPortBinding(ip, context)
	if err := portBinding.configurePortBindings(); err != nil {
		return nil, err
	}

	valueResolver := NewDockerComponentValueResolver(ip, context)
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

// https://play.golang.org/p/BDt3qEQ_2H
func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

