package celeritas

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const version = "1.0.0"

// Celeritas is the main application struct that holds configuration and logging.
type Celeritas struct {
	AppName  string
	Debug    bool
	Version  string
	ErrorLog *log.Logger
	InfoLog  *log.Logger
	RootPath string
	config   config
}

// config holds internal configuration settings for the application.
type config struct {
	port     string
	renderer string
}

// New initializes a new Celeritas instance with the given root path.
// It sets up directories, loads environment variables, and configures loggers.
func (c *Celeritas) New(rootPath string) error {
	pathConfig := initPaths{
		rootPath:    rootPath,
		folderNames: []string{"handlers", "migrations", "views", "data", "public", "tmp", "logs", "middleware"},
	}

	err := c.Init(pathConfig)
	if err != nil {
		return err
	}

	err = c.CheckDotEnv(rootPath)
	if err != nil {
		return err
	}

	// Read .env file to populate environment variables.
	err = godotenv.Load(rootPath + "/.env")
	if err != nil {
		return err
	}

	// Create loggers for info and error output.
	infoLog, errorLog, err := c.startLoggers()
	if err != nil {
		return err
	}

	c.InfoLog = infoLog
	c.ErrorLog = errorLog
	c.Debug, err = strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		c.Debug = false // Default to false if DEBUG env var is invalid.
	}
	c.Version = version
	c.RootPath = rootPath

	c.config = config{
		port:     os.Getenv("PORT"),
		renderer: os.Getenv("RENDERER"),
	}
	return nil
}

// Init creates the necessary directory structure for the application based on the provided paths.
func (c *Celeritas) Init(p initPaths) error {
	root := p.rootPath
	for _, path := range p.folderNames {
		// Create a folder if it doesn’t already exist.
		err := c.CreateDirIfNotExist(root + "/" + path)
		if err != nil {
			return err
		}
	}
	return nil
}

// CheckDotEnv ensures a .env file exists at the specified path, creating it if necessary.
func (c *Celeritas) CheckDotEnv(path string) error {
	err := c.CreateFileIfNotExists(fmt.Sprintf("%s/.env", path))
	if err != nil {
		return err
	}
	return nil
}