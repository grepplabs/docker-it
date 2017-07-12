package redis

import (
	"fmt"
	dit "github.com/cloud-42/docker-it"
	"github.com/garyburd/redigo/redis"
	"log"
	"os"
	"time"
)

const (
	DefaultAtMost = 15 * time.Second
	DefaultDelay  = time.Second
)

type Wait struct {
	PortName string
	AtMost   time.Duration
	Delay    time.Duration
	Logger   *log.Logger
}

// implements dockerit.Callback
func (r *Wait) Call(componentName string, resolver dit.ValueResolver) error {
	if port, err := resolver.Port(componentName, r.PortName); err != nil {
		return err
	} else {
		host := resolver.Host()
		err := r.pollRedis(componentName, host, port)
		if err != nil {
			return fmt.Errorf("redis wait: failed to connect to %s:%s: %v ", host, port, err)
		}
		return nil
	}
}
func (r *Wait) getLogger() *log.Logger {
	if r.Logger != nil {
		return r.Logger
	} else {
		return log.New(os.Stdout, "REDIS WAIT: ", log.Ldate|log.Ltime)
	}
}
func (r *Wait) getTimeout() time.Duration {
	if r.AtMost == 0 {
		return DefaultAtMost
	} else {
		return r.AtMost
	}
}

func (r *Wait) getDelay() time.Duration {
	if r.Delay == 0 {
		return DefaultDelay
	} else {
		return r.Delay
	}
}

func (r *Wait) pollRedis(componentName string, host string, port string) error {

	logger := r.getLogger()
	logger.Println("Waiting for", componentName, fmt.Sprintf("%s:%s", host, port))

	timeout := r.getTimeout()
	delay := r.getDelay()

	var err error
	var start time.Time
	for start = time.Now(); time.Since(start) < timeout; {
		err = r.ping(host, port)
		if err == nil {
			logger.Println(componentName, "is up after", time.Since(start))
			return nil
		}
		time.Sleep(delay)
	}
	if err != nil {
		return fmt.Errorf("ping failed after %s with error '%v'", time.Since(start), err)
	}
	return nil
}

func (r *Wait) ping(host string, port string) error {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", host, port),
		redis.DialConnectTimeout(time.Second), redis.DialReadTimeout(time.Second), redis.DialWriteTimeout(time.Second))
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Do("PING")
	return err
}
