package dockerit

type DockerComponent struct {
	Name                    string
	Image                   string
	ForcePull               bool
	RemoveImageAfterDestroy bool
	ExposedPorts            []Port
	EnvironmentVariables    map[string]string
	FollowLogs              bool
	AfterStart              Callback
}

type Callback interface {
	Call(componentName string, resolver ValueResolver) error
}

type ValueResolver interface {
	Resolve(template string) (string, error)
	Host() string
	Port(componentName string, portName string) (int, error)
}

type Port struct {
	Name          string
	ContainerPort int
	HostPort      int
}
