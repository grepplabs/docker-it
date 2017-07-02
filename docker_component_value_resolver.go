package dockerit

import (
	"bytes"
	"fmt"
	"strconv"
	"text/template"
)

const (
	qualifierHost          = "Host"          // host
	qualifierContainerPort = "ContainerPort" // exposed port within container
	qualifierTargetPort    = "TargetPort"    // exposed port within container
	qualifierHostPort      = "HostPort"      // mapped port on host
	qualifierPort          = "Port"          // mapped port on host
)

type DockerComponentValueResolver struct {
	containers map[string]*DockerContainer
}

func NewDockerComponentValueResolver(containers map[string]*DockerContainer) *DockerComponentValueResolver {
	return &DockerComponentValueResolver{
		containers: containers,
	}
}

func (r *DockerComponentValueResolver) value(m map[string]interface{}, key string) (interface{}, error) {
	if val, ok := m[key]; !ok {
		return nil, fmt.Errorf("Unknown key '%s'", key)
	} else {
		return val, nil
	}
}

func (r *DockerComponentValueResolver) configureContainersEnv(host string) error {

	contextVariables, err := r.getEnvironmentContextVariables(host)
	if err != nil {
		return err
	}
	var funcMap = template.FuncMap{
		"value": r.value,
	}
	for containerName, container := range r.containers {
		if container.EnvironmentVariables == nil {
			continue
		}
		env := make(map[string]string)
		for k, v := range container.EnvironmentVariables {
			t := template.New(fmt.Sprintf("DockerComponent %s Env %s", containerName, k)).Funcs(funcMap).Option("missingkey=error")

			t, err := t.Parse(v)
			if err != nil {
				return err
			}
			var b bytes.Buffer
			err = t.Execute(&b, &contextVariables)
			if err != nil {
				return err
			}
			env[k] = b.String()
		}
		// assign env to container
		container.env = env
	}
	return nil
}

func (r *DockerComponentValueResolver) getEnvironmentContextVariables(host string) (map[string]interface{}, error) {

	result := make(map[string]interface{})

	for containerName, container := range r.containers {
		if container.portBindings == nil {
			return nil, fmt.Errorf("portBindings for '%s' is not defined", containerName)
		}

		result[fmt.Sprintf("%s.%s", containerName, qualifierHost)] = host

		for _, port := range container.portBindings {
			if port.Name == "" || port.Name == containerName {
				result[fmt.Sprintf("%s.%s", containerName, qualifierPort)] = strconv.Itoa(port.HostPort)
				result[fmt.Sprintf("%s.%s", containerName, qualifierHostPort)] = strconv.Itoa(port.HostPort)
				result[fmt.Sprintf("%s.%s", containerName, qualifierContainerPort)] = strconv.Itoa(port.ContainerPort)
				result[fmt.Sprintf("%s.%s", containerName, qualifierTargetPort)] = strconv.Itoa(port.ContainerPort)
			}
			if port.Name != "" {
				result[fmt.Sprintf("%s.%s.%s", containerName, port.Name, qualifierPort)] = strconv.Itoa(port.HostPort)
				result[fmt.Sprintf("%s.%s.%s", containerName, port.Name, qualifierHostPort)] = strconv.Itoa(port.HostPort)
				result[fmt.Sprintf("%s.%s.%s", containerName, port.Name, qualifierContainerPort)] = strconv.Itoa(port.ContainerPort)
				result[fmt.Sprintf("%s.%s.%s", containerName, port.Name, qualifierTargetPort)] = strconv.Itoa(port.ContainerPort)
			}
		}

		// original names (no lowercase)
		for _, exposedPorts := range container.DockerComponent.ExposedPorts {
			if exposedPorts.Name == "" || toPortName(exposedPorts.Name) == toContainerName(containerName) {
				result[fmt.Sprintf("%s.%s", container.DockerComponent.Name, qualifierPort)] = result[fmt.Sprintf("%s.%s", containerName, qualifierPort)]
				result[fmt.Sprintf("%s.%s", container.DockerComponent.Name, qualifierHostPort)] = result[fmt.Sprintf("%s.%s", containerName, qualifierHostPort)]
				result[fmt.Sprintf("%s.%s", container.DockerComponent.Name, qualifierContainerPort)] = result[fmt.Sprintf("%s.%s", containerName, qualifierContainerPort)]
				result[fmt.Sprintf("%s.%s", container.DockerComponent.Name, qualifierTargetPort)] = result[fmt.Sprintf("%s.%s", containerName, qualifierTargetPort)]
			}
			if exposedPorts.Name != "" {
				result[fmt.Sprintf("%s.%s.%s", container.DockerComponent.Name, exposedPorts.Name, qualifierPort)] = fmt.Sprintf("%s.%s.%s", containerName, toPortName(exposedPorts.Name), qualifierPort)
				result[fmt.Sprintf("%s.%s.%s", container.DockerComponent.Name, exposedPorts.Name, qualifierHostPort)] = fmt.Sprintf("%s.%s.%s", containerName, toPortName(exposedPorts.Name), qualifierHostPort)
				result[fmt.Sprintf("%s.%s.%s", container.DockerComponent.Name, exposedPorts.Name, qualifierContainerPort)] = fmt.Sprintf("%s.%s.%s", containerName, toPortName(exposedPorts.Name), qualifierContainerPort)
				result[fmt.Sprintf("%s.%s.%s", container.DockerComponent.Name, exposedPorts.Name, qualifierTargetPort)] = fmt.Sprintf("%s.%s.%s", containerName, toPortName(exposedPorts.Name), qualifierTargetPort)
			}
		}

	}
	return result, nil
}
