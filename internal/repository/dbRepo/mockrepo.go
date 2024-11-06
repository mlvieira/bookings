package dbrepo

import (
	"errors"
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
