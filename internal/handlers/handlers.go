package handlers

import (
	"encoding/json"
	"fmt"
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
	rooms, err := m.DB.GetAllRooms(6)
	if err != nil {
		helpers.ServerError(w, err)
	}

	data := make(map[string]any)
	data["rooms"] = rooms

	render.Template(w, r, "landing.page.html", &models.TemplateData{
		Data: data,
	})
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
		m.App.Session.Put(r.Context(), "error", "Error parsing form")
		http.Redirect(w, r, "/availability", http.StatusSeeOther)
		return
	}

	start := r.FormValue("start_date")
	end := r.FormValue("end_date")

	if start == "" || end == "" {
		m.App.Session.Put(r.Context(), "error", "Dates cannot be empty")
		http.Redirect(w, r, "/availability", http.StatusSeeOther)
		return
	}

	layout := "01-02-2006"

	startDate, err := time.Parse(layout, start)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Error parsing dates")
		http.Redirect(w, r, "/availability", http.StatusSeeOther)
		return
	}

	endDate, err := time.Parse(layout, end)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Error parsing dates")
		http.Redirect(w, r, "/availability", http.StatusSeeOther)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Error searching database")
		http.Redirect(w, r, "/availability", http.StatusSeeOther)
		return
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
		m.App.Session.Put(r.Context(), "error", "Error getting reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	room, err := m.DB.GetRoomByID(res.RoomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Room not found")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.Room.RoomName = room.RoomName

	m.App.Session.Put(r.Context(), "reservation", res)

	data := make(map[string]any)
	data["reservation"] = res

	render.Template(w, r, "reservation.page.html", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

// PostBooking handles the POST request for booking
func (m *Repository) PostBooking(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Error getting reservation from session")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Error parsing form")
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
		m.App.Session.Put(r.Context(), "error", "Error inserting reservation in the database")
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
		m.App.Session.Put(r.Context(), "error", "Error inserting room restriction in the database")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	htmlMsg := fmt.Sprintf(`
		<strong>Reservation Confirmation</strong><br/>
		Dear %s, <br>
		This is a confirmation of your reservation from %s to %s for the room %s.
	`, reservation.FirstName, reservation.StartDate.Format("01-02-2006"), reservation.EndDate.Format("01-02-2006"), reservation.Room.RoomName)

	msg := models.MailData{
		To:       reservation.Email,
		From:     "noreply@bookings.com",
		Subject:  "Your Reservation is Confirmed! 🎉",
		Content:  htmlMsg,
		Template: "confirmation.html",
	}

	m.App.MailChan <- msg

	htmlMsg = fmt.Sprintf(`
		<strong>Your room has been booked</strong><br/>
		We're here to tell you great news!
		Your room %s has been booked from %s to %s.
	`, reservation.Room.RoomName, reservation.StartDate.Format("01-02-2006"), reservation.EndDate.Format("01-02-2006"))

	msg = models.MailData{
		To:       "john@realstate",
		From:     "noreply@bookings.com",
		Subject:  fmt.Sprintf("Your room %s has been booked! 🎉", reservation.Room.RoomName),
		Content:  htmlMsg,
		Template: "confirmation.html",
	}

	m.App.MailChan <- msg

	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/book/summary", http.StatusSeeOther)
}

// ReservationSummary handles the GET request with the data from the reservation sent
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Error getting reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]any)

	data["reservation"] = reservation

	render.Template(w, r, "reservation-summary.page.html", &models.TemplateData{
		Data: data,
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
	roomIDstr := chi.URLParam(r, "id")
	roomID, err := strconv.Atoi(roomIDstr)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Invalid room id")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Error getting reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.RoomID = roomID

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/book", http.StatusSeeOther)
}

func (m *Repository) NotFound(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "not-found.page.html", &models.TemplateData{})
}

func (m *Repository) ShowLoginPage(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.html", &models.TemplateData{
		Form: forms.New(nil),
	})
}

func (m *Repository) PostShowLoginPage(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Error parsing form")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	if email == "" || password == "" {
		m.App.Session.Put(r.Context(), "error", "Email or Password cannot be empty")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")
	if !form.Valid() {
		http.Error(w, "error", http.StatusSeeOther)
		render.Template(w, r, "login.page.html", &models.TemplateData{
			Form: form,
		})
		return
	}

	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Invalid login credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-dashboard.page.html", &models.TemplateData{})
}

