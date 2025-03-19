package main

import (
	"errors"
	"github.com/fatih/color"
	"github.com/jorgeSader/devify"
	"os"
)

const version = "0.1.0"

var cel devify.Devify

func main() {
	arg1, arg2, arg3, err := validateInput()
	var message string

	if err != nil {
		exitGracefully(err)
	}

	setup()

	switch arg1 {
	case "help":
		showHelp()

	case "version":
		color.Yellow("Application Version: " + version)

	case "migrate":
		if arg2 == "" {
			arg2 = "up"
		}
		err = doMigrate(arg2, arg3)
		if err != nil {
			exitGracefully(err)
		}
		message = "Migrations complete!"
	case "make":
		if arg2 == "" {
			exitGracefully(errors.New("make requires a subcommand: (migration|model|handler)"))
		}
		err = doMake(arg2, arg3)
		if err != nil {
			exitGracefully(err)
		}

	default:
		showHelp()
	}
	exitGracefully(nil, message)
}

func validateInput() (string, string, string, error) {
	var arg1, arg2, arg3 string

	if len(os.Args) > 1 {
		arg1 = os.Args[1]
		if len(os.Args) > 2 {
			arg2 = os.Args[2]
		}
		if len(os.Args) > 3 {
			arg3 = os.Args[3]
		}
	} else {
		color.Red("Error: no arguments supplied")
		showHelp()
		return "", "", "", errors.New("no arguments supplied")
	}

	return arg1, arg2, arg3, nil
}

func exitGracefully(err error, msg ...string) {
	message := ""
	if len(msg) > 0 {
		message = msg[0]
	}

	if err != nil {
		color.Red("Error: %v\n", err)
	}

	if len(message) > 0 {
		color.Yellow(message)

	} else {
		color.Green("Finished!")
	}

	os.Exit(0)
}
