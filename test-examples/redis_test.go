package test_examples

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRedisCall(t *testing.T) {
	a := assert.New(t)

	dockerEnvironment.Start("it-redis")

	host, err := dockerEnvironment.Resolve(`{{ value . "it-redis.Host"}}`)
	a.Nil(err)
	port, err := dockerEnvironment.Resolve(`{{ value . "it-redis.Port"}}`)
	a.Nil(err)

	fmt.Println("redis host", host)
	fmt.Println("redis port", port)
}
