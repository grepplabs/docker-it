package postgres

import (
	dit "github.com/cloud-42/docker-it"
	"github.com/cloud-42/docker-it/wait"
	"github.com/cloud-42/docker-it/wait/database"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type PostgresWait struct {
	wait.Wait
	DatabaseUrl string
}

// implements dockerit.Callback
func (r *PostgresWait) Call(componentName string, resolver dit.ValueResolver) error {
	if r.DatabaseUrl == "" {
		return errors.New("postgres wait: DatabaseUrl must not be empty")
	}

	databaseWait := database.DatabaseWait{
		Wait:        r.Wait,
		DatabaseUrl: r.DatabaseUrl,
		DriverName:  "postgres",
	}
	return databaseWait.Call(componentName, resolver)
}
