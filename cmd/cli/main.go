package main

import (
	"errors"
	"github.com/fatih/color"
	"github.com/jorgeSader/celeritas"
	"log"
	"os"
)

const version = "0.1.0"

var cel celeritas.Celeritas

func main() {
	arg1, arg2, arg3, err := validateInput()
	if err != nil {
		exitGracefully(err)
	}

	setup()

	switch arg1 {
	case "help":
		showHelp()

	case "version":
		color.Yellow("Application Version: " + version)

	case "make":
		if arg2 == "" {
			exitGracefully(errors.New("make requires a subcommand: (migration|model|handler)"))
		}
		err = doMake(arg2, arg3)
		if err != nil {
			exitGracefully(err)
		}

	default:
		log.Println(arg1, arg2, arg3)
	}
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

func showHelp() {
	color.Yellow(`Available commands:
	help		- show this help
	version		- show version

	`)
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