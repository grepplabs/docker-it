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

type Options struct {
	AtMost time.Duration
	Delay  time.Duration
	Logger *log.Logger
}

type Wait struct {
	atMost time.Duration
	delay  time.Duration
	logger *log.Logger
}

func NewWait(options Options) Wait {
	atMost := options.AtMost
	if options.AtMost == 0 {
		atMost = DefaultAtMost
	}
	delay := options.Delay
	if options.Delay == 0 {
		delay = DefaultDelay
	}
	return Wait{
		atMost: atMost,
		delay:  delay,
		logger: options.Logger,
	}
}

func (r *Wait) GetLogger(componentName string) *log.Logger {
	if r.logger != nil {
		return r.logger
	} else {
		return log.New(os.Stdout, fmt.Sprintf("WAIT FOR %s: ", componentName), log.Ldate|log.Ltime)
	}
}

func (r *Wait) GetAtMost() time.Duration {
	return r.atMost
}

func (r *Wait) GetDelay() time.Duration {
	return r.delay
}

func (r *Wait) Poll(componentName string, readinessProbe func() error) error {
	timeout := r.GetAtMost()
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
