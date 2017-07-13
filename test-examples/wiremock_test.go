package test_examples

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestWiremockCall(t *testing.T) {
	a := assert.New(t)

	host := dockerEnvironment.Host()
	port, err := dockerEnvironment.Port("it-wiremock", "")
	a.Nil(err)

	url := fmt.Sprintf("http://%s:%s/__admin/requests", host, port)
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	req, err := http.NewRequest("GET", url, nil)
	a.Nil(err)
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	a.EqualValues(resp.StatusCode, http.StatusOK)
	_, err = ioutil.ReadAll(resp.Body)
	a.Nil(err)
}
