package dockerit

import (
	"fmt"
	"sync"
	"github.com/docker/go-connections/nat"
	"strconv"
)

type DockerComponent struct {
	containerId string

	Name                    string
	Image                   string
	ImageLocalOnly          bool
	RemoveImageAfterDestroy bool
	ExposedPorts            []Port
	EnvironmentVariables    map[string]string
	ExposeEnvAsSystemProps  bool
	ConnectToNetwork        bool
	FollowLogs              bool
	BeforeStart             interface{}
	AfterStart              interface{}
}

func (r *DockerComponent) Accept(context EnvironmentContext) error {
	return nil
}

type DockerLifecycleHandler struct {
	dockerClient   *DockerClient
	createdClients []*DockerClient

	context EnvironmentContext

	// guard createdClients
	mu sync.Mutex
}

func NewDockerLifecycleHandler(context EnvironmentContext) (*DockerLifecycleHandler, error) {
	dockerClient, err := NewDockerClient()
	if err != nil {
		return nil, err
	}
	createdClients := []*DockerClient{dockerClient}
	return &DockerLifecycleHandler{dockerClient: dockerClient, createdClients: createdClients, context: context}, nil
}

func (r *DockerLifecycleHandler) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, createdClient := range r.createdClients {
		// ignore close errors
		createdClient.Close()
	}
	r.createdClients = []*DockerClient{}
}

func (r *DockerLifecycleHandler) Create(component DockerComponent) error {
	if exists, err := r.containerExists(component); err != nil {
		return err
	} else if exists {
		// log: component name {} , container with id {} already exists
		return nil
	}

	if err := r.checkOrPullDockerImage(component); err != nil {
		return err
	}

	if err := r.createDockerContainer(component); err != nil {
		return err
	}

	return nil
}

func (r *DockerLifecycleHandler) Start(component DockerComponent) error {
	return nil
}

func (r *DockerLifecycleHandler) Pause(component DockerComponent) error {
	return nil
}

func (r *DockerLifecycleHandler) Unpause(component DockerComponent) error {
	return nil
}

func (r *DockerLifecycleHandler) Stop(component DockerComponent) error {
	return nil
}

func (r *DockerLifecycleHandler) Destroy(component DockerComponent) error {
	return nil
}

func (r *DockerLifecycleHandler) containerExists(component DockerComponent) (bool, error) {
	if component.containerId != "" {
		if container, err := r.dockerClient.GetContainerById(component.containerId); err != nil {
			if container != nil {
				return true, nil
			}
		} else {
			return false, err
		}
	}
	return false, nil
}

func (r *DockerLifecycleHandler) checkOrPullDockerImage(component DockerComponent) error {
	var imageExists bool
	if summary, err := r.dockerClient.GetImageByName(component.Image); err != nil {
		return err
	} else {
		imageExists = summary != nil
	}

	if component.ImageLocalOnly {
		if imageExists {
			return nil
		} else {
			return fmt.Errorf("Local images %s does not exist", component.Image)
		}
	}
	if err := r.dockerClient.PullImage(component.Image); err != nil {
		if imageExists {
			// log: image cannot be pulled
			return nil
		} else {
			return err
		}
	}
	return nil
}

func (r *DockerLifecycleHandler) createDockerContainer(component DockerComponent) error {
	containerName := r.getContainerName(component)

	exposedPorts := make(nat.PortSet)
	if component.ExposedPorts != nil {
		for _, exposedPort := range component.ExposedPorts {
			port, err := nat.NewPort("tcp", strconv.Itoa(exposedPort.ContainerPort))
			if err != nil {
				return err
			}
			exposedPorts[port] = struct{}{}

		}
	}
	r.dockerClient.CreateContainer(containerName, component.Image, nil, exposedPorts, nil)
	return nil
}

func (r *DockerLifecycleHandler) getContainerName(component DockerComponent) string {
	containerName := component.Name
	if r.context.ID != "" {
		containerName += "-" + r.context.ID
	}
	return containerName
}
