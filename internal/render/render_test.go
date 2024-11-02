package render

import (
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/mlvieira/bookings/internal/models"
)

func TestDefaultData(t *testing.T) {
	var td models.TemplateData

	res, err := getSession()
	if err != nil {
		t.Error(err)
	}

	session.Put(res.Context(), "flash", "123")

	result := AddDefaultData(&td, res)
	if result.Flash != "123" {
		t.Error("flash value of 123 not found in session data")
	}
}

func TestRenderTemplate(t *testing.T) {
	if err := os.Chdir("../.."); err != nil {
		t.Fatal("Could not change working directory:", err)
	}

	tc, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}

	app.TemplateCache = tc

	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()

	err = Template(rr, r, "home.page.html", &models.TemplateData{})
	if err != nil {
		t.Fatal(err)
	}

	err = Template(rr, r, "notexist.page.html", &models.TemplateData{})
	if err == nil {
		t.Fatal("rendered non existant template")
	}
}

func TestNewTemplates(t *testing.T) {
	NewRenderer(app)
}

func TestCreateTemplateCache(t *testing.T) {
	if err := os.Chdir("../.."); err != nil {
		t.Fatal("Could not change working directory:", err)
	}

	_, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}
}

func getSession() (*http.Request, error) {
	res, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		return nil, err
	}

	ctx := res.Context()
	ctx, _ = session.Load(ctx, res.Header.Get("X-Session"))
	res = res.WithContext(ctx)

	return res, nil
}

func TestDict(t *testing.T) {
	result := dict("name", "Alice", "age", 30)
	expected := map[string]any{"name": "Alice", "age": 30}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("dict() returned unexpected result: got %v, want %v", result, expected)
	}

	result = dict()
	expected = map[string]any{}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("dict() with empty input returned unexpected result: got %v, want %v", result, expected)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("dict() did not panic on odd number of arguments")
		}
	}()
	_ = dict("name")
}
