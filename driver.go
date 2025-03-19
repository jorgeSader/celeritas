package devify

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5"
)

func (d *Devify) OpenDB(dbType, dsn string) (*sql.DB, error) {
	if dbType == "postgres" || dbType == "postgresql" {
		dbType = "pgx"
	}

	db, err := sql.Open(dbType, dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
