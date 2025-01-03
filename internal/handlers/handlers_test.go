package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
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

func createTestUser(userID, accessLevel int) models.User {
	return models.User{
		ID:          userID,
		AccessLevel: accessLevel,
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

func handleAdminHandlers(
	t *testing.T,
	req *http.Request,
	useSssion bool,
	expectedCode int,
	user models.User,
	handler http.HandlerFunc,
) {
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	if useSssion {
		app.Session.Put(ctx, "user", user)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	const errMessage = "Handler returned wrong response code: got %d, wanted %d"
	if rr.Code != expectedCode {
		t.Errorf(errMessage, rr.Code, expectedCode)
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
	{"Valid login page GET", "/user/login", http.StatusOK},
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

func TestRepository_PostAvailability(t *testing.T) {
	execPostAvailability := func(
		t *testing.T,
		useForm bool,
		expectedCode int,
		expectedLocation string,
		form url.Values,
	) {
		var body io.Reader = nil
		if useForm {
			body = strings.NewReader(form.Encode())
		}

		req, err := http.NewRequest("POST", "/availability", body)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		handleBookingRequest(t, req, false, expectedCode, expectedLocation, models.Reservation{}, http.HandlerFunc(Repo.PostAvailability))
	}

	t.Run("Valid request", func(t *testing.T) {
		form := url.Values{}
		form.Add("start_date", "12-17-2049")
		form.Add("end_date", "12-20-2049")
		execPostAvailability(t, true, http.StatusOK, "", form)
	})

	t.Run("Invalid Form", func(t *testing.T) {
		execPostAvailability(t, false, http.StatusSeeOther, "/availability", nil)
	})

	t.Run("Empty Start Date", func(t *testing.T) {
		form := url.Values{}
		form.Add("start_date", "12-17-2049")
		execPostAvailability(t, true, http.StatusSeeOther, "/availability", form)
	})

	t.Run("Invalid Start Date", func(t *testing.T) {
		form := url.Values{}
		form.Add("start_date", "2050-02-01")
		form.Add("end_date", "12-20-2049")
		execPostAvailability(t, true, http.StatusSeeOther, "/availability", form)
	})

	t.Run("Invalid End Date", func(t *testing.T) {
		form := url.Values{}
		form.Add("start_date", "12-20-2049")
		form.Add("end_date", "2050-02-01")
		execPostAvailability(t, true, http.StatusSeeOther, "/availability", form)
	})

	t.Run("Database Error: Error searching DB", func(t *testing.T) {
		form := url.Values{}
		form.Add("start_date", "12-17-2050")
		form.Add("end_date", "12-20-2050")
		execPostAvailability(t, true, http.StatusSeeOther, "/availability", form)
	})

	t.Run("No room available", func(t *testing.T) {
		form := url.Values{}
		form.Add("start_date", "12-18-2050")
		form.Add("end_date", "12-18-2050")
		execPostAvailability(t, true, http.StatusSeeOther, "/availability", form)
	})
}

func TestRepository_ReservationSummary(t *testing.T) {
	executeBookingTest := func(
		t *testing.T,
		useSession bool,
		expectedCode int,
		expectedLocation string,
		reservation models.Reservation,
	) {
		req, err := http.NewRequest("GET", "/book/summary", nil)
		if err != nil {
			t.Fatal(err)
		}

		handleBookingRequest(t, req, useSession, expectedCode, expectedLocation, reservation, http.HandlerFunc(Repo.ReservationSummary))
	}

	t.Run("Valid request", func(t *testing.T) {
		executeBookingTest(t, true, http.StatusOK, "", createTestReservation(1, "test"))
	})

	t.Run("Missing session", func(t *testing.T) {
		executeBookingTest(t, false, http.StatusTemporaryRedirect, "/", createTestReservation(1, "test"))
	})
}

func TestRepository_ChooseRoom(t *testing.T) {
	executeBookingTest := func(
		t *testing.T,
		useSession bool,
		expectedCode int,
		expectedLocation,
		roomID string,
		reservation models.Reservation,
	) {
		req, err := http.NewRequest("GET", fmt.Sprintf("/rooms/book/%s", roomID), nil)
		if err != nil {
			t.Fatal(err)
		}

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", roomID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		handleBookingRequest(t, req, useSession, expectedCode, expectedLocation, reservation, http.HandlerFunc(Repo.ChooseRoom))
	}

	t.Run("Valid request", func(t *testing.T) {
		executeBookingTest(t, true, http.StatusSeeOther, "/book", "1", createTestReservation(1, "test"))
	})

	t.Run("Invalid RoomID in path", func(t *testing.T) {
		executeBookingTest(t, true, http.StatusTemporaryRedirect, "/", "err", models.Reservation{})
	})

	t.Run("Missing session", func(t *testing.T) {
		executeBookingTest(t, false, http.StatusTemporaryRedirect, "/", "1", models.Reservation{})
	})
}

func TestRepository_PostShowLoginPage(t *testing.T) {
	executeLoginPage := func(
		t *testing.T,
		useForm bool,
		expectedCode int,
		expectedLocation string,
		form url.Values,
	) {
		var body io.Reader = nil
		if useForm {
			body = strings.NewReader(form.Encode())
		}
		req, err := http.NewRequest("POST", "/users/login", body)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		handleBookingRequest(t, req, false, expectedCode, expectedLocation, models.Reservation{}, http.HandlerFunc(Repo.PostShowLoginPage))
	}

	t.Run("Valid login", func(t *testing.T) {
		form := url.Values{}
		form.Add("email", "john@example.com")
		form.Add("password", "password")
		executeLoginPage(t, true, http.StatusSeeOther, "/", form)
	})

	t.Run("Invalid form", func(t *testing.T) {
		executeLoginPage(t, false, http.StatusSeeOther, "/user/login", nil)
	})

	t.Run("Empty email", func(t *testing.T) {
		form := url.Values{}
		form.Add("email", "")
		form.Add("password", "password")
		executeLoginPage(t, true, http.StatusSeeOther, "/user/login", form)
	})

	t.Run("Invalid email", func(t *testing.T) {
		form := url.Values{}
		form.Add("email", "john")
		form.Add("password", "password")
		executeLoginPage(t, true, http.StatusSeeOther, "", form)
	})

	t.Run("Wrong credentials", func(t *testing.T) {
		form := url.Values{}
		form.Add("email", "johhn@at.com")
		form.Add("password", "password")
		executeLoginPage(t, true, http.StatusSeeOther, "/user/login", form)
	})
}

func TestRepository_Logout(t *testing.T) {
	execLogout := func(
		t *testing.T,
		expectedCode int,
		expectedLocation string,
	) {
		req, err := http.NewRequest("GET", "/logout", nil)
		if err != nil {
			t.Fatal(err)
		}

		handleBookingRequest(t, req, false, expectedCode, expectedLocation, models.Reservation{}, http.HandlerFunc(Repo.Logout))
	}

	t.Run("Logout", func(t *testing.T) {
		execLogout(t, http.StatusSeeOther, "/user/login")
	})
}

func TestRepository_AdminDashboard(t *testing.T) {
	execLogout := func(
		t *testing.T,
		path string,
		expectedCode int,
		useSession bool,
		user models.User,
		handler http.HandlerFunc,
	) {
		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			t.Fatal(err)
		}

		handleAdminHandlers(t, req, useSession, expectedCode, user, handler)
	}

	t.Run("GET - Dashboard", func(t *testing.T) {
		execLogout(t, "/admin/dashboard", http.StatusOK, true, createTestUser(1, 1), http.HandlerFunc(Repo.AdminDashboard))
	})

	t.Run("GET - Dashboard Session not found", func(t *testing.T) {
		execLogout(t, "/admin/dashboard", http.StatusSeeOther, false, models.User{}, http.HandlerFunc(Repo.AdminDashboard))
	})

	t.Run("GET - New reservations", func(t *testing.T) {
		execLogout(t, "/admin/reservations/new", http.StatusOK, true, createTestUser(1, 3), http.HandlerFunc(Repo.AdminNewReservations))
	})

	t.Run("GET - All reservations", func(t *testing.T) {
		execLogout(t, "/admin/reservations/all", http.StatusOK, true, createTestUser(1, 3), http.HandlerFunc(Repo.AdminAllReservations))
	})

	t.Run("GET - Calendar - HTML", func(t *testing.T) {
		execLogout(t, "/admin/reservations/calendar", http.StatusOK, true, createTestUser(1, 3), http.HandlerFunc(Repo.AdminCalendarReservations))
	})

}

func TestRepository_AdminCalendarReservations(t *testing.T) {
	testJSONResponse := func(
		t *testing.T,
		path string,
		useSession bool,
		user models.User,
		handler http.HandlerFunc,
		expectedCode int,
		expectedResponse []models.CalendarResponse,
	) {
		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			t.Fatal(err)
		}

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		if useSession {
			app.Session.Put(ctx, "user", user)
		}

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != expectedCode {
			t.Errorf("Handler returned wrong status code: got %d, expected %d", rr.Code, expectedCode)
		}

		var actualResponse []models.CalendarResponse
		err = json.NewDecoder(rr.Body).Decode(&actualResponse)
		if err != nil {
			t.Fatalf("Failed to decode JSON response: %v", err)
		}

		if len(actualResponse) != len(expectedResponse) {
			t.Fatalf("Expected %d reservations, got %d", len(expectedResponse), len(actualResponse))
		}

		for i, res := range actualResponse {
			if res.ID != expectedResponse[i].ID ||
				res.Title != expectedResponse[i].Title ||
				!res.Start.Equal(expectedResponse[i].Start) ||
				!res.End.Equal(expectedResponse[i].End) {
				t.Errorf("Mismatch in reservation %d: expected %+v, got %+v", i, expectedResponse[i], res)
			}
		}
	}

	t.Run("GET - All reservations JSON - Valid", func(t *testing.T) {
		expectedResponse := []models.CalendarResponse{
			{
				ID:       "1",
				Title:    "Test room reservation",
				Start:    time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
				End:      time.Date(2024, 12, 8, 0, 0, 0, 0, time.UTC),
				AllDay:   true,
				Url:      "/admin/reservations/details/1",
				Editable: false,
				ExtendedProps: map[string]any{
					"name":        "John Doe",
					"room":        "Test Room",
					"lastUpdated": time.Now(),
				},
			},
		}

		path := "/admin/reservations/calendar/json?start=2024-12-01T00:00:00Z&end=2024-12-08T00:00:00Z"
		testJSONResponse(t, path, true, createTestUser(1, 3), http.HandlerFunc(Repo.JsonAdminCalendarReservations), http.StatusOK, expectedResponse)
	})

	t.Run("GET - All Reservations JSON - Wrong Start Format", func(t *testing.T) {
		path := "/admin/reservations/calendar/json?start=01-12-2024T00:00:00Z&end=2024-12-08T00:00:00Z"
		expectedErrorResponse := map[string]string{
			"error": "Invalid start date format. Use RFC3339 format.",
		}

		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		app.Session.Put(ctx, "user", createTestUser(1, 3))

		Repo.JsonAdminCalendarReservations(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Handler returned wrong status code: got %d, expected %d", rr.Code, http.StatusBadRequest)
		}

		var actualErrorResponse map[string]string
		err = json.NewDecoder(rr.Body).Decode(&actualErrorResponse)
		if err != nil {
			t.Fatalf("Failed to decode JSON response: %v", err)
		}

		if actualErrorResponse["error"] != expectedErrorResponse["error"] {
			t.Errorf("Expected error message: %q, got: %q", expectedErrorResponse["error"], actualErrorResponse["error"])
		}
	})

	t.Run("GET - All Reservations JSON - Wrong End Format", func(t *testing.T) {
		path := "/admin/reservations/calendar/json?start=2024-12-01T00:00:00Z&end=01-12-2024T00:00:00Z"
		expectedErrorResponse := map[string]string{
			"error": "Invalid start date format. Use RFC3339 format.",
		}

		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		app.Session.Put(ctx, "user", createTestUser(1, 3))

		Repo.JsonAdminCalendarReservations(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Handler returned wrong status code: got %d, expected %d", rr.Code, http.StatusBadRequest)
		}

		var actualErrorResponse map[string]string
		err = json.NewDecoder(rr.Body).Decode(&actualErrorResponse)
		if err != nil {
			t.Fatalf("Failed to decode JSON response: %v", err)
		}

		if actualErrorResponse["error"] != expectedErrorResponse["error"] {
			t.Errorf("Expected error message: %q, got: %q", expectedErrorResponse["error"], actualErrorResponse["error"])
		}
	})

	t.Run("GET - All Reservations JSON - DB Error", func(t *testing.T) {
		path := "/admin/reservations/calendar/json?start=0001-01-01T00:00:00Z&end=2024-12-08T00:00:00Z"

		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		app.Session.Put(ctx, "user", createTestUser(1, 3))

		Repo.JsonAdminCalendarReservations(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("Handler returned wrong status code: got %d, expected %d", rr.Code, http.StatusInternalServerError)
		}

		var actualErrorResponse map[string]string
		err = json.NewDecoder(rr.Body).Decode(&actualErrorResponse)
		if err != nil {
			t.Fatalf("Failed to decode JSON response: %v", err)
		}

		expectedErrorMessage := "Internal Server Error"
		if actualErrorResponse["error"] != expectedErrorMessage {
			t.Errorf("Expected error message: %q, got: %q", expectedErrorMessage, actualErrorResponse["error"])
		}
	})

}

func TestRepository_PostAdminReservationSummary(t *testing.T) {
	execLogout := func(
		t *testing.T,
		reservationID,
		expectedLocation string,
		expectedCode int,
		useForm,
		useSession bool,
		form url.Values,
		user models.User,
		handler http.HandlerFunc,
	) {
		var body io.Reader = nil
		if useForm {
			body = strings.NewReader(form.Encode())
		}
		req, err := http.NewRequest("POST", fmt.Sprintf("/admin/reservations/details/%s", reservationID), body)
		if err != nil {
			t.Fatal(err)
		}

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", reservationID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		if useSession {
			app.Session.Put(ctx, "user", user)
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

	t.Run("POST - Reservation Summary - Valid", func(t *testing.T) {
		form := url.Values{}
		form.Add("first_name", "John")
		form.Add("last_name", "Doe")
		form.Add("email", "john@example.com")
		form.Add("phone", "555555555")
		execLogout(t, "1", "/admin/reservations/details/1", http.StatusSeeOther, true, true, form, models.User{}, http.HandlerFunc(Repo.PostAdminReservationSummary))
	})

	t.Run("POST - Reservation Summary - Invalid Form", func(t *testing.T) {
		execLogout(t, "1", "/admin/dashboard", http.StatusTemporaryRedirect, false, true, nil, models.User{}, http.HandlerFunc(Repo.PostAdminReservationSummary))
	})

	t.Run("POST - Reservation Summary - Invalid ID", func(t *testing.T) {
		form := url.Values{}
		form.Add("first_name", "John")
		form.Add("last_name", "Doe")
		form.Add("email", "john@example.com")
		form.Add("phone", "555555555")
		execLogout(t, "a", "/admin/dashboard", http.StatusTemporaryRedirect, true, true, form, models.User{}, http.HandlerFunc(Repo.PostAdminReservationSummary))
	})

	t.Run("POST - Reservation Summary - DB GetReservation error", func(t *testing.T) {
		form := url.Values{}
		form.Add("first_name", "John")
		form.Add("last_name", "Doe")
		form.Add("email", "john@example.com")
		form.Add("phone", "555555555")
		execLogout(t, "2", "/admin/dashboard", http.StatusSeeOther, true, true, form, models.User{}, http.HandlerFunc(Repo.PostAdminReservationSummary))
	})

	t.Run("POST - Reservation Summary - DB UpdateReservation error", func(t *testing.T) {
		form := url.Values{}
		form.Add("first_name", "John")
		form.Add("last_name", "Doe")
		form.Add("email", "john@example.com")
		form.Add("phone", "555555555")
		execLogout(t, "2", "/admin/dashboard", http.StatusSeeOther, true, true, form, models.User{}, http.HandlerFunc(Repo.PostAdminReservationSummary))
	})

}
