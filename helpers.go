package celeritas

import (
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
