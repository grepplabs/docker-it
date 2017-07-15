package database

import (
	"database/sql"
	"fmt"
	dit "github.com/cloud-42/docker-it"
	"github.com/cloud-42/docker-it/wait"
	"github.com/pkg/errors"
)

type DatabaseWait struct {
	wait.Wait
	DatabaseUrl string
	DriverName  string
}

// implements dockerit.Callback
func (r *DatabaseWait) Call(componentName string, resolver dit.ValueResolver) error {
	if r.DatabaseUrl == "" {
		return errors.New("database wait: DatabaseUrl must not be empty")
	}
	if r.DriverName == "" {
		return errors.New("database wait: DriverName must not be empty")
	}
	if url, err := resolver.Resolve(r.DatabaseUrl); err != nil {
		return err
	} else {
		err := r.pollConnect(componentName, url)
		if err != nil {
			return fmt.Errorf("%s wait: failed to connect to %s %v ", r.DriverName, url, err)
		}
		return nil
	}
}

func (r *DatabaseWait) pollConnect(componentName string, url string) error {

	logger := r.GetLogger(componentName)
	logger.Printf("Waiting for %s %s\n", r.DriverName, url)

	f := func() error {
		return r.connect(url)
	}
	return r.Poll(componentName, f)
}

func (r *DatabaseWait) connect(url string) error {
	db, err := sql.Open(r.DriverName, url)
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
