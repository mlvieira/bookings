package dbrepo

import (
	"context"
	"errors"
	"time"

	"github.com/mlvieira/bookings/internal/models"
	"golang.org/x/crypto/bcrypt"
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

// GetRoomByURL gets a room by url path
func (m *mysqlDBRepo) GetRoomByUrl(url string) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room models.Room

	stmt, err := m.DB.Prepare(`
				SELECT
					id
					, room_name
					, room_description
					, room_url
				FROM
					rooms
				WHERE
					room_url = ?
			`)
	if err != nil {
		return room, err
	}

	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, url)
	err = row.Scan(&room.ID, &room.RoomName, &room.RoomDescription, &room.RoomURL)
	if err != nil {
		return room, err
	}

	return room, nil
}

// GetUserByID fetch user information by ID
func (m *mysqlDBRepo) GetUserByID(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var u models.User

	stmt, err := m.DB.Prepare(`
				SELECT
					id
					, first_name 
					, last_name
					, email
					, password
					, access_level
					, created_at
					, updated_at
				FROM
					users
				WHERE
					id = ?
			`)
	if err != nil {
		return u, err
	}

	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, id)
	err = row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.AccessLevel,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return u, err
	}

	return u, nil

}

// UpdateUser updates user information in the database
func (m *mysqlDBRepo) UpdateUser(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
				UPDATE
					users
				SET
					first_name = ?
					, last_name = ?
					, email = ?
					, access_level = ?
					, updated_at = ?
					, password = ?
				WHERE
					id = ?
			`)
	if err != nil {
		tx.Rollback()
		return err
	}

	defer stmt.Close()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 8)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = stmt.ExecContext(ctx,
		user.FirstName,
		user.LastName,
		user.Email,
		user.AccessLevel,
		time.Now(),
		hashedPassword,
		user.ID,
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

// Authenticate authenticates a user
func (m *mysqlDBRepo) Authenticate(email, testPassword string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var hashedPassword string
	var user models.User

	stmt, err := m.DB.Prepare(`
		SELECT
			id
			, password
			, access_level
		FROM
			users
		WHERE
			email = ?
	`)
	if err != nil {
		return user, err
	}

	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, email)
	err = row.Scan(
		&user.ID,
		&hashedPassword,
		&user.AccessLevel,
	)
	if err != nil {
		return user, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return user, errors.New("incorrect password")
	} else if err != nil {
		return user, err
	}

	return user, nil
}

// AllReservations returns a slice of all reservations. or a slice of all reservation during a time frame
func (m *mysqlDBRepo) AllReservations(start, end *time.Time) ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
		SELECT
			r.id
			, r.first_name
			, r.last_name
			, r.email
			, r.phone
			, r.start_date
			, r.end_date
			, r.processed
			, r.room_id
			, r.created_at
			, r.updated_at
			, rm.id
			, rm.room_name
		FROM
			reservations r
		LEFT JOIN
			rooms rm ON r.room_id = rm.id
		WHERE 1=1
	`

	args := []interface{}{}
	if start != nil && end != nil {
		query += " AND r.end_date > ? AND r.start_date < ?"
		args = append(args, *start, *end)
	}

	query += " ORDER BY r.start_date ASC"

	stmt, err := m.DB.Prepare(query)
	if err != nil {
		return reservations, err
	}

	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return reservations, err
	}

	for rows.Next() {
		var i models.Reservation
		err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.Processed,
			&i.RoomID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Room.ID,
			&i.Room.RoomName,
		)
		if err != nil {
			return reservations, err
		}

		reservations = append(reservations, i)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

// AllNewReservations returns a slice of all new reservations
func (m *mysqlDBRepo) AllNewReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	stmt, err := m.DB.Prepare(`
		SELECT
			r.id
			, r.first_name
			, r.last_name
			, r.email
			, r.phone
			, r.start_date
			, r.end_date
			, r.room_id
			, r.created_at
			, r.updated_at
			, rm.id
			, rm.room_name
		FROM
			reservations r
		LEFT JOIN
			rooms rm on r.room_id = rm.id
		WHERE
			r.processed = 0
		ORDER BY r.start_date asc
	`)
	if err != nil {
		return reservations, err
	}

	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return reservations, err
	}

	for rows.Next() {
		var i models.Reservation
		err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.RoomID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Room.ID,
			&i.Room.RoomName,
		)
		if err != nil {
			return reservations, err
		}

		reservations = append(reservations, i)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

