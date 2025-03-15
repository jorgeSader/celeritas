package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
)

func doSessionTable() error {
	// Validate cel initialization
	if cel.DB.DataType == "" {
		return errors.New("DATABASE_TYPE not set in .env or environment")
	}
	if cel.RootPath == "" {
		return errors.New("RootPath not set; setup failed")
	}

	dbType := cel.DB.DataType
	var templateFile string

	switch dbType {
	case "mysql", "mariadb":
		templateFile = "mysql_session"

	case "postgres", "postgresql":
		templateFile = "postgres_session"

	case "sqlite", "sqlite3", "libsql", "turso":
		templateFile = "sqlite_session"

	case "mongo", "mongodb":
		color.Yellow(`Note: MongoDB doesn't require SQL migrations.
To setup a MongoDB session store:
	1. Initialize a new session manager
	2. Configure it to use mongodbstore

Example:
	sessionManager = scs.New()
	sessionManager.Store = mongodbstore.New(client.Database("database"))

See full example: https://github.com/alexedwards/scs/tree/master/mongodbstore
`)
		return nil // No SQL migrations needed for MongoDB

	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	// Generate migration file names
	fileName := fmt.Sprintf("%d_create_sessions_table", time.Now().UnixNano())
	migrationDir := cel.RootPath + "/migrations"
	upFile := migrationDir + "/" + fileName + ".up.sql"
	downFile := migrationDir + "/" + fileName + ".down.sql"

	// Ensure migrations directory exists
	if err := os.MkdirAll(migrationDir, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %v", err)
	}

	// Copy up migration from embedded templates
	templatePath := fmt.Sprintf("templates/migrations/%s.sql", templateFile)
	err := copyFileFromTemplate(templatePath, upFile)
	if err != nil {
		return fmt.Errorf("failed to copy up migration from %s to %s: %v", templatePath, upFile, err)
	}

	// Write down migration
	err = copyDataToFile([]byte("DROP TABLE sessions"), downFile)
	if err != nil {
		return fmt.Errorf("failed to write down migration to %s: %v", downFile, err)
	}

	// Run migrations (apply all pending)
	err = doMigrate("up", "")
	if err != nil {
		return fmt.Errorf("migration failed: %v", err)
	}

	return nil
}
