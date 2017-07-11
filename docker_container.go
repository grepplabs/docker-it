package dockerit

type DockerContainer struct {
	DockerComponent

	containerID  string
	portBindings []Port
	env          map[string]string

	stopFollowLogsChannel chan struct{}
}

func NewDockerContainer(component DockerComponent) *DockerContainer {
	return &DockerContainer{DockerComponent: component, stopFollowLogsChannel: make(chan struct{}, 1)}
}

func (r *DockerContainer) StopFollowLogs() {
	select {
	case r.stopFollowLogsChannel <- struct{}{}:
	default:
	}
}
