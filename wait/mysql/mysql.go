package mysql

import (
	"errors"
	dit "github.com/cloud-42/docker-it"
	"github.com/cloud-42/docker-it/wait"
	"github.com/cloud-42/docker-it/wait/database"
	_ "github.com/go-sql-driver/mysql"
)

type Options struct {
	WaitOptions wait.Options
}

type mySQLWait struct {
	waitOptions wait.Options
	databaseUrl string
}

func NewMySQLWait(databaseUrl string, options Options) *mySQLWait {
	if databaseUrl == "" {
		panic(errors.New("mysql wait: DatabaseUrl must not be empty"))
	}
	return &mySQLWait{
		waitOptions: options.WaitOptions,
		databaseUrl: databaseUrl,
	}
}

// implements dockerit.Callback
func (r *mySQLWait) Call(componentName string, resolver dit.ValueResolver) error {
	databaseWait := database.NewDatabaseWait(
		"mysql", r.databaseUrl,
		database.Options{
			WaitOptions: r.waitOptions,
		})
	return databaseWait.Call(componentName, resolver)
}
