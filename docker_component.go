package dockerit

// DockerComponent holds parameters defining docker component.
type DockerComponent struct {
	// Name of the docker component
	Name string
	// Docker image name
	Image string
	// Pull an image from a registry
	ForcePull bool
	// After destroy remove image from the docker host
	RemoveImageAfterDestroy bool
	// List of exposed ports
	ExposedPorts []Port
	// Container environment variables
	EnvironmentVariables map[string]string
	// Command to run when starting the container
	Cmd []string
	// List of volume bindings for this container
	Binds []string
	// Follow container log output
	FollowLogs bool
	// Callback invoked after start container command was invoked.
	AfterStart Callback
}

// Callback provides a way for the callee to invoke the code inside the caller
type Callback interface {
	// Callback method invoked with the current component name and value resolver
	Call(componentName string, resolver ValueResolver) error
}

// ValueResolver allows resolution of container parameters
type ValueResolver interface {
	// Resolve applies a parsed template to the docker environment context
	Resolve(template string) (string, error)
	// Hosts provides external IP of the container
	Host() string
	// Port provides a host port for a given component and named port
	Port(componentName string, portName string) (int, error)
}

// Port holds definition of a port mapping
type Port struct {
	// Optional port name. If not specified, the lower-cased component name is used.
	// Each named port in a component must have a unique name.
	Name string
	// Container port mapped to the host port
	ContainerPort int
	// Number of port to expose on the host. If not specified, ephemeral port is used.
	HostPort int
}
