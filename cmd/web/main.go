package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mlvieira/bookings/internal/config"
	"github.com/mlvieira/bookings/internal/render"
	"github.com/mlvieira/bookings/internal/routes"
)

var app config.AppConfig

// main is the main logic
func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Starting aplication on http://localhost%s\n", app.Port)

	srv := &http.Server{
		Addr:    app.Port,
		Handler: routes.Routes(&app),
	}

	log.Fatal(srv.ListenAndServe())
}

func run() error {

	app = *config.SetupAppConfig(false)

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal(err)
	}

	app.TemplateCache = tc
	app.UseCache = app.InProduction
	app.Port = ":8080"

	return nil
}
