package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type postData struct {
	key   string
	value string
}

type testsDataType struct {
	name               string
	url                string
	method             string
	params             []postData
	expectedStatusCode int
}

var testsData = []testsDataType{
	{"home", "/", "GET", []postData{}, http.StatusOK},
	{"about", "/about", "GET", []postData{}, http.StatusOK},
	{"generalsquarters", "/rooms/generals-quarter", "GET", []postData{}, http.StatusOK},
	{"majorssuite", "/rooms/majors-suite", "GET", []postData{}, http.StatusOK},
	{"contact", "/contact", "GET", []postData{}, http.StatusOK},
	{"availabilityGET", "/availability", "GET", []postData{}, http.StatusOK},
	{"bookingGET", "/book", "GET", []postData{}, http.StatusOK},
	{"bookingsummary", "/book/summary", "GET", []postData{}, http.StatusOK},

	{"availabilityPOST", "/availability", "POST", []postData{
		{key: "start_date", value: "02-02-2000"},
		{key: "end_date", value: "02-04-2000"},
	}, http.StatusOK},
	{"availabilityPOSTJSON", "/availability/json", "POST", []postData{
		{key: "start_date", value: "02-02-2000"},
		{key: "end_date", value: "02-04-2000"},
	}, http.StatusOK},
	{"bookingPOST", "/book", "POST", []postData{
		{key: "first_name", value: "John"},
		{key: "last_name", value: "Doe"},
		{key: "email", value: "john@example.com"},
		{key: "phone", value: "5555555555"},
	}, http.StatusOK},

	{"notfound", "/a", "GET", []postData{}, http.StatusNotFound},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)

	defer ts.Close()

	for _, test := range testsData {
		if test.method == "GET" {
			getRequest(ts, test, t)
		} else {
			postRequest(ts, test, t)
		}
	}
}

func getRequest(ts *httptest.Server, test testsDataType, t *testing.T) {
	res, err := ts.Client().Get(ts.URL + test.url)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != test.expectedStatusCode {
		t.Errorf("%s handler returned wrong status code: got %v want %v", test.url, res.StatusCode, test.expectedStatusCode)
	}
}

func postRequest(ts *httptest.Server, test testsDataType, t *testing.T) {
	values := url.Values{}

	for _, param := range test.params {
		values.Add(param.key, param.value)
	}

	res, err := ts.Client().PostForm(ts.URL+test.url, values)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != test.expectedStatusCode {
		t.Errorf("%s handler returned wrong status code: got %v want %v", test.url, res.StatusCode, test.expectedStatusCode)
	}
}
