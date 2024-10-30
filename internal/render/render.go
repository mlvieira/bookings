package render

import (
	"bytes"
	"log"
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/justinas/nosurf"
	"github.com/mlvieira/bookings/internal/config"
	"github.com/mlvieira/bookings/internal/models"
)

var app *config.AppConfig

// NewTemplates set the config for the template cache
func NewTemplates(a *config.AppConfig) {
	app = a
}

func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.CSRFToken = nosurf.Token(r)
	return td
}

// RenderTemplate render a view
func RenderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) {
	var tc map[string]*template.Template

	if app.UseCache {
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}

	t, ok := tc[tmpl]
	if !ok {
		log.Fatal("could not get template from template cache")
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
	}
}

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
		"dict": dict,
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
