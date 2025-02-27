package celeritas

import (
	"log"
	"os"
)

// CreateDirIfNotExist creates a directory at the given path with permissions 0755 if it doesn’t exist.
func (c *Celeritas) CreateDirIfNotExist(path string) error {
	const mode = 0o755
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, mode)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateFileIfNotExists creates a file at the given path if it doesn’t exist.
func (c *Celeritas) CreateFileIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				return // Swallow close errors to avoid masking creation errors.
			}
		}(file)
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