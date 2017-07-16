package elastic

import (
	"fmt"
	dit "github.com/cloud-42/docker-it"
	"github.com/cloud-42/docker-it/wait"
	"github.com/pkg/errors"
	v5 "gopkg.in/olivere/elastic.v5"
)

type Options struct {
	wait.Wait
}

type elasticWait struct {
	wait.Wait
	urlTemplate string
}

func NewElasticWait(urlTemplate string, options Options) *elasticWait {
	if urlTemplate == "" {
		panic(errors.New("elastic wait: UrlTemplate must not be empty"))
	}
	return &elasticWait{
		Wait:        options.Wait,
		urlTemplate: urlTemplate,
	}
}

// implements dockerit.Callback
func (r *elasticWait) Call(componentName string, resolver dit.ValueResolver) error {
	if url, err := resolver.Resolve(r.urlTemplate); err != nil {
		return err
	} else {
		err := r.pollElastic(componentName, url)
		if err != nil {
			return fmt.Errorf("elastic wait: failed to connect to %s %v ", url, err)
		}
		return nil
	}
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
	client, err := v5.NewClient(v5.SetURL(url))
	if err != nil {
		return err
	}
	client.Stop()
	return client.WaitForGreenStatus("1s")
}
