package session

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/postgresstore"
	//"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
)

type Session struct {
	CookieLifetime string
	CookiePersist  string
	CookieSecure   string
	CookieName     string
	CookieDomain   string
	SessionType    string
	BDPool         *sql.DB
}

func (c *Session) InitSession() *scs.SessionManager {
	var persist, secure bool

	// how long should sessions last? (defaults to 60min)
	minutes, err := strconv.Atoi(c.CookieLifetime)
	if err != nil {
		minutes = 60
	}

	// should cookies persist? (defaults to false)
	if strings.ToLower(c.CookiePersist) == "true" {
		persist = true
	}

	// must cookies be secure? (defaults to false)
	if strings.ToLower(c.CookieSecure) == "true" {
		secure = true
	}

	// create session
	session := scs.New()
	session.Lifetime = time.Duration(minutes) * time.Minute
	session.Cookie.Persist = persist
	session.Cookie.Name = c.CookieName
	session.Cookie.Secure = secure
	session.Cookie.Domain = c.CookieDomain
	session.Cookie.SameSite = http.SameSiteLaxMode

	// which session store?
	switch strings.ToLower(c.SessionType) {
	case "redis":
		//session.Store = redisstore.New(c.BDPool)

	case "mysql", "mariadb":
		session.Store = mysqlstore.New(c.BDPool)

	case "postgres", "postgresql":
		session.Store = postgresstore.New(c.BDPool)

	case "sqlite", "sqlite3", "libsql", "turso", "tursodb":
		session.Store = sqlite3store.New(c.BDPool)

	default:
		// cookie
	}

	return session
}
