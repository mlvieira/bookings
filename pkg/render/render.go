package render

import (
	"bytes"
	"log"
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/mlvieira/bookings/pkg/config"
	"github.com/mlvieira/bookings/pkg/models"
)

var app *config.AppConfig

// NewTemplates set the config for the template cache
func NewTemplates(a *config.AppConfig) {
	app = a
}

func AddDefaultData(td *models.TemplateData) *models.TemplateData {
	return td
}

// RenderTemplate render a view
func RenderTemplate(w http.ResponseWriter, tmpl string, td *models.TemplateData) {
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

	td = AddDefaultData(td)

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

	layouts, err := filepath.Glob("./templates/*.layout.html")
	if err != nil {
		return templates, err
	}

	pages, err := filepath.Glob("./templates/*.page.html")
	if err != nil {
		return templates, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		filenames := make([]string, 0, len(layouts)+1)
		filenames = append(filenames, page)
		filenames = append(filenames, layouts...)

		ts, err := template.New(name).ParseFiles(filenames...)
		if err != nil {
			return templates, err
		}

		templates[name] = ts
	}

	return templates, nil
}
