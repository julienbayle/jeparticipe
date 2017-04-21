package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/julienbayle/jeparticipe/app"
)

func main() {

	var (
		// Path to the database
		dbFile = flag.String("db", "jeparticipe.db", "Path to the BoltDB file")

		// Application port
		port = flag.String("port", "8090", "Server port")

		// Application base path
		baseUrl = flag.String("baseurl", "", "Base URL on the server (example : /api)")
	)

	flag.Parse()

	app := app.NewApp(*dbFile)
	defer app.ShutDown()

	api := app.BuildApi(app.ProdMode, *baseUrl)
	log.Fatal(http.ListenAndServe(":"+*port, api.MakeHandler()))
}
