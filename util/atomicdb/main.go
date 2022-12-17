package main

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"

	"crg.eti.br/go/config"
	_ "crg.eti.br/go/config/ini"
)

type Config struct {
	DBFile string `json:"db_file" ini:"db_file" cfg:"db_file" cfgRequired:"true"`
	New    bool   `json:"new" ini:"new" cfg:"new" cfgDefault:"false"`
}

func main() {

	// load config
	var cfg = Config{}
	config.PrefixEnv = "ATOMIC"
	config.File = "config.ini"
	err := config.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	// open file
	connectionString := fmt.Sprintf(`file:%s?mode=rwc&_journal_mode=WAL&_busy_timeout=10000`, cfg.DBFile)
	db, err := sqlx.Open("sqlite", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	/*
		err = db.Ping()
		if err != nil {
			log.Fatal(err)
		}
	*/

	// get SQLite version
	var version string
	err = db.Get(&version, "SELECT sqlite_version()")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("SQLite version: %s", version)

	// create tables
	if cfg.New {
		// migration table
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
			)`)

		var lastMigration = 0
		// count migrations
		var count int
		err = db.Get(&count, `SELECT COUNT(*) FROM migrations`)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Migrations: %d", count)
		if count != 0 {
			err = db.Get(&lastMigration, "SELECT MAX(id) as max FROM migrations")
			if err != nil {
				log.Fatal(err)
			}
		}
		log.Printf("Last migration: %d", lastMigration)

		// begin transaction
		tx, err := db.Beginx()
		// run migrations
		switch lastMigration {
		case 0:
			log.Println("running migration 1")
			_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS users (
				id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				nickname TEXT NOT NULL,
				email TEXT NOT NULL,
				password TEXT NOT NULL,
				ssh_public_key TEXT NOT NULL,
				created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
			)`)
			if err != nil {
				log.Printf("error: %v", err)
				err = tx.Rollback()
				if err != nil {
					log.Fatal(err)
				}
				return
			}
			_, err = tx.Exec(`INSERT INTO migrations (id) VALUES (1)`)
			if err != nil {
				log.Printf("error: %v", err)
				err = tx.Rollback()
				if err != nil {
					log.Fatal(err)
				}
				return
			}
			log.Println("done migration 1")
			fallthrough
		default:
			log.Println("no migrations to run")
		}

		// commit transaction
		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
		}

	}

}
