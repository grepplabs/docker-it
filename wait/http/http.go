package http

import (
	"context"
	"fmt"
	dit "github.com/cloud-42/docker-it"
	"github.com/cloud-42/docker-it/wait"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

const (
	DefaultMethod = "GET"
)

type Options struct {
	wait.Wait
	UrlTemplate string
	Method      string
}

type httpWait struct {
	Options
}

func NewHttpWait(options Options) *httpWait {
	method := options.Method
	if method == "" {
		method = DefaultMethod
	}
	return &httpWait{
		Options{
			Wait:        options.Wait,
			UrlTemplate: options.UrlTemplate,
			Method:      method,
		},
	}
}

// implements dockerit.Callback
func (r *httpWait) Call(componentName string, resolver dit.ValueResolver) error {
	if r.UrlTemplate == "" {
		return errors.New("http wait: UrlTemplate must not be empty")
	}
	if url, err := resolver.Resolve(r.UrlTemplate); err != nil {
		return err
	} else {
		err := r.pollHttp(componentName, url)
		if err != nil {
			return fmt.Errorf("http wait: failed to connect to %s %v ", url, err)
		}
		return nil
	}
}

func (r *httpWait) pollHttp(componentName string, url string) error {

	logger := r.GetLogger(componentName)
	logger.Println("Waiting for http", url)

	f := func() error {
		return r.getRequest(url)
	}
	return r.Poll(componentName, f)
}

func (r *httpWait) getRequest(url string) error {
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	req, err := http.NewRequest(r.Method, url, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	} else {
		return fmt.Errorf("server %s returned status: %v", url, resp.Status)
	}
}
