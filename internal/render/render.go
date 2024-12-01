package render

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"path/filepath"
	"text/template"
	"time"

	"github.com/justinas/nosurf"
	"github.com/mlvieira/bookings/internal/config"
	"github.com/mlvieira/bookings/internal/models"
)

var app *config.AppConfig

// NewRenderer set the config for the template cache
func NewRenderer(a *config.AppConfig) {
	app = a
}

// AddDefaultData adds common data for all requests
func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.CSRFToken = nosurf.Token(r)
	if app.Session.Exists(r.Context(), "user") {
		td.IsAuthenticated = 1
	}
	return td
}

// Template render a view
func Template(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) error {
	var tc map[string]*template.Template

	if app.UseCache {
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}

	t, ok := tc[tmpl]
	if !ok {
		return errors.New("cant get template from cache")
	}

	buf := new(bytes.Buffer)

	td = AddDefaultData(td, r)

	err := t.Execute(buf, td)
	if err != nil {
		log.Fatal(err)
	}

	_, err = buf.WriteTo(w)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// CreateTemplateCache creates cache for the templates
func CreateTemplateCache() (map[string]*template.Template, error) {
	templates := map[string]*template.Template{}

	layouts, err := filepath.Glob("./templates/partials/*.layout.html")
	if err != nil {
		return templates, err
	}

	pages, err := filepath.Glob("./templates/*.page.html")
	if err != nil {
		return templates, err
	}

	funcMap := template.FuncMap{
		"humanDate": humanDate,
		"dict":      dict,
		"concat":    concat,
		"seq":       seq,
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(funcMap).ParseFiles(append(layouts, page)...)
		if err != nil {
			return templates, err
		}

		templates[name] = ts
	}

	return templates, nil
}

// dict creates a map from a variadic list of key-value pairs.
func dict(values ...any) map[string]any {
	m := make(map[string]any)
	for i := 0; i < len(values); i += 2 {
		key := values[i].(string)
		value := values[i+1]
		m[key] = value
	}
	return m
}

// humanDate return time in mm-dd-yyyy format
func humanDate(t time.Time) string {
	return t.Format("01-02-2006")
}

// concat Concat two strings
func concat(x, y string) string {
	return x + " " + y
}

// seq generates a slice of integers from start to end (inclusive).
func seq(start, end int) []int {
	s := make([]int, end-start+1)
	for i := range s {
		s[i] = start + i
	}
	return s
}
