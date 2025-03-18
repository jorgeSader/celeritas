package celeritas

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
	"github.com/jorgeSader/celeritas/cache"
	"github.com/jorgeSader/celeritas/render"
	"github.com/jorgeSader/celeritas/session"

	"github.com/joho/godotenv"
)

const version = "1.0.0"

// Celeritas is the main application struct that holds configuration and logging.
type Celeritas struct {
	AppName       string
	Debug         bool
	Version       string
	ErrorLog      *log.Logger
	InfoLog       *log.Logger
	RootPath      string
	Routes        *chi.Mux
	Render        *render.Render
	Session       *scs.SessionManager
	DB            Database
	JetViews      *jet.Set
	config        config
	EncryptionKey string
	Cache         cache.Cache
}

// config holds internal configuration settings for the application.
type config struct {
	port        string
	renderer    string
	cookie      cookieConfig
	sessionType string
	database    databaseConfig
	redis       redisConfig
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

	// connect to database
	dbType := os.Getenv("DATABASE_TYPE")
	if dbType != "" {
		db, err := c.OpenDB(dbType, c.BuildDSN())
		if err != nil {
			errorLog.Println(err)
			os.Exit(1)
		}
		c.DB = Database{
			DataType: dbType,
			Pool:     db,
		}

	}

	if strings.ToLower(os.Getenv("CACHE")) == "redis" {
		myRedisCache := c.createClientRedisCache()
		c.Cache = myRedisCache
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
		cookie: cookieConfig{
			name:     os.Getenv("COOKIE_NAME"),
			lifeTime: os.Getenv("COOKIE_LIFETIME"),
			persist:  os.Getenv("COOKIE_PERSIST"),
			secure:   os.Getenv("COOKIE_SECURE"),
			domain:   os.Getenv("COOKIE_DOMAIN"),
		},
		sessionType: os.Getenv("SESSION_TYPE"),

		database: databaseConfig{
			database: dbType,
			dsn:      c.BuildDSN(),
		},
		redis: redisConfig{
			host:     os.Getenv("REDIS_HOST"),
			password: os.Getenv("REDIS_PASSWORD"),
			prefix:   os.Getenv("REDIS_PREFIX"),
		},
	}

	// create session
	sess := session.Session{
		CookieName:     c.config.cookie.name,
		CookieLifetime: c.config.cookie.lifeTime,
		CookiePersist:  c.config.cookie.persist,
		CookieSecure:   c.config.cookie.secure,
		CookieDomain:   c.config.cookie.domain,
		SessionType:    c.config.sessionType,
		BDPool:         c.DB.Pool,
	}

	c.Session = sess.InitSession()
	c.EncryptionKey = os.Getenv("ENCRYPTION_KEY")

	var views = jet.NewSet(
		jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
		jet.InDevelopmentMode(),
	)

	c.JetViews = views

	c.createRenderer()

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
		Handler:      c.Routes,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
	}

	defer c.DB.Pool.Close()

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

func (c *Celeritas) createRenderer() {
	myRenderer := render.Render{
		RootPath: c.RootPath,
		Renderer: c.config.renderer,
		Port:     c.config.port,
		JetViews: c.JetViews,
		Session:  c.Session,
		UseCache: false, // TODO: Enable caching by default and/or add to config file
	}

	// Initialize template cache for Go templates
	if strings.ToLower(c.config.renderer) == "go" {
		cache, err := myRenderer.CreateTemplateCache()
		if err != nil {
			c.ErrorLog.Fatalf("Failed to create template cache: %v", err)
		}
		myRenderer.TemplateCache = cache
	}

	c.Render = &myRenderer
}

func (c *Celeritas) createClientRedisCache() *cache.RedisCache {
	cacheClient := cache.RedisCache{
		Conn:   c.createRedisPool(),
		Prefix: c.config.redis.prefix,
	}
	return &cacheClient
}

func (c *Celeritas) createRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50,
		MaxActive:   10000,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp",
				c.config.redis.host,
				redis.DialPassword(c.config.redis.password))
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func (c *Celeritas) BuildDSN() string {
	var dsn string

	dbType := strings.ToLower(os.Getenv("DATABASE_TYPE"))

	switch dbType {
	case "postgres", "postgresql":
		dsn = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5",
			os.Getenv("DATABASE_HOST"),
			os.Getenv("DATABASE_PORT"),
			os.Getenv("DATABASE_USER"),
			os.Getenv("DATABASE_NAME"),
			os.Getenv("DATABASE_SSL_MODE"))

		if os.Getenv("DATABASE_PASS") != "" {
			dsn = fmt.Sprintf("%s password=%s", dsn, os.Getenv("DATABASE_PASS"))
		}

	case "mariadb", "mysql":

	default:

	}
	return dsn
}
