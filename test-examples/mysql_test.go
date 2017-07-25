package test_examples

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMySQLCall(t *testing.T) {
	a := assert.New(t)

	host := dockerEnvironment.Host()
	port, err := dockerEnvironment.Port("it-mysql", "")
	a.Nil(err)

	url := fmt.Sprintf("root:mypassword@tcp(%s:%d)/", host, port)

	db, err := sql.Open("mysql", url)
	a.Nil(err)

	rows, err := db.Query("SELECT 1")
	a.Nil(err)

	for rows.Next() {
		var val int
		err = rows.Scan(&val)
		a.Nil(err)
		fmt.Println(val)
	}
	defer db.Close()

}
