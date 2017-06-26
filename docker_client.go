package dockerit

import (
	"context"
	dockerTypes "github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	dockerFilters "github.com/docker/docker/api/types/filters"
	dockerNetwork "github.com/docker/docker/api/types/network"
	dockerClient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"io"
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

func (r *DockerClient) GetContainerByID(containerID string) (*dockerTypes.Container, error) {
	options := dockerTypes.ContainerListOptions{All: true}
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

func (r *DockerClient) GetImageByName(imageName string) (*dockerTypes.ImageSummary, error) {
	// https://docs.docker.com/engine/api/v1.29/#operation/ImageList
	imageFilters := dockerFilters.NewArgs()
	imageFilters.Add("reference", imageName)
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

func (r *DockerClient) RemoveImageByName(imageName string) error {
	if summary, err := r.GetImageByName(imageName); err != nil {
		return err
	} else if summary != nil {
		return r.RemoveImage(summary.ID)
	}
	return nil
}

func (r *DockerClient) RemoveImage(imageID string) error {
	options := dockerTypes.ImageRemoveOptions{}
	_, err := r.client.ImageRemove(context.Background(), imageID, options)
	return err

}

func (r *DockerClient) PullImage(imageName string) error {
	options := dockerTypes.ImagePullOptions{}
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

func (r *DockerClient) StartContainer(containerID string) error {
	options := dockerTypes.ContainerStartOptions{}
	return r.client.ContainerStart(context.Background(), containerID, options)
}

func (r *DockerClient) GetNetworkIDByName(networkName string) (string, error) {
	// https://docs.docker.com/engine/api/v1.29/#operation/NetworkList
	networkFilters := dockerFilters.NewArgs()
	networkFilters.Add("name", networkName)
	options := dockerTypes.NetworkListOptions{Filters: networkFilters}
	if networkResources, err := r.client.NetworkList(context.Background(), options); err != nil {
		return "", err
	} else {
		if len(networkResources) != 0 {
			return networkResources[0].ID, nil
		}
	}
	return "", nil
}

func (r *DockerClient) CreateNetwork(networkName string) (string, error) {
	options := dockerTypes.NetworkCreate{}
	if response, err := r.client.NetworkCreate(context.Background(), networkName, options); err != nil {
		return "", err
	} else {
		return response.ID, nil
	}
}

func (r *DockerClient) ConnectToNetwork(networkID string, containerID string, aliases []string) error {
	options := &dockerNetwork.EndpointSettings{NetworkID: networkID, Aliases: aliases}
	return r.client.NetworkConnect(context.Background(), networkID, containerID, options)
}

func (r *DockerClient) ContainerLogs(containerID string, follow bool) (io.ReadCloser, error) {
	options := dockerTypes.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: follow}
	return r.client.ContainerLogs(context.Background(), containerID, options)
}

func (r *DockerClient) StopContainer(containerID string) error {
	return r.client.ContainerStop(context.Background(), containerID, nil)
}

func (r *DockerClient) RemoveContainer(containerID string) error {
	options := dockerTypes.ContainerRemoveOptions{RemoveVolumes: true}
	return r.client.ContainerRemove(context.Background(), containerID, options)
}
