package test_examples

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRedisCall(t *testing.T) {
	a := assert.New(t)

	host := dockerEnvironment.Host()
	port, err := dockerEnvironment.Port("it-redis", "")
	a.Nil(err)

	fmt.Println("redis host", host)
	fmt.Println("redis port", port)
}
