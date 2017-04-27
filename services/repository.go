package services

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/boltdb/bolt"
)

type RepositoryService struct {
	Db *bolt.DB
}

// NewRepositoryService opens existing database or creates a new one
func NewRepositoryService(dbFilePath string) *RepositoryService {
	Db, err := bolt.Open(dbFilePath, 0600, nil)
	if err != nil {
		panic("Unable to create database " + err.Error())
	}

	return &RepositoryService{Db: Db}
}

// ShutDown closes the database (do defer this)
func (rs *RepositoryService) ShutDown() {
	rs.Db.Close()
}

// CreateCollectionIfNotExists creates a new document collection
func (rs *RepositoryService) CreateCollectionIfNotExists(collection string) error {
	return rs.Db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(collection))
		if err != nil {
			panic("Unable to create bucket " + err.Error())
		}
		return nil
	})
}

// GetDocument gets a document from a collection
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

// CommitDocument commits a document to the database (create / update)
func (rs *RepositoryService) CommitDocument(collection string, identifier string, document interface{}) error {
	return rs.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(collection))
		data, _ := json.Marshal(document)
		return b.Put([]byte(identifier), data)
	})
}

// GetBackup returns the database dump
func (es *RepositoryService) Backup(w rest.ResponseWriter, r *rest.Request) {
	if !hasSuperAdminPriviledge(r) {
		rest.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	err := es.Db.View(func(tx *bolt.Tx) error {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="backup.db"`)
		w.Header().Set("Content-Length", strconv.Itoa(int(tx.Size())))
		_, err := tx.WriteTo(w.(io.Writer))
		return err
	})
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
