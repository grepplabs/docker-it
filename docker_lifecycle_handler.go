package dockerit

import (
	"errors"
	"fmt"
	"github.com/docker/docker/pkg/stdcopy"
	"io"
	"strings"
)

type dockerLifecycleHandler struct {
	dockerClient *dockerClient
	context      *dockerEnvironmentContext
}

func newDockerLifecycleHandler(context *dockerEnvironmentContext) (*dockerLifecycleHandler, error) {
	dockerClient, err := newDockerClient()
	if err != nil {
		return nil, err
	}
	return &dockerLifecycleHandler{dockerClient: dockerClient, context: context}, nil
}

func (r *dockerLifecycleHandler) Close() {
	r.context.logger.Info.Println("Closing docker lifecycle handler")

	for _, container := range r.context.containers {
		container.stopFollowLogs()
	}
	r.dockerClient.Close()
}

func (r *dockerLifecycleHandler) Create(container *dockerContainer) error {
	if exists, err := r.containerExists(container.containerID); err != nil {
		return err
	} else if exists {
		r.context.logger.Info.Println("Component", container.Name, "already exists, container", TruncateID(container.containerID))
		return nil
	}

	if err := r.checkOrPullDockerImage(container.Image, container.ForcePull); err != nil {
		return err
	}

	if err := r.createDockerContainer(container); err != nil {
		return err
	}
	r.context.logger.Info.Println("Created new container", TruncateID(container.containerID), "for", container.Name)

	return nil
}

func (r *dockerLifecycleHandler) Start(container *dockerContainer) error {
	r.context.logger.Info.Println("Start component", container.Name)

	if container.containerID == "" {
		if err := r.Create(container); err != nil {
			return err
		}
	}
	if running, err := r.isContainerRunning(container.containerID); err != nil {
		return err
	} else if running {
		r.context.logger.Info.Println("Component", container.Name, "is already running", TruncateID(container.containerID))
		return nil
	}

	r.context.logger.Info.Println("Starting container", TruncateID(container.containerID), "for", container.Name)
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
		if err := container.AfterStart.Call(container.Name, r.context); err != nil {
			return err

		}
	}
	return nil
}

func (r *dockerLifecycleHandler) Stop(container *dockerContainer) error {
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

func (r *dockerLifecycleHandler) Destroy(container *dockerContainer) error {
	r.context.logger.Info.Println("Destroy component", container.Name, "container", TruncateID(container.containerID))

	container.stopFollowLogs()

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

	r.context.logger.Info.Println("Remove container", TruncateID(container.containerID))
	if err := r.dockerClient.RemoveContainer(container.containerID); err != nil {
		return err
	}
	container.containerID = ""

	if container.RemoveImageAfterDestroy {
		r.context.logger.Info.Println("Remove image", container.Image)
		if err := r.dockerClient.RemoveImageByName(container.Image); err != nil {
			return err
		}
	}
	return nil
}

func (r *dockerLifecycleHandler) isContainerRunning(containerID string) (bool, error) {
	if containerID == "" {
		return false, errors.New("isContainerRunning: containerID must not be empty")
	}
	container, err := r.dockerClient.GetContainerByID(containerID)
	if err != nil {
		return false, err
	}
	if container != nil {
		return strings.ToLower(container.State) == "running", nil
	}
	return false, fmt.Errorf("Container with ID %s does not exist", containerID)

}

func (r *dockerLifecycleHandler) containerExists(containerID string) (bool, error) {
	if containerID != "" {
		container, err := r.dockerClient.GetContainerByID(containerID)
		if err != nil {
			return false, err
		}
		if container != nil {
			return true, nil
		}
	}
	return false, nil
}

func (r *dockerLifecycleHandler) checkOrPullDockerImage(image string, forcePull bool) error {
	summary, err := r.dockerClient.GetImageByName(image)
	if err != nil {
		return err
	}
	imageExists := summary != nil

	if !forcePull {
		if imageExists {
			return nil
		}
		return fmt.Errorf("Local images %s does not exist", image)
	}
	r.context.logger.Info.Println("Pulling image", image)
	if err := r.dockerClient.PullImage(image); err != nil {
		if imageExists {
			r.context.logger.Info.Println("Image", image, "cannot be pulled, using existing one")
			return nil
		}
		return err
	}
	return nil
}

func (r *dockerLifecycleHandler) createDockerContainer(container *dockerContainer) error {
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
	containerID, err := r.dockerClient.CreateContainer(containerName, container.Image, env, portSpecs)
	if err != nil {
		return err
	}
	container.containerID = containerID
	return nil
}

func (r *dockerLifecycleHandler) getContainerName(name string) string {
	var containerName string
	if r.context.ID != "" {
		containerName += name + "-" + r.context.ID
	} else {
		containerName = name
	}
	return normalizeName(containerName)
}

func (r *dockerLifecycleHandler) fetchLogs(containerID string, dstout, dsterr io.Writer) error {
	reader, err := r.dockerClient.ContainerLogs(containerID, false)
	if err != nil {
		return err
	}
	_, err = stdcopy.StdCopy(dstout, dsterr, reader)
	return err
}

func (r *dockerLifecycleHandler) followLogs(container *dockerContainer, dstout, dsterr io.Writer) error {
	followClient, err := newDockerClient()
	if err != nil {
		return err
	}

	reader, err := followClient.ContainerLogs(container.containerID, true)
	if err != nil {
		return err
	}
	r.context.logger.Info.Println("Start follow logs", TruncateID(container.containerID))
	go func() {
		defer followClient.Close()
		for {
			select {
			case <-container.stopFollowLogsChannel:
				r.context.logger.Info.Println("Received stop follow logs", TruncateID(container.containerID))
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
