package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mlvieira/bookings/internal/config"
	"github.com/mlvieira/bookings/internal/forms"
	"github.com/mlvieira/bookings/internal/models"
	"github.com/mlvieira/bookings/internal/render"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Landing handles the GET request for the landing page
func (m *Repository) Landing(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "landing.page.html", &models.TemplateData{})
}

// Contact handles the GET request for the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "contact.page.html", &models.TemplateData{})
}

// RoomMajors handles the GET request for Major's Suite room page
func (m *Repository) RoomMajors(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["room_name"] = "Major's Suite"
	stringMap["image_path"] = "/static/images/marjors-suite.png"
	stringMap["room_url"] = "majors-suite"

	render.RenderTemplate(w, r, "rooms.page.html", &models.TemplateData{
		StringMap: stringMap,
	})
}

// RoomGenerals handles the GET request for General's Quarters room page
func (m *Repository) RoomGenerals(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["room_name"] = "General's quarter"
	stringMap["image_path"] = "/static/images/generals-quarters.png"
	stringMap["room_url"] = "generals-quarter"

	render.RenderTemplate(w, r, "rooms.page.html", &models.TemplateData{
		StringMap: stringMap,
	})
}

// Availability handles the GET request for the availability page
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "availability.page.html", &models.TemplateData{})
}

// jsonResponse defines the structure of a JSON response with status and message
type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

// AvailabilityJSON handles the POST request and returns a JSON response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	resp := jsonResponse{
		OK:      true,
		Message: "Available",
	}

	out, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// PostAvailability handles the POST request to search if room is available
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start_date")
	end := r.Form.Get("end_date")

	w.Write([]byte(fmt.Sprintf("Posted. Start date %s and End date is %s", start, end)))
}

// Booking handles the GET request for booking form
func (m *Repository) Booking(w http.ResponseWriter, r *http.Request) {
	var emptyReservation models.Reservation
	data := make(map[string]any)

	data["reservation"] = emptyReservation

	render.RenderTemplate(w, r, "reservation.page.html", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

// PostBooking handles the POST request for booking
func (m *Repository) PostBooking(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
	}

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email", "phone")
	form.MinLength("first_name", 3, r)
	form.MinLength("last_name", 3, r)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]any)
		data["reservation"] = reservation

		render.RenderTemplate(w, r, "reservation.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})

		return
	}

	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/book/summary", http.StatusSeeOther)
}

// ReservationSummary handles the GET request with the data from the reservation sent
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		log.Println("Error getting session value")
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]any)
	data["reservation"] = reservation

	render.RenderTemplate(w, r, "reservation-summary.page.html", &models.TemplateData{
		Data: data,
	})
}

// About handles the GET request for the about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {

	stringMap := make(map[string]string)
	stringMap["test"] = "Hello, again."

	remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIP

	render.RenderTemplate(w, r, "about.page.html", &models.TemplateData{
		StringMap: stringMap,
	})
}
