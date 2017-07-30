package dockerit

import (
	"context"
	"github.com/docker/docker/api/types"
	typesContainer "github.com/docker/docker/api/types/container"
	typesFilters "github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stringid"
	"github.com/docker/go-connections/nat"
	"io"
	"io/ioutil"
	"os"
)

const (
	DEFAULT_DOCKER_API_VERSION = "1.23"
)

type dockerClient struct {
	client *client.Client
}

func SetDefaultDockerApiVersion() {
	// ensure docker API version
	if os.Getenv("DOCKER_API_VERSION") == "" {
		os.Setenv("DOCKER_API_VERSION", DEFAULT_DOCKER_API_VERSION)
	}
}

func NewDockerClient() (*dockerClient, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return &dockerClient{client: cli}, nil
}

func (r *dockerClient) Close() error {
	return r.client.Close()
}

func (r *dockerClient) GetContainerByID(containerID string) (*types.Container, error) {
	options := types.ContainerListOptions{All: true}
	if containers, err := r.client.ContainerList(context.Background(), options); err != nil {
		return nil, err
	} else {
		for _, container := range containers {
			if container.ID == containerID {
				return &container, nil
			}
		}
	}
	return nil, nil
}

func (r *dockerClient) GetImageByName(imageName string) (*types.ImageSummary, error) {
	// https://docs.docker.com/engine/api/v1.29/#operation/ImageList
	imageFilters := typesFilters.NewArgs()
	imageFilters.Add("reference", imageName)
	options := types.ImageListOptions{All: false, Filters: imageFilters}
	if summaries, err := r.client.ImageList(context.Background(), options); err != nil {
		return nil, err
	} else {
		if len(summaries) != 0 {
			return &summaries[0], nil
		}
	}
	return nil, nil
}

func (r *dockerClient) RemoveImageByName(imageName string) error {
	// https://docs.docker.com/engine/api/v1.29/#operation/ImageList
	imageFilters := typesFilters.NewArgs()
	imageFilters.Add("reference", imageName)
	options := types.ImageListOptions{All: false, Filters: imageFilters}
	if summaries, err := r.client.ImageList(context.Background(), options); err != nil {
		return err
	} else {
		for _, summary := range summaries {
			if err = r.RemoveImage(summary.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *dockerClient) RemoveImage(imageID string) error {
	options := types.ImageRemoveOptions{Force: true}
	_, err := r.client.ImageRemove(context.Background(), imageID, options)
	return err

}

func (r *dockerClient) PullImage(imageName string) error {
	options := types.ImagePullOptions{}
	resp, err := r.client.ImagePull(context.Background(), imageName, options)
	if err != nil {
		return err
	}
	_, err = ioutil.ReadAll(resp)
	if err != nil {
		return err
	}
	return nil
}

func (r *dockerClient) CreateContainer(containerName string, image string, env []string, portSpecs []string) (string, error) {
	// ip:public:private/proto
	exposedPorts, portBindings, err := nat.ParsePortSpecs(portSpecs)
	if err != nil {
		return "", err
	}
	config := typesContainer.Config{
		Image:        image,
		Env:          env,
		ExposedPorts: exposedPorts,
	}
	hostConfig := typesContainer.HostConfig{
		PortBindings: portBindings,
	}

	if body, err := r.client.ContainerCreate(context.Background(), &config, &hostConfig, nil, containerName); err != nil {
		return "", err
	} else {
		// return container ID
		return body.ID, nil
	}
}

func (r *dockerClient) StartContainer(containerID string) error {
	options := types.ContainerStartOptions{}
	return r.client.ContainerStart(context.Background(), containerID, options)
}

func (r *dockerClient) ContainerLogs(containerID string, follow bool) (io.ReadCloser, error) {
	options := types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: follow}
	return r.client.ContainerLogs(context.Background(), containerID, options)
}

func (r *dockerClient) StopContainer(containerID string) error {
	return r.client.ContainerStop(context.Background(), containerID, nil)
}

func (r *dockerClient) RemoveContainer(containerID string) error {
	options := types.ContainerRemoveOptions{RemoveVolumes: true, Force: true}
	return r.client.ContainerRemove(context.Background(), containerID, options)
}

func TruncateID(id string) string {
	return stringid.TruncateID(id)
}
