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

func (d *Session) InitSession() *scs.SessionManager {
	var persist, secure bool

	// how long should sessions last? (defaults to 60min)
	minutes, err := strconv.Atoi(d.CookieLifetime)
	if err != nil {
		minutes = 60
	}

	// should cookies persist? (defaults to false)
	if strings.ToLower(d.CookiePersist) == "true" {
		persist = true
	}

	// must cookies be secure? (defaults to false)
	if strings.ToLower(d.CookieSecure) == "true" {
		secure = true
	}

	// create session
	session := scs.New()
	session.Lifetime = time.Duration(minutes) * time.Minute
	session.Cookie.Persist = persist
	session.Cookie.Name = d.CookieName
	session.Cookie.Secure = secure
	session.Cookie.Domain = d.CookieDomain
	session.Cookie.SameSite = http.SameSiteLaxMode

	// which session store?
	switch strings.ToLower(d.SessionType) {
	case "redis":
		//session.Store = redisstore.New(d.BDPool)

	case "mysql", "mariadb":
		session.Store = mysqlstore.New(d.BDPool)

	case "postgres", "postgresql":
		session.Store = postgresstore.New(d.BDPool)

	case "sqlite", "sqlite3", "libsql", "turso", "tursodb":
		session.Store = sqlite3store.New(d.BDPool)

	default:
		// cookie
	}

	return session
}
