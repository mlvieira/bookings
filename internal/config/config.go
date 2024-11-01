package config

import (
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/alexedwards/scs/v2"
)

// AppConfig holds the application config
type AppConfig struct {
	UseCache      bool
	TemplateCache map[string]*template.Template
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
	InProduction  bool
	Port          string
	Session       *scs.SessionManager
}

// SetupAppConfig initializes the main application configuration
func SetupAppConfig(inProduction bool) *AppConfig {

	app := AppConfig{
		InProduction: inProduction,
		InfoLog:      log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		ErrorLog:     log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
	}

	session := scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	app.Session = session

	return &app
}
