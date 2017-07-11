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
	return &dockerEnvironmentContext{ID: id, logger: logger, externalIP: externalIP, containers: make(map[string]*dockerContainer)}, nil
}

func normalizeName(name string) string {
	return strings.ToLower(name)
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
	if container, exits := r.containers[normalizeName(name)]; !exits {
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
