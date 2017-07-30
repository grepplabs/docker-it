package testexamples

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPostgesCall(t *testing.T) {
	a := assert.New(t)

	urlTemplate := `postgres://postgres:postgres@{{ value . "it-postgres.Host"}}:{{ value . "it-postgres.Port"}}/postgres?sslmode=disable`
	url, err := dockerEnvironment.Resolve(urlTemplate)
	a.Nil(err)
	fmt.Println(url)

	db, err := sql.Open("postgres", url)
	a.Nil(err)
	defer db.Close()

	rows, err := db.Query("SELECT now()")
	a.Nil(err)
	defer rows.Close()

	for rows.Next() {
		var date time.Time
		err = rows.Scan(&date)
		a.Nil(err)
		fmt.Println(date)
	}
}
