package dockerit

import (
	"fmt"
	"net"
)

type dockerEnvironmentPortBinding struct {
	bindIP  string
	context *dockerEnvironmentContext
}

func newDockerEnvironmentPortBinding(bindIP string, context *dockerEnvironmentContext) *dockerEnvironmentPortBinding {
	return &dockerEnvironmentPortBinding{
		bindIP:  bindIP,
		context: context,
	}
}

func (r *dockerEnvironmentPortBinding) configurePortBindings() error {
	componentPorts, err := r.getNormalizedExposedPorts()
	if err != nil {
		return err
	}

	portBindings, err := getPortBindings(r.bindIP, componentPorts)
	if err != nil {
		return err
	}
	for containerName, ports := range portBindings {
		container, err := r.context.getContainer(containerName)
		if err != nil {
			return err
		}
		// container port bindings
		container.portBindings = ports
	}
	return nil
}

func (r *dockerEnvironmentPortBinding) getNormalizedExposedPorts() (map[string][]Port, error) {
	componentPorts := make(map[string][]Port)

	for containerName, container := range r.context.containers {
		if _, exists := componentPorts[containerName]; exists {
			return nil, fmt.Errorf("DockerComponent '%s' is configured twice", containerName)
		}
		ports := make([]Port, 0)
		if container.DockerComponent.ExposedPorts != nil {
			namedPorts := make(map[string]struct{})
			for _, exposedPort := range container.DockerComponent.ExposedPorts {
				if exposedPort.ContainerPort <= 0 {
					return nil, fmt.Errorf("DockerComponent '%s' ContainerPort '%d' is invalid",
						containerName, exposedPort.ContainerPort)
				}

				portName := normalizeName(exposedPort.Name)
				if portName == "" {
					portName = containerName
				}

				if _, exists := namedPorts[portName]; exists {
					return nil, fmt.Errorf("DockerComponent '%s' port name '%s' is configured twice",
						containerName, portName)
				}
				namedPorts[portName] = struct{}{}

				ports = append(ports,
					Port{
						Name:          portName,
						ContainerPort: exposedPort.ContainerPort,
						HostPort:      exposedPort.HostPort},
				)
			}
		}
		componentPorts[containerName] = ports
	}
	return componentPorts, nil
}

func getPortBindings(host string, componentPorts map[string][]Port) (map[string][]Port, error) {
	listeners := make([]*net.TCPListener, 0)
	result := make(map[string][]Port)
	for componentName, ports := range componentPorts {
		bindings := make([]Port, 0)
		for _, port := range ports {
			listener, hostPort, err := listenTCP(host, port.HostPort)
			if err != nil {
				closeTCPListeners(listeners)
				return nil, err
			}
			listeners = append(listeners, listener)
			binding := Port{
				Name:          port.Name,
				ContainerPort: port.ContainerPort,
				HostPort:      hostPort}
			bindings = append(bindings, binding)

		}
		result[componentName] = bindings
	}

	closeTCPListeners(listeners)
	return result, nil
}

func closeTCPListeners(listeners []*net.TCPListener) {
	for _, listener := range listeners {
		listener.Close()
	}
}

func listenTCP(host string, port int) (*net.TCPListener, int, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, 0, err
	}
	return l, l.Addr().(*net.TCPAddr).Port, nil
}
