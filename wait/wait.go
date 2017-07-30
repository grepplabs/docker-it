package wait

import (
	"fmt"
	"log"
	"os"
	"time"
)

const (
	defaultAtMost       = 60 * time.Second
	defaultPollInterval = time.Second
)

// Options defines wait parameters.
type Options struct {
	// Maximal wait duration
	AtMost time.Duration
	// Poll interval used to delay next invocation of readinessProbe function
	PollInterval time.Duration
	// Logger used in the wait
	Logger *log.Logger
}

// Wait holds wait parameters and provides defaults when Options parameters were not defined.
type Wait struct {
	atMost       time.Duration
	pollInterval time.Duration
	logger       *log.Logger
}

// NewWait creates a new Wait
func NewWait(options Options) Wait {
	atMost := options.AtMost
	if options.AtMost == 0 {
		atMost = defaultAtMost
	}
	delay := options.PollInterval
	if options.PollInterval == 0 {
		delay = defaultPollInterval
	}
	return Wait{
		atMost:       atMost,
		pollInterval: delay,
		logger:       options.Logger,
	}
}

// GetLogger provides wait logger
func (r *Wait) GetLogger(componentName string) *log.Logger {
	if r.logger != nil {
		return r.logger
	}
	return log.New(os.Stdout, fmt.Sprintf("WAIT FOR %s: ", componentName), log.Ldate|log.Ltime)
}

// GetAtMost provides maximal wait duration
func (r *Wait) GetAtMost() time.Duration {
	return r.atMost
}

// GetPollInterval provides poll interval between readinessProbe invocation
func (r *Wait) GetPollInterval() time.Duration {
	return r.pollInterval
}

// Poll invokes readinessProbe until it provides no error or timeout is reached
func (r *Wait) Poll(componentName string, readinessProbe func() error) error {
	timeout := r.GetAtMost()
	pollInterval := r.GetPollInterval()

	var err error
	var start time.Time
	for start = time.Now(); time.Since(start) < timeout; {
		err = readinessProbe()
		if err == nil {
			r.GetLogger(componentName).Println("Component is up after", time.Since(start))
			return nil
		}
		time.Sleep(pollInterval)
	}
	if err != nil {
		return fmt.Errorf("Readiness probe of '%s' failed after %s with error '%v'", componentName, time.Since(start), err)
	}
	return nil
}