// GetReservationById return reservation associated by ID
func (m *mysqlDBRepo) GetReservationById(id int) (models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservation models.Reservation

	stmt, err := m.DB.Prepare(`
		SELECT
			r.id
			, r.first_name
			, r.last_name
			, r.email
			, r.phone
			, r.start_date
			, r.end_date
			, r.room_id
			, r.created_at
			, r.updated_at
			, r.processed
			, rm.id
			, rm.room_name
		FROM
			reservations r
		LEFT JOIN rooms rm ON r.room_id = rm.id
		WHERE
			r.id = ?
	`)
	if err != nil {
		return reservation, err
	}

	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, id)
	err = row.Scan(
		&reservation.ID,
		&reservation.FirstName,
		&reservation.LastName,
		&reservation.Email,
		&reservation.Phone,
		&reservation.StartDate,
		&reservation.EndDate,
		&reservation.RoomID,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
		&reservation.Processed,
		&reservation.Room.ID,
		&reservation.Room.RoomName,
	)
	if err != nil {
		return reservation, err
	}

	return reservation, nil
}

// UpdateReservation updates user information in the database
func (m *mysqlDBRepo) UpdateReservation(res models.Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
				UPDATE
					reservations
				SET
					first_name = ?
					, last_name = ?
					, email = ?
					, phone = ?
					, updated_at = ?
				WHERE
					id = ?
			`)
	if err != nil {
		tx.Rollback()
		return err
	}

	defer stmt.Close()

	_, err = stmt.ExecContext(ctx,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		time.Now(),
		res.ID,
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

// DeleteReservation deletes reservation by id
func (m *mysqlDBRepo) DeleteReservation(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
				DELETE FROM	
					reservations
				WHERE
					id = ?
			`)
	if err != nil {
		tx.Rollback()
		return err
	}

	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

// UpdateProcessedForReservation updates processed for a reservation ID
func (m *mysqlDBRepo) UpdateProcessedForReservation(id, processed int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
				UPDATE	
					reservations
				SET
					processed = ?
					, updated_at = ?
				WHERE
					id = ?
			`)
	if err != nil {
		tx.Rollback()
		return err
	}

	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, processed, time.Now(), id)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

// GetAllRooms gets all rooms registred
func (m *mysqlDBRepo) GetAllRooms(limit int) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	stmt, err := m.DB.Prepare(`
				SELECT
					id
					, room_name
					, room_description
					, room_url
				FROM
					rooms
				LIMIT ?
			`)
	if err != nil {
		return rooms, err
	}

	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, limit)
	if err != nil {
		return rooms, err
	}

	for rows.Next() {
		var r models.Room
		err := rows.Scan(
			&r.ID,
			&r.RoomName,
			&r.RoomDescription,
			&r.RoomURL,
		)
		if err != nil {
			return rooms, err
		}

		rooms = append(rooms, r)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

// CreateUser creates a user
func (m *mysqlDBRepo) CreateUser(user models.User) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.Begin()
	if err != nil {
		return 0, err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO
			users
			(first_name, last_name, email, password,
			access_level, created_at, updated_at)
		VALUES
			(?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	defer stmt.Close()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 8)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	ret, err := stmt.ExecContext(ctx,
		user.FirstName,
		user.LastName,
		user.Email,
		hashedPassword,
		user.AccessLevel,
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

func (m *mysqlDBRepo) ListUsers() ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var users []models.User

	stmt, err := m.DB.Prepare(`
		SELECT
			id
			, first_name
			, last_name
			, email
			, access_level
		FROM
			users
	`)
	if err != nil {
		return users, err
	}

	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return users, err
	}

	for rows.Next() {
		var u models.User
		err := rows.Scan(
			&u.ID,
			&u.FirstName,
			&u.LastName,
			&u.Email,
			&u.AccessLevel,
		)
		if err != nil {
			return users, err
		}

		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return users, err
	}

	return users, nil
}

// DeleteUser deletes user by id
func (m *mysqlDBRepo) DeleteUser(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
				DELETE FROM	
					users
				WHERE
					id = ?
			`)
	if err != nil {
		tx.Rollback()
		return err
	}

	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
