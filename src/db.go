package main

import (
	"log"

	"github.com/boltdb/bolt"
)

// DB represents a Bolt-backed data store.
type DB struct {
	*bolt.DB
}

// Initializes and opens the database.
func (db *DB) Open(path string) error {

	// Open the database in the current directory.
	// It will be created if it doesn't exist yet.
	var err error
	db.DB, err = bolt.Open(path, 0600, nil)
	if err != nil {
		log.Fatal(err)
		return err
	}

	if err != nil {
		db.Close()
		return err
	}

	return nil
}

