package dbrepo

import (
	"errors"
	"fmt"
	"time"

	"github.com/mlvieira/bookings/internal/models"
)

// InsertReservation inserts a reservation into the database
func (m *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	if res.Email == "john@at.com" {
		return 0, errors.New("err")
	}

	return 1, nil
}

// InsertRoomRestriction inserts a room restriction in the database
func (m *testDBRepo) InsertRoomRestriction(res models.RoomRestriction) error {
	if res.RoomID == 404 {
		return errors.New("err")
	}
	return nil
}

// SearchAvailabilityByDatesByRoomID returns true if availability exists for roomID and false if no availability
func (m *testDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	if roomID == 404 {
		return false, errors.New("err")
	}

	t := time.Date(2050, 12, 17, 0, 0, 0, 0, &time.Location{})

	return t.Equal(start), nil
}

// SearchAvailabilityForAllRooms return a slice of available rooms, if any, for given date range
func (m *testDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	var rooms []models.Room

	ta := time.Date(2050, 12, 17, 0, 0, 0, 0, &time.Location{})
	if ta.Equal(start) {
		return rooms, errors.New("err")
	}

	te := time.Date(2050, 12, 18, 0, 0, 0, 0, &time.Location{})
	if te.Equal(end) {
		return rooms, nil
	}

	room := models.Room{
		ID: 1,
	}
	rooms = append(rooms, room)

	return rooms, nil
}

// GetRoomByID gets a room by id
func (m *testDBRepo) GetRoomByID(id int) (models.Room, error) {
	var room models.Room
	if id > 2 {
		return room, errors.New("err")
	}
	return room, nil
}

// GetRoomByURL gets a room by url path
func (m *testDBRepo) GetRoomByUrl(url string) (models.Room, error) {
	var room models.Room
	if url != "majors-suite" {
		return room, errors.New("err")
	}

	return room, nil
}

// GetUserByID fetch user information by ID
func (m *testDBRepo) GetUserByID(id int) (models.User, error) {
	var u models.User

	return u, nil
}

// UpdateUser updates user information in the database
func (m *testDBRepo) UpdateUser(user models.User) error {
	return nil
}

// Authenticate authenticates a user
func (m *testDBRepo) Authenticate(email, testPassword string) (models.User, error) {
	var user models.User
	if email != "john@example.com" {
		return user, errors.New("err")
	}

	return user, nil
}

// TODO
func (m *testDBRepo) AllReservations(start, end *time.Time) ([]models.Reservation, error) {
	if start != nil && start.IsZero() {
		return nil, fmt.Errorf("simulated database error")
	}

	var reservations []models.Reservation

	defaultStart := time.Now()
	defaultEnd := time.Now().AddDate(0, 0, 7)

	if start == nil {
		start = &defaultStart
	}
	if end == nil {
		end = &defaultEnd
	}

	reservations = []models.Reservation{
		{
			ID:        1,
			FirstName: "John",
			LastName:  "Doe",
			Email:     "johndoe@example.com",
			Phone:     "123-456-7890",
			StartDate: *start,
			EndDate:   *end,
			Room: models.Room{
				ID:       1,
				RoomName: "Test",
			},
			UpdatedAt: time.Now(),
		},
	}

	return reservations, nil
}

func (m *testDBRepo) AllNewReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation

	return reservations, nil
}

func (m *testDBRepo) GetReservationById(id int) (models.Reservation, error) {
	var reservation models.Reservation

	return reservation, nil
}

func (m *testDBRepo) UpdateReservation(res models.Reservation) error {
	return nil
}

func (m *testDBRepo) DeleteReservation(id int) error {

	return nil
}

func (m *testDBRepo) UpdateProcessedForReservation(id, processed int) error {
	return nil
}

func (m *testDBRepo) GetAllRooms(limit int) ([]models.Room, error) {
	var rooms []models.Room

	return rooms, nil
}

func (m *testDBRepo) CreateUser(user models.User) (int, error) {
	return 1, nil
}

func (m *testDBRepo) ListUsers() ([]models.User, error) {
	var users []models.User
	return users, nil
}

func (m *testDBRepo) DeleteUser(id int) error {
	return nil
}
