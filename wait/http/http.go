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
	defaultMethod = "GET"
)

// Options defines Http wait parameters.
type Options struct {
	WaitOptions wait.Options
	Method      string
}

type httpWait struct {
	wait.Wait
	method      string
	urlTemplate string
}

// NewHttpWait creates a new Http wait
func NewHttpWait(urlTemplate string, options Options) *httpWait {
	if urlTemplate == "" {
		panic(errors.New("http wait: UrlTemplate must not be empty"))
	}

	method := options.Method
	if method == "" {
		method = defaultMethod
	}
	return &httpWait{
		Wait:        wait.NewWait(options.WaitOptions),
		urlTemplate: urlTemplate,
		method:      method,
	}
}

// implements dockerit.Callback
func (r *httpWait) Call(componentName string, resolver dit.ValueResolver) error {
	url, err := resolver.Resolve(r.urlTemplate)
	if err != nil {
		return err
	}
	err = r.pollHttp(componentName, url)
	if err != nil {
		return fmt.Errorf("http wait: failed to connect to %s %v ", url, err)
	}
	return nil
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
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	req, err := http.NewRequest(r.method, url, nil)
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
	}
	return fmt.Errorf("server %s returned status: %v", url, resp.Status)
}
