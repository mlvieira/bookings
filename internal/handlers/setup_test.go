package handlers

import (
	"encoding/gob"
	"log"
	"net/http"
	"os"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mlvieira/bookings/internal/config"
	"github.com/mlvieira/bookings/internal/helpers"
	"github.com/mlvieira/bookings/internal/models"
	"github.com/mlvieira/bookings/internal/render"
)

var app config.AppConfig

func getRoutes() http.Handler {
	if err := os.Chdir("../.."); err != nil {
		log.Fatal("Could not change working directory:", err)
	}

	gob.Register(models.Reservation{})

	app = *config.SetupAppConfig(false)

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal(err)
	}

	app.TemplateCache = tc
	app.UseCache = app.InProduction
	app.Port = ":8080"

	repo := NewRepo(&app)
	NewHandlers(repo)
	render.NewTemplates(&app)
	helpers.NewHelpers(&app)

	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(sessionLoad(app.Session))

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

	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	return mux
}

// sessionLoad loads and saves the session on every request
func sessionLoad(session *scs.SessionManager) func(http.Handler) http.Handler {
	return session.LoadAndSave
}
