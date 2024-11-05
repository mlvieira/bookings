package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mlvieira/bookings/internal/config"
	"github.com/mlvieira/bookings/internal/driver"
	"github.com/mlvieira/bookings/internal/forms"
	"github.com/mlvieira/bookings/internal/helpers"
	"github.com/mlvieira/bookings/internal/models"
	"github.com/mlvieira/bookings/internal/render"
	"github.com/mlvieira/bookings/internal/repository"
	dbrepo "github.com/mlvieira/bookings/internal/repository/dbRepo"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewMysqlRepo(db.SQL, a),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Landing handles the GET request for the landing page
func (m *Repository) Landing(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "landing.page.html", &models.TemplateData{})
}

// Contact handles the GET request for the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.html", &models.TemplateData{})
}

// RoomsPage handles the GET request for individual room page
func (m *Repository) RoomsPage(w http.ResponseWriter, r *http.Request) {
	room, err := m.DB.GetRoomByUrl(chi.URLParam(r, "room"))
	if err != nil {
		helpers.ClientError(w, http.StatusNotFound)
		return
	}

	data := make(map[string]any)
	data["room"] = room

	render.Template(w, r, "rooms.page.html", &models.TemplateData{
		Data: data,
	})
}

// Availability handles the GET request for the availability page
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "availability.page.html", &models.TemplateData{})
}

// jsonResponse defines the structure of a JSON response with status and message
type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

// AvailabilityJSON handles the POST request and returns a JSON response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Internal server error",
		}
		out, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	sd := r.FormValue("start_date")
	ed := r.FormValue("end_date")
	if sd == "" || ed == "" {
		resp := jsonResponse{
			OK:      false,
			Message: "Dates cannot be empty",
		}
		out, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	layout := "01-02-2006"

	startDate, _ := time.Parse(layout, sd)

	endDate, _ := time.Parse(layout, ed)

	roomID, _ := strconv.Atoi(r.FormValue("room_id"))

	available, err := m.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomID)
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Error searching in the database",
		}
		out, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	var msg string

	if available {
		msg = "Available"
	} else {
		msg = "Unavailable"
	}

	resp := jsonResponse{
		OK:      available,
		Message: msg,
	}

	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    roomID,
	}

	m.App.Session.Put(r.Context(), "reservation", res)

	out, _ := json.Marshal(resp)

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// PostAvailability handles the POST request to search if room is available
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
	}

	start := r.FormValue("start_date")
	end := r.FormValue("end_date")
	m.App.InfoLog.Println("ROOM:", start, end)

	if start == "" || end == "" {
		helpers.ClientError(w, http.StatusBadRequest)
	}

	layout := "01-02-2006"

	startDate, err := time.Parse(layout, start)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	endDate, err := time.Parse(layout, end)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	for _, room := range rooms {
		m.App.InfoLog.Println("ROOM:", room.ID, room.RoomName)
	}

	if len(rooms) == 0 {
		m.App.Session.Put(r.Context(), "error", "No availability")
		http.Redirect(w, r, "/availability", http.StatusSeeOther)
		return
	}

	data := make(map[string]any)
	data["rooms"] = rooms

	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	m.App.Session.Put(r.Context(), "reservation", res)

	render.Template(w, r, "choose-room.page.html", &models.TemplateData{
		Data: data,
	})
}

// Booking handles the GET request for booking form
func (m *Repository) Booking(w http.ResponseWriter, r *http.Request) {
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		err := errors.New("can't get reservation from session")
		m.App.Session.Put(r.Context(), "error", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	room, err := m.DB.GetRoomByID(res.RoomID)
	if err != nil {
		err := errors.New("can't find room")
		m.App.Session.Put(r.Context(), "error", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.Room.RoomName = room.RoomName

	m.App.Session.Put(r.Context(), "reservation", res)

	sd := res.StartDate.Format("01-02-2006")
	ed := res.EndDate.Format("01-02-2006")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	data := make(map[string]any)
	data["reservation"] = res

	render.Template(w, r, "reservation.page.html", &models.TemplateData{
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
}

// PostBooking handles the POST request for booking
func (m *Repository) PostBooking(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		err := errors.New("can't get reservation from session")
		m.App.Session.Put(r.Context(), "error", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	err := r.ParseForm()
	if err != nil {
		err := errors.New("can't parse form")
		m.App.Session.Put(r.Context(), "error", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Email = r.Form.Get("email")
	reservation.Phone = r.Form.Get("phone")

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email", "phone")
	form.MinLength("first_name", 3)
	form.MinLength("last_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]any)
		data["reservation"] = reservation
		http.Error(w, "error", http.StatusSeeOther)

		render.Template(w, r, "reservation.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})

		return
	}

	lastID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		err := errors.New("can't insert reservation in the database")
		m.App.Session.Put(r.Context(), "error", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        reservation.RoomID,
		ReservationID: lastID,
		RestrictionID: 1,
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		err := errors.New("can't insert room restriction in the database")
		m.App.Session.Put(r.Context(), "error", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/book/summary", http.StatusSeeOther)
}

// ReservationSummary handles the GET request with the data from the reservation sent
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		err := errors.New("can't get reservation from session")
		helpers.ServerError(w, err)
		m.App.Session.Put(r.Context(), "error", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Remove(r.Context(), "reservation")

	stringMap := make(map[string]string)
	data := make(map[string]any)

	data["reservation"] = reservation

	sd := reservation.StartDate.Format("01-02-2006")
	ed := reservation.EndDate.Format("01-02-2006")

	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	render.Template(w, r, "reservation-summary.page.html", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
}

// About handles the GET request for the about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {

	stringMap := make(map[string]string)
	stringMap["test"] = "Hello, again."

	remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIP

	render.Template(w, r, "about.page.html", &models.TemplateData{
		StringMap: stringMap,
	})
}

func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, errors.New("error getting value from the session"))
	}

	res.RoomID = roomID

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/book", http.StatusSeeOther)
}
