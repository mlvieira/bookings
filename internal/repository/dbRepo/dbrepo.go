package dbrepo

import (
	"database/sql"

	"github.com/mlvieira/bookings/internal/config"
	"github.com/mlvieira/bookings/internal/repository"
)

type mysqlDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

func NewMysqlRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	return &mysqlDBRepo{
		App: a,
		DB:  conn,
	}
}
