package handlers

import (
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mlvieira/bookings/internal/config"
	"github.com/mlvieira/bookings/internal/helpers"
	"github.com/mlvieira/bookings/internal/models"
	"github.com/mlvieira/bookings/internal/render"
	dbrepo "github.com/mlvieira/bookings/internal/repository/dbRepo"
)

var app config.AppConfig

func TestMain(m *testing.M) {
	if err := os.Chdir("../.."); err != nil {
		log.Fatal("Could not change working directory:", err)
	}

	gob.Register(models.Reservation{})

	app = *config.SetupAppConfig(false)

	defer close(app.MailChan)

	listenForMail()

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal(err)
	}

	app.TemplateCache = tc
	app.UseCache = app.InProduction
	app.Port = ":8080"

	repo := newTestRepo(&app)
	NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	os.Exit(m.Run())
}

func getRoutes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(sessionLoad(app.Session))

	mux.Get("/", Repo.Landing)
	mux.Get("/about", Repo.About)
	mux.Get("/contact", Repo.Contact)
	mux.Get("/rooms/{room}", Repo.RoomsPage)
	mux.Get("/rooms/book/{id}", Repo.ChooseRoom)
	mux.Get("/availability", Repo.Availability)
	mux.Post("/availability", Repo.PostAvailability)
	mux.Post("/availability/json", Repo.AvailabilityJSON)
	mux.Get("/book", Repo.Booking)
	mux.Post("/book", Repo.PostBooking)
	mux.Get("/book/summary", Repo.ReservationSummary)
	mux.Get("/user/login", Repo.ShowLoginPage)
	mux.Post("/user/login", Repo.PostShowLoginPage)
	mux.Get("/user/logout", Repo.Logout)

	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(auth(app.Session))
		mux.Get("/dashboard", Repo.AdminDashboard)
		mux.Get("/reservations/new", Repo.AdminNewReservations)
		mux.Get("/reservations/all", Repo.AdminAllReservations)
		mux.Get("/reservations/calendar", Repo.AdminCalendarReservations)
		mux.Get("/reservations/{src}/{id}", Repo.AdminReservationSummary)
	})

	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	return mux
}

// sessionLoad loads and saves the session on every request
func sessionLoad(session *scs.SessionManager) func(http.Handler) http.Handler {
	return session.LoadAndSave
}

func newTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestRepo(a),
	}
}

func listenForMail() {
	go func() {
		for {
			_ = <-app.MailChan
		}
	}()
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
