package mysql

import (
	"errors"
	dit "github.com/cloud-42/docker-it"
	"github.com/cloud-42/docker-it/wait"
	"github.com/cloud-42/docker-it/wait/database"
	_ "github.com/go-sql-driver/mysql"
)

type Options struct {
	wait.Wait
	DatabaseUrl string
}

type mySQLWaitWait struct {
	Options
}

func NewMySQLWait(options Options) *mySQLWaitWait {
	return &mySQLWaitWait{
		Options{
			Wait:        options.Wait,
			DatabaseUrl: options.DatabaseUrl,
		},
	}
}

// implements dockerit.Callback
func (r *mySQLWaitWait) Call(componentName string, resolver dit.ValueResolver) error {
	if r.DatabaseUrl == "" {
		return errors.New("mysql wait: DatabaseUrl must not be empty")
	}
	databaseWait := database.NewDatabaseWait(database.Options{
		Wait:        r.Wait,
		DatabaseUrl: r.DatabaseUrl,
		DriverName:  "mysql",
	})
	return databaseWait.Call(componentName, resolver)
}
