package devify

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
	"github.com/jorgeSader/devify/cache"
	"github.com/jorgeSader/devify/render"
	"github.com/jorgeSader/devify/session"

	"github.com/joho/godotenv"
)

const version = "1.0.0"

// Devify is the main application struct that holds configuration and logging.
type Devify struct {
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

// New initializes a new Devify instance with the given root path.
// It sets up directories, loads environment variables, and configures loggers.
func (d *Devify) New(rootPath string) error {
	pathConfig := initPaths{
		rootPath:    rootPath,
		folderNames: []string{"handlers", "migrations", "views", "data", "public", "tmp", "logs", "middleware"},
	}

	err := d.Init(pathConfig)
	if err != nil {
		return err
	}

	err = d.CheckDotEnv(rootPath)
	if err != nil {
		return err
	}

	// Read .env file to populate environment variables.
	err = godotenv.Load(rootPath + "/.env")
	if err != nil {
		return err
	}

	// Create loggers for info and error output.
	infoLog, errorLog, err := d.startLoggers()
	if err != nil {
		return err
	}

	// connect to database
	dbType := os.Getenv("DATABASE_TYPE")
	if dbType != "" {
		db, err := d.OpenDB(dbType, d.BuildDSN())
		if err != nil {
			errorLog.Println(err)
			os.Exit(1)
		}
		d.DB = Database{
			DataType: dbType,
			Pool:     db,
		}

	}

	if strings.ToLower(os.Getenv("CACHE")) == "redis" {
		myRedisCache := d.createClientRedisCache()
		d.Cache = myRedisCache
	}

	d.InfoLog = infoLog
	d.ErrorLog = errorLog
	d.Debug, err = strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		d.Debug = false // Default to false if DEBUG env var is invalid.
	}
	d.Version = version
	d.RootPath = rootPath
	d.Routes = d.routes().(*chi.Mux)

	d.config = config{
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
			dsn:      d.BuildDSN(),
		},
		redis: redisConfig{
			host:     os.Getenv("REDIS_HOST"),
			password: os.Getenv("REDIS_PASSWORD"),
			prefix:   os.Getenv("REDIS_PREFIX"),
		},
	}

	// create session
	sess := session.Session{
		CookieName:     d.config.cookie.name,
		CookieLifetime: d.config.cookie.lifeTime,
		CookiePersist:  d.config.cookie.persist,
		CookieSecure:   d.config.cookie.secure,
		CookieDomain:   d.config.cookie.domain,
		SessionType:    d.config.sessionType,
		BDPool:         d.DB.Pool,
	}

	d.Session = sess.InitSession()
	d.EncryptionKey = os.Getenv("ENCRYPTION_KEY")

	var views = jet.NewSet(
		jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
		jet.InDevelopmentMode(),
	)

	d.JetViews = views

	d.createRenderer()

	return nil
}

// Init creates the necessary directory structure for the application based on the provided paths.
func (d *Devify) Init(p initPaths) error {
	root := p.rootPath
	for _, path := range p.folderNames {
		// Create a folder if it doesn’t already exist.
		err := d.CreateDirIfNotExist(root + "/" + path)
		if err != nil {
			return err
		}
	}
	return nil
}

// ListenAndServe Starts the webserver
func (d *Devify) ListenAndServe() {
	srv := &http.Server{
		Addr:         ":" + d.config.port,
		ErrorLog:     d.ErrorLog,
		Handler:      d.Routes,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
	}

	defer d.DB.Pool.Close()

	d.InfoLog.Printf("Server listening on port %s", d.config.port)
	err := srv.ListenAndServe()
	if err != nil {
		d.ErrorLog.Fatal(err)
	}
}

// CheckDotEnv ensures a .env file exists at the specified path, creating it if necessary.
func (d *Devify) CheckDotEnv(path string) error {
	err := d.CreateFileIfNotExists(fmt.Sprintf("%s/.env", path))
	if err != nil {
		return err
	}
	return nil
}

// startLoggers initializes info and error loggers with timestamps and file info for errors.
func (d *Devify) startLoggers() (*log.Logger, *log.Logger, error) {
	var infoLog *log.Logger
	var errorLog *log.Logger

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	return infoLog, errorLog, nil
}

func (d *Devify) createRenderer() {
	myRenderer := render.Render{
		RootPath: d.RootPath,
		Renderer: d.config.renderer,
		Port:     d.config.port,
		JetViews: d.JetViews,
		Session:  d.Session,
		UseCache: false, // TODO: Enable caching by default and/or add to config file
	}

	// Initialize template cache for Go templates
	if strings.ToLower(d.config.renderer) == "go" {
		cache, err := myRenderer.CreateTemplateCache()
		if err != nil {
			d.ErrorLog.Fatalf("Failed to create template cache: %v", err)
		}
		myRenderer.TemplateCache = cache
	}

	d.Render = &myRenderer
}

func (d *Devify) createClientRedisCache() *cache.RedisCache {
	cacheClient := cache.RedisCache{
		Conn:   d.createRedisPool(),
		Prefix: d.config.redis.prefix,
	}
	return &cacheClient
}

func (d *Devify) createRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50,
		MaxActive:   10000,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp",
				d.config.redis.host,
				redis.DialPassword(d.config.redis.password))
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func (d *Devify) BuildDSN() string {
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
