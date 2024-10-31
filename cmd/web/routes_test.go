package main

import (
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mlvieira/bookings/internal/config"
)

func TestRoutes(t *testing.T) {
	var app config.AppConfig

	mux := routes(&app)

	if _, ok := mux.(*chi.Mux); !ok {
		t.Errorf("type is not *chi.Mux. got %T", mux)
	}

}
