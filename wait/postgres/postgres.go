package postgres

import (
	dit "github.com/cloud-42/docker-it"
	"github.com/cloud-42/docker-it/wait"
	"github.com/cloud-42/docker-it/wait/database"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type Options struct {
	wait.Wait
}

type postgresWait struct {
	wait.Wait
	databaseUrl string
}

func NewPostgresWait(databaseUrl string, options Options) *postgresWait {
	if databaseUrl == "" {
		panic(errors.New("postgres wait: DatabaseUrl must not be empty"))
	}
	return &postgresWait{
		Wait:        options.Wait,
		databaseUrl: databaseUrl,
	}
}

// implements dockerit.Callback
func (r *postgresWait) Call(componentName string, resolver dit.ValueResolver) error {
	databaseWait := database.NewDatabaseWait(
		"postgres", r.databaseUrl,
		database.Options{
			Wait: r.Wait,
		})
	return databaseWait.Call(componentName, resolver)
}
