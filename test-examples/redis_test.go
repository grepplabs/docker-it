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

	address := fmt.Sprintf("%s:%d", host, port)
	fmt.Println(address)

	conn, err := redis.Dial("tcp", address)
	a.Nil(err)
	defer conn.Close()

	_, err = conn.Do("SET", "test-key", "test-value")
	a.Nil(err)

	value, err := redis.String(conn.Do("GET", "test-key"))
	a.Nil(err)
	a.EqualValues("test-value", value)
}
