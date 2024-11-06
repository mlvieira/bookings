package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/mlvieira/bookings/internal/driver"
	"github.com/mlvieira/bookings/internal/models"
)

func mockDB() *driver.DB {
	return &driver.DB{
		SQL: nil,
	}
}

func getCtx(req *http.Request) context.Context {
	ctx, err := app.Session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}

	return ctx
}

func createTestReservation(roomID int, roomName string) models.Reservation {
	return models.Reservation{
		RoomID: roomID,
		Room: models.Room{
			ID:              roomID,
			RoomName:        roomName,
			RoomDescription: "Test",
			RoomURL:         "test-url",
		},
	}
}

func handleBookingRequest(
	t *testing.T,
	req *http.Request,
	useSession bool,
	expectedCode int,
	expectedLocation string,
	reservation models.Reservation,
	handler http.HandlerFunc,
) {
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	if useSession {
		app.Session.Put(ctx, "reservation", reservation)
	}

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	app.Session.Destroy(req.Context())

	const errMessage = "Handler returned wrong response code: got %d, wanted %d"
	if rr.Code != expectedCode {
		t.Errorf(errMessage, rr.Code, expectedCode)
	}

	if expectedLocation != "" {
		location := rr.Header().Get("Location")
		if location != expectedLocation {
			t.Errorf("Handler redirected to wrong URL: got %s, wanted %s", location, expectedLocation)
		}
	}
}

func TestNewRepo(t *testing.T) {
	db := mockDB()
	testRepo := NewRepo(&app, db)

	if testRepo.App != Repo.App {
		t.Errorf("expected app config to be %v, got %v", Repo.App, testRepo.App)
	}

	if testRepo.DB == nil {
		t.Error("expected DB repo to be initialized, got nil")
	}

	if reflect.TypeOf(testRepo).String() != "*handlers.Repository" {
		t.Errorf("Did not get correct type from NewRepo: got %s, wanted *Repository", reflect.TypeOf(testRepo).String())
	}

}

type testsDataType struct {
	name               string
	url                string
	expectedStatusCode int
}

var testsData = []testsDataType{
	{"Valid Home", "/", http.StatusOK},
	{"Valid About", "/about", http.StatusOK},
	{"Valid Room Page", "/rooms/majors-suite", http.StatusOK},
	{"Valid Contact", "/contact", http.StatusOK},
	{"Valid Availability GET", "/availability", http.StatusOK},
	{"Page not found", "/a", http.StatusNotFound},
	{"Room not found", "/rooms/a", http.StatusNotFound},
}

func TestHandlers(t *testing.T) {
	getRequest := func(t *testing.T, ts *httptest.Server, test testsDataType) {
		res, err := ts.Client().Get(ts.URL + test.url)
		if err != nil {
			t.Fatal(err)
		}

		if res.StatusCode != test.expectedStatusCode {
			t.Errorf("%s handler returned wrong status code: got %v want %v", test.url, res.StatusCode, test.expectedStatusCode)
		}
	}

	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)

	defer ts.Close()

	for _, test := range testsData {
		t.Run(test.name, func(t *testing.T) {
			getRequest(t, ts, test)
		})
	}
}

func TestRepository_Booking(t *testing.T) {
	executeBookingTest := func(
		t *testing.T,
		useSession bool,
		expectedCode int,
		expectedLocation string,
		reservation models.Reservation,
	) {
		req, err := http.NewRequest("GET", "/book", nil)
		if err != nil {
			t.Fatal(err)
		}

		handleBookingRequest(t, req, useSession, expectedCode, expectedLocation, reservation, http.HandlerFunc(Repo.Booking))
	}

	t.Run("Valid request", func(t *testing.T) {
		executeBookingTest(t, true, http.StatusOK, "", createTestReservation(1, "test"))
	})

	t.Run("Missing session", func(t *testing.T) {
		executeBookingTest(t, false, http.StatusTemporaryRedirect, "/", createTestReservation(1, "test"))
	})

	t.Run("Nonexistent room ID", func(t *testing.T) {
		executeBookingTest(t, true, http.StatusTemporaryRedirect, "/", createTestReservation(3, "test"))
	})
}

