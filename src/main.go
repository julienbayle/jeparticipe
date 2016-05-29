package main

import (
	"log"
	"net/http"
	"github.com/ant0ine/go-json-rest/rest"
	"flag"
)

// Path to the database
var dbFile = flag.String("db", "/tmp/jeparticipe.db", "Path to the BoltDB file")

func main() {

	flag.Parse()

	// Initialize db.
	var database DB
	if err := database.Open(*dbFile); err != nil {
		log.Fatal(err)
		return
	}
	defer database.Close()

	// Init application
	app := NewApp(database)

	// Init API endpoint
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)

	router, err := rest.MakeRouter(
		rest.Get("/activity_bucket/:code", app.GetActivityBucket),
		rest.Get("/activity_bucket/:code/add_participant/:name", app.AddAParticipantToAnActivityBucket ),
		rest.Get("/activity_bucket/:code/remove_participant/:name", app.RemoveAParticipantFromAnActivityBucket),
	)
	if err != nil {
		log.Fatal(err)
		return
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))

}

