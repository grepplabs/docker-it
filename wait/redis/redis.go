package redis

import (
	"fmt"
	dit "github.com/cloud-42/docker-it"
	"github.com/cloud-42/docker-it/wait"
	"github.com/garyburd/redigo/redis"
	"time"
)

type Options struct {
	WaitOptions wait.Options
	PortName    string
}

type redisWait struct {
	wait.Wait
	portName string
}

func NewRedisWait(options Options) *redisWait {
	return &redisWait{
		Wait:     wait.NewWait(options.WaitOptions),
		portName: options.PortName,
	}
}

// implements dockerit.Callback
func (r *redisWait) Call(componentName string, resolver dit.ValueResolver) error {
	if port, err := resolver.Port(componentName, r.portName); err != nil {
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

func (r *redisWait) pollRedis(componentName string, host string, port int) error {

	logger := r.GetLogger(componentName)
	logger.Println("Waiting for redis", fmt.Sprintf("%s:%d", host, port))

	f := func() error {
		return r.ping(host, port)
	}
	return r.Poll(componentName, f)
}

func (r *redisWait) ping(host string, port int) error {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", host, port),
		redis.DialConnectTimeout(time.Second), redis.DialReadTimeout(time.Second), redis.DialWriteTimeout(time.Second))
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Do("PING")
	return err
}
