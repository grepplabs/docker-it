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

type Callback interface {
	Call() error
}

type Port struct {
	Name          string
	ContainerPort int
	HostPort      int
}
