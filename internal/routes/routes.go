package routes

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/justinas/nosurf"
	"github.com/mlvieira/bookings/internal/config"
	"github.com/mlvieira/bookings/internal/handlers"
	"github.com/mlvieira/bookings/internal/helpers"
)

func Routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(noSurf(app))
	mux.Use(sessionLoad(app.Session))

	mux.Get("/", handlers.Repo.Landing)
	mux.Get("/about", handlers.Repo.About)
	mux.Get("/contact", handlers.Repo.Contact)
	mux.Get("/rooms/{room}", handlers.Repo.RoomsPage)
	mux.Get("/rooms/book/{id}", handlers.Repo.ChooseRoom)
	mux.Get("/availability", handlers.Repo.Availability)
	mux.Post("/availability", handlers.Repo.PostAvailability)
	mux.Post("/availability/json", handlers.Repo.AvailabilityJSON)
	mux.Get("/book", handlers.Repo.Booking)
	mux.Post("/book", handlers.Repo.PostBooking)
	mux.Get("/book/summary", handlers.Repo.ReservationSummary)
	mux.Get("/user/login", handlers.Repo.ShowLoginPage)
	mux.Post("/user/login", handlers.Repo.PostShowLoginPage)
	mux.Get("/user/logout", handlers.Repo.Logout)

	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(auth(app.Session))

		mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/admin/dashboard", http.StatusFound)
		})
		mux.Get("/dashboard", handlers.Repo.AdminDashboard)
		mux.Get("/reservations/new", handlers.Repo.AdminNewReservations)
		mux.Get("/reservations/all", handlers.Repo.AdminAllReservations)
		mux.Get("/reservations/calendar", handlers.Repo.AdminCalendarReservations)
		mux.Get("/reservations/calendar/json", handlers.Repo.JsonAdminCalendarReservations)
		mux.Get("/reservations/details/{id}", handlers.Repo.AdminReservationSummary)
		mux.Post("/reservations/details/{id}", handlers.Repo.PostAdminReservationSummary)
		mux.Post("/reservations/processed", handlers.Repo.PostJsonAdminChangeResStatus)
		mux.Post("/reservations/delete", handlers.Repo.PostJsonAdminDeleteRes)
		mux.Get("/users/new", handlers.Repo.AdminCreateUser)
		mux.Post("/users/new", handlers.Repo.PostAdminCreateUser)
	})

	mux.NotFound(handlers.Repo.NotFound)

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

func auth(session *scs.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !helpers.IsAuthenticated(r) {
				session.Put(r.Context(), "error", "Not logged in")
				http.Redirect(w, r, "/user/login", http.StatusSeeOther)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
