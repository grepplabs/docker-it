package dockerit

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net"
	"strings"
)

type dockerEnvironmentContext struct {
	ID         string
	logger     *logger
	externalIP string
	containers map[string]*dockerContainer
}

func newDockerEnvironmentContext() (*dockerEnvironmentContext, error) {
	externalIP, err := externalIP()
	if err != nil {
		return nil, err
	}
	logger := newLogger()
	logger.Info.Println("Using IP", externalIP)
	id := uuid.New().String()
	id = id[len(id)-12:]

	return &dockerEnvironmentContext{ID: id, logger: logger, externalIP: externalIP, containers: make(map[string]*dockerContainer)}, nil
}

func normalizeName(name string) string {
	return strings.ToLower(name)
}

func (r *dockerEnvironmentContext) configurePortBindings() error {
	// we could use 0.0.0.0
	portBinding := newDockerEnvironmentPortBinding(r.externalIP, r)
	return portBinding.configurePortBindings()
}

func (r *dockerEnvironmentContext) configureContainersEnv() error {
	return r.getValueResolver().configureContainersEnv()
}

func (r *dockerEnvironmentContext) getValueResolver() *dockerEnvironmentValueResolver {
	return newDockerComponentValueResolver(r.externalIP, r)
}

func (r *dockerEnvironmentContext) addContainer(component DockerComponent) (*dockerContainer, error) {
	if component.Name == "" || component.Image == "" {
		return nil, errors.New("DockerComponent Name and Image must not be empty")
	}
	name := normalizeName(component.Name)
	container := newDockerContainer(component)
	if _, exits := r.containers[name]; exits {
		return nil, fmt.Errorf("DockerComponent [%s] is configured twice", name)
	}
	r.containers[name] = container
	return container, nil
}

func (r *dockerEnvironmentContext) getContainer(name string) (*dockerContainer, error) {
	container, exits := r.containers[normalizeName(name)]
	if !exits {
		return nil, fmt.Errorf("DockerComponent [%s] is not configured", name)
	}
	return container, nil
}

// implements ValueResolver
func (r *dockerEnvironmentContext) Resolve(templateText string) (string, error) {
	return r.getValueResolver().resolve(templateText)
}

// implements ValueResolver
func (r *dockerEnvironmentContext) Host() string {
	return r.externalIP
}

// implements ValueResolver
func (r *dockerEnvironmentContext) Port(componentName string, portName string) (int, error) {
	return r.getValueResolver().Port(componentName, portName)
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
