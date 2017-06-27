package dockerit

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type DockerLifecycleHandler struct {
	dockerClient *DockerClient
	context      DockerEnvironmentContext
}

func NewDockerLifecycleHandler(context DockerEnvironmentContext) (*DockerLifecycleHandler, error) {
	dockerClient, err := NewDockerClient()
	if err != nil {
		return nil, err
	}
	// TODO: use EnvironmentContext to create a DockerEnvironment context
	return &DockerLifecycleHandler{dockerClient: dockerClient, context: context}, nil
}

func (r *DockerLifecycleHandler) Close() {
	for _, container := range r.context.containers {
		container.StopFollowLogs()
	}
	r.dockerClient.Close()
}

func (r *DockerLifecycleHandler) Create(container *DockerContainer) error {
	if exists, err := r.containerExists(container.containerID); err != nil {
		return err
	} else if exists {
		// log: component name {} , container with id {} already exists
		return nil
	}

	if err := r.checkOrPullDockerImage(container.Image, container.ImageLocalOnly); err != nil {
		return err
	}

	if err := r.createDockerContainer(container); err != nil {
		return err
	}
	// log: component.containerId
	if container.ConnectToNetwork {
		return r.connectToNetwork(container.containerID, container.Name)
	}

	return nil
}

func (r *DockerLifecycleHandler) Start(container *DockerContainer) error {
	if container.containerID == "" {
		if err := r.Create(container); err != nil {
			return err
		}
	}
	if running, err := r.isContainerRunning(container.containerID); err != nil {
		return err
	} else if running {
		// log: container is running
		return nil
	}

	if container.BeforeStart != nil {
		if err := container.BeforeStart.Call(); err != nil {
			return err
		}
	}
	if err := r.dockerClient.StartContainer(container.containerID); err != nil {
		// try to fetch logs from container
		r.fetchLogs(container.containerID, os.Stderr)
		return err
	}
	if container.FollowLogs {
		if err := r.followLogs(container, os.Stdout); err != nil {
			return err
		}
	}
	if container.AfterStart != nil {
		if err := container.AfterStart.Call(); err != nil {
			return err

		}
	}
	return nil
}

func (r *DockerLifecycleHandler) Stop(container *DockerContainer) error {
	if container.containerID == "" {
		return nil
	}
	if result, err := r.isContainerRunning(container.containerID); err != nil {
		return err
	} else if result {
		return r.dockerClient.StopContainer(container.containerID)
	}
	return nil
}

func (r *DockerLifecycleHandler) Destroy(container *DockerContainer) error {
	//TODO: disallow Start() et.c operations after destroy
	//TODO: could send to channel and doOnce is not required
	container.StopFollowLogs()

	if container.containerID == "" {
		return nil
	}

	if exists, err := r.containerExists(container.containerID); err != nil {
		return err
	} else if !exists {
		return nil
	}

	if running, err := r.isContainerRunning(container.containerID); err != nil {
		return err
	} else if running {
		r.Stop(container)
	}

	if err := r.dockerClient.RemoveContainer(container.containerID); err != nil {
		return err
	}

	if container.RemoveImageAfterDestroy {
		if err := r.dockerClient.RemoveImageByName(container.Image); err != nil {
			return err
		}
	}
	// TODO: should delete network if a new one (not default) was created
	return nil
}

func (r *DockerLifecycleHandler) isContainerRunning(containerID string) (bool, error) {
	if containerID == "" {
		return false, errors.New("isContainerRunning: containerID must not be empty")
	}
	if container, err := r.dockerClient.GetContainerByID(containerID); err != nil {
		return false, err
	} else {
		if container != nil {
			return strings.ToLower(container.Status) == "up", nil
		} else {
			return false, fmt.Errorf("Container with ID %s does not exist", containerID)
		}
	}
}

func (r *DockerLifecycleHandler) containerExists(containerID string) (bool, error) {
	if containerID != "" {
		if container, err := r.dockerClient.GetContainerByID(containerID); err != nil {
			return false, err
		} else {
			if container != nil {
				return true, nil
			}
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

func (r *DockerLifecycleHandler) createDockerContainer(container *DockerContainer) error {
	containerName := r.getContainerName(container.Name)
	// bind on all interfaces
	// TODO: can be provider as variable
	ip := "0.0.0.0"

	portSpecs := make([]string, 0)
	if container.ExposedPorts != nil {
		for _, exposedPort := range container.ExposedPorts {
			// ip:public:private/proto
			portSpec := fmt.Sprintf("%s:%d:%d/%s", ip, exposedPort.HostPort, exposedPort.ContainerPort, "tcp")
			portSpecs = append(portSpecs, portSpec)
		}
	}
	env := make([]string, 0)
	if container.EnvironmentVariables != nil {
		for k, v := range container.EnvironmentVariables {
			env = append(env, k+"="+v)
		}
	}
	if containerID, err := r.dockerClient.CreateContainer(containerName, container.Image, env, portSpecs); err != nil {
		return err
	} else {
		container.containerID = containerID
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

func (r *DockerLifecycleHandler) fetchLogs(containerID string, dst io.Writer) error {
	reader, err := r.dockerClient.ContainerLogs(containerID, false)
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, reader)
	return err
}

func (r *DockerLifecycleHandler) followLogs(container *DockerContainer, dst io.Writer) error {
	followClient, err := NewDockerClient()
	if err != nil {
		return err
	}
	reader, err := followClient.ContainerLogs(container.containerID, true)
	if err != nil {
		return err
	}
	go func() {
		defer followClient.Close()
		for {
			select {
			case <-container.stopFollowLogsChannel:
				return
			}
		}

	}()
	go func() {
		// TODO: prefix lines / e.g. special logger for each 'name'
		// TODO: Writer to copy the log from
		_, err = io.Copy(dst, reader)
		if err != nil && err != io.EOF {
			// TODO: log error
		}
	}()
	return nil
}

func (r *DockerLifecycleHandler) getPublicFacingIP() string {
	//TODO:  implement - required for host resolution in value resolver
	//       or it can be provider as variable
	return "127.0.0.1"
}
