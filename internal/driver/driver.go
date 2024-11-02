package driver

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// DB holds the database connection pool
type DB struct {
	SQL *sql.DB
}

var dbConn = &DB{}

const maxOpenDbConn = 10
const maxIdleDbConn = 5
const maxDbLifetime = 5 * time.Minute

// ConnectSQL creates database pool
func ConnectSQL(dsn string) (*DB, error) {
	db, err := NewDatabase(dsn)
	if err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(maxOpenDbConn)
	db.SetConnMaxIdleTime(maxIdleDbConn)
	db.SetConnMaxLifetime(maxDbLifetime)

	dbConn.SQL = db

	if err = testDB(db); err != nil {
		return nil, err
	}

	return dbConn, nil
}

// testDB tries to ping database
func testDB(db *sql.DB) error {
	if err := db.Ping(); err != nil {
		return err
	}

	return nil
}

// NewDatabase create a new database for the application
func NewDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = testDB(db); err != nil {
		return nil, err
	}

	return db, nil
}
