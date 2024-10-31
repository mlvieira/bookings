package handlers

import (
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mlvieira/bookings/internal/config"
	"github.com/mlvieira/bookings/internal/models"
	"github.com/mlvieira/bookings/internal/render"
)

var app config.AppConfig
var session *scs.SessionManager

func getRoutes() http.Handler {
	if err := os.Chdir("../.."); err != nil {
		log.Fatal("Could not change working directory:", err)
	}

	gob.Register(models.Reservation{})

	app.InProduction = true

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	app.Session = session

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal(err)
	}

	app.TemplateCache = tc
	app.UseCache = app.InProduction

	repo := NewRepo(&app)
	NewHandlers(repo)
	render.NewTemplates(&app)

	if repo == nil {
		log.Fatal("repo is nil")
	}
	if app.TemplateCache == nil {
		log.Fatal("app.TemplateCache is nil")
	} else {
		log.Println("Template cache initialized successfully")
	}

	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(SessionLoad)

	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	mux.Get("/", Repo.Landing)
	mux.Get("/about", Repo.About)
	mux.Get("/contact", Repo.Contact)
	mux.Get("/rooms/majors-suite", Repo.RoomMajors)
	mux.Get("/rooms/generals-quarter", Repo.RoomGenerals)
	mux.Get("/availability", Repo.Availability)
	mux.Post("/availability", Repo.PostAvailability)
	mux.Post("/availability/json", Repo.AvailabilityJSON)
	mux.Get("/book", Repo.Booking)
	mux.Post("/book", Repo.PostBooking)
	mux.Get("/book/summary", Repo.ReservationSummary)

	return mux
}

// SessionLoad loads and save the session on every request
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}
