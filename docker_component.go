package dockerit

type DockerComponent struct {
	Name                    string
	Image                   string
	ImageLocalOnly          bool
	RemoveImageAfterDestroy bool
	ExposedPorts            []Port
	EnvironmentVariables    map[string]string
	FollowLogs              bool
	BeforeStart             Callback
	AfterStart              Callback
}

type DockerContainer struct {
	DockerComponent

	containerID  string
	portBindings []Port
	env          map[string]string

	stopFollowLogsChannel chan struct{}
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
	select {
	case r.stopFollowLogsChannel <- struct{}{}:
	default:
	}
}
