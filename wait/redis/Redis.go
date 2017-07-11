package redis

import (
	dit "github.com/cloud-42/docker-it"
	"time"
)

type Wait struct {
	PortName      string
	AtMost	      time.Duration
}
// implements dockerit.Callback
func (r *Wait) Call(componentName string, resolver dit.ValueResolver) error {
	if _, err := resolver.Port(componentName, r.PortName); err != nil {
		return err
	} else {
		// TODO:
		return nil
	}
}