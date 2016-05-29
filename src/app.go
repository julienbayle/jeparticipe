package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/boltdb/bolt"
	"fmt"
)

type App struct {
	Db *DB
}

func NewApp(Db *DB) *App {

	app := &App{Db:&Db}

	// Init data bucket if needed
	app.Db.Update(func(tx *bolt.Tx) error {
		activityBucket := &ActivityBucket{Code:"nocode"}
		_, err := tx.CreateBucketIfNotExists(activityBucket.Bucket())
		if err != nil {
			return fmt.Errorf("Unable to create bucket: %s", err)
		}
		fmt.Println("Bolt database ready")

		return nil
	})

	return app
}

func (app *App) GetActivityBucket(w rest.ResponseWriter, r *rest.Request) {
	code := r.PathParam("code")
	activityBucket := &ActivityBucket{Code:code}
	activityBucket.load(app.Db)
	w.WriteJson(activityBucket)
}

func (app *App) AddAParticipantToAnActivityBucket(w rest.ResponseWriter, r *rest.Request) {
	code := r.PathParam("code")
	participantName := r.PathParam("name")

	activityBucket := &ActivityBucket{Code:code}
	activityBucket.load(app.Db)
	activityBucket.AddAParticipant(participantName)
	activityBucket.save(app.Db)
	w.WriteJson("OK")
}

func (app *App) RemoveAParticipantFromAnActivityBucket(w rest.ResponseWriter, r *rest.Request) {
	code := r.PathParam("code")
	participantName := r.PathParam("name")

	activityBucket := &ActivityBucket{Code:code}
	activityBucket.load(app.Db)
	activityBucket.RemoveAParticipant(participantName)
	activityBucket.save(app.Db)
	w.WriteJson("OK")
}