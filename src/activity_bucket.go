package main

import (
	"encoding/json"
	"log"
	"fmt"
	"github.com/boltdb/bolt"
)


// ActivityBucket represents a "tableau des inscriptions"
type ActivityBucket struct {
	Code string
	Participants []*Participant
}

type Participant struct {
	Name string `json:"text"`
}

// Bolt database bucket name
func (activityBucket *ActivityBucket) Bucket() []byte {
	return []byte("participants")
}

func (activityBucket *ActivityBucket) AddAParticipant(name string) {
	activityBucket.Participants = append(activityBucket.Participants, &Participant{Name:name})
}

func (activityBucket *ActivityBucket) RemoveAParticipant(name string) {
	var toRemove int = -1
	for k, v := range activityBucket.Participants {
		if v.Name == name {
			toRemove = k
		}
	}
	if toRemove >=0 {
		activityBucket.Participants = append(activityBucket.Participants[:toRemove],
			activityBucket.Participants[toRemove+1:]...)
	}
}

func (activityBucket *ActivityBucket) ToJson() []byte {
	js, err := json.Marshal(activityBucket)
	if err != nil {
		log.Fatal("Unable to marshal data")
		return nil
	}
	fmt.Printf("data %+v\n", activityBucket)
	fmt.Println("json : ", string(js[:]))
	return js
}

// Loads data from database
func (activityBucket *ActivityBucket) load(db *DB) error {

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(activityBucket.Bucket())
		v := b.Get([]byte(activityBucket.Code))
		if v == nil {
			return nil
		}
		if err := json.Unmarshal(v, activityBucket); err != nil {
			panic(err)
		}
		return nil
	})

	return err
}

// Saves data to database
func (activityBucket *ActivityBucket) save(db *DB) error {

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(activityBucket.Bucket())
		return b.Put([]byte(activityBucket.Code), activityBucket.ToJson())
	})

	return err
}



