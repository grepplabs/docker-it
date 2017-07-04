package dockerit

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net"
	"strings"
)

type DockerEnvironmentContext struct {
	ID         string
	externalIP string
	containers map[string]*DockerContainer
}

func NewDockerEnvironmentContext() (*DockerEnvironmentContext, error) {
	externalIP, err := externalIP()
	if err != nil {
		return nil, err
	}
	id := uuid.New().String()
	return &DockerEnvironmentContext{ID: id, externalIP: externalIP, containers: make(map[string]*DockerContainer)}, nil
}

func toContainerName(name string) string {
	return strings.ToLower(name)
}

func (r *DockerEnvironmentContext) addContainer(component DockerComponent) error {
	if component.Name == "" || component.Image == "" {
		return errors.New("DockerComponent Name and Image must not be empty")
	}
	name := toContainerName(component.Name)
	container := NewDockerContainer(component)
	if _, exits := r.containers[name]; exits {
		return fmt.Errorf("DockerComponent [%s] is configured twice", name)
	}
	r.containers[name] = container
	return nil
}

func (r *DockerEnvironmentContext) getContainer(name string) (*DockerContainer, error) {
	if container, exits := r.containers[toContainerName(name)]; !exits {
		return nil, fmt.Errorf("DockerComponent [%s] is not configured", name)
	} else {
		return container, nil
	}
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
