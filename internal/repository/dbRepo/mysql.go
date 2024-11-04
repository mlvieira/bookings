package dbrepo

import (
	"context"
	"time"

	"github.com/mlvieira/bookings/internal/models"
)

// InsertReservation inserts a reservation into the database
func (m *mysqlDBRepo) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.Begin()
	if err != nil {
		return 0, err
	}

	stmt, err := tx.Prepare(`
				INSERT INTO
					reservations 
					(first_name, last_name, email, phone, start_date,
					end_date, room_id, created_at, updated_at) 
				VALUES
					(?, ?, ?, ?, ?, ?, ?, ?, ?)
				`)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	defer stmt.Close()

	ret, err := stmt.ExecContext(ctx,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	lastID, _ := ret.LastInsertId()

	return int(lastID), nil
}

// InsertRoomRestriction inserts a room restriction in the database
func (m *mysqlDBRepo) InsertRoomRestriction(res models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
				INSERT INTO
					room_restrictions
					(start_date, end_date, room_id, reservation_id, created_at, updated_at, restriction_id)
				VALUES 
					(?, ?, ?, ?, ?, ?, ?)
				`)
	if err != nil {
		tx.Rollback()
		return err
	}

	defer stmt.Close()

	_, err = stmt.ExecContext(ctx,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		res.ReservationID,
		time.Now(),
		time.Now(),
		res.RestrictionID,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

// SearchAvailabilityByDatesByRoomID returns true if availability exists for roomID and false if no availability
func (m *mysqlDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var numRows int

	stmt, err := m.DB.Prepare(`
				SELECT
					count(id)
				FROM
					room_restrictions
				WHERE 1=1
				AND	room_id = ?
				AND ? < end_date
				AND ? > start_date
				`)
	if err != nil {
		return false, err
	}

	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, roomID, start, end)
	err = row.Scan(&numRows)
	if err != nil {
		return false, nil
	}

	return numRows == 0, nil
}

// SearchAvailabilityForAllRooms return a slice of available rooms, if any, for given date range
func (m *mysqlDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	stmt, err := m.DB.Prepare(`
				SELECT
					r.id
					, r.room_name
					, r.room_description
					, r.room_url
				FROM
					rooms r
				WHERE
					r.id NOT IN (
						SELECT
							rr.room_id
						FROM
							room_restrictions rr
						WHERE 1=1
						AND ? < rr.end_date
						AND ? > rr.start_date
					)
				`)
	if err != nil {
		return rooms, err
	}

	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, start, end)
	if err != nil {
		return rooms, err
	}

	for rows.Next() {
		var room models.Room
		err := rows.Scan(&room.ID, &room.RoomName, &room.RoomDescription, &room.RoomURL)
		if err != nil {
			return rooms, err
		}

		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil

}

// GetRoomByID gets a room by id
func (m *mysqlDBRepo) GetRoomByID(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room models.Room

	stmt, err := m.DB.Prepare(`
				SELECT
					id
					, room_name
				FROM
					rooms
				WHERE
					id = ?
			`)
	if err != nil {
		return room, err
	}

	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, id)
	err = row.Scan(&room.ID, &room.RoomName)
	if err != nil {
		return room, err
	}

	return room, nil
}
