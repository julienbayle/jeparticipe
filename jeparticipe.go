package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

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

	jeparticipe := app.NewApp(*dbFile)
	defer jeparticipe.ShutDown()

	fmt.Println("Super admin password is " + jeparticipe.SuperAdminPassword)

	api := jeparticipe.BuildApi(app.ProdMode, *baseUrl)
	log.Fatal(http.ListenAndServe(":"+*port, api.MakeHandler()))
}
