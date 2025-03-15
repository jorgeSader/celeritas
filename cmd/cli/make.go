package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
)

func doMake(arg2, arg3 string) error {
	switch arg2 {
	case "migration":
		dbType := cel.DB.DataType
		if arg3 == "" {
			exitGracefully(errors.New("migration name is required"))
		}

		fileName := fmt.Sprintf("%d_%s", time.Now().UnixMicro(), arg3)

		upFile := cel.RootPath + "/migrations/" + fileName + dbType + ".up.sql"
		downFile := cel.RootPath + "/migrations/" + fileName + dbType + ".down.sql"

		err := copyFileFromTemplate("templates/migrations/migration."+dbType+".up.sql", upFile)
		if err != nil {
			exitGracefully(err)
		}

		err = copyFileFromTemplate("templates/migrations/migration."+dbType+".down.sql", downFile)
		if err != nil {
			exitGracefully(err)
		}

	case "auth":
		err := doAuth(arg2, arg3)
		if err != nil {
			exitGracefully(err)
		}

	case "handler":
		if arg3 == "" {
			exitGracefully(errors.New("handler name is required"))
		}

		fileName := cel.RootPath + "/handlers/" + strings.ToLower(arg3) + ".go"
		if fileExists(fileName) {
			exitGracefully(errors.New(fileName + " already exists"))
		}

		data, err := templateFS.ReadFile("templates/handlers/handler.go.txt")
		if err != nil {
			exitGracefully(err)
		}

		handler := string(data)
		handler = strings.ReplaceAll(handler, "$HANDLERNAME$", strcase.ToCamel(arg3))

		err = ioutil.WriteFile(fileName, []byte(handler), 0644)
		if err != nil {
			exitGracefully(err)
		}

	}
	return nil
}
