package services

import (
	"encoding/json"

	"github.com/boltdb/bolt"
)

type RepositoryService struct {
	Db *bolt.DB
}

// Opens database
func NewRepositoryService(dbFilePath string) *RepositoryService {
	Db, err := bolt.Open(dbFilePath, 0600, nil)
	if err != nil {
		panic("Unable to create database " + err.Error())
	}

	return &RepositoryService{Db: Db}
}

// Close database (defer this)
func (rs *RepositoryService) ShutDown() {
	rs.Db.Close()
}

// Creates a new document collection
func (rs *RepositoryService) CreateCollectionIfNotExists(collection string) error {
	return rs.Db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(collection))
		if err != nil {
			panic("Unable to create bucket " + err.Error())
		}
		return nil
	})
}

// Get a document from a collection
func (rs *RepositoryService) GetDocument(collection string, identifier string, document interface{}) error {
	return rs.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(collection))
		v := b.Get([]byte(identifier))
		if v != nil {
			if err := json.Unmarshal(v, document); err != nil {
				panic(err)
			}
		}
		return nil
	})
}

// Commits a document to the database (create / update)
func (rs *RepositoryService) CommitDocument(collection string, identifier string, document interface{}) error {
	return rs.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(collection))
		data, _ := json.Marshal(document)
		return b.Put([]byte(identifier), data)
	})
}
