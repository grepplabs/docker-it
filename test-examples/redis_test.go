package test_examples

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRedisCall(t *testing.T) {
	a := assert.New(t)

	host := dockerEnvironment.Host()
	port, err := dockerEnvironment.Port("it-redis", "")
	a.Nil(err)

	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	a.Nil(err)
	defer conn.Close()

	_, err = conn.Do("SET", "test-key", "test-value")
	a.Nil(err)

	value, err := redis.String(conn.Do("GET", "test-key"))
	a.Nil(err)
	a.EqualValues("test-value", value)
}
