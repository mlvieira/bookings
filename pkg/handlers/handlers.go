package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mlvieira/bookings/pkg/config"
	"github.com/mlvieira/bookings/pkg/models"
	"github.com/mlvieira/bookings/pkg/render"
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

func (m *Repository) Landing(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "landing.page.html", &models.TemplateData{})
}

func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "contact.page.html", &models.TemplateData{})
}

func (m *Repository) RoomMajors(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["room_name"] = "Major's Suite"
	stringMap["image_path"] = "/static/images/marjors-suite.png"
	stringMap["room_url"] = "majors-suite"

	render.RenderTemplate(w, r, "rooms.page.html", &models.TemplateData{
		StringMap: stringMap,
	})
}

func (m *Repository) RoomGenerals(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["room_name"] = "General's quarter"
	stringMap["image_path"] = "/static/images/generals-quarters.png"
	stringMap["room_url"] = "generals-quarter"

	render.RenderTemplate(w, r, "rooms.page.html", &models.TemplateData{
		StringMap: stringMap,
	})
}

func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "availability.page.html", &models.TemplateData{})
}

type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

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

func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start_date")
	end := r.Form.Get("end_date")

	w.Write([]byte(fmt.Sprintf("Posted. Start date %s and End date is %s", start, end)))
}

func (m *Repository) Booking(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "reservation.page.html", &models.TemplateData{})
}

// About is the about page handler
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {

	stringMap := make(map[string]string)
	stringMap["test"] = "Hello, again."

	remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIP

	render.RenderTemplate(w, r, "about.page.html", &models.TemplateData{
		StringMap: stringMap,
	})
}
