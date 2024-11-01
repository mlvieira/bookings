package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexedwards/scs/v2"
	"github.com/justinas/nosurf"
	"github.com/mlvieira/bookings/internal/config"
)

func TestRoutes(t *testing.T) {
	app := &config.AppConfig{
		InProduction: false,
	}

	mux := Routes(app)

	if mux == nil {
		t.Fatal("Routes() returned nil, expected a valid http.Handler")
	}

	_, ok := mux.(http.Handler)
	if !ok {
		t.Error("Routes() does not implement http.Handler")
	}

}

func TestSessionLoad(t *testing.T) {
	session := scs.New()

	app := &config.AppConfig{
		InProduction: false,
		Session:      session,
	}

	setHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.Session.Put(r.Context(), "message", "Hello, World!")
		w.WriteHeader(http.StatusOK)
	})

	getHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		message := app.Session.GetString(r.Context(), "message")
		if message != "Hello, World!" {
			t.Errorf("Expected 'Hello, World!', got '%s'", message)
		}
		w.WriteHeader(http.StatusOK)
	})

	wrappedSetHandler := sessionLoad(session)(setHandler)
	wrappedGetHandler := sessionLoad(session)(getHandler)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/set", nil)
	if err != nil {
		t.Fatal(err)
	}
	wrappedSetHandler.ServeHTTP(rr, req)

	cookies := rr.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("Expected a session cookie to be set")
	}
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "session" {
			sessionCookie = cookie
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("Session cookie not found")
	}

	rr2 := httptest.NewRecorder()
	req2, err := http.NewRequest("GET", "/get", nil)
	if err != nil {
		t.Fatal(err)
	}
	req2.AddCookie(sessionCookie)
	wrappedGetHandler.ServeHTTP(rr2, req2)
}

func TestNoSurf(t *testing.T) {
	app := &config.AppConfig{
		InProduction: false,
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := nosurf.Token(r)

		w.Header().Set("X-CSRF-Token", token)
		w.Write([]byte("OK"))
	})

	wrappedHandler := noSurf(app)(testHandler)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	wrappedHandler.ServeHTTP(rr, req)

	var csrfCookie *http.Cookie
	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == nosurf.CookieName {
			csrfCookie = cookie
			break
		}
	}

	if csrfCookie == nil {
		t.Fatal("CSRF cookie not found")
	}

	csrfToken := rr.Header().Get("X-CSRF-Token")
	if csrfToken == "" {
		t.Fatal("CSRF token is empty")
	}

	rr2 := httptest.NewRecorder()
	req2, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	wrappedHandler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusBadRequest && rr2.Code != http.StatusForbidden {
		t.Errorf("Expected status 400 or 403 when CSRF token is missing, got %d", rr2.Code)
	}

	rr3 := httptest.NewRecorder()
	req3, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req3.Header.Set("X-CSRF-Token", csrfToken)

	req3.AddCookie(csrfCookie)

	wrappedHandler.ServeHTTP(rr3, req3)
	if rr3.Code != http.StatusOK {
		t.Errorf("Expected status 200 when CSRF token is valid, got %d", rr3.Code)
	}
}
