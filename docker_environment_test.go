package dockerit

import (
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestNewDockerEnvironmentFailsWhenComponentListIsEmpty(t *testing.T) {
	a := assert.New(t)
	_, err := NewDockerEnvironment()
	a.NotNil(err)
	a.EqualError(err, "Component list is empty")
}

func TestNewDockerEnvironmentStartFails(t *testing.T) {
	a := assert.New(t)

	env, err := NewDockerEnvironment(
		DockerComponent{
			Name:       "it-busybox",
			Image:      "busybox",
			ForcePull:  true,
			FollowLogs: false,
		},
	)
	a.Nil(err)

	a.EqualError(env.Start(), "No component was provided to start")
	a.EqualError(env.StartParallel(), "No component was provided to start in parallel")
	a.EqualError(env.Start("it-unknown"), "DockerComponent [it-unknown] is not configured")
	a.EqualError(env.StartParallel("it-unknown"), "DockerComponent [it-unknown] is not configured")
}

func TestNewDockerEnvironmentLifeCycle(t *testing.T) {
	a := assert.New(t)

	env, err := NewDockerEnvironment(
		DockerComponent{
			Name:       "it-busybox",
			Image:      "busybox",
			ForcePull:  true,
			FollowLogs: true,
		},
	)
	a.Nil(err)

	err = env.Start("it-busybox")
	a.Nil(err)

	err = env.Stop("it-busybox")
	a.Nil(err)

	err = env.Start("it-busybox")
	a.Nil(err)

	err = env.Stop("it-busybox")
	a.Nil(err)

	err = env.Destroy("it-busybox")
	a.Nil(err)

	var counter uint32

	beforeShutdown := func() {
		atomic.AddUint32(&counter, 1)
	}
	env.Shutdown(beforeShutdown)
	env.Shutdown(beforeShutdown)

	// shutdown is invoked only once
	a.Equal(uint32(1), atomic.LoadUint32(&counter))

}

func TestNewDockerEnvironmentWithShutdown(t *testing.T) {
	a := assert.New(t)

	env, err := NewDockerEnvironment(
		DockerComponent{
			Name:       "it-busybox",
			Image:      "busybox",
			ForcePull:  true,
			FollowLogs: true,
		},
	)
	var counter uint32
	doneChannel := env.WithShutdown(func() {
		atomic.AddUint32(&counter, 1)
	})

	err = env.Start("it-busybox")
	a.Nil(err)

	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)

	select {
	case <-doneChannel:
	case <-time.After(time.Second * 3):
	}
	a.Equal(uint32(1), atomic.LoadUint32(&counter))
}
