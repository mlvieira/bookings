package main

import (
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mlvieira/bookings/internal/config"
)

func TestRoutes(t *testing.T) {
	var app config.AppConfig

	mux := routes(&app)

	switch v := mux.(type) {
	case *chi.Mux:
	default:
		t.Errorf("type is not *chi.Mux. its %T", v)
	}
}
