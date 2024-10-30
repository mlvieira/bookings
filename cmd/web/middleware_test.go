package main

import (
	"net/http"
	"testing"
)

func TestNoSurf(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// just a dummy to wrap
	})

	h := NoSurf(&handler)
	switch v := h.(type) {
	case http.Handler:
	default:
		t.Errorf("type is not Http.Handler. its %s", v)
	}
}

func TestSessionLoad(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// just a dummy to wrap
	})

	h := SessionLoad(&handler)
	switch v := h.(type) {
	case http.Handler:
	default:
		t.Errorf("type is not Http.Handler. its %T", v)
	}
}
