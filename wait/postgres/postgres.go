package postgres

import (
	dit "github.com/cloud-42/docker-it"
	"github.com/cloud-42/docker-it/wait"
	"github.com/cloud-42/docker-it/wait/database"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type Options struct {
	WaitOptions wait.Options
}

type postgresWait struct {
	waitOptions wait.Options
	databaseUrl string
}

func NewPostgresWait(databaseUrl string, options Options) *postgresWait {
	if databaseUrl == "" {
		panic(errors.New("postgres wait: DatabaseUrl must not be empty"))
	}
	return &postgresWait{
		waitOptions: options.WaitOptions,
		databaseUrl: databaseUrl,
	}
}

// implements dockerit.Callback
func (r *postgresWait) Call(componentName string, resolver dit.ValueResolver) error {
	databaseWait := database.NewDatabaseWait(
		"postgres", r.databaseUrl,
		database.Options{
			WaitOptions: r.waitOptions,
		})
	return databaseWait.Call(componentName, resolver)
}
