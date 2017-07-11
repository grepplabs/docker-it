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

type dockerEnvironmentValueResolver struct {
	ip      string
	context *dockerEnvironmentContext
}

func newDockerComponentValueResolver(ip string, context *dockerEnvironmentContext) *dockerEnvironmentValueResolver {
	return &dockerEnvironmentValueResolver{
		ip:      ip,
		context: context,
	}
}

func (r *dockerEnvironmentValueResolver) value(m map[string]interface{}, key string) (interface{}, error) {
	if val, ok := m[key]; !ok {
		return nil, fmt.Errorf("Unknown key '%s'", key)
	} else {
		return val, nil
	}
}

func (r *dockerEnvironmentValueResolver) configureContainersEnv() error {

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

// implements ValueResolver
func (r *dockerEnvironmentValueResolver) Resolve(templateText string) (string, error) {

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

func (r *dockerEnvironmentValueResolver) resolveValue(templateName string, templateText string, contextVariables map[string]interface{}) (string, error) {

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

func (r *dockerEnvironmentValueResolver) getEnvironmentContextVariables() (map[string]interface{}, error) {

	result := make(map[string]interface{})
	for containerName, container := range r.context.containers {
		if container.portBindings == nil {
			return nil, fmt.Errorf("portBindings for '%s' is not defined", containerName)
		}
		r.appendContainerContextVariables(container.DockerComponent.Name, r.ip, result, container)
		r.appendContainerContextVariables(containerName, r.ip, result, container)
	}
	return result, nil
}

func (r *dockerEnvironmentValueResolver) appendContainerContextVariables(name string, ip string, result map[string]interface{}, container *dockerContainer) {
	result[fmt.Sprintf("%s.%s", name, qualifierHost)] = ip

	for _, port := range container.portBindings {
		if port.Name == "" || normalizeName(port.Name) == normalizeName(name) {
			result[fmt.Sprintf("%s.%s", name, qualifierPort)] = strconv.Itoa(port.HostPort)
			result[fmt.Sprintf("%s.%s", name, qualifierHostPort)] = strconv.Itoa(port.HostPort)
			result[fmt.Sprintf("%s.%s", name, qualifierContainerPort)] = strconv.Itoa(port.ContainerPort)
			result[fmt.Sprintf("%s.%s", name, qualifierTargetPort)] = strconv.Itoa(port.ContainerPort)
		}
		if port.Name != "" {
			result[fmt.Sprintf("%s.%s.%s", name, port.Name, qualifierPort)] = strconv.Itoa(port.HostPort)
			result[fmt.Sprintf("%s.%s.%s", name, port.Name, qualifierHostPort)] = strconv.Itoa(port.HostPort)
			result[fmt.Sprintf("%s.%s.%s", name, port.Name, qualifierContainerPort)] = strconv.Itoa(port.ContainerPort)
			result[fmt.Sprintf("%s.%s.%s", name, port.Name, qualifierTargetPort)] = strconv.Itoa(port.ContainerPort)
		}
	}

	// original names (no lowercase)
	for _, exposedPorts := range container.DockerComponent.ExposedPorts {
		if exposedPorts.Name == "" || normalizeName(exposedPorts.Name) == normalizeName(name) {
			result[fmt.Sprintf("%s.%s", name, qualifierPort)] = result[fmt.Sprintf("%s.%s", name, qualifierPort)]
			result[fmt.Sprintf("%s.%s", name, qualifierHostPort)] = result[fmt.Sprintf("%s.%s", name, qualifierHostPort)]
			result[fmt.Sprintf("%s.%s", name, qualifierContainerPort)] = result[fmt.Sprintf("%s.%s", name, qualifierContainerPort)]
			result[fmt.Sprintf("%s.%s", name, qualifierTargetPort)] = result[fmt.Sprintf("%s.%s", name, qualifierTargetPort)]
		}
		if exposedPorts.Name != "" {
			result[fmt.Sprintf("%s.%s.%s", name, exposedPorts.Name, qualifierPort)] = result[fmt.Sprintf("%s.%s.%s", name, normalizeName(exposedPorts.Name), qualifierPort)]
			result[fmt.Sprintf("%s.%s.%s", name, exposedPorts.Name, qualifierHostPort)] = result[fmt.Sprintf("%s.%s.%s", name, normalizeName(exposedPorts.Name), qualifierHostPort)]
			result[fmt.Sprintf("%s.%s.%s", name, exposedPorts.Name, qualifierContainerPort)] = result[fmt.Sprintf("%s.%s.%s", name, normalizeName(exposedPorts.Name), qualifierContainerPort)]
			result[fmt.Sprintf("%s.%s.%s", name, exposedPorts.Name, qualifierTargetPort)] = result[fmt.Sprintf("%s.%s.%s", name, normalizeName(exposedPorts.Name), qualifierTargetPort)]
		}
	}
}

func (r *dockerEnvironmentValueResolver) getSystemContextVariables() map[string]interface{} {
	result := make(map[string]interface{})
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		result[pair[0]] = pair[1]
	}
	return result
}
