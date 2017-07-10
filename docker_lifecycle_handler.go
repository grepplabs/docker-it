package dockerit

import (
	"errors"
	"fmt"
	"github.com/docker/docker/pkg/stdcopy"
	"io"
	"strings"
)

type DockerLifecycleHandler struct {
	dockerClient *DockerClient
	context      *DockerEnvironmentContext
}

func NewDockerLifecycleHandler(context *DockerEnvironmentContext) (*DockerLifecycleHandler, error) {
	dockerClient, err := NewDockerClient()
	if err != nil {
		return nil, err
	}
	return &DockerLifecycleHandler{dockerClient: dockerClient, context: context}, nil
}

func (r *DockerLifecycleHandler) Close() {
	r.context.logger.Info.Println("Closing docker lifecycle handler")

	for _, container := range r.context.containers {
		container.StopFollowLogs()
	}
	r.dockerClient.Close()
}

func (r *DockerLifecycleHandler) Create(container *DockerContainer) error {
	if exists, err := r.containerExists(container.containerID); err != nil {
		return err
	} else if exists {
		r.context.logger.Info.Println("Component", container.Name, "already exists, container", container.containerID)
		return nil
	}

	if err := r.checkOrPullDockerImage(container.Image, container.ImageLocalOnly); err != nil {
		return err
	}

	if err := r.createDockerContainer(container); err != nil {
		return err
	}
	r.context.logger.Info.Println("Created new container", container.containerID, "for", container.Name)

	return nil
}

func (r *DockerLifecycleHandler) Start(container *DockerContainer) error {
	r.context.logger.Info.Println("Start component", container.Name)

	if container.containerID == "" {
		if err := r.Create(container); err != nil {
			return err
		}
	}
	if running, err := r.isContainerRunning(container.containerID); err != nil {
		return err
	} else if running {
		r.context.logger.Info.Println("Component", container.Name, "is already running", container.containerID)
		return nil
	}

	if container.BeforeStart != nil {
		if err := container.BeforeStart.Call(); err != nil {
			return err
		}
	}
	r.context.logger.Info.Println("Starting container", container.containerID, "for", container.Name)
	if err := r.dockerClient.StartContainer(container.containerID); err != nil {
		// try to fetch logs from container
		out := stdoutWriter(container.Name)
		r.fetchLogs(container.containerID, out, out)
		return err
	}
	if container.FollowLogs {
		out := stdoutWriter(container.Name)
		if err := r.followLogs(container, out, out); err != nil {
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
	r.context.logger.Info.Println("Stop component", container.Name)

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
	r.context.logger.Info.Println("Destroy component", container.Name)

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

	r.context.logger.Info.Println("Remove container", container.containerID)
	if err := r.dockerClient.RemoveContainer(container.containerID); err != nil {
		return err
	}

	if container.RemoveImageAfterDestroy {
		r.context.logger.Info.Println("Remove image", container.Image)
		if err := r.dockerClient.RemoveImageByName(container.Image); err != nil {
			return err
		}
	}
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
			return strings.ToLower(container.State) == "running", nil
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
	r.context.logger.Info.Println("Pulling image", image)
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

	portSpecs := make([]string, 0)
	if container.portBindings != nil {
		for _, portBinding := range container.portBindings {
			// ip:public:private/proto
			portSpec := fmt.Sprintf("%s:%d:%d/%s", r.context.externalIP, portBinding.HostPort, portBinding.ContainerPort, "tcp")
			portSpecs = append(portSpecs, portSpec)
		}
	}
	env := make([]string, 0)
	if container.env != nil {
		for k, v := range container.env {
			env = append(env, k+"="+v)
		}
	}
	r.context.logger.Info.Println("Creating container for", container.Name, "name", containerName, "env", env, "portSpecs", portSpecs)
	if containerID, err := r.dockerClient.CreateContainer(containerName, container.Image, env, portSpecs); err != nil {
		return err
	} else {
		container.containerID = containerID
		return nil
	}
}

func (r *DockerLifecycleHandler) getContainerName(name string) string {
	var containerName string
	if r.context.ID != "" {
		containerName += name + "-" + r.context.ID
	} else {
		containerName = name
	}
	return normalizeName(containerName)
}

func (r *DockerLifecycleHandler) getNetworkName() string {
	networkName := "docker-environment"
	if r.context.ID != "" {
		networkName += "-" + r.context.ID
	}
	return networkName
}

func (r *DockerLifecycleHandler) fetchLogs(containerID string, dstout, dsterr io.Writer) error {
	reader, err := r.dockerClient.ContainerLogs(containerID, false)
	if err != nil {
		return err
	}
	_, err = stdcopy.StdCopy(dstout, dsterr, reader)
	return err
}

func (r *DockerLifecycleHandler) followLogs(container *DockerContainer, dstout, dsterr io.Writer) error {
	followClient, err := NewDockerClient()
	if err != nil {
		return err
	}

	reader, err := followClient.ContainerLogs(container.containerID, true)
	if err != nil {
		return err
	}
	r.context.logger.Info.Println("Start follow logs", container.containerID)
	go func() {
		defer followClient.Close()
		for {
			select {
			case <-container.stopFollowLogsChannel:
				r.context.logger.Info.Println("Received stop follow logs", container.containerID)
				return
			}
		}

	}()
	go func() {
		_, err = stdcopy.StdCopy(dstout, dsterr, reader)
		if err != nil && err != io.EOF {
			r.context.logger.Error.Println("Follow logs error", err)
		}
	}()
	return nil
}
