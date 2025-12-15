// config/database.go
package config

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func ConnectDB(url string) (*sql.DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
