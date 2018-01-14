package mysql

import (
	"errors"
	dit "github.com/grepplabs/docker-it"
	"github.com/grepplabs/docker-it/wait"
	"github.com/grepplabs/docker-it/wait/database"
	// the initialization registers mysql as a driver for the SQL interface.
	_ "github.com/go-sql-driver/mysql"
)

// Options defines MySQL wait parameters.
type Options struct {
	WaitOptions wait.Options
}

type mySQLWait struct {
	waitOptions wait.Options
	databaseURL string
}

// NewMySQLWait creates a new MySQL wait
func NewMySQLWait(databaseURL string, options Options) *mySQLWait {
	if databaseURL == "" {
		panic(errors.New("mysql wait: DatabaseUrl must not be empty"))
	}
	return &mySQLWait{
		waitOptions: options.WaitOptions,
		databaseURL: databaseURL,
	}
}

// implements dockerit.Callback
func (r *mySQLWait) Call(componentName string, resolver dit.ValueResolver) error {
	databaseWait := database.NewDatabaseWait(
		"mysql", r.databaseURL,
		database.Options{
			WaitOptions: r.waitOptions,
		})
	return databaseWait.Call(componentName, resolver)
}
