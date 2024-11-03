package routes

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/justinas/nosurf"
	"github.com/mlvieira/bookings/internal/config"
	"github.com/mlvieira/bookings/internal/handlers"
)

func Routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(noSurf(app))
	mux.Use(sessionLoad(app.Session))

	mux.Get("/", handlers.Repo.Landing)
	mux.Get("/about", handlers.Repo.About)
	mux.Get("/contact", handlers.Repo.Contact)
	mux.Get("/rooms/majors-suite", handlers.Repo.RoomMajors)
	mux.Get("/rooms/generals-quarters", handlers.Repo.RoomGenerals)
	mux.Get("/rooms/book/{id}", handlers.Repo.ChooseRoom)
	mux.Get("/availability", handlers.Repo.Availability)
	mux.Post("/availability", handlers.Repo.PostAvailability)
	mux.Post("/availability/json", handlers.Repo.AvailabilityJSON)
	mux.Get("/book", handlers.Repo.Booking)
	mux.Post("/book", handlers.Repo.PostBooking)
	mux.Get("/book/summary", handlers.Repo.ReservationSummary)

	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	return mux
}

// sessionLoad loads and saves the session on every request
func sessionLoad(session *scs.SessionManager) func(http.Handler) http.Handler {
	return session.LoadAndSave
}

// noSurf adds CSRF protection to all POST requests
func noSurf(app *config.AppConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		csrfHandler := nosurf.New(next)
		csrfHandler.SetBaseCookie(http.Cookie{
			HttpOnly: true,
			Path:     "/",
			Secure:   app.InProduction,
			SameSite: http.SameSiteLaxMode,
		})
		return csrfHandler
	}
}