func (m *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllNewReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]any)
	data["reservations"] = reservations

	render.Template(w, r, "admin-new-reservations.page.html", &models.TemplateData{
		Data: data,
	})
}

func (m *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]any)
	data["reservations"] = reservations

	render.Template(w, r, "admin-all-reservations.page.html", &models.TemplateData{
		Data: data,
	})
}

func (m *Repository) JsonAdminCalendarReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllReservations()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Internal Server error")
		http.Redirect(w, r, "/admin/dashboard", http.StatusTemporaryRedirect)
		return
	}

	var calendarResponses []models.CalendarResponse

	for _, res := range reservations {
		calendarResponse := models.CalendarResponse{
			ID:       fmt.Sprintf("%d", res.ID),
			Title:    fmt.Sprintf("%s room reservation", res.Room.RoomName),
			Start:    res.StartDate,
			End:      res.EndDate,
			AllDay:   true,
			Url:      fmt.Sprintf("/admin/reservations/details/%d", res.ID),
			Editable: false,
			ExtendedProps: map[string]any{
				"name":        fmt.Sprintf("%s %s", res.FirstName, res.LastName),
				"room":        res.Room.RoomName,
				"lastUpdated": res.UpdatedAt,
			},
		}
		calendarResponses = append(calendarResponses, calendarResponse)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(calendarResponses); err != nil {
		helpers.ServerError(w, err)
	}

}

func (m *Repository) AdminCalendarReservations(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-calendar-reservations.page.html", &models.TemplateData{})
}

func (m *Repository) AdminReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservationIDstr := chi.URLParam(r, "id")
	resStr, err := strconv.Atoi(reservationIDstr)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Invalid reservation id")
		http.Redirect(w, r, "/admin/dashboard", http.StatusTemporaryRedirect)
		return
	}

	res, err := m.DB.GetReservationById(resStr)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Reservation not found")
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	data := make(map[string]any)
	data["reservation"] = res

	render.Template(w, r, "admin-reservations-summary.page.html", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

func (m *Repository) PostAdminReservationSummary(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservationIDstr := chi.URLParam(r, "id")
	resStr, err := strconv.Atoi(reservationIDstr)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Invalid reservation id")
		http.Redirect(w, r, "/admin/dashboard", http.StatusTemporaryRedirect)
		return
	}

	res, err := m.DB.GetReservationById(resStr)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Reservation not found")
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email", "phone")
	form.MinLength("first_name", 3)
	form.MinLength("last_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		m.App.Session.Put(r.Context(), "error", "Invalid form values")
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations/details/%d", resStr), http.StatusSeeOther)
		return
	}

	err = m.DB.UpdateReservation(res)
	if err != nil {
		helpers.ServerError(w, err)
	}

	m.App.Session.Put(r.Context(), "flash", "Reservation updated successfully")

	http.Redirect(w, r, fmt.Sprintf("/admin/reservations/details/%d", resStr), http.StatusSeeOther)
}

type payloadStatus struct {
	ID string `json:"id"`
}

func (m *Repository) PostJsonAdminChangeResStatus(w http.ResponseWriter, r *http.Request) {
	var payload payloadStatus

	err := json.NewDecoder(r.Body).Decode(&payload)
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

	id, err := strconv.Atoi(payload.ID)
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Invalid reservation ID",
		}
		out, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	err = m.DB.UpdateProcessedForReservation(id, 1)
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Error updating database",
		}
		out, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	resp := jsonResponse{
		OK:      true,
		Message: "Reservation status has been changed!",
	}

	out, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (m *Repository) PostJsonAdminDeleteRes(w http.ResponseWriter, r *http.Request) {
	var payload payloadStatus

	err := json.NewDecoder(r.Body).Decode(&payload)
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

	id, err := strconv.Atoi(payload.ID)
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Invalid reservation ID",
		}
		out, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	err = m.DB.DeleteReservation(id)
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Error updating database",
		}
		out, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	resp := jsonResponse{
		OK:      true,
		Message: "Reservation has been deleted!",
	}

	out, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}
