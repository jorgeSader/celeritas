package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

func setup() {
	err := godotenv.Load()
	if err != nil {
		exitGracefully(err)
	}

	path, err := os.Getwd()
	if err != nil {
		exitGracefully(err)
	}

	cel.RootPath = path
	cel.DB.DataType = strings.ToLower(os.Getenv("DATABASE_TYPE"))
}

func getDSN() string {
	dbType := cel.DB.DataType

	var dsn string

	switch dbType {
	case "pgx", "postgres", "postgresql":
		if os.Getenv("DATABASE_PASS") != "" {
			dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
				os.Getenv("DATABASE_USER"),
				os.Getenv("DATABASE_PASS"),
				os.Getenv("DATABASE_HOST"),
				os.Getenv("DATABASE_PORT"),
				os.Getenv("DATABASE_NAME"),
				os.Getenv("DATABASE_SSL_MODE"),
			)
		} else {
			dsn = fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s",
				os.Getenv("DATABASE_USER"),
				os.Getenv("DATABASE_HOST"),
				os.Getenv("DATABASE_PORT"),
				os.Getenv("DATABASE_NAME"),
				os.Getenv("DATABASE_SSL_MODE"),
			)
		}
		return dsn

	case "mysql", "mariadb":
		return "mysql://" + cel.BuildDSN()

	case "sql", "sqlite", "sqlite3", "turso":
		//TODO: build dsn

	case "mongo", "mongodb":
	//TODO: build dsn

	default:
		exitGracefully(errors.New("database type not supported: " + dbType))
	}
	return dsn
}

func showHelp() {
	color.Yellow(`Available commands:

	help                    - show this help
	version	                - show version
	migrate                 - runs all up migrations that have not been applied
	migrate down            - reverses the most recent migration
	migrate reset           - runs all down migrations in reverse order, and then all up migrations
	make migration <name>   - creates two new migrations(one up & one down) in the migrations folder
	make auth 				- creates and runs migrations for authentication tables, and creates models and middleware
	make handler <name>		- creates a stub handler in the handlers directory
	make model <name>		- creates a new model in the data directory
	

	`)
}
