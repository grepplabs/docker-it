package testexamples

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestHttpCall(t *testing.T) {
	a := assert.New(t)

	host := dockerEnvironment.Host()
	port, err := dockerEnvironment.Port("it-http", "")
	a.Nil(err)

	url := fmt.Sprintf("http://%s:%d/__admin/requests", host, port)
	fmt.Println(url)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, err := http.NewRequest("GET", url, nil)
	a.Nil(err)

	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	a.Nil(err)
	defer resp.Body.Close()

	a.EqualValues(resp.StatusCode, http.StatusOK)
	_, err = ioutil.ReadAll(resp.Body)
	a.Nil(err)
}
