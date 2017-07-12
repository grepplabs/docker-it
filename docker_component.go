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
	Call(componentName string, resolver ValueResolver) error
}

type ValueResolver interface {
	Resolve(template string) (string, error)
	Host() string
	Port(componentName string, portName string) (string, error)
}

type Port struct {
	Name          string
	ContainerPort int
	HostPort      int
}
