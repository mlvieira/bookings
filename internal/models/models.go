package models

import (
	"time"
)

// User create struct for handling user data
type User struct {
	ID          int
	FirstName   string
	LastName    string
	Email       string
	Password    string
	AccessLevel int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Room create struct for handling room data
type Room struct {
	ID              int
	RoomName        string
	RoomDescription string
	RoomURL         string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Restriction create struct for handling restriction data
type Restriction struct {
	ID              int
	RestrictionName string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Reservation create struct for handling reservation data
type Reservation struct {
	ID        int
	FirstName string
	LastName  string
	Email     string
	Phone     string
	StartDate time.Time
	EndDate   time.Time
	RoomID    int
	CreatedAt time.Time
	UpdatedAt time.Time
	Room      Room
	Processed int
}

// RoomRestriction create struct for handling room restriction data
type RoomRestriction struct {
	ID            int
	StartDate     time.Time
	EndDate       time.Time
	RoomID        int
	ReservationID int
	RestrictionID int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Room          Room
	Reservation   Reservation
	Restriction   Restriction
}

// MailData holds an email message
type MailData struct {
	To       string
	From     string
	Subject  string
	Content  string
	Template string
}

type CalendarResponse struct {
	ID            string         `json:"id"`
	Title         string         `json:"title"`
	Start         time.Time      `json:"start"`
	End           time.Time      `json:"end"`
	AllDay        bool           `json:"allDay"`
	Url           string         `json:"url"`
	LastUpdated   time.Time      `json:"lastUpdated"`
	Editable      bool           `json:"editable"`
	ExtendedProps map[string]any `json:"extendedProps"`
}
