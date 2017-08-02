package database

import (
	"database/sql"
	"fmt"
	dit "github.com/cloud-42/docker-it"
	"github.com/cloud-42/docker-it/wait"
	"github.com/pkg/errors"
)

// Options defines database wait parameters.
type Options struct {
	WaitOptions wait.Options
}

type databaseWait struct {
	wait.Wait
	driverName  string
	databaseUrl string
}

// NewDatabaseWait creates a new database wait
func NewDatabaseWait(driverName string, databaseUrl string, options Options) *databaseWait {
	if driverName == "" {
		panic(errors.New("database wait: DatabaseUrl must not be empty"))
	}
	if databaseUrl == "" {
		panic(errors.New("database wait: DriverName must not be empty"))
	}
	return &databaseWait{
		Wait:        wait.NewWait(options.WaitOptions),
		driverName:  driverName,
		databaseUrl: databaseUrl,
	}
}

// implements dockerit.Callback
func (r *databaseWait) Call(componentName string, resolver dit.ValueResolver) error {
	url, err := resolver.Resolve(r.databaseUrl)
	if err != nil {
		return err
	}
	err = r.pollConnect(componentName, url)
	if err != nil {
		return fmt.Errorf("%s wait: failed to connect to %s %v ", r.driverName, url, err)
	}
	return nil
}

func (r *databaseWait) pollConnect(componentName string, url string) error {

	logger := r.GetLogger(componentName)
	logger.Printf("Waiting for %s %s\n", r.driverName, url)

	f := func() error {
		return r.connect(url)
	}
	return r.Poll(componentName, f)
}

func (r *databaseWait) connect(url string) error {
	db, err := sql.Open(r.driverName, url)
	if err != nil {
		return err
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		return err
	}
	return nil
}
