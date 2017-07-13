package postgres

import (
	"database/sql"
	"fmt"
	dit "github.com/cloud-42/docker-it"
	"github.com/cloud-42/docker-it/wait"
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
		return errors.New("postgres wait: UrlTemplate must not be empty")
	}
	if url, err := resolver.Resolve(r.DatabaseUrl); err != nil {
		return err
	} else {
		err := r.pollConnect(componentName, url)
		if err != nil {
			return fmt.Errorf("postgres wait: failed to connect to %s %v ", url, err)
		}
		return nil
	}
}

func (r *PostgresWait) pollConnect(componentName string, url string) error {

	logger := r.GetLogger(componentName)
	logger.Println("Waiting for postgres", url)

	f := func() error {
		return r.connect(url)
	}
	return r.Poll(componentName, f)
}

func (r *PostgresWait) connect(url string) error {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return err
	}
	defer db.Close()
	rows, err := db.Query("SELECT 1")
	if err != nil {
		return err
	}
	defer rows.Close()
	return err
}
