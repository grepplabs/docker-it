package dockerit

import (
	"sync"
)

type DockerComponent struct {
	Name                    string
	Image                   string
	ImageLocalOnly          bool
	RemoveImageAfterDestroy bool
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

	containerID  string
	portBindings []Port
	env          map[string]string

	stopFollowLogsChannel chan struct{}
	stopFollowLogsOnce    sync.Once
}

type Callback interface {
	Call() error
}

type Port struct {
	Name          string
	ContainerPort int
	HostPort      int
}

func NewDockerContainer(component DockerComponent) *DockerContainer {
	return &DockerContainer{DockerComponent: component, stopFollowLogsChannel: make(chan struct{}, 1)}
}

func (r *DockerContainer) StopFollowLogs() {
	r.stopFollowLogsOnce.Do(func() {
		close(r.stopFollowLogsChannel)
	})
}
