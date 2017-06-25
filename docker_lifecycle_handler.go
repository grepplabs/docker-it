package dockerit

import (
	"fmt"
	"sync"
)

type DockerComponent struct {
	containerID string

	Name                    string
	Image                   string
	ImageLocalOnly          bool
	RemoveImageAfterDestroy bool
	// TODO: rename ExposedPorts to PortBindings
	ExposedPorts           []Port
	EnvironmentVariables   map[string]string
	ExposeEnvAsSystemProps bool
	ConnectToNetwork       bool
	FollowLogs             bool
	BeforeStart            interface{}
	AfterStart             interface{}
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

func (r *DockerLifecycleHandler) Create(component *DockerComponent) error {
	if exists, err := r.containerExists(component.containerID); err != nil {
		return err
	} else if exists {
		// log: component name {} , container with id {} already exists
		return nil
	}

	if err := r.checkOrPullDockerImage(component.Image, component.ImageLocalOnly); err != nil {
		return err
	}

	if err := r.createDockerContainer(component); err != nil {
		return err
	}
	// // log: component.containerId
	if component.ConnectToNetwork {
		return r.connectToNetwork(component.containerID, component.Name)
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
	// TODO: should delete network if a new one (not default) was created
	return nil
}

func (r *DockerLifecycleHandler) containerExists(containerID string) (bool, error) {
	if containerID != "" {
		if container, err := r.dockerClient.GetContainerByID(containerID); err != nil {
			if container != nil {
				return true, nil
			}
		} else {
			return false, err
		}
	}
	return false, nil
}

func (r *DockerLifecycleHandler) checkOrPullDockerImage(image string, imageLocalOnly bool) error {
	var imageExists bool
	if summary, err := r.dockerClient.GetImageByName(image); err != nil {
		return err
	} else {
		imageExists = summary != nil
	}

	if imageLocalOnly {
		if imageExists {
			return nil
		} else {
			return fmt.Errorf("Local images %s does not exist", image)
		}
	}
	if err := r.dockerClient.PullImage(image); err != nil {
		if imageExists {
			// log: image cannot be pulled
			return nil
		} else {
			return err
		}
	}
	return nil
}

func (r *DockerLifecycleHandler) createDockerContainer(component *DockerComponent) error {
	containerName := r.getContainerName(component.Name)
	ip := r.getIP()

	portSpecs := make([]string, 0)
	if component.ExposedPorts != nil {
		for _, exposedPort := range component.ExposedPorts {
			// ip:public:private/proto
			portSpec := fmt.Sprintf("%s:%d:%d/%s", ip, exposedPort.HostPort, exposedPort.ContainerPort, "tcp")
			portSpecs = append(portSpecs, portSpec)
		}
	}
	env := make([]string, 0)
	if component.EnvironmentVariables != nil {
		for k, v := range component.EnvironmentVariables {
			env = append(env, k+"="+v)
		}
	}
	if containerID, err := r.dockerClient.CreateContainer(containerName, component.Image, env, portSpecs); err != nil {
		return err
	} else {
		component.containerID = containerID
		return nil
	}
}

func (r *DockerLifecycleHandler) connectToNetwork(containerID string, name string) error {
	networkID, err := r.getOrCreateNetwork()
	if err != nil {
		return err
	} else {
		return r.dockerClient.ConnectToNetwork(networkID, containerID, []string{name})
	}
}
func (r *DockerLifecycleHandler) getOrCreateNetwork() (string, error) {
	networkName := r.getNetworkName()
	networkID, err := r.dockerClient.GetNetworkIDByName(networkName)
	if err != nil {
		return "", err
	} else if networkID == "" {
		return r.dockerClient.CreateNetwork(networkName)
	}
	return networkID, nil
}

func (r *DockerLifecycleHandler) getContainerName(name string) string {
	containerName := name
	if r.context.ID != "" {
		containerName += "-" + r.context.ID
	}
	return containerName
}

func (r *DockerLifecycleHandler) getNetworkName() string {
	networkName := "docker-environment"
	if r.context.ID != "" {
		networkName += "-" + r.context.ID
	}
	return networkName
}

func (r *DockerLifecycleHandler) getIP() string {
	return "0.0.0.0"
}
