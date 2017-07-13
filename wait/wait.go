package wait

import (
	"fmt"
	"log"
	"os"
	"time"
)

const (
	DefaultAtMost = 15 * time.Second
	DefaultDelay  = time.Second
)

type Wait struct {
	AtMost time.Duration
	Delay  time.Duration
	Logger *log.Logger
}

func (r *Wait) GetLogger(componentName string) *log.Logger {
	if r.Logger != nil {
		return r.Logger
	} else {
		return log.New(os.Stdout, fmt.Sprintf("WAIT FOR %s: ", componentName), log.Ldate|log.Ltime)
	}
}

func (r *Wait) GetTimeout() time.Duration {
	if r.AtMost == 0 {
		return DefaultAtMost
	} else {
		return r.AtMost
	}
}

func (r *Wait) GetDelay() time.Duration {
	if r.Delay == 0 {
		return DefaultDelay
	} else {
		return r.Delay
	}
}
func (r *Wait) Poll(componentName string, readinessProbe func() error) error {
	timeout := r.GetTimeout()
	delay := r.GetDelay()

	var err error
	var start time.Time
	for start = time.Now(); time.Since(start) < timeout; {
		err = readinessProbe()
		if err == nil {
			r.GetLogger(componentName).Println("Component is up after", time.Since(start))
			return nil
		}
		time.Sleep(delay)
	}
	if err != nil {
		return fmt.Errorf("Readiness probe of '%s' failed after %s with error '%v'", componentName, time.Since(start), err)
	}
	return nil
}
