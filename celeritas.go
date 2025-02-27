package celeritas

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/jorgeSader/celeritas/render"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

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
	Routes   *chi.Mux
	Render   *render.Render
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
	c.Routes = c.routes().(*chi.Mux)

	c.config = config{
		port:     os.Getenv("PORT"),
		renderer: os.Getenv("RENDERER"),
	}

	c.Render = c.createRenderer(c)

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

// ListenAndServe Starts the webserver
func (c *Celeritas) ListenAndServe() {
	srv := &http.Server{
		Addr:         ":" + c.config.port,
		ErrorLog:     c.ErrorLog,
		Handler:      c.routes(),
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
	}
	c.InfoLog.Printf("Server listening on port %s", c.config.port)
	err := srv.ListenAndServe()
	if err != nil {
		c.ErrorLog.Fatal(err)
	}
}

// CheckDotEnv ensures a .env file exists at the specified path, creating it if necessary.
func (c *Celeritas) CheckDotEnv(path string) error {
	err := c.CreateFileIfNotExists(fmt.Sprintf("%s/.env", path))
	if err != nil {
		return err
	}
	return nil
}

// startLoggers initializes info and error loggers with timestamps and file info for errors.
func (c *Celeritas) startLoggers() (*log.Logger, *log.Logger, error) {
	var infoLog *log.Logger
	var errorLog *log.Logger

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	return infoLog, errorLog, nil
}

func (c *Celeritas) createRenderer(cel *Celeritas) *render.Render {
	myRenderer := render.Render{
		RootPath: cel.RootPath,
		Renderer: cel.config.renderer,
		Port:     cel.config.port,
	}
	return &myRenderer
}
