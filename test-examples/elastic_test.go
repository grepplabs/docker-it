package test_examples

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	elastic "gopkg.in/olivere/elastic.v5"
	"testing"
)

func TestElasticCall(t *testing.T) {
	a := assert.New(t)

	host := dockerEnvironment.Host()
	port, err := dockerEnvironment.Port("it-es", "")
	a.Nil(err)

	url := fmt.Sprintf("http://%s:%d", host, port)
	// https://www.elastic.co/guide/en/x-pack/current/security-getting-started.html
	client, err := elastic.NewSimpleClient(elastic.SetURL(url), elastic.SetBasicAuth("elastic", "changeme"))
	a.Nil(err)
	defer client.Stop()

	ctx := context.Background()
	indexName := "docker-it-" + uuid.New().String()
	_, err = client.CreateIndex(indexName).Do(ctx)
	a.Nil(err)
}
