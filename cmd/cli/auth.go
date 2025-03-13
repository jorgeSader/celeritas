package main

import (
	"fmt"
	"time"
)

func doAuth(arg1, arg2 string) error {
	// Create migrations
	dbType := cel.DB.DataType
	fileName := fmt.Sprintf("%d_create_auth_tables", time.Now().UnixMicro())
	upFile := cel.RootPath + "/migrations/" + fileName + ".up.sql"
	downFile := cel.RootPath + "/migrations/" + fileName + ".down.sql"

	err := copyFileFromTemplate("templates/migrations/auth_tables."+dbType+".up.sql", upFile)
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/migrations/auth_tables."+dbType+".down.sql", downFile)
	if err != nil {
		exitGracefully(err)
	}

	// Run migrations
	err = doMigrate("up", "")
	if err != nil {
		exitGracefully(err)
	}

	return nil
}