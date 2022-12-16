package main

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func main() {
	// connect
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatal(err)
	}

	// get SQLite version
	var version string
	err = db.Get(&version, "SELECT sqlite_version()")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("SQLite version: %s", version)

}
