package dockerit

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"strings"
	"strconv"
	"net"
)

type DockerEnvironmentContext struct {
	ID         string
	containers map[string]*DockerContainer
}

func NewDockerEnvironmentContext() *DockerEnvironmentContext {
	id := uuid.New().String()
	return &DockerEnvironmentContext{ID: id, containers: make(map[string]*DockerContainer)}
}

func containerName(name string) string {
	return strings.ToLower(name)
}

func (r *DockerEnvironmentContext) addContainer(component DockerComponent) error {
	if component.Name == "" || component.Image == "" {
		return errors.New("DockerComponent Name and Image must not be empty")
	}
	name := containerName(component.Name)
	container := NewDockerContainer(component)
	container.DockerComponent.Name = name
	if _, exits := r.containers[name]; exits {
		return fmt.Errorf("DockerComponent [%s] is configured twice", name)
	}
	r.containers[name] = container
	return nil
}

func (r *DockerEnvironmentContext) getContainer(name string) (*DockerContainer, error) {
	if container, exits := r.containers[containerName(name)]; !exits {
		return nil, fmt.Errorf("DockerComponent [%s] is not configured", name)
	} else {
		return container, nil
	}
}

func (r *DockerEnvironmentContext) configurePortBindings(host string) error {
	componentPorts, err := r.getNormalizedExposedPorts()
	if err != nil {
		return err
	}

	portBindings, err := getPortBindings(host, componentPorts)
	if err != nil {
		return err
	}
	for containerName, ports := range portBindings {
		if container, err := r.getContainer(containerName); err != nil {
			return err
		} else {
			// container port bindings
			container.portBindings = ports
		}
	}
	return nil
}


func (r *DockerEnvironmentContext) getNormalizedExposedPorts() (map[string][]Port, error) {
	componentPorts := make(map[string][]Port)

	for containerName, container := range r.containers {
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

				portName := strings.ToLower(exposedPort.Name)
				if portName == "" {
					portName = containerName
				}

				if _, exists := namedPorts[portName]; exists {
					return nil, fmt.Errorf("DockerComponent '%s' port name '%s' is configured twice",
						containerName, portName)
				} else {
					namedPorts[portName] = struct{}{}
				}

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
	addr, err := net.ResolveTCPAddr("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		return nil, 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, 0, err
	}
	return l, l.Addr().(*net.TCPAddr).Port, nil
}