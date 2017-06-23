package dockerit

type EnvironmentContext struct {
	ID string
}

type Port struct {
	Name          string
	ContainerPort int
	HostPort      int
}
