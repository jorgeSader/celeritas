package celeritas

// initPaths defines the root path and folder structure for initializing the application.
type initPaths struct {
	rootPath    string
	folderNames []string
}

type cookieConfig struct {
	name     string
	lifeTime string
	persist  string
	secure   string
	domain   string
}
