package dockerit

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDockerLifecycleHandler(t *testing.T) {
	a := assert.New(t)

	context, err := newDockerEnvironmentContext()
	a.Nil(err)

	container, err := context.addContainer(DockerComponent{
		Name:                    testImage,
		Image:                   testImage,
		ForcePull:               true,
		RemoveImageAfterDestroy: true,
		ExposedPorts: []Port{
			{
				ContainerPort: 4711,
			},
		},
	})
	a.Nil(err)
	a.Empty(container.containerID)

	handler, err := newDockerLifecycleHandler(context)
	a.Nil(err)

	err = handler.Create(container)
	containerId1 := container.containerID
	a.Nil(err)
	a.NotEmpty(container.containerID)

	// next create has no effect
	err = handler.Create(container)
	a.Nil(err)
	a.Equal(containerId1, container.containerID)

	err = handler.checkOrPullDockerImage(testImage, false)
	a.Nil(err)

	running, err := handler.isContainerRunning(container.containerID)
	a.Nil(err)
	a.False(running)

	exists, err := handler.containerExists(container.containerID)
	a.Nil(err)
	a.True(exists)

	err = handler.Start(container)
	a.Nil(err)

	running, err = handler.isContainerRunning(container.containerID)
	a.Nil(err)
	a.True(running)

	// next start has no effect
	err = handler.Start(container)
	a.Nil(err)

	err = handler.fetchLogs(container.containerID, stdoutWriter(""), stdoutWriter(""))
	a.Nil(err)

	err = handler.Stop(container)
	a.Nil(err)

	exists, err = handler.containerExists(container.containerID)
	a.Nil(err)
	a.True(exists)

	running, err = handler.isContainerRunning(container.containerID)
	a.Nil(err)
	a.False(running)

	// next stop has no effect
	err = handler.Stop(container)
	a.Nil(err)

	err = handler.Destroy(container)
	a.Nil(err)
	a.Empty(container.containerID)

	exists, err = handler.containerExists(container.containerID)
	a.Nil(err)
	a.False(exists)

	// images should be deleted as RemoveImageAfterDestroy is set to true
	err = handler.checkOrPullDockerImage(testImage, false)
	a.EqualError(err, "Local images "+testImage+" does not exist")

	err = handler.checkOrPullDockerImage(testImage, true)
	a.Nil(err)

	err = handler.Destroy(container)
	a.Nil(err)

	// next destroy has no effect
	err = handler.Destroy(container)
	a.Nil(err)

	err = handler.Start(container)
	a.Nil(err)
	a.NotEqual(containerId1, container.containerID)

	err = handler.Destroy(container)
	a.Nil(err)

	handler.Close()

}
