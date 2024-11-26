package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mlvieira/bookings/internal/config"
	"github.com/mlvieira/bookings/internal/driver"
	"github.com/mlvieira/bookings/internal/handlers"
	"github.com/mlvieira/bookings/internal/helpers"
	"github.com/mlvieira/bookings/internal/render"
	"github.com/mlvieira/bookings/internal/routes"
)

var app config.AppConfig

// main is the main logic
func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}

	defer db.SQL.Close()

	defer close(app.MailChan)

	app.InfoLog.Println("Starting mail server")
	listenForMail()

	fmt.Printf("Starting aplication on http://localhost%s\n", app.Port)

	srv := &http.Server{
		Addr:    app.Port,
		Handler: routes.Routes(&app),
	}

	log.Fatal(srv.ListenAndServe())
}

func run() (*driver.DB, error) {

	app = *config.SetupAppConfig(false)

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	app.TemplateCache = tc
	app.UseCache = app.InProduction
	app.Port = ":8080"

	db, err := driver.ConnectSQL("dev:dev@/bookings?parseTime=true")
	if err != nil {
		return nil, err
	}

	log.Println("Connected to the database")

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return db, nil
}
