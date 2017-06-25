package dockerit

import (
	"context"
	dockerTypes "github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	dockerFilters "github.com/docker/docker/api/types/filters"
	dockerClient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"io/ioutil"
)

type DockerClient struct {
	client *dockerClient.Client
}

func NewDockerClient() (*DockerClient, error) {
	cli, err := dockerClient.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return &DockerClient{client: cli}, nil
}

func (r *DockerClient) Close() error {
	return r.client.Close()
}

func (r *DockerClient) GetContainerById(containerId string) (*dockerTypes.Container, error) {
	options := dockerTypes.ContainerListOptions{All: true}
	if containers, err := r.client.ContainerList(context.Background(), options); err != nil {
		return nil, err
	} else {
		for _, container := range containers {
			if container.ID == containerId {
				return &container, nil
			}
		}
	}
	return nil, nil
}

func (r *DockerClient) GetImageByName(image string) (*dockerTypes.ImageSummary, error) {
	imageFilters := dockerFilters.NewArgs()
	imageFilters.Add("reference", image)
	options := dockerTypes.ImageListOptions{Filters: imageFilters}
	if summaries, err := r.client.ImageList(context.Background(), options); err != nil {
		return nil, err
	} else {
		if len(summaries) != 0 {
			return &summaries[0], nil
		}
	}
	return nil, nil
}

func (r *DockerClient) PullImage(image string) error {
	options := dockerTypes.ImagePullOptions{}
	resp, err := r.client.ImagePull(context.Background(), image, options)
	if err != nil {
		return err
	}
	_, err = ioutil.ReadAll(resp)
	if err != nil {
		return err
	}
	return nil
}

func (r *DockerClient) CreateContainer(containerName string, image string, env []string, portSpecs []string) (string, error) {
	// ip:public:private/proto
	exposedPorts, portBindings, err := nat.ParsePortSpecs(portSpecs)
	if err != nil {
		return "", err
	}
	config := dockerContainer.Config{
		Image:        image,
		Env:          env,
		ExposedPorts: exposedPorts,
	}
	hostConfig := dockerContainer.HostConfig{
		PortBindings: portBindings,
	}

	if body, err := r.client.ContainerCreate(context.Background(), &config, &hostConfig, nil, containerName); err != nil {
		return "", err
	} else {
		// return container ID
		return body.ID, nil
	}
}
