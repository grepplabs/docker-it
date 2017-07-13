package redis

import (
	"fmt"
	dit "github.com/cloud-42/docker-it"
	"github.com/cloud-42/docker-it/wait"
	"github.com/garyburd/redigo/redis"
	"time"
)

type RedisWait struct {
	wait.Wait
	PortName string
}

// implements dockerit.Callback
func (r *RedisWait) Call(componentName string, resolver dit.ValueResolver) error {
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

func (r *RedisWait) pollRedis(componentName string, host string, port string) error {

	logger := r.GetLogger(componentName)
	logger.Println("Waiting for redis", fmt.Sprintf("%s:%s", host, port))

	f := func() error {
		return r.ping(host, port)
	}
	return r.Poll(componentName, f)
}

func (r *RedisWait) ping(host string, port string) error {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", host, port),
		redis.DialConnectTimeout(time.Second), redis.DialReadTimeout(time.Second), redis.DialWriteTimeout(time.Second))
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Do("PING")
	return err
}
