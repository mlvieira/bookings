package handlers

import (
	"net/http"
	"net/http/cookiejar"
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
	{"bookingsummary", "/book/summary", "GET", []postData{}, http.StatusInternalServerError},

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

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	client := &http.Client{
		Jar: jar,
	}

	ts := httptest.NewTLSServer(routes)

	defer ts.Close()

	client.Transport = &http.Transport{
		TLSClientConfig: ts.Client().Transport.(*http.Transport).TLSClientConfig,
	}

	for _, test := range testsData {
		if test.method == "GET" {
			getRequest(client, ts.URL, test, t)
		} else {
			postRequest(client, ts.URL, test, t)
		}
	}
}

func getRequest(client *http.Client, baseUrl string, test testsDataType, t *testing.T) {
	res, err := client.Get(baseUrl + test.url)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != test.expectedStatusCode {
		t.Errorf("%s handler returned wrong status code: got %v want %v", test.url, res.StatusCode, test.expectedStatusCode)
	}
}

func postRequest(client *http.Client, baseUrl string, test testsDataType, t *testing.T) {
	values := url.Values{}

	for _, param := range test.params {
		values.Add(param.key, param.value)
	}

	res, err := client.PostForm(baseUrl+test.url, values)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != test.expectedStatusCode {
		t.Errorf("%s handler returned wrong status code: got %v want %v", test.url, res.StatusCode, test.expectedStatusCode)
	}
}
