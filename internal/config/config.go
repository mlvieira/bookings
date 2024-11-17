package config

import (
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/mlvieira/bookings/internal/models"
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
	MailChan      chan models.MailData
}

// SetupAppConfig initializes the main application configuration
func SetupAppConfig(inProduction bool) *AppConfig {

	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})

	mailChan := make(chan models.MailData)

	app := AppConfig{
		InProduction: inProduction,
		InfoLog:      log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		ErrorLog:     log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		MailChan:     mailChan,
	}

	session := scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	app.Session = session

	return &app
}
