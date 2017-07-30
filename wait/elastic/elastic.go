package elastic

import (
	"fmt"
	dit "github.com/cloud-42/docker-it"
	"github.com/cloud-42/docker-it/wait"
	"github.com/pkg/errors"
	v5 "gopkg.in/olivere/elastic.v5"
)

// Options defines Elasticsearch wait parameters.
type Options struct {
	WaitOptions        wait.Options
	Username, Password string
}

type elasticWait struct {
	wait.Wait
	urlTemplate        string
	username, password string
}

// NewElasticWait creates a new Elasticsearch wait
func NewElasticWait(urlTemplate string, options Options) *elasticWait {
	if urlTemplate == "" {
		panic(errors.New("elastic wait: UrlTemplate must not be empty"))
	}
	return &elasticWait{
		Wait:        wait.NewWait(options.WaitOptions),
		urlTemplate: urlTemplate,
		username:    options.Username,
		password:    options.Password,
	}
}

// implements dockerit.Callback
func (r *elasticWait) Call(componentName string, resolver dit.ValueResolver) error {
	url, err := resolver.Resolve(r.urlTemplate)
	if err != nil {
		return err
	}
	err = r.pollElastic(componentName, url)
	if err != nil {
		return fmt.Errorf("elastic wait: failed to connect to %s %v ", url, err)
	}
	return nil
}

func (r *elasticWait) pollElastic(componentName string, url string) error {

	logger := r.GetLogger(componentName)
	logger.Println("Waiting for elastic", url)

	f := func() error {
		return r.waitForGreenStatus(url)
	}
	return r.Poll(componentName, f)
}

func (r *elasticWait) waitForGreenStatus(url string) error {
	clientOptions := make([]v5.ClientOptionFunc, 0)
	clientOptions = append(clientOptions, v5.SetURL(url))
	if r.password != "" && r.username != "" {
		clientOptions = append(clientOptions, v5.SetBasicAuth(r.username, r.password))
	}
	client, err := v5.NewClient(clientOptions...)
	if err != nil {
		return err
	}
	client.Stop()
	return client.WaitForGreenStatus("1s")
}
