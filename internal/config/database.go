package config

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

func NewDB(env *Env) (*sql.DB, error) {
	db, err := sql.Open("postgres", env.DBURL)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	db.SetMaxOpenConns(15)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(2 * time.Hour)
	db.SetConnMaxIdleTime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db.Ping: %w", err)
	}

	return db, nil
}
