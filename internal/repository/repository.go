package repository

import (
	"time"

	"github.com/mlvieira/bookings/internal/models"
)

type DatabaseRepo interface {
	InsertReservation(res models.Reservation) (int, error)
	InsertRoomRestriction(res models.RoomRestriction) error
	SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error)
	SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error)
	GetRoomByID(id int) (models.Room, error)
	GetRoomByUrl(url string) (models.Room, error)
	GetUserByID(id int) (models.User, error)
	UpdateUser(user models.User) error
	Authenticate(email, testPassword string) (models.User, error)
	AllReservations(start, end *time.Time) ([]models.Reservation, error)
	AllNewReservations() ([]models.Reservation, error)
	GetReservationById(id int) (models.Reservation, error)
	UpdateReservation(res models.Reservation) error
	DeleteReservation(id int) error
	UpdateProcessedForReservation(id, processed int) error
	GetAllRooms(limit int) ([]models.Room, error)
	CreateUser(user models.User) (int, error)
	ListUsers() ([]models.User, error)
	DeleteUser(id int) error
}
