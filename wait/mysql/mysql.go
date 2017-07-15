package mysql

import (
	"errors"
	dit "github.com/cloud-42/docker-it"
	"github.com/cloud-42/docker-it/wait"
	"github.com/cloud-42/docker-it/wait/database"
	_ "github.com/go-sql-driver/mysql"
)

type MySQLWait struct {
	wait.Wait
	DatabaseUrl string
}

// implements dockerit.Callback
func (r *MySQLWait) Call(componentName string, resolver dit.ValueResolver) error {
	if r.DatabaseUrl == "" {
		return errors.New("mysql wait: DatabaseUrl must not be empty")
	}

	databaseWait := database.DatabaseWait{
		Wait:        r.Wait,
		DatabaseUrl: r.DatabaseUrl,
		DriverName:  "mysql",
	}
	return databaseWait.Call(componentName, resolver)
}
