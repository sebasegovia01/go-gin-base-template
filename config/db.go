package config

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

func NewPostgresDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}

	// Connection pool
	db.SetMaxOpenConns(25)                 // max number of open conns
	db.SetMaxIdleConns(25)                 // max number of inactive conns
	db.SetConnMaxLifetime(5 * time.Minute) // max time life

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
