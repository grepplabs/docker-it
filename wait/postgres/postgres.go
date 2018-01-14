package postgres

import (
	dit "github.com/grepplabs/docker-it"
	"github.com/grepplabs/docker-it/wait"
	"github.com/grepplabs/docker-it/wait/database"
	// the initialization registers pq as a driver for the SQL interface.
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

// Options defines PostgreSQL wait parameters.
type Options struct {
	WaitOptions wait.Options
}

type postgresWait struct {
	waitOptions wait.Options
	databaseURL string
}

// NewPostgresWait creates a new PostgreSQL wait
func NewPostgresWait(databaseURL string, options Options) *postgresWait {
	if databaseURL == "" {
		panic(errors.New("postgres wait: DatabaseUrl must not be empty"))
	}
	return &postgresWait{
		waitOptions: options.WaitOptions,
		databaseURL: databaseURL,
	}
}

// implements dockerit.Callback
func (r *postgresWait) Call(componentName string, resolver dit.ValueResolver) error {
	databaseWait := database.NewDatabaseWait(
		"postgres", r.databaseURL,
		database.Options{
			WaitOptions: r.waitOptions,
		})
	return databaseWait.Call(componentName, resolver)
}
