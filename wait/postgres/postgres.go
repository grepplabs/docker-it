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
	DatabaseUrl string
}

type postgresWait struct {
	Options
}

func NewPostgresWait(options Options) *postgresWait {
	return &postgresWait{
		Options{
			Wait:        options.Wait,
			DatabaseUrl: options.DatabaseUrl,
		},
	}
}

// implements dockerit.Callback
func (r *postgresWait) Call(componentName string, resolver dit.ValueResolver) error {
	if r.DatabaseUrl == "" {
		return errors.New("postgres wait: DatabaseUrl must not be empty")
	}
	databaseWait := database.NewDatabaseWait(database.Options{
		Wait:        r.Wait,
		DatabaseUrl: r.DatabaseUrl,
		DriverName:  "postgres",
	})
	return databaseWait.Call(componentName, resolver)
}
