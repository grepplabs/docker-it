package dockerit

import (
	"sync"
)

type DockerComponent struct {
	Name                    string
	Image                   string
	ImageLocalOnly          bool
	RemoveImageAfterDestroy bool
	// TODO: rename ExposedPorts to PortBindings
	// TODO: mount volumes ?
	ExposedPorts           []Port
	EnvironmentVariables   map[string]string
	ExposeEnvAsSystemProps bool
	ConnectToNetwork       bool
	FollowLogs             bool
	BeforeStart            Callback
	AfterStart             Callback
}

type DockerContainer struct {
	DockerComponent

	containerID           string
	stopFollowLogsChannel chan struct{}
	stopFollowLogsOnce    sync.Once
}

func NewDockerContainer(component DockerComponent) *DockerContainer {
	return &DockerContainer{DockerComponent: component, stopFollowLogsChannel: make(chan struct{}, 1)}
}

func (r *DockerContainer) StopFollowLogs() {
	r.stopFollowLogsOnce.Do(func() {
		close(r.stopFollowLogsChannel)
	})
}
