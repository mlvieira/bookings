package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mlvieira/bookings/internal/config"
	"github.com/mlvieira/bookings/internal/handlers"
)

// routes sets up application routes and middleware.
func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	mux.Get("/", handlers.Repo.Landing)
	mux.Get("/about", handlers.Repo.About)
	mux.Get("/contact", handlers.Repo.Contact)
	mux.Get("/rooms/majors-suite", handlers.Repo.RoomMajors)
	mux.Get("/rooms/generals-quarter", handlers.Repo.RoomGenerals)
	mux.Get("/book/majors-suite", handlers.Repo.Booking)
	mux.Get("/availability", handlers.Repo.Availability)
	mux.Post("/availability", handlers.Repo.PostAvailability)
	mux.Post("/availability/json", handlers.Repo.AvailabilityJSON)

	return mux
}