func TestRepository_PostBooking(t *testing.T) {
	executePostBookingTest := func(
		t *testing.T,
		useForm,
		useSession bool,
		expectedCode int,
		expectedLocation string,
		form url.Values,
		reservation models.Reservation,
	) {
		var body io.Reader = nil
		if useForm {
			body = strings.NewReader(form.Encode())
		}

		req, err := http.NewRequest("POST", "/book", body)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		handleBookingRequest(t, req, useSession, expectedCode, expectedLocation, reservation, http.HandlerFunc(Repo.PostBooking))
	}

	t.Run("Valid form", func(t *testing.T) {
		form := url.Values{}
		form.Add("first_name", "John")
		form.Add("last_name", "Doe")
		form.Add("email", "john@example.com")
		form.Add("phone", "55555555")
		executePostBookingTest(t, true, true, http.StatusSeeOther, "/book/summary", form, createTestReservation(1, "test"))
	})

	t.Run("Missing session", func(t *testing.T) {
		executePostBookingTest(t, false, false, http.StatusSeeOther, "/", nil, models.Reservation{})
	})

	t.Run("Invalid form", func(t *testing.T) {
		executePostBookingTest(t, false, true, http.StatusSeeOther, "/", nil, createTestReservation(1, "test"))
	})

	t.Run("Invalid email", func(t *testing.T) {
		form := url.Values{}
		form.Add("first_name", "John")
		form.Add("last_name", "Doe")
		form.Add("email", "john@example")
		form.Add("phone", "55555555")
		executePostBookingTest(t, true, true, http.StatusSeeOther, "", form, createTestReservation(1, "test"))
	})

	t.Run("Database error: InsertReservation", func(t *testing.T) {
		form := url.Values{}
		form.Add("first_name", "John")
		form.Add("last_name", "Doe")
		form.Add("email", "john@at.com")
		form.Add("phone", "55555555")
		executePostBookingTest(t, true, true, http.StatusSeeOther, "/", form, createTestReservation(1, "test"))
	})

	t.Run("Database error: InsertRoomRestriction", func(t *testing.T) {
		form := url.Values{}
		form.Add("first_name", "John")
		form.Add("last_name", "Doe")
		form.Add("email", "john@example.com")
		form.Add("phone", "55555555")
		executePostBookingTest(t, true, true, http.StatusSeeOther, "/", form, createTestReservation(404, "test"))
	})
}

func TestRepository_AvailabilityJSON(t *testing.T) {
	executeAvailabilityJSONTest := func(t *testing.T, useForm bool, expectedOK bool, expectedMessage string, startDate, endDate, roomID string) {
		var body io.Reader = nil
		if useForm {
			form := url.Values{}
			form.Add("start_date", startDate)
			form.Add("end_date", endDate)
			form.Add("room_id", roomID)
			body = strings.NewReader(form.Encode())
		}

		req, err := http.NewRequest("POST", "/availability/json", body)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		ctx := getCtx(req)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.AvailabilityJSON)
		handler.ServeHTTP(rr, req)

		var j jsonResponse
		err = json.Unmarshal(rr.Body.Bytes(), &j)
		if err != nil {
			t.Error("failed parsing json")
		}

		if j.OK != expectedOK && j.Message != expectedMessage {
			t.Errorf("Expected OK: %v, Message: '%s', got OK: %v, Message: '%s'", expectedOK, expectedMessage, j.OK, j.Message)
		}
	}

	t.Run("Invalid Form", func(t *testing.T) {
		executeAvailabilityJSONTest(t, false, false, "Internal server error", "", "", "")
	})

	t.Run("Empty Start Date", func(t *testing.T) {
		executeAvailabilityJSONTest(t, true, false, "Dates cannot be empty", "", "05-16-2050", "")
	})

	t.Run("Database Error", func(t *testing.T) {
		executeAvailabilityJSONTest(t, true, false, "Error searching in the database", "05-15-2050", "05-16-2050", "404")
	})

	t.Run("Room Unavailable", func(t *testing.T) {
		executeAvailabilityJSONTest(t, true, false, "Unavailable", "05-10-2050", "05-16-2050", "1")
	})

	t.Run("Room Available", func(t *testing.T) {
		executeAvailabilityJSONTest(t, true, true, "Available", "12-17-2050", "12-18-2050", "1")
	})
}
