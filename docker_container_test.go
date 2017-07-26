package dockerit

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestStopFollowLogsIsNonBlocking(t *testing.T) {
	a := assert.New(t)

	container := newDockerContainer(DockerComponent{
		Name:       "it-redis",
		Image:      "redis",
		FollowLogs: true,
	})
	a.NotNil(container)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		for {
			select {
			case <-container.stopFollowLogsChannel:
				wg.Done()
				return
			}
		}
	}()

	container.StopFollowLogs()
	wg.Wait()

	for i := 0; i < 3; i++ {
		container.StopFollowLogs()
	}
}
