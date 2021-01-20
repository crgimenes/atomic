package jsonfiles

import (
	"encoding/json"
	"log"

	"github.com/crgimenes/atomic/database"
)

type Database struct {
}

type Bucket struct {
	name string
}

func (db Database) List() ([]database.Bucket, error) {
	return nil, nil
}

func (db Database) Use(buketName string) (database.Bucket, error) {
	b := Bucket{
		name: buketName,
	}
	return b, nil
}

func (b Bucket) Name() string {
	return b.name
}

func (b Bucket) List() ([]string, error) {
	return []string{}, nil
}

func (b Bucket) Save(ID string, value database.Data) error {
	return nil
}

func (b Bucket) Load(ID string, value database.Data) error {
	return nil
}

func (b Bucket) String() string {
	return b.name
}

func ToString(i interface{}) string {
	b, err := json.MarshalIndent(i, "", "\t")
	if err != nil {
		log.Println(err)
	}
	return string(b)
}
