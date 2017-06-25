package dockerit

type EnvironmentContext struct {
	ID string
}

type Callback interface {
	Call(EnvironmentContext) error
}

type Port struct {
	Name          string
	ContainerPort int
	HostPort      int
}
