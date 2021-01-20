package jsonfiles

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/crgimenes/atomic/config"
	"github.com/crgimenes/atomic/database"
)

type Database struct {
	path string
}

type Bucket struct {
	path string
	name string
}

func NewDatabase(cfg config.Config) (Database, error) {
	path, err := filepath.Abs(cfg.DatabasePath)
	if err != nil {
		return Database{}, err
	}
	db := Database{
		path: path,
	}
	return db, nil
}

func (db Database) ListBuckets() ([]string, error) {
	dir, err := ioutil.ReadDir(db.path)
	if err != nil {
		log.Fatal(err)
	}
	var r []string
	for _, v := range dir {
		if !v.IsDir() {
			continue
		}
		r = append(r, v.Name())
	}
	return r, nil
}

func (db Database) UseBucket(name string) (database.Bucket, error) {
	path := filepath.Join(db.path, name)
	b := Bucket{
		name: name,
		path: path,
	}
	return b, nil
}

func (b Bucket) Name() string {
	return b.name
}

func (b Bucket) List() ([]string, error) {
	files, err := ioutil.ReadDir(b.path)
	if err != nil {
		log.Fatal(err)
	}
	var r []string
	for _, v := range files {
		if v.IsDir() {
			continue
		}
		r = append(r, v.Name())
	}
	return r, nil
}

func (b Bucket) Save(ID string, value database.Data) error {
	p, err := json.MarshalIndent(value, "", "\t")
	if err != nil {
		return err
	}
	file := filepath.Join(b.path, ID) + ".json"
	err = ioutil.WriteFile(file, p, 0644)
	return err
}

func (b Bucket) Load(ID string, value database.Data) error {
	file := filepath.Join(b.path, ID) + ".json"
	p, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	err = json.Unmarshal(p, value)
	return err
}

func (b Bucket) String() string {
	return b.name
}
