package dockerit

type dockerContainer struct {
	DockerComponent

	containerID  string
	portBindings []Port
	env          map[string]string

	stopFollowLogsChannel chan struct{}
}

func newDockerContainer(component DockerComponent) *dockerContainer {
	return &dockerContainer{DockerComponent: component, stopFollowLogsChannel: make(chan struct{}, 1)}
}

func (r *dockerContainer) StopFollowLogs() {
	select {
	case r.stopFollowLogsChannel <- struct{}{}:
	default:
	}
}
