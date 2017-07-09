package dockerit

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"
)

const (
	qualifierHost          = "Host"          // host
	qualifierContainerPort = "ContainerPort" // exposed port within container
	qualifierTargetPort    = "TargetPort"    // exposed port within container
	qualifierHostPort      = "HostPort"      // mapped port on host
	qualifierPort          = "Port"          // mapped port on host
)

type DockerEnvironmentValueResolver struct {
	ip      string
	context *DockerEnvironmentContext
}

func NewDockerComponentValueResolver(ip string, context *DockerEnvironmentContext) *DockerEnvironmentValueResolver {
	return &DockerEnvironmentValueResolver{
		ip:      ip,
		context: context,
	}
}

func (r *DockerEnvironmentValueResolver) value(m map[string]interface{}, key string) (interface{}, error) {
	if val, ok := m[key]; !ok {
		return nil, fmt.Errorf("Unknown key '%s'", key)
	} else {
		return val, nil
	}
}

func (r *DockerEnvironmentValueResolver) configureContainersEnv() error {

	contextVariables, err := r.getEnvironmentContextVariables()
	if err != nil {
		return err
	}
	for containerName, container := range r.context.containers {
		if container.EnvironmentVariables == nil {
			continue
		}
		env := make(map[string]string)
		for k, v := range container.EnvironmentVariables {
			value, err := r.resolveValue(fmt.Sprintf("DockerComponent %s Env %s", containerName, k), v, contextVariables)
			if err != nil {
				return err
			}
			env[k] = value
		}
		// assign env to container
		container.env = env
	}
	return nil
}

func (r *DockerEnvironmentValueResolver) resolve(templateText string) (string, error) {

	contextVariables := r.getSystemContextVariables()

	envContextVariables, err := r.getEnvironmentContextVariables()
	if err != nil {
		return "", err
	} else {
		// overwrites the same key
		for k, v := range envContextVariables {
			contextVariables[k] = v
		}
	}
	return r.resolveValue("resolve", templateText, contextVariables)
}

func (r *DockerEnvironmentValueResolver) resolveValue(templateName string, templateText string, contextVariables map[string]interface{}) (string, error) {

	var funcMap = template.FuncMap{
		"value": r.value,
	}

	t := template.New(templateName).Funcs(funcMap).Option("missingkey=error")
	t, err := t.Parse(templateText)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	err = t.Execute(&b, &contextVariables)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func (r *DockerEnvironmentValueResolver) getEnvironmentContextVariables() (map[string]interface{}, error) {

	result := make(map[string]interface{})

	for containerName, container := range r.context.containers {
		if container.portBindings == nil {
			return nil, fmt.Errorf("portBindings for '%s' is not defined", containerName)
		}

		result[fmt.Sprintf("%s.%s", containerName, qualifierHost)] = r.ip

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
				result[fmt.Sprintf("%s.%s.%s", container.DockerComponent.Name, exposedPorts.Name, qualifierPort)] = result[fmt.Sprintf("%s.%s.%s", containerName, toPortName(exposedPorts.Name), qualifierPort)]
				result[fmt.Sprintf("%s.%s.%s", container.DockerComponent.Name, exposedPorts.Name, qualifierHostPort)] = result[fmt.Sprintf("%s.%s.%s", containerName, toPortName(exposedPorts.Name), qualifierHostPort)]
				result[fmt.Sprintf("%s.%s.%s", container.DockerComponent.Name, exposedPorts.Name, qualifierContainerPort)] = result[fmt.Sprintf("%s.%s.%s", containerName, toPortName(exposedPorts.Name), qualifierContainerPort)]
				result[fmt.Sprintf("%s.%s.%s", container.DockerComponent.Name, exposedPorts.Name, qualifierTargetPort)] = result[fmt.Sprintf("%s.%s.%s", containerName, toPortName(exposedPorts.Name), qualifierTargetPort)]
			}
		}

	}
	return result, nil
}

func (r *DockerEnvironmentValueResolver) getSystemContextVariables() map[string]interface{} {
	result := make(map[string]interface{})
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		result[pair[0]] = pair[1]
	}
	return result
}
